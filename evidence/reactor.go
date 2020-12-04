package evidence

import (
	"fmt"
	"time"

	clist "github.com/tendermint/tendermint/libs/clist"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/service"
	"github.com/tendermint/tendermint/p2p"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
)

var (
	_ service.Service = (*Reactor)(nil)

	// ChannelShims contains a map of ChannelDescriptorShim objects, where each
	// object wraps a reference to a legacy p2p ChannelDescriptor and the corresponding
	// p2p proto.Message the new p2p Channel is responsible for handling.
	//
	//
	// TODO: Remove once p2p refactor is complete.
	// ref: https://github.com/tendermint/tendermint/issues/5670
	ChannelShims = map[p2p.ChannelID]*p2p.ChannelDescriptorShim{
		EvidenceChannel: {
			MsgType: new(tmproto.EvidenceList),
			Descriptor: &p2p.ChannelDescriptor{
				ID:                  byte(EvidenceChannel),
				Priority:            5,
				RecvMessageCapacity: maxMsgSize,
			},
		},
	}
)

const (
	EvidenceChannel = p2p.ChannelID(0x38)

	maxMsgSize = 1048576 // 1MB TODO make it configurable

	// broadcast all uncommitted evidence this often. This sets when the reactor
	// goes back to the start of the list and begins sending the evidence again.
	// Most evidence should be committed in the very next block that is why we wait
	// just over the block production rate before sending evidence again.
	broadcastEvidenceIntervalS = 10
)

// Reactor handles evpool evidence broadcasting amongst peers.
type Reactor struct {
	service.BaseService

	evpool          *Pool
	eventBus        *types.EventBus
	evidenceCh      *p2p.Channel
	evidenceChDone  chan bool
	peerUpdates     p2p.PeerUpdates
	peerUpdatesDone chan bool
}

// NewReactor returns a new Reactor with the given config and evpool.
func NewReactor(logger log.Logger, evidenceCh *p2p.Channel, peerUpdates p2p.PeerUpdates, evpool *Pool) *Reactor {
	r := &Reactor{
		evpool:          evpool,
		evidenceCh:      evidenceCh,
		evidenceChDone:  make(chan bool),
		peerUpdates:     peerUpdates,
		peerUpdatesDone: make(chan bool),
	}

	r.BaseService = *service.NewBaseService(logger, "Evidence", r)
	return r
}

// OnStart starts separate go routines for each p2p Channel and listens for
// envelopes on each. In addition, it also listens for peer updates and handles
// messages on that p2p channel accordingly. The caller must be sure to execute
// OnStop to ensure the outbound p2p Channels are closed. No error is returned.
func (r *Reactor) OnStart() error {
	go r.processEvidenceCh()
	go r.processPeerUpdates()

	return nil
}

// OnStop stops the reactor by signaling to all spawned goroutines to exit and
// blocking until they all exit. Finally, it will close the evidence p2p Channel.
func (r *Reactor) OnStop() {
	// Wait for all goroutines to safely exit before proceeding to close all p2p
	// Channels. After closing, the router can be signaled that it is safe to stop
	// sending on the inbound In channel and close it.
	r.evidenceChDone <- true
	r.peerUpdatesDone <- true

	if err := r.evidenceCh.Close(); err != nil {
		panic(fmt.Sprintf("failed to close evidence channel: %s", err))
	}
}

func (r *Reactor) processPeerUpdates() {
	for {
		select {
		case peerUpdate := <-r.peerUpdates:
			r.Logger.Debug("peer update", "peer", peerUpdate.PeerID.String())

			switch peerUpdate.Status {
			case p2p.PeerStatusNew, p2p.PeerStatusUp:
				go r.broadcastEvidenceRoutine(peerUpdate.PeerID)
			}

		case <-r.peerUpdatesDone:
			r.Logger.Debug("stopped listening on peer updates channel")
			return
		}
	}
}

// SetEventBus implements events.Eventable.
func (r *Reactor) SetEventBus(b *types.EventBus) {
	r.eventBus = b
}

