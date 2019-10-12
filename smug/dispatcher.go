
// main dispatch
// handles sending/receiving all events between brokers

package smug

import "fmt"


type CentralDispatch struct {
    log *Logger
    brokers []Broker
}


func NewCentralDispatch() *CentralDispatch {
    return &CentralDispatch{log: NewLogger("dispatch")}
}


func (cd *CentralDispatch) Broadcast(ev *Event) {
    // is this a directed message in reply from earlier??
    if ev.ReplyBroker != nil {
        ev.ReplyBroker.Publish(ev, cd)
        return
    }
    // hand to all
    for _,b := range cd.brokers {
        if ev.Origin != b {
            b.Publish(ev, cd)
        }
    }
}


func (cd *CentralDispatch) NumBrokers() int {
    return len(cd.brokers)
}


func (cd *CentralDispatch) AddBroker(b Broker) {
    cd.brokers = append(cd.brokers, b)
}


func (cd *CentralDispatch) RemoveBroker(b Broker) error {
    found := false
    for i,n := range cd.brokers {
        if n == b {
            cd.brokers[i] = cd.brokers[len(cd.brokers)-1]
            cd.brokers = cd.brokers[:len(cd.brokers)-1]
            found = true
            break
        }
    }
    if !found {
        return fmt.Errorf("broker not found: %s", b.Name())
    }
    return nil
}


