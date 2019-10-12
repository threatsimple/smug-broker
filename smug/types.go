// shared types across smug, makes it easy to ref and share


package smug

import "time"


type Broker interface {
    Name() string
    Publish(*Event, Dispatcher)
}


type Dispatcher interface {
    Broadcast(*Event)
    AddBroker(Broker)
    RemoveBroker(Broker) error
    NumBrokers() int
}


type Event struct {
    IsCmdOutput bool
    Origin Broker
    ReplyBroker Broker // all brokers will see message but may choose to ignore
                       // unless beneficial (bot handlers, etc)
    ReplyNick string // replyBroker will use this to target a specific user
                     // either privately or some other mechanism. this should
                     // not be changed once set by the originating event as it
                     // may specific to a given broker's format
    Nick string
    Avatar string
    Text string
    RawText string
    ts time.Time
}