// processEvidenceCh implements a blocking event loop where we listen for p2p
// Envelope messages from the evidenceCh.
func (r *Reactor) processEvidenceCh() {
	for {
		select {
		case envelope := <-r.evidenceCh.In:
			switch msg := envelope.Message.(type) {
			case *tmproto.EvidenceList:
				r.Logger.Debug(
					"received evidence list",
					"num_evidence", len(msg.Evidence),
					"peer", envelope.From.String(),
				)

				for _, protoEv := range msg.Evidence {
					ev, err := types.EvidenceFromProto(&protoEv)
					if err != nil {
						r.Logger.Error("failed to convert evidence", "peer", envelope.From.String(), "err", err)
						continue
					}

					if err := r.evpool.AddEvidence(ev); err != nil {
						r.Logger.Error("failed to add evidence", "peer", envelope.From.String(), "err", err)

						// If we're given invalid evidence by the peer, notify the router
						// that we should remove this peer.
						if _, ok := err.(*types.ErrInvalidEvidence); ok {
							r.evidenceCh.Error <- p2p.PeerError{
								PeerID:   envelope.From,
								Err:      err,
								Severity: p2p.PeerErrorSeverityLow,
							}
						}
					}
				}

			default:
				r.Logger.Error("received unknown message", "msg", msg, "peer", envelope.From.String())
				r.evidenceCh.Error <- p2p.PeerError{
					PeerID:   envelope.From,
					Err:      fmt.Errorf("received unknown message: %T", msg),
					Severity: p2p.PeerErrorSeverityLow,
				}
			}

		case <-r.evidenceChDone:
			r.Logger.Debug("stopped listening on evidence channel")
			return
		}
	}
}

