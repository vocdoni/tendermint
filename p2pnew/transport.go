package p2p

import (
	"context"
	"io"
	"net/url"

	"github.com/tendermint/tendermint/crypto"
)

// StreamID represents a single stream ID. It is up to the transport
// to separate streams as appropriate.
type StreamID uint8

// Endpoint represents a node endpoint used by Transport to dial a peer. A node
// can have multiple endpoints, and are usually resolved from a PeerAddress.
//
// Endpoints are represented as URLs, where some fields have special meaning:
//
// - Host: if set, must be an IP address (v4 or v6), and defines this as a networked endpoint.
// - Port: if set, Host must be set as well, and can be interpreted both as a TCP or UDP port
//   (this may be used e.g. when configuring NAT routers via UPnP). If not given, UPnP and
//   such cannot be used, but the
//
// Host has implications for how the endpoint is advertised to peers, e.g.
// if set the IP address should only be advertised to peers that have access
// to that network (so 192.168.0.0 should only be advertised to peers on that
// network, while public IPs can be advertised to anyone). If Host is not set,
// this is considered a non-networked transport, and should only be advertised
// to peers using other non-networked transports.
type Endpoint struct {
	url.URL
	// may contain additional fields to track e.g. failure statistics,
	// unless we store this in the Router.
}

// PeerAddress converts the endpoint into a peer address, used e.g. to advertise
// the local node's transport endpoints to other peers.
func (e Endpoint) PeerAddress() PeerAddress { return PeerAddress{} }

// Transport represents a network transport that can provide both inbound
// and outbound connections to peer endpoints. Connections must be encrypted
// (i.e. the peer must be able to provide a public key), and must support
// separate logical streams of raw bytes.
type Transport interface {
	// Accept waits for the next inbound connection until the context is cancelled.
	Accept(context.Context) (Connection, error)

	// Dial creates an outbound connection to an endpoint.
	Dial(context.Context, Endpoint) (Connection, error)

	// Endpoints returns a list of endpoints the transport is listening on.
	// Any IP addresses do not need to be normalized or otherwise preprocessed
	// by the transport (this will be done elsewhere before advertising them to
	// peers, e.g. by expanding out 0.0.0.0 to local interface addresses).
	Endpoints() []Endpoint

	// Protocols returns a list of protocols (aka schemes) that this Transport
	// can handle. Only one Transport can use a given protocol. It is used by
	// Router when looking up a Transport for an Endpoint.
	Protocols() []string
}

// Connection represents a single secure connection to an address. It contains
// separate logical streams that can read or write raw bytes.
//
// Callers are responsible for authenticating the remote peer's pubkey
// against known information, i.e. the peer ID. Otherwise they are vulnerable
// to MitM attacks.
type Connection interface {
	// Stream returns a reference to a stream within the connection, identified
	// by an arbitrary stream ID. Multiple calls return the same stream. Any
	// errors should be returned via the Stream interface.
	Stream(StreamID) Stream

	// LocalEndpoint returns the local endpoint for the connection.
	LocalEndpoint() Endpoint

	// RemoteEndpoint returns the remote endpoint for the connection.
	RemoteEndpoint() Endpoint

	// PubKey returns the public key of the remote peer. It should not change
	// after the connection has been established.
	PubKey() crypto.PubKey

	// Close closes the connection.
	Close() error
}

// Stream represents a single logical IO stream within a connection.
//
// FIXME For compatibility with the old MConn protocol, a single Write or Read
// call must correspond to a single logical message such that we can set
// PacketMsg.EOF at the end of the message. Once we can change the protocol or
// remove MConn, we should drop this requirement such that the given byte slices
// are arbitrary data.
type Stream interface {
	io.Reader // Read([]byte) (int, error)
	io.Writer // Write([]byte) (int, error)
	io.Closer // Close() error
}
