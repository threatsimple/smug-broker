package smug


import (
    "fmt"
    "testing"
)


type TestDispatch struct {
    lastbroadcast *Event
}


// our mock Broadcast captures the last event broadcast to it in the
// lastbroadcast member. this is not threadsafe in any way
func (td *TestDispatch) Broadcast(ev *Event) {
    td.lastbroadcast = ev
}
func (td *TestDispatch) AddBroker(Broker) {}
func (td *TestDispatch) RemoveBroker(Broker) error { return fmt.Errorf("wat?") }
func (td *TestDispatch) NumBrokers() int { return 0 }


func TestLocalVersionCommand(t *testing.T) {
    myver := "99.99.99"
    vc := &VersionCommand{Version:myver}
    td := &TestDispatch{}

    // test our version command match
    if ! vc.match(&Event{Text:"..version yo"}) {
        t.Errorf("did not match on version command")
    }

    // now test the version string returns properly
    e := &Event{}
    vc.exec(e,e,td)
    if td.lastbroadcast.Text != fmt.Sprintf("version: %s",myver) {
        t.Errorf(
            "version not returned.  expected: version: %s\ngot %s",
            myver,
            td.lastbroadcast.Text )
    }
}


