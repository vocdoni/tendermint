package p2p

import (
	"context"
	"fmt"

	"github.com/tendermint/tendermint/proto/tendermint/statesync"
)

// The P2P stack does not have an explicit concept of a Reactor. Instead,
// reactors are just "something that listens on a channel", and can be
// implemented in any number of ways. Below is an example.

// ExampleMessage is an example Protobuf message. We just use statesync.Message.
type ExampleMessage = statesync.Message

// ExampleReactor is a minimal example reactor, implemented as a simple function.
// The reactor will exit when the context is cancelled.
func ExampleReactor(
	ctx context.Context, channel *Channel, peerUpdates PeerUpdates, peerErrors PeerErrors,
) {
	select {
	case env := <-channel.In:
		switch msg := env.Message.(type) {
		case *statesync.SnapshotsRequest:
			channel.Out <- Envelope{To: env.From, Message: &statesync.SnapshotsResponse{}}

		default:
			peerErrors <- PeerError{
				ID:     env.From,
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
