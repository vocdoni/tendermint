// nolint: structcheck,unused
package p2p

import (
	"context"
	"net/url"

	dbm "github.com/tendermint/tm-db"
)

// URL represents an externally provided peer address, which is resolved into a
// set of specific transport endpoints. For example, the URL may contain a DNS
// hostname which is resolved into a set of IP addresses. The URL scheme
// specifies which transport to use, e.g. mconn://validator.foo.com:26657.
type URL url.URL

// Resolve resolves the URL into a set of endpoints (e.g. resolving a DNS name
// into a list of IP addresses).
func (u URL) Resolve(ctx context.Context) ([]Endpoint, error) {
	return nil, nil
}

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

// Peer contains information about a peer.
type Peer struct {
	ID        PeerID
	Status    PeerStatus
	URLs      []URL      // Peer URLs (e.g. from config), resolved to endpoints when appropriate.
	Endpoints []Endpoint // Network endpoints, e.g. IP/port pairs.
}

// Peers tracks information about known peers for the Router.
//
// FIXME Needs a way to get pending peers and endpoints, and to report endpoint
// failures.
type Peers struct {
	peers map[PeerID]*Peer // Peers, entire set cached in memory.
	db    dbm.DB           // Database for persistence, if non-nil.
}

// NewPeers creates a new peer set, using db for persistence if non-nil.
func NewPeers(db dbm.DB) (*Peers, error) { return nil, nil }

// Get fetches a peer from the set, and whether it existed or not.
func (p *Peers) Get(id PeerID) (Peer, bool) { return Peer{}, false }

// Merge merges a peer with an existing entry (by ID) if any. This is useful
// e.g. to add peer URLs from a config file or endpoints from DNS lookups while
// also keeping reported peer endpoints obtained via peer exchange.
func (p *Peers) Merge(peer Peer) error { return nil }

// Set sets a peer, replacing the existing entry (by ID) if any.
func (p *Peers) Set(peer Peer) error { return nil }
