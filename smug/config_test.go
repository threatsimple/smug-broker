package smug

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	fixcfg := LoadConfig("test_fixtures/test.yaml")
	if fixcfg.Brokers["tester"].Server != "some.example.com" {
		t.Errorf("err: should not have matched")
	}
}

func TestEnvConfig(t *testing.T) {
	const sv = "blerg"
	os.Setenv("SMUG_TESTER_SERVER", sv)
	fixcfg := LoadConfig("test_fixtures/test.yaml")
	if fixcfg.Brokers["tester"].Server != sv {
		t.Errorf(
			"err: setting via env. got %s wanted %s",
			fixcfg.Brokers["tester"].Server, sv,
		)
	}
}
