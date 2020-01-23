# General thoughts

Smug consists of several brokers with specialized functionality. A dispatcher
receives events from any broker and immediately publishes it to all other
brokers.  The dispatcher is quite simple by design, with minimal functionality
and is built around the Observer Pattern.

Each broker acts as one or more of a: sink, tap, and/or filter.

*sink* The sink function acts like a destination producer.

*tap* The tap function acts like a source consumer for an arbitrary service.

*filter* The filter function transforms each event

Events are individual messages sent to any broker.  A broker creates an event
and then sends that to the Dispatcher.  The dispatcher then forwards that event
to any other broker.

## Events
An event is quite simple, consisting of the originating broker, a string
representing the author, the simple text of the event and any formatted blocks.
Any broker can decide to ignore formatted blocks so all events should have a
simple text representation to fall back to.

# broker types

At present, there are three types of brokers:  irc, slack, pattern-router.

## irc broker

This broker consumes, and produces to, one irc channel.  Anything sent to this
channel is dispatched to any other active broker.  Likewise, anything sent to
the other active brokers gets published to the irc channel.

**note** the format of the server connection string is
`server.domain.com:portnum`.  So if it's connecting on 6697 for ssl, you'd use
`irc.example.com:6697`.  If you don't specify a port, `:6667` will be appended
for you.

## slack broker

This broker connects to a slack network and brokers between slack and other
brokers.  Anything sent to the slack channel gets dispatched to the other
brokers.  Anything sent to the other brokers gets dispatched to the slack
channel.

Some simple slack formatting is available in the form of simple blocks.

# Configuration File

**quickstart** copy and edit the smug.yaml.template file provided.

At a minimum you will need a configuration file spelling out the brokers.

This file consists of two top level stanzas.

- `brokers` - a collection of broker configuration stanzas
- `active-brokers` - a list of brokerkeys corresponding to a defined broker in
  the brokers collection.

## Environmental Overrides

For each of these brokers, you can override any value in the configuration
file via an environment variable.  This allows you to set tokens and passwords
if needed, without leaving them lying around in a file or passed on the command
line.

To set a config variable via the environment, the name of the broker config
stanza is needed.

The environment variable name consists of the upper case string:

`SMUG_+brokerkey+configvalue`

For example, given the following yaml stanza for an irc broker;

```
brokers:
    ircbroker:
        type    : "slack"
        token   : "fake-token-here"
        channel : "#general"
```

The brokerkey above is `ircbroker`.  This could be any valid slug string.

If you set the environment variable of `SMUG_IRCBROKER_TOKEN=REALTOKEN` then the
environment variable value will be used when the broker is created and connects
to slack.


