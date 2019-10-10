// broker: irc
// owns communication between irc and the dispatcher


package smug


import (
    "crypto/tls"
    "fmt"
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
            ib.Put(fmt.Sprintf("%s online", ib.botname))
        } )
    // ib.conn.AddCallback("366", func(e *irc.Event) { }) // ignore end of names
    err := ib.conn.Connect(ib.server)
    if err != nil {
        ib.log.Errorf("ERR %s", err)
        ib.conn = nil // error'd here, set this connection to nil XXX
    }
}


func (ib *IrcBroker) Put(msg string) {
    ib.conn.Privmsg(ib.channel, msg)
}


func (ib *IrcBroker) Publish(ev *Event, dis Dispatcher) {
    if ev.IsCmdOutput {
        ib.Put(fmt.Sprintf("%s", ev.Text))
    } else {
        ib.Put(fmt.Sprintf("|%s| %s", ev.Nick, ev.Text))
    }
}


func (ib *IrcBroker) Run(dis Dispatcher) {
    // XXX this should ensure some sort of singleton to ensure Run should only
    // ever be called once...
    ib.conn.AddCallback("PRIVMSG", func (e *libirc.Event) {
        ev := &Event{
            Origin: ib,
            Nick: e.Nick,
            Text: e.Message(),
            ts: time.Now(),
        }
        dis.Broadcast(ev)
    })
    ib.conn.AddCallback("CTCP_ACTION", func (e *libirc.Event) {
        if e.Arguments[0] == ib.channel {
            ev := &Event{
                Origin: ib,
                Nick: e.Nick,
                Text: fmt.Sprintf("_ %s %s _", e.Nick, e.Arguments[1]),
                ts: time.Now(),
            }
            dis.Broadcast(ev)
        }
    })
}


