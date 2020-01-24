package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	smug "github.com/threatsimple/smug-broker/smug"
)

var version string

type RuntimeOpts struct {
	configFile  string
	loglevel    string
	showVersion bool
}

func buildRuntimeOpts() *RuntimeOpts {
	opts := &RuntimeOpts{
		configFile: "smug.conf",
		loglevel:   "warning",
	}
	flag.StringVar(&opts.configFile,
		"config", "smug.conf", "config file path")
	flag.StringVar(&opts.loglevel, "loglevel", "warning", "logging level")
	flag.BoolVar(&opts.showVersion, "version", false,
		"display version and exit")
	return opts
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

func parseConfig() (*RuntimeOpts, *smug.Config) {
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

	cfg := smug.LoadConfig(runopts.configFile)
	return runopts, cfg
}

func MakeIrcBroker(cfg *smug.BrokerConfig) smug.Broker {
	server := cfg.Server
	// cfg.GetBool("ssl")
	nick := cfg.Nick
	channel := cfg.Channel
	ib := &smug.IrcBroker{}
	ib.Setup(
		server,
		channel,
		nick,
		fmt.Sprintf("%s-%s", "smug", version),
	)
	return ib
}

func MakeSlackBroker(cfg *smug.BrokerConfig) smug.Broker {
	token := cfg.ApiToken
	channel := cfg.Channel
	sb := &smug.SlackBroker{}
	sb.Setup(token, channel)
	return sb
}

func MakePatternBroker(cfg *smug.BrokerConfig) smug.Broker {
	pb := &smug.PatternRoutingBroker{}
	pb.Setup()
	for _, p := range cfg.Patterns {
		if p.RegEx == "" {
			ErrorAndExit("pattern broker pattern.regex must not be blank")
		}
		if p.Url == "" {
			ErrorAndExit("pattern broker pattern.url must not be blank")
		}
		if p.Method == "" {
			ErrorAndExit("pattern broker pattern.method must not be blank")
		}
		// now build our pattern
		newp, _ := smug.NewExtendedPattern(
			p.Name, p.RegEx, p.Url, p.Headers, p.Vars, p.Method, p.Help)
		pb.AddPattern(newp)
	}
	return pb
}

type BrokerBuilder func(*smug.BrokerConfig) smug.Broker

func makeBroker(brokerKey string, cfg *smug.BrokerConfig) (smug.Broker, error) {
	broker_types := map[string]BrokerBuilder{
		"irc":     MakeIrcBroker,
		"pattern": MakePatternBroker,
		"slack":   MakeSlackBroker,
	}
	brokerType := cfg.Type
	if broker_factory, ok := broker_types[brokerType]; ok {
		// valid broker, make it up!
		return broker_factory(cfg), nil
	} else {
		return nil, fmt.Errorf("invalid broker type: %s", brokerType)
	}
}

func createBrokers(cfg *smug.Config) []smug.Broker {
	brokers := []smug.Broker{}
	for _, ab := range cfg.ActiveBrokers {
		brcfg, found := cfg.Brokers[ab]
		if !found {
			panic(fmt.Sprintf("missing broker config: %s", ab))
		}
		b, err := makeBroker(ab, brcfg)
		if err != nil {
			panic(err)
		}
		brokers = append(brokers, b)
	}
	return brokers
}

func main() {
	opts, cfg := parseConfig()

	// setup logging first
	smug.SetupLogging(opts.loglevel)

	log := smug.NewLogger("smug")
	maxprocs := runtime.GOMAXPROCS(-1)
	log.Infof("starting smug ver:%s gomaxprocs:%d", version, maxprocs)

	dispatcher := smug.NewCentralDispatch()

	// setup our localcmdbroker first
	lc := &smug.LocalCmdBroker{}
	lc.Setup("smug", "", version)
	dispatcher.AddBroker(lc)
	defer dispatcher.RemoveBroker(lc)

	// now brokers from config
	brokers := createBrokers(cfg)
	for _, b := range brokers {
		dispatcher.AddBroker(b)
		defer dispatcher.RemoveBroker(b)
	}

	// just loop here for now so others can run like happy little trees
	for true {
		time.Sleep(400 * time.Millisecond)
	}

}
