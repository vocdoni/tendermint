package p2p

import (
	"context"

	"github.com/gogo/protobuf/proto"
)

// Router maintains connections to peers and route Protobuf messages between
// them and local reactors.
type Router struct {
}

// Broadcast broadcasts a message to all peers on a channel.
func (r *Router) Broadcast(ctx context.Context, ch ChannelID, msg proto.Message) error {
	return nil
}

// Receive receives a message from a peer on a channel.
func (r *Router) Receive(ctx context.Context, ch ChannelID) (*Peer, proto.Message, error) {
	return nil, nil, nil
}

// Send sends a message to a peer on a channel.
func (r *Router) Send(ctx context.Context, to *Peer, ch ChannelID, msg proto.Message) error {
	return nil
}

// Peer contains information about a peer.
type Peer struct {
	ID        []byte
	Addresses []Address
}
