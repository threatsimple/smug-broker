// main dispatch
// handles sending/receiving all events between brokers

package smug

import (
	"fmt"
	"sync"
)

type CentralDispatch struct {
	mux     sync.RWMutex
	log     *Logger
	brokers []Broker
}

func NewCentralDispatch() *CentralDispatch {
	return &CentralDispatch{log: NewLogger("dispatch")}
}

func (cd *CentralDispatch) Broadcast(ev *Event) {
	// hand to all
	cd.mux.RLock()
	for _, b := range cd.brokers {
		if ev.Origin != b {
			go b.HandleEvent(ev, cd)
		}
	}
	cd.mux.RUnlock()
}

func (cd *CentralDispatch) NumBrokers() int {
	return len(cd.brokers)
}

func (cd *CentralDispatch) AddBroker(b Broker) {
	go b.Activate(cd)
	cd.mux.Lock()
	cd.brokers = append(cd.brokers, b)
	cd.mux.Unlock()
}

func (cd *CentralDispatch) RemoveBroker(b Broker) error {
	found := false
	cd.mux.Lock()
	for i, n := range cd.brokers {
		if n == b {
			cd.brokers[i] = cd.brokers[len(cd.brokers)-1]
			cd.brokers = cd.brokers[:len(cd.brokers)-1]
			found = true
			break
		}
	}
	cd.mux.Unlock()
	if !found {
		return fmt.Errorf("broker not found: %s", b.Name())
	}
	return nil
}
