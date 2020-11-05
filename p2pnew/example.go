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

// ExampleMessage is an example Protobuf message. We just "inherit" from
// statesync.Message, but this should usually be a Protobuf-generated message.
type ExampleMessage struct{ *statesync.Message }

// Wrap implements Wrapper, which allows this message to wrap a variety
// of other messages. This is useful since a channel can only pass messages
// of a single type.
func (m *ExampleMessage) Wrap(inner proto.Message) error {
	switch inner := inner.(type) {
	case *statesync.SnapshotsRequest:
		m.Message.Sum = &statesync.Message_SnapshotsRequest{SnapshotsRequest: inner}
	case *statesync.SnapshotsResponse:
		m.Message.Sum = &statesync.Message_SnapshotsResponse{SnapshotsResponse: inner}
	// These just handle the cases where the message is already wrapped.
	case *ExampleMessage:
		*m = *inner
	case *statesync.Message:
		m.Message = inner
	default:
		return fmt.Errorf("unknown message %T", inner)
	}
	return nil
}

// Unwrap implements Unwrapper, which unwraps the inner message contained in this.
// It is the inverse of Wrap.
func (m *ExampleMessage) Unwrap() (proto.Message, error) {
	switch inner := m.Message.Sum.(type) {
	case *statesync.Message_SnapshotsRequest:
		return inner.SnapshotsRequest, nil
	case *statesync.Message_SnapshotsResponse:
		return inner.SnapshotsResponse, nil
	default:
		return nil, fmt.Errorf("unknown message %T", inner)
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

	ExampleReactor(ctx, channel, router.PeerUpdates(ctx), router.PeerErrors())
	return nil
}
