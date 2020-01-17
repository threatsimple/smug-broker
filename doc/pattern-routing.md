# Pattern Routing on Text Input

There is a fairly flexible module that is defined in the config file as:

  `type = "pattern"`

This allows you to define regular expressions that if they match against input,
will then create a specially formatted payload and forward to an HTTP API
handler.  If that handler replies back with a certain formatted payload, you can
return that results to the brokers (and display output in slack, irc, etc).


The regex is simply just a standard golang based regular expression, but with
the added ability to extract named matches and include those in the api payload.


For instance, a regex of `^..echo` would match ANY input with a starting string
of `..echo`.

If you used the regex `^..echo\b(?P<what>.+)$` then any line started with
`..echo ` and following text would match and a var of `what` would include the
text following the `..echo ` portion.

## help text

The command `..list` will provide a message containing the help text from any
defined Pattern with a non-empty `help` attribute.

## API Payload

The API body will be a json encoded payload, and a content-type header of
`application/json` will be set on the request.

The body will always include two members:

- `actor` - if irc or slack, this is the nick of the user speaking
- `text` - the raw text of the entire input

Any other named regex groups will also be included.  In our example of the echo
`what` match from above, an input of `..echo hello` from `joe` would result in the
following payload:

```
{
  "actor": "joe",
  "text": "..echo hello",
  "what": "hello"
}
```

## Return Messags from API

If the API return value is a json body with a key of `text`, that will be
returned as a message to the other brokers.

Our echo example would return:

```
{
  "text": "hello"
}
```

For more advanced formatting or communication options, an optional `blocks`
member can be returned also.  The `blocks` attribute should be a json list of
`block` objects, with optional members of:

- `title` - optional title of the formatted block
- `text` - a markdown formatted block of text
- `img` - an image url to show as an accesssory for the formatted block

A more advanced echo member might be:

```
{
  "text": "hello",
  "blocks": [
    {
      "title": "echoing!"
    },
    {
      "text": "hello",
      "img": "http://example.com/some/speaking/clipart.jpg"
    }
  ]
}
```

