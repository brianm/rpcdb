# Distributed RPC Debugger (drdb)

## The Debugger Client

Debugger interface is in a web browser talking to drdbd service
instance. Can issue requests directly from the debugger with addition
of a browser plugin, otherwise will need to C&P to curl most likely
ugh. Maybe client should be an [electrum](http://electron.atom.io/)
based application instead of browser based. Browser based is super
convenient, but rather impinges on ability to initiate arbitrary HTTP,
which is likely what folks will want. Breaking out to a real client
via electron also allows for easier non-HTTP requests to issue
directly from the client.

Alternately, we initiate debug rpc chains from the drdbd service
instance itself. This means the drdbd needs to know how to fire off
arbitrary requests, which might be annoying, but it is also amenable
to the microservice plugin approach -- if you want a new type of
request, implement a microservice that knows how to do it and have it
register with drdbd. Actually, I kind of like this model.

So the debugger client is back to being an in-browser webapp which
communicates with drdbd to fire off initiating requests. RPC protocols
which can be debugged pretty much require the ability to add arbitrary
header information in an envelope. I am going to push out of scope
things that don't have an envelope. As almost everyone tunnels over
http now, even grpc does, let's just simplify and start with HTTP.

So, debug request is fired, it includes a debug header, or headers,
something like:

```
POST /wombat HTTP/1.1
Host: api.xn.io
Debug-Session: https://drdbd1.internal/abc123
Debug-Signature: <digital signature>
Debug-Breakpoint: <some breakpoint definition expression>
Debug-Breakpoint: <another breakpoint>,<a third breakpoint>
Content-type: application/json
Connection: close

{"hello":"world"}
```

The `Debug-Session` header tells the debugger middleware what to
attach to when a breakpoint is tripped.

The `Debug-Breakpoint` headers define a breakpoint. There will need to
be some trivial-to-implement breakpoint expression language used here.
Importantly, the expression language must not use `,` as
[multiple http headers may be combined with `,`](http://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html)
and it is nice to not break http!

The `Debug-Signature` header signs the debug request/session from
drdbd in such a manner that each piece of middleware can confirm this
is a bona fide debug request. This is needed to prevent attacks via
debug middleware.

# Debug Middleware

Services instrumented for debugging will generally do so through
transparent middleware. This requires two types of middleware, and a
means for them to pass information between each other.

Server middleware needs to be able to intercept inbound RPCs and
invoke the debugger if a breakpoint is triggered. They capture the RPC
body and send it off to the debug session, blocking the request until
a response is received. The debugger may send back an alternate RPC
payload, additional breakpoint definitions (including wildcard
breakpoints, to allow things like step-into), request termination,
etc. Once the middleware receives a response from the debugger it
resumes the RPC invocation.

Client middleware instruments RPC clients so that debug information is
passed down the distributed call tree, to catch client rpc breakpoints
(break on the client side of the RPC), etc. Client middleware needs to
be able to do the same kinds of things as server middleware -- send
the payload, and response of RPCs off to the debugger, receive
instructions back, etc.

In addition to manipulating the RPC messages, the middleware will need
to be able to manipulate timeout behavior in order make debug sessions
amenable to human time scales.

# RPC Debug Hooks

The general flow of RPC debug hooks is:

* RPC REQUEST hook in client middleware
* Client sends request to the server
* RPC RECEIVE hook in server middleware
* Server handles RPC
* RPC REPLY hook in server middleware
* Server sends the reply back to client
* RPC RESPONSE hook in client middleware
* Client receives the response

At each step the debugger can manipulate the input and output, inject
additional breakpoints, remove breakpoints, or terminate the flow.

# Debugger Interface

The debugger interface needs to be able to support concurrent RPCs in
the system under debug, including multiple breakpoints tripped at the
same time. It makes sense to visually represent the distributed call
tree, and open some kind of breakpoint interaction interface when a
breakpoint is triggered. The interaction mode should provide the
standard step-in, step-out, step-over debugger options, as well as the
ability to set arbitrary and contingent breakpoints downstream or
upstream of the currently triggered breakpoint.

Additionally, the debugger should be able to trigger trace captures
without actually tripping a breakpoint. That is, any or all of the
debugger hooks should be able to send the rpc details back to the
debugger for analysis without pausing execution in a breakpoint. A
typical workflow might be to issue a trace, then set a breakpoint on a
subsequent request based on the view of the trace.

The desire to set breakpoints based on past requests or traces implies
that the debug ui needs to support some kind of aggregate or historic
view -- the "source code" of the rpc tree which allows for breakpoints
to be set. Because particular call trees tend to change depending on
request & data conditions (respond from cache vs database, etc) we
don't have a nice model for the source code visual breakpoint
establishment folks are used to. GDB style breakpoint expressions,
except on logical service/rpc/endpoint names instead of functions or
line numbers might be handy. Additionally, AOP style cut point
definitions might be useful to mine for ideas.

# Breakpoint Expression Language

We will need some kind of expression language that can describe
breakpoints in a determistic way which is very easy to implement in
every language and rpc system. This might take some thought.

# DRDBD Server Architecture

Because the debugger server needs to support arbitrarily many RPC
systems and many are proprietary it seems the drbdb daemon needs to
support a pretty flexible plugin model. A particular user might want
to support an RPC system completely unique to their infrastructure,
and should be able to.

One way of doing this would be to fully embrace REST and microservices
to support microservice based plugins. Adding an RPC system, message
body type, etc becomes writing a microservice that understands it and
can interact with the debugger client. The plugin registers with the
drdbd service and then can be made use of as needed. This implies an
[I-Tier style](https://engineering.groupon.com/2013/misc/i-tier-dismantling-the-monoliths/)
composite client interface where parts of the debugger interaction
interface are owned completely by plugins. For instance, understanding
legal manipulations of stumpy messages is not something that can be
implemented in the core because unless you are Amazon you have no idea
how it should work.

A nice aspect of this model is that the core RPC mechanisms
(http/json, thift, grpc -- basically the widely used OSS ones) can be
bundled in and started automagically, then proprietary or experimental
ones can be run completely seperately and register with the drdbd
instance.

# Pseudo-Random Notes

Middleware will probably want to reserve a very small amount of
resources for debug sessions, for instance supporting up to at most 5
(configurable) simultaneous active debug sessions before rejecting the
debug requests. This is because the ability to debug in production is
very powerful and possible, but we need to protect the system from
accidental resource starvation through debugging!

Speaking of debugging in production, middleware SHOULD take
appropriate measures to not impact non-debug RPCs in any way, and
prevent debug sessions from impacting non-debug sessions. They must
not obtain or hold locks or resources which are also used or contended
for by non-debug requests.

The design of the system should be sufficiently robust and secure to
support running debug middleware, and attaching debug sessions, to
production systems. The ability to poke through the real system is
exceptionally powerful for tracing down strange conditions, and should
be supported if at all possible.
