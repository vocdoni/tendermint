# ADR 062: P2P Abstractions

## Changelog

- 2020-11-06: Initial version (@erikgrinaker)

## Context

[ADR 061](adr-061-p2p-refactor-scope.md) decided to refactor the peer-to-peer (P2P) networking stack. The first phase of this is to redesign and refactor the internal P2P architecture and implementation, while retaining protocol compatibility as far as possible.

## Alternative Approaches

> This section contains information around alternative options that are considered before making a decision. It should contain a explanation on why the alternative approach(es) were not chosen.

## Decision

The P2P stack will be redesigned as a message-oriented architecture, primarily relying on Go channels for communication and scheduling. It will use arbitrary stream transports to communicate with peers, peer-addressable channels to pass Protobuf messages between peers, and a router that routes messages between reactors and peers. Message passing is asynchronous with at-most-once delivery.

## Detailed Design

This ADR is primarily concerned with the architecture and interfaces of the P2P stack, not their internal implementation details. Since implementations can be non-trivial, separate ADRs may be submitted for these. The APIs described here should therefore be considered a rough architecture outline, not a complete and final design.

Primary design objectives have been:

* Loose coupling between components, for a simpler architecture with increased testability.
* Better quality-of-service scheduling of messages, with backpressure and increased performance.
* Centralized peer lifecycle and connection management.
* Better peer address detection, advertisement, and exchange.
* Pluggable transports (not necessarily networked).
* Backwards compatibility with the current P2P network protocols.

The main abstractions in the new stack (following [Go visibility rules](https://golang.org/ref/spec#Exported_identifiers)) are:

* `peer`: A node in the network, uniquely identified by a `PeerID`.

* `Transport`: An arbitrary mechanism to exchange raw bytes with a peer.

* `Channel`: A bidirectional channel to exchange Protobuf messages with arbitrary peers, addressed by `PeerID`. There can be any number of channels, each of which can pass one specific message type.

* `Router`: Maintains transport connections to peers and routes channel messages.

* `peerStore`: Stores peer data for the router, in memory and/or on disk.

* Reactor: While this was a first-class concept in the old P2P stack, it is simply a design pattern in the new design, loosely defined as "something which listens on a channel and reacts to messages" (e.g. as simple as a function).

These concepts and related entities are described in detail below, in a bottom-up order.

### Transports

Transports are arbitrary mechanisms for exchanging raw bytes with a peer. For example, a gRPC transport would connect to a peer over TCP/IP and send bytes using the gRPC protocol, while an in-memory transport might connect to a peer running in another goroutine using internal byte buffers. Note that transports don't have a notion of a `peer` as such - instead, they use arbitrary endpoint addresses, to decouple them from P2P stack internals.

Transports must satisfy a few requirements:

* Be connection-oriented, and support both listening for inbound connections and making outbound connections, using arbitrary endpoint addresses.

* Support multiple logical byte streams within a single connection. For example, QUIC has native support for separate independent streams, while HTTP/2 and MConn multiplex streams over a single TCP connection. This is necessary in order to take advantage of native stream support in transport protocols such as QUIC.

* Provide the public key of the peer, and possibly encrypt or sign the traffic as appropriate. This should be compared with known data (e.g. the peer ID) to authenticate the peer and avoid man-in-the-middle attacks.

The initial transport implementation will be a port of the current MConn protocol currently used by Tendermint, which should be backwards-compatible at the wire level.

The `Transport` interface is:

```go
type Transport interface {
	// Accept waits for the next inbound connection, or until the context is
	// cancelled.
	Accept(context.Context) (Connection, error)

	// Dial creates an outbound connection to an endpoint.
	Dial(context.Context, Endpoint) (Connection, error)

    // Endpoints lists endpoints the transport is listening on. Any endpoint IP
    // addresses do not need to be normalized in any way (e.g. 0.0.0.0 is
    // valid), as they will be preprocessed before being advertised.
	Endpoints() []Endpoint
}
```

#### Endpoints

`Endpoint` represents a transport endpoint. A connection is always between two endpoints: one at the local node and one at the remote peer. An outbound connection to a remote endpoint is made via a `Dial()` call, and inbound connections to a local endpoint the transport is listening on (as reported by `Endpoints()`) are returned via `Accept()`.

The `Endpoint` struct and related types is:

```go
type Endpoint struct {
    // Protocol specifies the transport protocol, used by the router to pick a
    // transport for an endpoint.
	Protocol Protocol

	// Path is an optional, arbitrary transport-specific path or identifier.
	Path string

	// IP is an IP address (v4 or v6) to connect to. If set, this defines the
    // endpoint as a networked endpoint.
	IP net.IP

    // Port is a network port (either TCP or UDP). If not set, a default port
    // may be used depending on the protocol.
	Port uint16
}

type Protocol string
```

Endpoints are arbitrary transport-specific addresses, but if they are networked they must use IP addresses, and thus rely on IP as a fundamental packet routing protocol. This is to be able to make certain decisions about address discovery, advertisement, and exchange - for example, a `192.168.0.0/24` IP address should only be advertised to peers on that IP network, while `8.8.8.8` may be advertised to all peers. Similarly, any port numbers if given must represent TCP and/or UDP port numbers, in order to use [UPnP](https://en.wikipedia.org/wiki/Universal_Plug_and_Play) to autoconfigure e.g. NAT gateways.

Non-networked endpoints (i.e. with no IP address) are considered local, and will only be advertised to other peers connecting via the same protocol.

## Status

Proposed

## Consequences

> This section describes the consequences, after applying the decision. All consequences should be summarized here, not just the "positive" ones.

### Positive

### Negative

### Neutral

## References

> Are there any relevant PR comments, issues that led up to this, or articles referenced for why we made the given design choice? If so link them here!

- {reference link}
