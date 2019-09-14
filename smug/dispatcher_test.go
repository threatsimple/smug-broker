package smug

import (
    "testing"
)

func TestManageBrokers(t *testing.T) {
    cc := &CentralDispatch{}
    b0 := &SlackBroker{}
    cc.AddBroker(b0)
    if cc.NumBrokers() != 1 {
        t.Errorf("expected 1 broker")
    }
    b1 := &SlackBroker{}
    cc.AddBroker(b1)
    if cc.NumBrokers() != 2 {
        t.Errorf("expected 2 brokers")
    }
    cc.RemoveBroker(b1)
    if cc.NumBrokers() != 1 {
        t.Errorf("removing broker errored")
    }
}
