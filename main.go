package main

import (
    "flag"
    "fmt"
    "log"
    "time"

    smug "github.com/nod/smug/smug"
)


var version string


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


type ReadThisOpts struct {
    apibase string
    authcode string
    prefix string
}


func parseReadThisOpts() *ReadThisOpts {
    opts := &ReadThisOpts{}
    flag.StringVar(&opts.apibase, "rt-api", "", "readthis api base url")
    flag.StringVar(&opts.authcode, "rt-auth", "", "readthis auth code")
    flag.StringVar(&opts.prefix, "rt-prefix", "", "readthis prefix trigger")
    return opts
}


type Opts struct {
    irc *IrcOpts
    slack *SlackOpts
    rt *ReadThisOpts
}


func parseOpts() *Opts {
    iopts := parseIrcOpts()
    sopts := parseSlackOpts()
    rtopts := parseReadThisOpts()
    opts := &Opts{
        irc: iopts,
        slack: sopts,
        rt: rtopts,
    }
    flag.Parse()
    return opts
}


func main() {
    log.Printf("starting smug ver:%s", version)
    dispatcher := &smug.CentralDispatch{}

    opts := parseOpts()

    // slack setup
    sb := &smug.SlackBroker{}
    sb.Setup(opts.slack.token, opts.slack.channel)
    go sb.Run(dispatcher)
    dispatcher.AddBroker(sb)

    // irc
    ib := &smug.IrcBroker{}
    ib.Setup(
        opts.irc.server,
        opts.irc.channel,
        opts.irc.nick,
        fmt.Sprintf("%s-%s", "smug", version),
    )
    go ib.Run(dispatcher)
    dispatcher.AddBroker(ib)

    rtb := &smug.ReadThisBroker{}
    rtb.Setup(opts.rt.apibase, opts.rt.prefix, opts.rt.authcode)
    dispatcher.AddBroker(rtb)

    // just loop here for now so others can run
    for true {
        time.Sleep(200 * time.Millisecond)
    }

}

