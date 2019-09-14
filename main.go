package main

import (
    "flag"
    "fmt"
    "time"

    smug "github.com/nod/smug/smug"
)


type SlackOpts struct {
    token string
    channel string
}


func parseSlackOpts() *SlackOpts {
    opts := &SlackOpts{}
    flag.StringVar(&opts.token, "slack-token", "", "slack api token")
    flag.StringVar(&opts.channel, "slack-channel", "", "slack channel to join")
    return opts
}


type IrcOpts struct {
    server string
    is_ssl bool
    channel string
    nick string
}


func parseIrcOpts() *IrcOpts {
    opts := &IrcOpts{}
    flag.StringVar(&opts.server, "irc-server", "", "irc server")
    flag.StringVar(&opts.channel, "irc-channel", "", "irc channel")
    flag.BoolVar(&opts.is_ssl, "irc-ssl", true, "irc ssl")
    flag.StringVar(&opts.nick, "irc-nick", "", "irc nick")
    return opts
}


type Opts struct {
    irc *IrcOpts
    slack *SlackOpts
}


func parseOpts() *Opts {
    iopts := parseIrcOpts()
    sopts := parseSlackOpts()
    opts := &Opts{
        irc: iopts,
        slack: sopts,
    }
    flag.Parse()
    return opts
}


func main() {
    dispatcher := &smug.CentralDispatch{}

    opts := parseOpts()

    sb := &smug.SlackBroker{}
    sb.Setup(opts.slack.token, opts.slack.channel)
    go sb.Run(dispatcher)
    dispatcher.AddBroker(sb)

    ib := &smug.IrcBroker{}
    fmt.Println("server", opts.irc.server)
    fmt.Println("chan", opts.irc.channel)
    fmt.Println("nick", opts.irc.nick)
    ib.Setup(opts.irc.server, opts.irc.channel, opts.irc.nick)
    go ib.Run(dispatcher)
    dispatcher.AddBroker(ib)

    time.Sleep(3 * time.Second)
    ib.Put("manana na")

    // just loop here for now so others can run
    for true {
        time.Sleep(200 * time.Millisecond)
    }

}


