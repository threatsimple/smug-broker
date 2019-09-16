package smug

import "time"


type Broker interface {
    Name() string
    Publish(*Event)
}


type Dispatcher interface {
    Broadcast(*Event)
    AddBroker(Broker)
    RemoveBroker(Broker) error
    NumBrokers() int
}


type Event struct {
    Origin Broker
    Nick string
    Avatar string
    Text string
    ts time.Time
}


