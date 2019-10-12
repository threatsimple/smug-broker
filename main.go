package main

import (
    "flag"
    "fmt"
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


type RuntimeOpts struct {
    loglevel string
    showVersion bool
}


func parseRuntimeOpts() *RuntimeOpts {
    opts := &RuntimeOpts{
        loglevel: "warning",
    }
    flag.StringVar(&opts.loglevel, "loglevel", "warning", "logging level")
    flag.BoolVar(&opts.showVersion, "version", false,
        "display version and exit")
    return opts
}


type Opts struct {
    irc *IrcOpts
    slack *SlackOpts
    rt *ReadThisOpts
    runtime *RuntimeOpts
}


func parseOpts() *Opts {
    iopts := parseIrcOpts()
    sopts := parseSlackOpts()
    rtopts := parseReadThisOpts()
    runopts := parseRuntimeOpts()
    opts := &Opts{
        irc: iopts,
        slack: sopts,
        rt: rtopts,
        runtime: runopts,
    }
    flag.Parse()
    return opts
}


func main() {
    opts := parseOpts()

    // show version and exit?
    if opts.runtime.showVersion {
        fmt.Printf("version: %s\n", version)
        return
    }

    // setup logging first
    smug.SetupLogging("debug")

    log := smug.NewLogger("smug")

    log.Infof("starting smug ver:%s", version)

    dispatcher := smug.NewCentralDispatch()

    // slack setup
    sb := &smug.SlackBroker{}
    sb.Setup(opts.slack.token, opts.slack.channel)
    go sb.Run(dispatcher)
    dispatcher.AddBroker(sb)
    defer dispatcher.RemoveBroker(sb)

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
    defer dispatcher.RemoveBroker(ib)

    lc := &smug.LocalCmdBroker{}
    lc.Setup("smug", "", version)
    dispatcher.AddBroker(lc)
    defer dispatcher.RemoveBroker(lc)

    rtb := &smug.ReadThisBroker{}
    rtb.Setup(opts.rt.apibase, opts.rt.prefix, opts.rt.authcode)
    dispatcher.AddBroker(rtb)
    defer dispatcher.RemoveBroker(rtb)

    // just loop here for now so others can run
    for true {
        time.Sleep(400 * time.Millisecond)
    }

}

