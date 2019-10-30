package main

import (
    "flag"
    "fmt"
    "os"
    "time"

    hocon "github.com/go-akka/configuration"

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
    configFile string
    loglevel string
    showVersion bool
}


func buildRuntimeOpts() *RuntimeOpts {
    opts := &RuntimeOpts{
        configFile: "smug.conf",
        loglevel: "warning",
    }
    flag.StringVar(&opts.configFile,
        "config", "smug.conf", "config file path")
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


func ErrorAndExit(msg string) {
    fmt.Printf(
        "ERR %s\nusage: %s",
        msg,
        os.Args[0],
    )
    flag.PrintDefaults()
    os.Exit(2)
}


func parseOpts() (*RuntimeOpts, *hocon.Config) {
    runopts := buildRuntimeOpts()
    flag.Parse()

    // show version and exit?  short-circuit here
    if runopts.showVersion {
        fmt.Printf("version: %s\n", version)
        os.Exit(1)
    }

    // is configfile specified or blank?
    if runopts.configFile == "" {
        ErrorAndExit(fmt.Sprintf("missing required config file"))
    }
    // does the file at least exist?
    if _, err := os.Stat(runopts.configFile); os.IsNotExist(err) {
        ErrorAndExit(fmt.Sprintf(
            "config file not found: %s\n",
            runopts.configFile,
        ))
    }

    cfg := hocon.LoadConfig(runopts.configFile)
    return runopts, cfg
}


func MakeIrcBroker(cfg *hocon.Config) smug.Broker {
    server := cfg.GetString("server")
    // cfg.GetBool("ssl")
    nick := cfg.GetString("nick")
    channel := cfg.GetString("channel")
    ib := &smug.IrcBroker{}
    ib.Setup(
        server,
        channel,
        nick,
        fmt.Sprintf("%s-%s", "smug", version),
    )
    return ib
}


func MakeSlackBroker(cfg *hocon.Config) smug.Broker {
    token := cfg.GetString("token")
    channel := cfg.GetString("channel")
    sb := &smug.SlackBroker{}
    sb.Setup(token, channel)
    return sb
}

func MakePatternBroker(cfg *hocon.Config) smug.Broker {
    pb := &smug.PatternRoutingBroker{}
    pats := cfg.GetNode("patterns").GetObject()
    for _,k := range pats.GetKeys() {
        name := cfg.GetString(fmt.Sprintf("patterns.%s.name", k))
        if name == "" {
            ErrorAndExit("pattern broker pattern.name must not be blank")
        }
        p := cfg.GetConfig(fmt.Sprintf("patterns.%s", k))
        reg := p.GetString("regex")
        if reg == "" {
            ErrorAndExit("pattern broker pattern.regex must not be blank")
        }
        url := p.GetString("url")
        if url == "" {
            ErrorAndExit("pattern broker pattern.url must not be blank")
        }
        method := p.GetString("method")
        if method == "" {
            ErrorAndExit("pattern broker pattern.method must not be blank")
        }
        // do we have headers that get attached?
        hdrs := map[string]string{}
        hdrNode := p.GetNode("headers")
        if hdrNode != nil {
            hdrObj := hdrNode.GetObject()
            for _,hk := range hdrObj.GetKeys() {
                hdrs[hk] = p.GetString(fmt.Sprintf("headers.%s",hk))
            }
        }
        // are there additional vars to attach
        vars := map[string]string{}
        varsNode := p.GetNode("vars")
        if varsNode != nil {
            varsObj := p.GetNode("vars").GetObject()
            for _,vk := range varsObj.GetKeys() {
                vars[vk] = p.GetString(fmt.Sprintf("vars.%s",vk))
            }
        }
        // now build our pattern
        newp,_ := smug.NewExtendedPattern(name, reg, url, hdrs, vars, method)
        pb.AddPattern(newp)
    }
    return pb
}

type BrokerBuilder func(*hocon.Config) smug.Broker

func makeBroker(brokerKey string, cfg *hocon.Config) (smug.Broker,error) {
    broker_types := map[string]BrokerBuilder {
        "irc": MakeIrcBroker,
        "pattern": MakePatternBroker,
        "slack": MakeSlackBroker,
    }

    brobj := cfg.GetConfig(fmt.Sprintf("brokers.%s",brokerKey))
    if brobj == nil {
        return nil,fmt.Errorf(
            "broker config brokers.%s object not found", brokerKey,
        )
    }

    brokerType := brobj.GetString("type")
    if broker_factory,ok := broker_types[brokerType]; ok {
        // valid broker, create it up!
        return broker_factory(brobj),nil
    } else {
        return nil, fmt.Errorf("invalid broker type: %s", brokerType)
    }
}


func createBrokers(cfg *hocon.Config) []smug.Broker {
    active_brokers := cfg.GetStringList("active-brokers")
    brokers := []smug.Broker{}
    for _,ab := range active_brokers {
        b,err := makeBroker(ab, cfg)
        if err != nil {
            panic(err)
        }
        brokers = append(brokers, b)
    }
    return brokers
}


func main() {
    opts,cfg := parseOpts()

    // setup logging first
    smug.SetupLogging(opts.loglevel)

    log := smug.NewLogger("smug")
    log.Infof("starting smug ver:%s", version)

    dispatcher := smug.NewCentralDispatch()

    // setup our localcmdbroker first
    lc := &smug.LocalCmdBroker{}
    lc.Setup("smug", "", version)
    dispatcher.AddBroker(lc)
    defer dispatcher.RemoveBroker(lc)

    // now brokers from config
    brokers := createBrokers(cfg)
    for _,b := range brokers {
        dispatcher.AddBroker(b)
        defer dispatcher.RemoveBroker(b)
    }

    /*
    // slack setup
    sb := &smug.SlackBroker{}
    sb.Setup(opts.slack.token, opts.slack.channel)
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
    dispatcher.AddBroker(ib)
    defer dispatcher.RemoveBroker(ib)

    prb := &smug.PatternRoutingBroker{}
    // prbrtb.Setup(opts.rt.apibase, opts.rt.prefix, opts.rt.authcode)
    dispatcher.AddBroker(prb)
    defer dispatcher.RemoveBroker(prb)
    */

    // just loop here for now so others can run like happy little trees
    for true {
        time.Sleep(400 * time.Millisecond)
    }

}

