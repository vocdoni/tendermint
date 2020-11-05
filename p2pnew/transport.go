package p2p

import (
	"context"
	"io"
	"net"

	"github.com/tendermint/tendermint/crypto"
)

// StreamID represents a single stream ID. It is up to the transport
// to separate streams as appropriate.
type StreamID uint8

// Endpoint represents a node endpoint used by Transport to dial a peer. A node
// can have multiple endpoints, and are usually resolved from a PeerAddress.
type Endpoint struct {
	// Protocol specifies the transport protocol, used by the router to pick a transport.
	Protocol string

	// Path is a transport-specific path or identifier to connect via. This
	// corresponds to the path, query-string, fragment, and opaque portions of a
	// PeerAddress URL. For example, an in-memory transport might use the URL
	// "memory:foo" to identify a "foo" peer to connect to, where "foo" would
	// become Path. Similarly, "http://host/path" would place "/path" in Path.
	Path string

	// IP is the IP address (v4 or v6) to connect to. If set, this defines the
	// endpoint as a networked endpoint. Networked endpoints are advertised to
	// peers depending on the IP visibility (e.g. private 192.168.0.0 IPs are
	// only advertised to peers on this network, but public IPs are advertised
	// to all peers). Non-networked endpoints are only advertised to other peers
	// connecting via the same protocol.
	IP net.IP

	// Port is the network port (either TCP or UDP). If not set, but IP is set,
	// a default port for the protocol will be used. Note that endpoints
	// returned by Transport.Endpoints() must set a Port number if they wish to
	// autoconfigure e.g. UPnP for NAT traversal.
	Port uint16
}

// PeerAddress converts the endpoint into a peer address, used e.g. to advertise
// the local node's transport endpoints to other peers.
func (e Endpoint) PeerAddress(id PeerID) PeerAddress { return PeerAddress{} }

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
// FIXME For compatibility with the old MConn protocol, a single Write call must
// correspond to a single logical message such that we can set PacketMsg.EOF at
// the end of the message. Once we can change the protocol or remove MConn, we
// should drop this requirement such that the given byte slices are arbitrary
// data.
type Stream interface {
	io.Reader // Read([]byte) (int, error)
	io.Writer // Write([]byte) (int, error)
	io.Closer // Close() error
}
