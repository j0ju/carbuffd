# carbuffd - Carbon Buffer

carbuffd was written to enable small IoT devices to deliver metrics using
a simplyfied Carbon line protocol.
The Carbon line protocol is very simple to implement.

Simply said, it is just another Carbon relay, just augmenting carbon metrics without
valid epoch from devices, that might not have a local time.

## Compiling

Get it, `go build` and/or `go install`, do it your prefered way.

## Usage

The basic usage is
>        ./carbuffd ListenAddress:Port RemoteAddress:Port

for example
>        ./carbuffd :2003 203.0.113.1:2003

## Use patterns

### Simple usage

Lets assume you have a network that is dedicated for IoT sensor devices
Just run
>         carbuffd :2003 203.0.113.1:2003
Where
* `:2003` is the listen address.
* `203.0.113.1:2003` is your carbon endpoint.
Now let the devices send data to the service IP carbuffd is listening on.

### Transparent proxy

If you have a small network, you might want to consider running this on Linux based
network routers.
With `iptables` you can intercept all traffic going to 203.0.113.1:2003 or just :2003
and let `carbuffd` augment metrics without time and forward them to your
Carbon Endpoint/Relay/Server you trust ;)

>        ./carbuffd :2003 203.0.113.1:2003
>        iptables -t nat -I PREROUTING -p tcp --dport 2003 -m --ctstate NEW j REDIRECT--to-ports 2003

Of course this might work on a BSD similary.

### Unified carbon transport service with small buffering capabilities

Lets say you have legacy^W longtime proven stable software environment, consisting of Shell, Perl,
YouNameIt(tm) scripts you would like to instrument.

Carbon metrics are easily generated in any language.
If you send these to `carbuffd`, you will get:
 * a simply fied transport facility, do not write the same `socat`, `netcat`, ...
   stuff over and over again, to send metrics to carbon server, AND think about error handling.
   A local daemon is easier to control.
 * a little buffering, if your Carbon endpoint is currently not available


You can run a local carbuffd on every machine, with Salt/Puppet/... this is easy.

## Protocol

### Orignal Carbon Line Protocol
See [documentation from the Graphite docs](http://graphite.readthedocs.io/en/latest/feeding-carbon.html#the-plaintext-protocol)

>         metric.path value epoch

Example

>         collectd.example_com.load.1m 0.10 1234567890

### Simplified IoT Carbon Line Protocol
>         metric.path value

Example

>         iot.by-id.5c-cf-7f-c0-ff-ee.dew_point 3.191501

This will augmented to above the above. The epoch will be the time the metric has been received.
>         iot.by-id.5c-cf-7f-c0-ff-ee.dew_point 3.191501 1234567890

Lines with valid epoch will not be altered.

## Inner workings
 * A largly size chan string is the buffer, maybe it might be adjusted via command line later.
 * Each connection gets it's go routine worker.

## ToDo & Ideas
 * A largly size chan string is the buffer, maybe it might be adjusted via command line later.
 * Add a device IoT device back channel for small control if IoT devices.

## Other
I used this project to learn some basics about Golang concurency.
I was inspired [here](https://divan.github.io/posts/go_concurrency_visualize/)

I do not want to compete with the original Carbon Relay written in python,
the fabulous [carbon-c-relay](https://github.com/grobian/carbon-c-relay) or
the [go-carbon-relay-ng](https://github.com/graphite-ng/carbon-relay-ng).
Maybe they pickup this protocol derivation sometime.

If you are using this project just drop me a message.
