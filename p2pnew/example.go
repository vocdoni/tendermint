package p2p

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/proto/tendermint/statesync"
)

// The P2P stack does not have an explicit concept of a Reactor. Instead,
// reactors are just "something that listens on a channel", and can be
// implemented in any number of ways. Below is an example.

// ExampleMessage is an example Protobuf message. We just use statesync.Message.
type ExampleMessage = statesync.Message

// ExampleWrapper is used to wrap outbound Protobuf messages in a container
// message, since channels can only pass a single Protobuf message type.
func ExampleWrapper(msg proto.Message) proto.Message {
	switch msg := msg.(type) {
	case *statesync.Message:
		return msg
	case *statesync.SnapshotsRequest:
		return &statesync.Message{Sum: &statesync.Message_SnapshotsRequest{SnapshotsRequest: msg}}
	case *statesync.SnapshotsResponse:
		return &statesync.Message{Sum: &statesync.Message_SnapshotsResponse{SnapshotsResponse: msg}}
	default:
		return nil
	}
}

// ExampleUnwrapper unwraps Protobuf messages from a container message, the
// inverse of ExampleWrapper.
func ExampleUnwrapper(msg proto.Message) proto.Message {
	switch msg := msg.(*statesync.Message).Sum.(type) {
	case *statesync.Message_SnapshotsRequest:
		return msg.SnapshotsRequest
	case *statesync.Message_SnapshotsResponse:
		return msg.SnapshotsResponse
	default:
		return nil
	}
}

// ExampleReactor is a minimal example reactor, implemented as a simple function.
// The reactor will exit when the context is cancelled.
func ExampleReactor(
	ctx context.Context, channel *Channel, peerUpdates PeerUpdates, peerErrors PeerErrors,
) {
	select {
	case envelope := <-channel.In:
		switch msg := envelope.Message.(type) {
		case *statesync.SnapshotsRequest:
			channel.Out <- Envelope{To: envelope.From, Message: &statesync.SnapshotsResponse{}}

		default:
			peerErrors <- PeerError{
				ID:     envelope.From,
				Err:    fmt.Errorf("unexpected message %T", msg),
				Action: PeerActionDisconnect,
			}
		}

	case peerUpdate := <-peerUpdates:
		fmt.Printf("Peer %q changed status to %q", peerUpdate.ID, peerUpdate.Status)

	case <-ctx.Done():
		break
	}
}

// RunExampleReactor builds and runs an example reactor connected to the given
// router. Generally all reactor setup should happen in one place (e.g. in the
// Node), such that all channels are defined in one place. But we could pass the
// router to the reactor as well.
func RunExampleReactor(router *Router) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	channel, err := router.Open(1, &ExampleMessage{})
	if err != nil {
		return err
	}
	channel.Wrapper = ExampleWrapper
	channel.Unwrapper = ExampleUnwrapper

	ExampleReactor(ctx, channel, router.PeerUpdates(ctx), router.PeerErrors())
	return nil
}
