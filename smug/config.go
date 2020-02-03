package smug

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type PatternConfig struct {
	Name    string            `yaml:"name"`
	Help    string            `yaml:"help"`
	RegEx   string            `yaml:"regex"`
	Url     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
	Vars    map[string]string `yaml:"vars"`
}

// NOTE this is a super set of broker config needs.
// not all brokers will use every member of this Config
// however, doing it this way allows the yaml unmarshal to Just Work(TM)
type BrokerConfig struct {
	Type     string          `yaml:"type"`
	Server   string          `yaml:"server" envcfg:"SERVER"`
	ApiToken string          `yaml:"token" envcfg:"APITOKEN"`
	UseSSL   bool            `yaml:"ssl" envcfg:"SSL"`
	Nick     string          `yaml:"nick" envcfg:"NICK"`
	Channel  string          `yaml:"channel" envcfg:"CHANNEL"`
	Patterns []PatternConfig `yaml:"patterns"`
}

type TelemetryConfig struct {
    Host    string          `yaml:"host"`
    Port    int          `yaml:"port"`
}

type Config struct {
	ActiveBrokers []string                 `yaml:"active-brokers"`
	Brokers       map[string]*BrokerConfig `yaml:"brokers"`
    Metrics       map[string]*TelemetryConfig `yaml:"metrics"`
}

func envOverrides(cfg *Config) {
	// populates from any environment variables
	for key, bcfg := range cfg.Brokers {
		b := reflect.TypeOf(*bcfg)
		for i := 0; i < b.NumField(); i++ {
			fld := b.Field(i)
			envkey := fld.Tag.Get("envcfg")
			if envkey == "" {
				continue
			}
			envnm := fmt.Sprintf(
				"SMUG_%s_%s", strings.ToUpper(key), envkey)
			val := os.Getenv(envnm)
			if val == "" {
				continue
			}
			bf := reflect.ValueOf(bcfg).Elem().Field(i)
			bf.SetString(val)
		}
	}
}

func LoadConfig(configFile string) *Config {
	cfg := Config{}
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		panic(err)
	}
	envOverrides(&cfg)
	return &cfg
}
