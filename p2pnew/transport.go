// nolint: structcheck,unused
package p2p

// Transports securely connect to network addresses using public key
// cryptography, and send/receive raw bytes across logically distinct streams.
// They don't know about peers or messages.

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
// can have multiple endpoints. Remote endpoints must always have an IP address,
// while local endpoints may not (e.g. for UNIX sockets or in-memory nodes).
type Endpoint struct {
	// protocol specifies the endpoint protocol, e.g. mconn or quic. The Router
	// uses this to map an endpoint onto a Transport.
	protocol string
	// address specifies the network address, e.g. an IP:port pair or UNIX file path.
	address string
	// ip contains the IP address of a remote endpoint, or nil if local. All
	// remote endpoints must have an IP address. It is primarily used for
	// endpoint filtering (i.e. don't advertise loopback endpoints to peers),
	// while transports should use address for dialing.
	ip net.IP
}

// Transport represents a network transport that can provide both inbound
// and outbound connections.
type Transport interface {
	// Accept waits for the next inbound connection.
	Accept(context.Context) (Connection, error)

	// Dial creates an outbound connection to an endpoint.
	Dial(context.Context, Endpoint) (Connection, error)
}

// Connection represents a single secure connection to an address. It contains
// separate logical streams that can read or write raw bytes.
type Connection interface {
	// Stream returns a reference to a stream within the connection, identified
	// by an arbitrary stream ID. Multiple calls return the same stream. Any
	// errors should be returned via the Stream interface.
	Stream(StreamID) Stream

	// PubKey returns the public key of the remote peer.
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
