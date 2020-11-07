# ADR 062: P2P Abstractions

## Changelog

- 2020-11-06: Initial version (@erikgrinaker)

## Context

In [ADR 061](adr-061-p2p-refactor-scope.md) we decided to refactor the peer-to-peer (P2P) networking stack. The first phase of this is to redesign and refactor the internal P2P architecture and implementation, while retaining protocol compatibility as far as possible.

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
// Transport is an arbitrary mechanism for exchanging bytes with a peer.
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

How the transport sets up listening is transport-dependent, and not covered by the interface. This is typically done e.g. during transport construction.

#### Endpoints

`Endpoint` represents a transport endpoint. A connection is always between two endpoints: one at the local node and one at the remote peer. An outbound connection to a remote endpoint is made via a `Dial()` call, and inbound connections received via a local endpoint the transport is listening on (as reported by `Endpoints()`) are returned via `Accept()`.

The `Endpoint` struct and related types is:

```go
// Endpoint represents a transport endpoint, either local or remote.
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

// Protocol identifies a transport protocol.
type Protocol string
```

Endpoints are arbitrary transport-specific addresses, but if they are networked they must use IP addresses, and thus rely on IP as a fundamental packet routing protocol. This is to enable certain policies about address discovery, advertisement, and exchange - for example, a private `192.168.0.0/24` IP address should only be advertised to peers on that IP network, while the public address `8.8.8.8` may be advertised to all peers. Similarly, any port numbers if given must represent TCP and/or UDP port numbers, in order to use [UPnP](https://en.wikipedia.org/wiki/Universal_Plug_and_Play) to autoconfigure e.g. NAT gateways.

Non-networked endpoints (i.e. with no IP address) are considered local, and will only be advertised to other peers connecting via the same protocol. For example, an in-memory transport might use `Endpoint{Protocol: "memory", Path: "foo"}` as an address for the node "foo", and this should only be advertised to other nodes using `Protocol: "memory"`.

#### Connections and Streams

A connection represents an established transport connection between two endpoints (and thus two nodes), which can be used to exchange arbitrary raw bytes across one or more logically distinct IO streams. Connections are set up either via `Transport.Dial()` (outbound connections) or `Transport.Accept()` (inbound connections). The caller is responsible for verifying the remote peer's public key as returned by the connection.

To exchange data, an arbitrary stream ID is picked and the returned stream can be used as an IO reader or writer. Transports are free to choose how to implement streams, e.g. as distinct communication pathways or multiplexed onto one common pathway.

`Connection` and the related `Stream` interfaces are:

```go
// Connection represents an established connection between two endpoints.
type Connection interface {
    // Stream returns a logically distinct IO stream within the connection,
    // using an arbitrary stream ID. Multiple calls with the same ID return the
    // same stream.
	Stream(StreamID) Stream

	// LocalEndpoint returns the local endpoint for the connection.
	LocalEndpoint() Endpoint

	// RemoteEndpoint returns the remote endpoint for the connection.
	RemoteEndpoint() Endpoint

	// PubKey returns the public key of the remote peer.
	PubKey() crypto.PubKey

	// Close closes the connection.
	Close() error
}

// StreamID is an arbitrary stream ID.
type StreamID uint8

// Stream represents a single logical IO stream within a connection.
type Stream interface {
	io.Reader // Read([]byte) (int, error)
	io.Writer // Write([]byte) (int, error)
	io.Closer // Close() error
}
```

A note on backwards-compatibility: the current MConn protocol takes whole messages expressed as byte slices and splits them up into `PacketMsg` messages, where the final packet of a message has `PacketMsg.EOF` set. In order to maintain wire-compatibility with this protocol, the MConn transport needs to be aware of message boundaries, even though it does not care what the messages actually are. One way to handle this is to break abstraction boundaries and have the transport decode the input's length-prefixed message framing and use this to determine message boundaries. Then at some point in the future when we can break protocol compatibility we either remove the MConn protocol completely or drop the `PacketMsg.EOF` field and rely on the length-prefix framing.

### Peers

Peers are remote nodes that we wish to communicate with. Each peer is identified by a unique binary `PeerID`, and has a set of `PeerAddress` addresses expressed as URLs that they can be reached via. Addresses are resolved into one or more transport endpoints, e.g. by resolving DNS hostnames into IP addresses (which should be refreshed periodically). Peers should always be expressed as address URLs, never as endpoints (which are a lower-level construct).

```go
// PeerID is a unique peer ID, generally used in hex form.
type PeerID []byte

// PeerAddress is a peer address URL. The User field, if set, gives the
// hex-encoded remove PeerID, which should be verified with the remote peer's
// public key as returned by the connection.
type PeerAddress url.URL

// Resolve resolves a PeerAddress into a set of Endpoints, typically by
// expanding out any DNS names given in URL.Host to IP addresses. Fields are
// mapped as follows:
//
//   Scheme → Endpoint.Protocol
//   Host   → Endpoint.IP
//   Port   → Endpoint.Port
//   Path+Query+Fragment,Opaque → Endpoint.Path
//
func (a PeerAddress) Resolve(ctx context.Context) []Endpoint { return nil }
```

The P2P stack needs to track a lot of internal information about peers, such as endpoints, status, priorities, and so on, which is done in an internal `peer` struct. This should not be exposed outside of the `p2p` package, to avoid race conditions, lock contention, and tight coupling. Other packages should only refer to peers by `PeerID`.

The `peer` struct might look like the below, but is intentionally underspecified and will depend on implementation requirements:

```go
// peer tracks internal status information about a peer.
type peer struct {
	ID        PeerID
	Status    PeerStatus
	Priority  PeerPriority
	Endpoints map[PeerAddress][]Endpoint // Resolved endpoints by address.
}

// PeerStatus specifies the status of a peer.
type PeerStatus string

const (
	PeerStatusNew     = "new"     // New peer which we haven't tried to contact yet.
	PeerStatusUp      = "up"      // Peer which we have an active connection to.
	PeerStatusDown    = "down"    // Peer which we're temporarily disconnected from.
	PeerStatusRemoved = "removed" // Peer which has been removed.
	PeerStatusBanned  = "banned"  // Peer which is banned for misbehavior.
)

// PeerPriority specifies peer priorities.
type PeerPriority int

const (
	PeerPriorityNormal PeerPriority = iota + 1
	PeerPriorityValidator
	PeerPriorityPersistent
)
```

Peer information is stored in a `peerStore`, which may be persisted in an underlying database. It is kept internal to avoid race conditions and tight coupling. The peer store will replace the current address book, and thus track e.g. connection statistics and such.

The `peerStore` should at the very least contain basic CRUD functionality as outlined below, but will possibly need to have additional functionality, and is intentionally underspecified:

```go
// peerStore contains information about peers, possibly persisted to disk.
type peerStore struct {
	peers map[string]*peer // Entire set in memory, with PeerID.String() keys.
	db    dbm.DB           // Database for persistence, if non-nil.
}

func (p *peerStore) Delete(id PeerID) error     { return nil }
func (p *peerStore) Get(id PeerID) (peer, bool) { return peer{}, false }
func (p *peerStore) List() []peer               { return nil }
func (p *peerStore) Set(peer peer) error        { return nil }
```

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
