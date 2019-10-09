// broker: slack
// provides our interface for consuming from, and publishing, to slack

// NOTE ABOUT CREDENTIALS
// this requires slack credentials to be created for your network and the bot to
// be invited to the specific channel.
// In particular, know that the slack api is wide open unless you go through the
// motions of creating a bot specific access token.  That means private messages
// can flow through this bot from the creating user and replies to that user,
// regardless if the bot is privy to those conversations or not.


package smug


import (
    "fmt"
    "regexp"
    "strings"
    "time"

    libsl "github.com/nlopes/slack"
)


/* ************************** *
 * fake the slacklib logger
 * ************************** */

type SlackLogger struct {
    *Logger
}


func (sl *SlackLogger) Output(lvl int, msg string) error {
    sl.Info(msg)
    return nil
}


/* ************************** *
 * repr our slack users
 * ************************** */


type SlackUser struct {
    Id string
    Nick string
    Avatar string
}


type SlackUserCache struct {
    users map[string]*SlackUser
    nicks map[string]*SlackUser
}


func (suc *SlackUserCache) CacheUser(user *SlackUser) {
    suc.users[user.Id] = user
    suc.nicks[user.Nick] = user
}


func (suc *SlackUserCache) UserFromAPI(
        sb *SlackBroker, ukey string) (*SlackUser,error) {
    user, err := sb.api.GetUserInfo(ukey)
    if err != nil {
        return nil,fmt.Errorf("err fetching user from slack: %+v", err)
    }
    suser := &SlackUser{
        Id: ukey,
        Nick: user.Name,
        Avatar: user.Profile.Image72,
    }
    suc.CacheUser(suser)
    return suser,nil
}


func (suc *SlackUserCache) UserNick(
        sb *SlackBroker, ukey string, cacheOnly bool) string {
    if val, ok := suc.users[ukey]; ok {
        return val.Nick
    }
    if cacheOnly { return "" }
    user,err := suc.UserFromAPI(sb, ukey)
    if err != nil {
        sb.log.Warnf("attempted to fetch %s but got err: %v", ukey, err)
        return ""
    }
    return user.Nick
}


func (suc *SlackUserCache) UserId(
        sb *SlackBroker, nick string, cacheOnly bool) string {
    if val, ok := suc.nicks[nick]; ok {
        return val.Id
    }
    if cacheOnly { return "" }
    user,err := suc.UserFromAPI(sb, nick)
    if err != nil {
        sb.log.Warnf("attempted to fetch %s but got err: %v", nick, err)
        return ""
    }
    return user.Id
}


func (suc *SlackUserCache) PopulateCache(sb *SlackBroker, mems []string) {
    for _,uid := range mems {
        suc.UserFromAPI(sb, uid)
    }
}


/* ************************** *
 * slack broker
 * ************************** */


type SlackBroker struct {
    log *Logger
    // components from slack lib
    api *libsl.Client
    rtm *libsl.RTM
    // internal plumbing
    usercache *SlackUserCache
    chanid string
    channel string
    token string
    mybotid string
    re_uids *regexp.Regexp
    re_usernick *regexp.Regexp
    re_atusers *regexp.Regexp
}



func (sb *SlackBroker) Name() string {
    return fmt.Sprintf("slack-%s", sb.channel)
}


// allows us to setup internal members without hitting the api
// let's us do certain tests that don't require api
func (sb *SlackBroker) SetupInternals() {
    sb.log = NewLogger("slack")
    sb.usercache = &SlackUserCache{}
    sb.usercache.users = make(map[string]*SlackUser)
    sb.usercache.nicks = make(map[string]*SlackUser)
    sb.re_uids = regexp.MustCompile(`<@(U\w+)>`) // get sub ids in msgs
    sb.re_usernick = regexp.MustCompile(`^(\w+):`)
    sb.re_atusers = regexp.MustCompile(`@(\w+)\b`)
}


func (sb *SlackBroker) ConvertRefsToUsers(s string, cacheOnly bool) string {
    matches := sb.re_uids.FindAllStringSubmatchIndex(s,-1)
    // will contain a uniq set of uids mentioned
    uids := make(map[string]bool)
    for i := len(matches)-1; i >= 0; i-- {
        // start,stop,sub0,sublen := matches[i]
        m := matches[i]
        uid := s[m[2]:m[3]]
        if ! uids[uid] { uids[uid] = true }
    }
    // now iterate over uids and replace them all
    for u,_ := range uids {
        s = strings.ReplaceAll(
            s,
            fmt.Sprintf("<@%s>",u),
            sb.usercache.UserNick(sb, u, cacheOnly),
        )
    }
    return s
}


func (sb *SlackBroker) ConvertUsersToRefs(s string, cacheOnly bool) string {
    //  first look for irc type  USER: at beginning of line
    matches := sb.re_usernick.FindAllStringSubmatchIndex(s,-1)
    for i := len(matches)-1; i >= 0; i-- {
        // start,stop,sub0,sublen := matches[i]
        m := matches[i]
        usernick := s[m[2]:m[3]]
        uid := sb.usercache.UserId(sb, usernick, cacheOnly)
        if len(uid) > 4 {
            s = strings.ReplaceAll(
                s,
                usernick,
                fmt.Sprintf("<@%s>",uid),
            )
        }
    }

    //  then do embedded @user replacements
    matches = sb.re_atusers.FindAllStringSubmatchIndex(s,-1)
    for i := len(matches)-1; i >= 0; i-- {
        // start,stop,sub0,sublen := matches[i]
        m := matches[i]
        usernick := s[m[2]:m[3]]
        uid := sb.usercache.UserId(sb, usernick, cacheOnly)
        if len(uid) > 1 {
            s = strings.ReplaceAll(
                s,
                "@" + usernick,
                fmt.Sprintf("<@%s>",uid),
            )
        }
    }

    return s
}


