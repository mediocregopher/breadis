# breadis

A proxy for redis cluster which automatically handles command redirection and
failover detection

To the client, acts just like a redis server. No changes necessary for any redis
client to communicate with it.

## Usage

Compile with

    go build

See command-line options with

    ./breadis --help

Load command-line options from config file with

    ./breadis --config <file>

See example config file with

    ./breadis --example

# Use case

breadis acts as a simple proxy to an entire redis cluster, acting exactly the
same as a normal single redis instance (so any existing redis driver can already
talk to breadis).

This might seem a bit silly, since you could just use a cluster-aware driver in
your application to interact with the redis cluster. But this has two downsides:

* You have to know you're interacting with a redis cluster in the first place,
  which may not always be the case. For instance, if in a dev environment an
  application commmunicates with a single redis on `localhost:6379`, but in
  production there is a cluster on multiple other boxes, that's something the
  application has to differentiate. Instead, the dev environment could keep
  using a single local redis, but in production the application server could
  have breadis listening on port 6379, meaning the application wouldn't have to
  know the difference.

* The redis cluster handling isn't trivial, and works best when there is a
  persistant process managing it. This may not be possible in stateless, request
  based languages (e.g. PHP) or if you're executing a bunch of one-off scripts
  which use redis cluster. For these cases having a breadis instance handle the
  hard parts and just having the application hit it instead of doing the hard
  parts itself.
