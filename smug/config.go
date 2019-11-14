package smug

import (
    "io/ioutil"

    yaml "gopkg.in/yaml.v2"
)


type PatternConfig struct {
    Name string                 `yaml:"name"`
    Help string                 `yaml:"help"`
    RegEx string                `yaml:"regex"`
    Url string                  `yaml:"url"`
    Method string               `yaml:"method"`
    Headers map[string]string   `yaml:"headers"`
    Vars map[string]string      `yaml:"vars"`
}


// NOTE this is a super set of broker config needs.
// not all brokers will use every member of this Config
// however, doing it this way allows the yaml unmarshal to Just Work(TM)
type BrokerConfig struct {
    Type string                 `yaml:"type"`
    Server string               `yaml:"server"`
    ApiToken string             `yaml:"token"`
    UseSSL bool                 `yaml:"ssl"`
    Nick string                 `yaml:"nick"`
    Channel string              `yaml:"channel"`
    Patterns []PatternConfig    `yaml:"patterns"`
}


type Config struct {
    ActiveBrokers []string          `yaml:"active-brokers"`
    Brokers map[string]BrokerConfig `yaml:"brokers"`
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
    return &cfg
}