// args [token, channel]
func (sb *SlackBroker) Setup(args ...string) {
    sb.SetupInternals()
    sb.token = args[0]
    sb.channel = args[1]
    sc := libsl.New(
        sb.token,
        libsl.OptionDebug(true),
        libsl.OptionLog(&SlackLogger{sb.log}),
    )
    sb.api = sc
    sb.rtm = sb.api.NewRTM()
    authtest,_ := sb.api.AuthTest() // gets our identity from slack api
    myuid := authtest.UserID
    myuser, err := sb.api.GetUserInfo(myuid)
    if err != nil {
        sb.log.Warnf("ERR occurred %+v", err)
    }
    sb.mybotid = myuser.Profile.BotID

    // populate my channel info
    // this is a bit ... lame. Should be better way?  XXX
    channels, _ := sb.api.GetChannels(false)
    for _, channel := range channels {
        if channel.Name == sb.channel {
            sb.chanid = channel.ID
            sb.usercache.PopulateCache(sb, channel.Members)
            break
        }
	}
    if sb.chanid == "" {
        sb.log.Warnf("ERR channel not found (%s)", sb.channel)
        return
    }

    // populate users in the chan

}


func (sb *SlackBroker) Put(msg string) {
    sb.rtm.SendMessage(sb.rtm.NewOutgoingMessage(msg, sb.chanid))
}


func (sb *SlackBroker) Publish(ev *Event, dis Dispatcher) {
    txt := sb.ConvertUsersToRefs(ev.Text, false)
    sb.api.PostMessage(
        sb.chanid,
        libsl.MsgOptionText(txt, false),
        libsl.MsgOptionUsername(ev.Nick),
        libsl.MsgOptionIconEmoji(fmt.Sprintf(":avatar_%s:", ev.Nick)),
    )
    // sb.rtm.SendMessage(sb.rtm.NewOutgoingMessage(msg, sb.chanid))
    // sb.Put(fmt.Sprintf("%s| %s", ev.Nick, ev.Text))
}


func (sb *SlackBroker) ParseToEvent(e *libsl.MessageEvent) *Event {
    sb.log.Debugf("SL MsgEv  %+v", e)
    outmsgs := []string{e.Text}
    if len(e.Files) > 0 {
        for _,f := range e.Files {
            outmsgs = append(outmsgs,
                fmt.Sprintf("%s(%s)",f.Name,f.URLPrivate) )
        }
    }
    if len(e.Attachments) > 0 {
        for _,a := range e.Attachments {
            outmsgs = append(outmsgs,
                fmt.Sprintf("%s - %s", a.Title, a.ImageURL) )
        }
    }
    outstr := strings.TrimSpace(strings.Join(outmsgs, " "))
    ev := &Event{
        Origin: sb,
        Nick: sb.usercache.UserNick(sb, e.User, false),
        RawText: outstr,
        Text: sb.ConvertRefsToUsers(outstr, false),
        ts: time.Now(),
    }
    return ev
}


func (sb *SlackBroker) Run(dis Dispatcher) {
    if sb.rtm == nil {
        // raise some error here XXX TODO
    }
    go sb.rtm.ManageConnection()
    for msg := range sb.rtm.IncomingEvents {
        switch e := msg.Data.(type) {
        case *libsl.HelloEvent:
            // ignore Hello
        case *libsl.UserTypingEvent:
            // ignore typing
        case *libsl.ConnectedEvent:
            sb.log.Infof("joining chan: %s", sb.channel)
        case *libsl.MessageEvent:
            // smugbot: 2019/09/14 08:47:44 websocket_managed_conn.go:369:
            // Incoming Event:
            // {"client_msg_id":"ed722fbc-5b37-4f78-9981-e3c9ce5c85a1","suppress_notification":false,"type":"message","text":"test","user":"U6CRHMXK4","team":"T6CRHMX5G","user_team":"T6CRHMX5G","source_team":"T6CRHMX5G","channel":"C6MR9CBGR","event_ts":"1568468854.004200","ts":"1568468854.004200"}
            if e.BotID != sb.mybotid && e.Channel == sb.chanid {
                ev := sb.ParseToEvent(e)
                dis.Broadcast(ev)
            }
        case *libsl.PresenceChangeEvent:
            sb.log.Infof("Presence Change: %v\n", e)
        case *libsl.LatencyReport:
            sb.log.Infof("Current latency: %v\n", e.Value)
        case *libsl.RTMError:
            sb.log.Warnf("Error: %s\n", e.Error())
        case *libsl.InvalidAuthEvent:
            sb.log.Fatalf("Invalid credentials")
            return
        default:
            // Ignore other events..
            sb.log.Infof("Unexpected: %v\n", msg.Data)
        }
    }
}


