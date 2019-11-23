
```
 ___ _ __ ___  _   _  __ _ 
/ __| '_ ` _ \| | | |/ _` |
\__ \ | | | | | |_| | (_| |
|___/_| |_| |_|\__,_|\__, |
                     |___/ 
```

# smug

Want to mirror everything from one irc to another with ease?

Want to proxy everything from a slack channel to an irc channel?

Want to proxy everything from a slack channel to an external RESTful api?

Broker communications between irc, slack, other services.

# usage

Download a release or build as documented below.

Copy the `smug.yaml.template` file to `smug.yaml`.  Edit the file to set your
connection parameters and keys.

Run like the following:

```bash
$ ./smug -conf=smug.yaml
```

# building

Assuming golang version 1.12+ installed, you can do the following:

```bash
$ cd $GOPATH/src
$ git clone https://github.com/nod/smug
$ cd smug
$ make test
$ make
```

If all goes well, this should create a `build/smug` as a compiled binary at that
point.