// Modeled after the mempool routine.
// - Evidence accumulates in a clist.
// - Each peer has a routine that iterates through the clist, sending available
//   evidence to the peer.
// - If we're waiting for new evidence and the list is not empty, we start
//   iterating from the beginning again.
//
// TODO: We need to handle peer stopping, otherwise, per peer, we can create
// multiple goroutines.
func (r *Reactor) broadcastEvidenceRoutine(peerID p2p.PeerID) {
	var next *clist.CElement
	for {
		// This happens because the CElement we were looking at got garbage
		// collected (removed). That is, .NextWaitChan() returned nil. So we can go
		// ahead and start from the beginning.
		if next == nil {
			select {
			case <-r.evpool.EvidenceWaitChan(): // wait until evidence is available
				if next = r.evpool.EvidenceFront(); next == nil {
					continue
				}

			case <-r.Quit():
				// we can safely exit the goroutine once the reactor stops
				return
			}
		}

		if ev, ok := next.Value.(types.Evidence); ok {
			evProto, err := types.EvidenceToProto(ev)
			if err != nil {
				panic(err)
			}

			r.Logger.Debug("gossiping evidence to peer", "ev", ev, "peer", peerID)
			r.evidenceCh.Out <- p2p.Envelope{
				To: peerID,
				Message: &tmproto.EvidenceList{
					Evidence: []tmproto.Evidence{*evProto},
				},
			}
		}

		afterCh := time.After(time.Second * broadcastEvidenceIntervalS)
		select {
		case <-afterCh:
			// start from the beginning every tick
			next = nil

		case <-next.NextWaitChan():
			next = next.Next()

		case <-r.Quit():
			// we can safely exit the goroutine once the reactor stops
			return
		}
	}
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

// // prepareEvidenceMessage returns a slice of Evidence objects to send to the peer,
// // or nil if the evidence is invalid for the peer. If message is nil, we should
// // sleep and try again.
// func (r Reactor) prepareEvidenceMessage(peerID p2p.PeerID, ev types.Evidence) []types.Evidence {
// 	// make sure the peer is up to date
// 	evHeight := ev.Height()
// 	peerState, ok := peer.Get(types.PeerStateKey).(PeerState)
// 	if !ok {
// 		// Peer does not have a state yet. We set it in the consensus reactor, but
// 		// when we add peer in Switch, the order we call reactors#AddPeer is
// 		// different every time due to us using a map. Sometimes other reactors
// 		// will be initialized before the consensus reactor. We should wait a few
// 		// milliseconds and retry.
// 		return nil
// 	}

// 	// NOTE: We only send evidence to peers where
// 	// peerHeight - maxAge < evidenceHeight < peerHeight
// 	var (
// 		peerHeight   = peerState.GetHeight()
// 		params       = r.evpool.State().ConsensusParams.Evidence
// 		ageNumBlocks = peerHeight - evHeight
// 	)

// 	if peerHeight <= evHeight { // peer is behind. sleep while he catches up
// 		return nil
// 	} else if ageNumBlocks > params.MaxAgeNumBlocks { // evidence is too old relative to the peer, skip
// 		// NOTE: if evidence is too old for an honest peer, then we're behind and
// 		// either it already got committed or it never will!
// 		r.Logger.Info(
// 			"not sending peer old evidence",
// 			"peerHeight", peerHeight,
// 			"evHeight", evHeight,
// 			"maxAgeNumBlocks", params.MaxAgeNumBlocks,
// 			"lastBlockTime", r.evpool.State().LastBlockTime,
// 			"maxAgeDuration", params.MaxAgeDuration,
// 			"peer", peer,
// 		)

// 		return nil
// 	}

// 	return []types.Evidence{ev}
// }

// Receive implements Reactor.
// It adds any received evidence to the evpool.
// XXX: do not call any methods that can block or incur heavy processing.
// https://github.com/tendermint/tendermint/issues/2888
// func (evR *Reactor) Receive(chID byte, src p2p.Peer, msgBytes []byte) {
// 	evis, err := decodeMsg(msgBytes)
// 	if err != nil {
// 		evR.Logger.Error("Error decoding message", "src", src, "chId", chID, "err", err, "bytes", msgBytes)
// 		evR.Switch.StopPeerForError(src, err)
// 		return
// 	}

// 	for _, ev := range evis {
// 		err := evR.evpool.AddEvidence(ev)
// 		switch err.(type) {
// 		case *types.ErrInvalidEvidence:
// 			evR.Logger.Error(err.Error())
// 			// punish peer
// 			evR.Switch.StopPeerForError(src, err)
// 			return
// 		case nil:
// 		default:
// 			// continue to the next piece of evidence
// 			evR.Logger.Error("Evidence has not been added", "evidence", evis, "err", err)
// 		}
// 	}
// }

// // AddPeer implements Reactor.
// func (evR *Reactor) AddPeer(peer p2p.Peer) {
// 	go evR.broadcastEvidenceRoutine(peer)
// }

// // PeerState describes the state of a peer.
// type PeerState interface {
// 	GetHeight() int64
// }

// // encodemsg takes a array of evidence
// // returns the byte encoding of the List Message
// func encodeMsg(evis []types.Evidence) ([]byte, error) {
// 	evi := make([]tmproto.Evidence, len(evis))
// 	for i := 0; i < len(evis); i++ {
// 		ev, err := types.EvidenceToProto(evis[i])
// 		if err != nil {
// 			return nil, err
// 		}
// 		evi[i] = *ev
// 	}
// 	epl := tmproto.EvidenceList{
// 		Evidence: evi,
// 	}

// 	return epl.Marshal()
// }

// // decodemsg takes an array of bytes
// // returns an array of evidence
// func decodeMsg(bz []byte) (evis []types.Evidence, err error) {
// 	lm := tmproto.EvidenceList{}
// 	if err := lm.Unmarshal(bz); err != nil {
// 		return nil, err
// 	}

// 	evis = make([]types.Evidence, len(lm.Evidence))
// 	for i := 0; i < len(lm.Evidence); i++ {
// 		ev, err := types.EvidenceFromProto(&lm.Evidence[i])
// 		if err != nil {
// 			return nil, err
// 		}
// 		evis[i] = ev
// 	}

// 	for i, ev := range evis {
// 		if err := ev.ValidateBasic(); err != nil {
// 			return nil, fmt.Errorf("invalid evidence (#%d): %v", i, err)
// 		}
// 	}

// 	return evis, nil
// }
