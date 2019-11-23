// broker: irc
// owns communication between irc and the dispatcher


package smug

 
import (
    "crypto/tls"
    "fmt"
    "strings"
    "time"

    libirc "github.com/thoj/go-ircevent"
)


type IrcBroker struct {
    log *Logger
    conn *libirc.Connection
    channel string
    nick string
    botname string
    prefix string
    server string
}


func (ib *IrcBroker) Name() string {
    return fmt.Sprintf("irc-%s-%s-as-%s", ib.server, ib.channel, ib.nick)
}


// args [server, channel, nick, botname]
func (ib *IrcBroker) Setup(args ...string) {
    ib.server = args[0]
    ib.channel = args[1]
    ib.nick = args[2]
    if len(args) > 3 {
        ib.botname = args[3]
    } else {
        ib.botname = "smug"
    }
    ib.log = NewLogger(ib.Name())
    ib.conn = libirc.IRC(ib.nick, ib.botname)
    // ib.conn.VerboseCallbackHandler = true
    ib.conn.UseTLS = true  // XXX should be a param
    if ib.conn.UseTLS {
        ib.conn.TLSConfig = &tls.Config{InsecureSkipVerify: true} // XXX
    }
    ib.conn.AddCallback(
        "001",
        func(e *libirc.Event) {
            ib.log.Infof("irc joining %s / %s", ib.server, ib.channel)
            ib.conn.Join(ib.channel)
            ib.conn.Privmsg(ib.channel, fmt.Sprintf("%s online", ib.botname))
        } )
    // ib.conn.AddCallback("366", func(e *irc.Event) { }) // ignore end of names
    err := ib.conn.Connect(ib.server)
    if err != nil {
        ib.log.Errorf("ERR %s", err)
        ib.conn = nil // error'd here, set this connection to nil XXX
    }
}


func (ib *IrcBroker) MsgTarget(target string, msg string, prefix string) {
    maxlen := 500
    for i,s := range strings.Split(msg, "\n") {
        if i > 6 { return } // just stop
        if len(s) > 0 {
            if len(s) > maxlen {
                outs := ChunkSplit(s, maxlen)
                for _,s := range outs {
                    ib.conn.Privmsg(target,prefix + s)
                    time.Sleep(100 * time.Millisecond) //slow down a flood
                }
            } else {
                ib.conn.Privmsg(target,prefix + s)
                if i > 0 { time.Sleep(100 * time.Millisecond) } //slow down a flood
            }
        }
    }
}


func (ib *IrcBroker) HandleEvent(ev *Event, dis Dispatcher) {
    if ev.ReplyBroker != nil && ev.ReplyBroker != ib {
        // not intended for us, just ignore silently
        return
    }
    var prefix string
    if ev.IsCmdOutput {
        prefix = ""
    } else {
        prefix = fmt.Sprintf("|%s| ", ev.Actor)
    }
    if ev.ReplyBroker == ib {
        // private message for a user
        go ib.MsgTarget(ev.ReplyTarget, ev.Text, prefix)
    } else {
        go ib.MsgTarget(ib.channel, ev.Text, prefix)
    }
}


func (ib *IrcBroker) Activate(dis Dispatcher) {
    // XXX this should ensure some sort of singleton to ensure Run should only
    // ever be called once...
    ib.conn.AddCallback("PRIVMSG", func (e *libirc.Event) {
        // are we a priv msg?

// event needs IsPrivateMsg which then responds with the RespondTo broker set on
// replies.  Is this best way?  Should all messages go to all brokers?  Or should a
// broker have an option that says "RecvsPrivate"

        ev := &Event{
            Origin: ib,
            Actor: e.Nick,
            Text: e.Message(),
            ts: time.Now(),
        }
        if len(e.Arguments) > 0 && e.Arguments[0] == ib.nick {
            ev.ReplyTarget = e.Nick
            ev.ReplyBroker = ib
        }
        dis.Broadcast(ev)
    })
    ib.conn.AddCallback("CTCP_ACTION", func (e *libirc.Event) {
        if e.Arguments[0] == ib.channel {
            ev := &Event{
                Origin: ib,
                Actor: e.Nick,
                Text: fmt.Sprintf("_ %s %s _", e.Nick, e.Arguments[1]),
                ts: time.Now(),
            }
            dis.Broadcast(ev)
        }
    })
}

func (ib *IrcBroker) Deactivate() { }

