# breadis

A proxy for a redis cluster which automatically handles sharding and failover.

## Features

* To the client, acts just like a redis server. No changes necessary for any
  redis client to communicate with it

* Automatic, transparent sharding and failover using sentinel

* Support for the `key{shardkey}` syntax, so shard keys can be specified

* No limit on number of breadis instances per cluster, no intercommunication
  between breadis instances required

## Usage

Compile with

    go build breadis.go

See command-line options with

    ./breadis --help

Load command-line options from config file with

    ./breadis --config <file>

See example config file with

    ./breadis --example

# Modes

There are two modes breadis can run in: single and multi

## Single mode

In single mode breadis only provides failover for a single redis master. You
give it the name this master is being monitored in sentinel with as the
`single-bucket` parameter. This mode is sufficient if you only care about
transparent failover of a single node, and not any sort of sharding or scaling.

## Multi mode

In this mode breadis will handle both failover and sharding. Keys will be
assigned a "bucket", where each bucket is the name of a master sentinel is
keeping track of. There is also a special bucket, the locator, which is another
redis master/slave whose sole purpose is to store the bucket names associated
with all keys being used.

Here is a high-level look at the process of performing a command in multi mode
(assuming default configuration):

* KEY in the command is found. If the key follows the `key{shardkey}` format,
  then `shardkey` is used as KEY for the rest of the steps.

* `get loc:KEY` is called on the locator node. If a bucket name is returned, the
  command is perfomed on the master of that bucket and the response returned to
  the client.

* If a bucket is not found for KEY, `SRANDMEMBER members` is called on the
  locator node. `members` is a key on the node containing a set of bucket names
  available for assigning new keys to.

* This bucket is set on the `loc:KEY` key in the locator for future commands,
  the command is forwarded to the bucket, and life goes on.

This configuration is different than most sharding schemes, and has the
following properties:

Cons:

* Extra round-trip for EVERY command. No sugar-coating this, it isn't optimal.
  But redis is fast, and in my experience this doesn't add *much* latency,
  especially if the locator node is on the same physical machine as breadis.

* Redistributing/cleaning existing keys is not easy, unless you only have a
  single instance of breadis, in which case it is easy (although not yet
  supported)

Pros:

* Simple to understand, and no locking issues or need for communication between
  breadis instances

* Easy and safe to add new masters (increasing capacity). Just add the master's
  name to the members pool on the locator

* Key locations can be (and are) cached in breadis, so the hit to the locator
  can be diminished for commonly used keys

# Limitations/behavior

* Cannot perform any commands whose first argument is not a key name, or any
  commands which take in more than one key as an argument (mget, mset, etc...)

* Will crash upon any communication failure to sentinel. Should be running in
  supervisor-type program (supervisord, upstart, systemd, etc...)
