// nolint: structcheck,unused
package p2p

import (
	"context"
	"net/url"

	dbm "github.com/tendermint/tm-db"
)

// PeerAddress represents a peer address, given as a URL. Peer addresses are used
// when expressing and exchanging peers (e.g. in config files and via PEX), but
// they are resolved into one or more Endpoints that are used when connecting.
//
// Some URL fields have special meaning:
//
// - Scheme: used by the Router to look up a Transport for this address.
// - Host: if non-empty, interpreted either as an IP address (v4 or v6) or DNS name.
// - User: if non-empty, interpreted as a PeerID checked against the peer's public key.
type PeerAddress url.URL

// Resolve resolves a PeerAddress into a set of Endpoints, typically by expanding
// out any DNS names given in Host.
func (a PeerAddress) Resolve(ctx context.Context) []Endpoint { return nil }

// PeerID is a unique peer ID.
type PeerID string

// PeerStatus contains the status of a peer.
type PeerStatus string

const (
	PeerStatusNew     = "new"     // New peer which we haven't tried to contact yet.
	PeerStatusUp      = "up"      // Peer which we have an active connection to.
	PeerStatusDown    = "down"    // Peer which we're temporarily disconnected from.
	PeerStatusRemoved = "removed" // Peer which has been removed.
	PeerStatusBanned  = "banned"  // Peer which is banned for misbehavior.
)

// PeerPriority contains peer priorities.
type PeerPriority int

const (
	PeerPriorityNormal PeerPriority = iota + 1
	PeerPriorityValidator
	PeerPriorityPersistent
)

// Peer contains information about a peer. It should only be used internally in
// the Router, while reactors only get access to the ID and status. This avoids
// race conditions and lock contention, and decouples reactors from P2P
// infrastructure.
type Peer struct {
	ID        PeerID
	Status    PeerStatus
	Priority  PeerPriority
	Addresses []PeerAddress              // Peer addresses, from e.g. config or PEX.
	Endpoints map[PeerAddress][]Endpoint // Resolved endpoints by address.
}

// PeerStore tracks information about known peers for the Router.
//
// FIXME The router needs to figure out which peers to connect to, which
// endpoints to use, and so on. This needs to be based on peer information such
// as peer priorities, number of connection failures, and so on, which should
// probably be tracked in PeerStore somehow so that it is persisted. This is
// left as an implementation detail, and probably requires additional methods.
type PeerStore struct {
	peers map[PeerID]*Peer // Entire set cached in memory.
	db    dbm.DB           // Database for persistence, if non-nil.
}

// NewPeerStore creates a new peer store, using db for persistence if non-nil.
func NewPeerStore(db dbm.DB) (*PeerStore, error) { return nil, nil }

// Delete removes a peer from the set.
func (p *PeerStore) Delete(id PeerID) error { return nil }

// Get fetches a peer from the set, and whether it existed or not.
func (p *PeerStore) Get(id PeerID) (Peer, bool) { return Peer{}, false }

// List returns a list of all peers.
func (p *PeerStore) List() []Peer { return nil }

// Set sets a peer, replacing the existing entry (by ID) if any.
func (p *PeerStore) Set(peer Peer) error { return nil }
