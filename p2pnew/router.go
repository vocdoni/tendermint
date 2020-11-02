package p2p

import (
	"github.com/gogo/protobuf/proto"
)

// ChannelID represents an arbitrary channel ID. Channel IDs generally map
// directly onto stream IDs.
type ChannelID StreamID

// Envelope is a wrapper for a message with a from/to address.
type Envelope struct {
	// From contains the message sender, or nil for outbound messages.
	From *Peer
	// To represents the message receiver, or nil for inbound messages.
	To *Peer
	// Broadcast sends the message to all known peers, ignoring To.
	Broadcast bool
	// Message is the payload.
	Message proto.Message
}

// Router maintains connections to peers and route Protobuf messages between
// them and local reactors.
type Router struct {
}

// Open opens a channel. A channel can only be opened once, until closed.
func (r *Router) Open(id ChannelID) (Channel, error) {
	return Channel{ID: id}, nil
}

// Channel represents a logically separate channel for Protobuf messages.
type Channel struct {
	// ID contains the Channel ID. It must not be changed.
	ID ChannelID
}

// Close closes the channel. It can no longer be used, and the ID can be
// reused by calling Router.Open().
func (c *Channel) Close() error {
	return nil
}

// Receive returns a Go channel that receives messages from peers.
// Envelope will always have From and Message set.
func (c *Channel) Receive() <-chan Envelope {
	return nil
}

// Send returns a Go channel that sends messages to peers.
// Envelope must have To (or Broadcast) and Message set, otherwise it is
// discarded.
func (c *Channel) Send() chan<- Envelope {
	return nil
}

// Peer contains information about a peer.
type Peer struct {
	ID        []byte
	Addresses []Address
}
