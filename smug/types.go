// shared types across smug, makes it easy to ref and share


package smug

import "time"


type Broker interface {
    Name() string
    HandleEvent(*Event, Dispatcher)
    Setup(...string) // at the end of this func, the broker should be able to
                     // Handle(event) as needed, whether that is a queue until
                     // Activate() is called by dispatcher.AddBroker
    Activate(Dispatcher) // this will setup a runloop if needed for the broker
    Deactivate() // must not return anything, will be called during destruction
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
    ReplyTarget string // replyBroker will use this to target a specific user
                       // either privately or some other mechanism. this should
                       // not be changed once set by the originating event as it
                       // may specific to a given broker's format
    Nick string
    Avatar string
    Text string
    RawText string
    ts time.Time
}

