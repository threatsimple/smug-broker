package smug


// 09:52:24 < smig> |marykarnes| congrats <@UHFUL9WUS> that sounds perfect


import (
    "fmt"
    "regexp"
    "strings"
    "time"

    libsl "github.com/nlopes/slack"
)


type SlackUser struct {
    Id string
    Nick string
    Avatar string
}


type SlackUserCache struct {
    users map[string]*SlackUser
}


func (suc *SlackUserCache) Username(sb *SlackBroker, ukey string) string {
    if val, ok := suc.users[ukey]; ok {
        return val.Nick
    }
    user, err := sb.api.GetUserInfo(ukey)
    if err != nil {
        return "piglet"
    } else {
        suc.users[ukey] = &SlackUser{
            Id: ukey,
            Nick: user.Name,
            Avatar: user.Profile.Image72,
        }
    }
    return user.Name
}


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
}


func (sb *SlackBroker) CachedUsername(ukey string) string {
    return sb.usercache.Username(sb,ukey)
}


func (sb *SlackBroker) Name() string {
    return fmt.Sprintf("slack-%s", sb.channel)
}


func (sb *SlackBroker) SetupInternals() {
    sb.log = NewLogger("slack")
    sb.usercache = &SlackUserCache{}
    sb.usercache.users = make(map[string]*SlackUser)
    sb.re_uids = regexp.MustCompile(`<@(U\w+)>`) // get sub ids in msgs
}


func (sb *SlackBroker) ConvertUserRefs(s string) string {
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
            sb.CachedUsername(u),
        )
    }
    return s
}


type SlackLogger struct {
    *Logger
}


func (sl *SlackLogger) Output(lvl int, msg string) error {
    sl.Info(msg)
    return nil
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

    channels, _ := sb.api.GetChannels(false)
    for _, channel := range channels {
        if channel.Name == sb.channel {
            sb.chanid = channel.ID
            break
        }
	}
    if sb.chanid == "" {
        sb.log.Warnf("ERR channel not found (%s)", sb.channel)
        return
    }

}


func (sb *SlackBroker) Put(msg string) {
    sb.rtm.SendMessage(sb.rtm.NewOutgoingMessage(msg, sb.chanid))
}


func (sb *SlackBroker) Publish(ev *Event, dis Dispatcher) {
    sb.api.PostMessage(
        sb.chanid,
        libsl.MsgOptionText(ev.Text, false),
        libsl.MsgOptionUsername(ev.Nick),
        libsl.MsgOptionIconEmoji(fmt.Sprintf(":avatar_%s:", ev.Nick)),
    )
    // sb.rtm.SendMessage(sb.rtm.NewOutgoingMessage(msg, sb.chanid))
    // sb.Put(fmt.Sprintf("%s| %s", ev.Nick, ev.Text))
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
                outmsgs := []string{e.Text}
                if len(e.Files) > 0 {
                    for _,f := range e.Files {
                        outmsgs = append(outmsgs,
                            fmt.Sprintf("%s(%s)",f.Name,f.URLPrivate))
                    }
                }
                outstr := strings.Join(outmsgs, " ")
                ev := &Event{
                    Origin: sb,
                    Nick: sb.CachedUsername(e.User),
                    RawText: outstr,
                    Text: sb.ConvertUserRefs(outstr),
                    ts: time.Now(),
                }
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


