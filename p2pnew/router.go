package p2p

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
)

// ChannelID is an arbitrary channel ID, and maps direcly onto a stream ID.
type ChannelID StreamID

// Envelope is a wrapper for a message with a from/to address. Inbound messages
// must have From set, and outbound messages must have To or Broadcast set.
type Envelope struct {
	From      PeerID        // Message sender, or empty for outbound messages.
	To        PeerID        // Message receiver, or empty for inbound messages.
	Broadcast bool          // Send message to all connected peers, ignoring To.
	Message   proto.Message // Payload.
}

// Wrapper is a container message that can contain a variety of inner messages.
// If a Channel message type implements Wrapper, the channel will try to wrap
// any other message types in the container message to support multiple types.
type Wrapper interface {
	proto.Message
	Wrap(proto.Message) error
	Unwrap() (proto.Message, error)
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

// PeerErrors is a channel for submitting peer errors.
type PeerErrors chan<- PeerError

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

// PeerUpdates is a channel for receiving peer updates.
type PeerUpdates <-chan PeerUpdate

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
// on any necessary interfaces.
func NewRouter(transports []Transport) *Router { return nil }

// Open opens a channel. A channel ID can only be used once, until closed. The
// messageType should be an empty Protobuf message of the type that will be
// passed through the channel. A channel only supports a single message type,
// since it needs to know what message type to unmarshal into.
//
// However, if messageType also implements Wrapper, then any other message types
// passed via the channel will be automatically wrapped and unwrapped by the
// outer message type (if possible). This allows the channel message type to
// be e.g. a oneof Protobuf message, and any message types that are supported
// by the oneof can be passed directly through the channel and are automatically
// wrapped/unwrapped in the container message.
//
// The channel automatically encodes and/or decodes Protobuf messages using
// length-prefixed (aka length-delimited) framing. Invalid encodings are dropped.
func (r *Router) Open(id ChannelID, messageType proto.Message) (*Channel, error) { return nil, nil }

// PeerErrors returns a channel that can be used to submit peer errors. The
// error specifies an action to take for the peer as well, e.g. disconnect
// or ban the peer. The sender should not close the channel.
func (r *Router) PeerErrors() chan<- PeerError { return nil }

// PeerUpdates returns a channel with peer updates. The caller must cancel
// the context to end the subscription, and keep consuming messages in a timely
// fashion until the channel is closed to avoid blocking updates.
//
// FIXME This should possibly be implemented via an PeerStore.OnUpdate() hook
// or something similar, to trigger notifications from the central data
// location rather than spread around the Router. This is left as an
// implementation detail.
func (r *Router) PeerUpdates(ctx context.Context) <-chan PeerUpdate { return nil }

// Channel represents a logically separate bidirectional channel to exchange
// Protobuf messages with any known peers. The router will use transport streams
// to send and receive messages with individual peer, where each channel uses
// its own distinct stream.
type Channel struct {
	// ID contains the channel ID.
	ID ChannelID

	// In is a channel for receiving inbound messages. Envelope will always have
	// From and Message set.
	//
	// The scheduling of incoming messages is an implementation detail that is
	// managed by the router. This could be done using any number of algorithms,
	// e.g. FIFO, round-robin, priority queues, or some other scheme.
	In <-chan Envelope

	// Out is a channel for sending outbound messages. Envelope must have To (or
	// Broadcast) and Message set, otherwise it is discarded.
	//
	// Messages are not guaranteed to be delivered, and may be dropped e.g. if
	// the peer goes offline, if the peer is overloaded, or for any other
	// reason.
	Out chan<- Envelope
}

// Close closes the channel, making it unusable. The ID can be reused. It is
// the caller's responsibility to close the channel. It is equivalent to
// close(Channel.Out). After closing, the Router will close Channel.In
func (c *Channel) Close() error { return nil }
