package smug

import (
	"testing"
)

type FakeBroker struct{}

func (fb *FakeBroker) Name() string                       { return "faker" }
func (fb *FakeBroker) HandleEvent(e *Event, d Dispatcher) {}
func (fb *FakeBroker) Setup(...string)                    {}
func (fb *FakeBroker) Activate(dis Dispatcher)            {}
func (fb *FakeBroker) Deactivate()                        {}

func TestManageBrokers(t *testing.T) {
	cc := &CentralDispatch{}
	b0 := &FakeBroker{}
	cc.AddBroker(b0)
	if cc.NumBrokers() != 1 {
		t.Errorf("expected 1 broker")
	}
	b1 := &FakeBroker{}
	cc.AddBroker(b1)
	if cc.NumBrokers() != 2 {
		t.Errorf("expected 2 brokers")
	}
	cc.RemoveBroker(b1)
	if cc.NumBrokers() != 1 {
		t.Errorf("removing broker errored")
	}
}
