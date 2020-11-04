package p2p

import (
	"github.com/gogo/protobuf/proto"
)

// ChannelID represents an arbitrary channel ID. Channel IDs generally map
// directly onto stream IDs.
type ChannelID StreamID

// Envelope is a wrapper for a message with a from/to address.
type Envelope struct {
	// From contains the message sender, or empty for outbound messages.
	From PeerID
	// To represents the message receiver, or empty for inbound messages.
	To PeerID
	// Broadcast sends an outbound message to all known peers, ignoring To.
	Broadcast bool
	// Message is the payload.
	Message proto.Message
}

// Router maintains connections to peers and route Protobuf messages between
// them and local reactors. It will handle e.g. connection retries and backoff.
// Some number of outbound messages per peer will be buffered, but once full
// any new outbound messages for that peer are discarded, and the queue may be
// discarded entirely if the peer is unreachable. Similarly, inbound messages
// will be buffered per channel, but once full any new inbound messages on
// that channel are discarded.
type Router struct{}

// NewRouter creates a new router. Transports must be pre-initialized to listen
// on any necessary interfaces, and keyed by endpoint protocol name.
func NewRouter(transports map[string]Transport) *Router { return nil }

// Open opens a channel. A channel ID can only be used once, until closed.
func (r *Router) Open(id ChannelID) (Channel, error) { return Channel{}, nil }

// Channel represents a logically separate bidirectional channel for Protobuf
// messages.
type Channel struct {
	// ID contains the channel ID.
	ID ChannelID
}

// Close closes the channel, making it unusable. The ID can be reused. It is
// the sender's responsibility to close the channel.
func (c *Channel) Close() error { return nil }

// Receive returns a Go channel that receives messages from peers. Envelope will
// always have From and Message set.
//
// The scheduling of incoming messages is an implementation detail that is
// managed by the router. This could be done using any number of algorithms,
// e.g. FIFO, round-robin, priority queues, or some other scheme.
func (c *Channel) Receive() <-chan Envelope { return nil }

// Send returns a Go channel that sends messages to peers. Envelope must have To
// (or Broadcast) and Message set, otherwise it is discarded.
//
// Messages are not guaranteed to be delivered, and may be dropped e.g. if the
// peer goes offline, if the peer is overloaded, or for any other reason.
func (c *Channel) Send() chan<- Envelope { return nil }
