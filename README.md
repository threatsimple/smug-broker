
```
 ___ _ __ ___  _   _  __ _ 
/ __| '_ ` _ \| | | |/ _` |
\__ \ | | | | | |_| | (_| |
|___/_| |_| |_|\__,_|\__, |
                     |___/ 
```

# smug

Broker communications between irc, slack, other services.

# usage

Download a release or build as documented below.

Copy the `smug.conf.template` file to `smug.conf`.  Edit the file to set your
connection parameters and keys.

Run like the following:

```bash
$ ./smug -conf=smug.conf
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


