package p2p

// Transports securely connect to network addresses using public key
// cryptography, and send/receive raw bytes across logically distinct streams.
// They don't know about peers or messages.

import (
	"context"
	"io"

	"github.com/tendermint/tendermint/crypto"
)

// Address is a transport-agnostic network address. An address should map
// onto exactly one Transport. It should not contain the peer ID, and a
// peer may have multiple addresses.
//
// Possibly use multiaddr? https://github.com/multiformats/multiaddr
type Address string

// StreamID represents a single stream ID. It is up to the transport
// to separate streams as appropriate.
type StreamID uint8

// Transport represents an underlying network transport. It creates connections
// to/from an address, and sends raw bytes across separate streams within this
// connection.
type Transport interface {
	// Accept waits for the next inbound connection.
	Accept() (Connection, error)

	// Dial creates an outbound connection to an address.
	Dial(context.Context, Address) (Connection, error)
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
type Stream interface {
	io.ReadWriteCloser
	// Read([]byte) (int, error)
	// Write([]byte) (int, error)
	// Close() error
}
