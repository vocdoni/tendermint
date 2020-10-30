package p2p

// Transports securely connect to network addresses using public key
// cryptography, and send/receive raw bytes across logically distinct channels.
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

// ChannelID represents a single channel ID. It is up to the transport
// to separate channels as appropriate (e.g. using separate streams).
type ChannelID uint8

// Transport represents an underlying network transport. It creates connections
// to/from an address, and sends raw bytes across separate channels within this
// connection.
type Transport interface {
	// Accept waits for the next inbound connection.
	Accept() (Connection, error)

	// Dial creates an outbound connection to an address.
	Dial(context.Context, Address) (Connection, error)
}

// Connection represents a single secure connection to an address. It contains
// separate logical channels that can read or write raw bytes.
type Connection interface {
	io.Closer // Close() error

	// Channel returns a reference to a channel within the connection (e.g. a
	// stream), identified by an arbitrary channel ID. Multiple calls with the
	// same channel ID should return the same channel. Any errors during channel
	// setup should be returned via the Channel interface as appropriate.
	Channel(ChannelID) Channel

	// PubKey returns the public key of the remote peer.
	PubKey() crypto.PubKey
}

// Channel represents a single IO channel within a connection.
type Channel interface {
	io.Reader // Read([]byte) (int, error)
	io.Writer // Write([]byte) (int, error)
	io.Closer // Close() error
}
