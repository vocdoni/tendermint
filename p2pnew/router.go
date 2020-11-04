package p2p

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
)

// ChannelID is an arbitrary channel ID, and maps direcly onto a stream ID.
type ChannelID StreamID

// Envelope is a wrapper for a message with a from/to address.
type Envelope struct {
	From      PeerID        // Message sender, or empty for outbound messages.
	To        PeerID        // Message receiver, or empty for inbound messages.
	Broadcast bool          // Send message to all connected peers, ignoring To.
	Message   proto.Message // Payload.
}

// PeerError is a peer error reported by a reactor via a Router.PeerErrors()
// channel. The error will be logged, and depending on the action the peer may
// be disconnected or banned.
type PeerError struct {
	ID     PeerID     // Peer which errored.
	Err    error      // The error which occurred.
	Action PeerAction // Action to take for peer.
}

func (e PeerError) Error() string { return fmt.Sprintf("Peer %q error: %v", e.ID, e.Err) }

// PeerAction is an action to take for a peer error.
type PeerAction string

const (
	PeerActionNone       PeerAction = "none"
	PeerActionDisconnect PeerAction = "disconnect"
	PeerActionBan        PeerAction = "ban"
)

// PeerUpdate notifies subscribers about peer status updates, via
// Router.PeerUpdates() channel.
type PeerUpdate struct {
	ID     PeerID
	Status PeerStatus
}

// Router maintains connections to peers and route Protobuf messages between
// them and local reactors.
//
// It will handle e.g. connection retries and backoff. Some number of outbound
// messages per peer will be buffered, but once full any new outbound messages
// for that peer are discarded, and the queue may be discarded entirely if the
// peer is unreachable. Similarly, inbound messages will be buffered per
// channel, but once full any new inbound messages on that channel are
// discarded.
//
// The router also sends peer status change updates to subscribers, and receives
// peer errors from e.g. reactors and takes requested action (e.g. disconnect
// or ban).
type Router struct{}

// NewRouter creates a new router. Transports must be pre-initialized to listen
// on any necessary interfaces, and keyed by endpoint protocol name.
func NewRouter(transports map[string]Transport) *Router { return nil }

// Open opens a channel. A channel ID can only be used once, until closed.
func (r *Router) Open(id ChannelID) (Channel, error) { return Channel{}, nil }

// PeerErrors returns a channel that can be used to submit peer errors. The
// error specifies an action to take for the peer as well, e.g. disconnect
// or ban the peer. The sender should not close the channel.
func (r *Router) PeerErrors() chan<- PeerError { return nil }

// PeerUpdates returns a channel with peer updates. The caller must cancel
// the context to end the subscription, and keep consuming messages in a timely
// fashion until the channel is closed to avoid blocking updates.
func (r *Router) PeerUpdates(ctx context.Context) <-chan PeerUpdate { return nil }

// Channel represents a logically separate bidirectional channel for Protobuf
// messages.
type Channel struct {
	// ID contains the channel ID.
	ID ChannelID
}

// Close closes the channel, making it unusable. The ID can be reused. It is
// the caller's responsibility to close the channel.
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
