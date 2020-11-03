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
	// PeerStatusNew represents a new peer which we haven't tried to contact yet.
	PeerStatusNew = "new"
	// PeerStatusUp represents a peer which we have an active connection to.
	PeerStatusUp = "up"
	// PeerStatusDown represents a peer which we're temporarily disconnected from.
	PeerStatusDown = "down"
	// PeerStatusRemove represents a peer which has been removed.
	PeerStatusRemoved = "removed"
	// PeerStatusBanned represents a peer which is banned for misbehavior.
	PeerStatusBanned = "banned"
)

// PeerUpdate is emitted whenever a peer's status changes.
type PeerUpdate struct {
	ID     PeerID
	Status PeerStatus
}

// PeerUpdates is a subscription for peer updates.
type PeerUpdates <-chan PeerUpdate

// PeerBanner is a channel that can be used to ban peers.
type PeerBanner chan<- PeerID

// Peer contains information about a peer.
type Peer struct {
	// ID contains the peer's node ID.
	ID PeerID
	// Status contains the peer's status.
	Status PeerStatus
	// URLs is a list of locally provided peer URLs (e.g. given in config file). These
	// are kept since we may need to periodically refresh the endpoints from the URLs,
	// e.g. to pick up DNS changes.
	URLs []URL
	// Endpoints contains network endpoints for the node, e.g. IP/port pairs.
	Endpoints []Endpoint
}

// Peers tracks information about known peers, used both by Router to keep track
// of peers, and to gossip peer information to other peers (e.g. via PEX).
//
// FIXME Needs a way to get pending peers and endpoints, and to report endpoint
// failures.
type Peers struct{}

// NewPeers creates a new peer set, using db for persistence if non-nil.
func NewPeers(db dbm.DB) (*Peers, error) { return nil, nil }

// Add adds a peer to the set, or merges it with an existing peer.
func (p *Peers) Add(peer *Peer) error { return nil }

// Get fetches a peer from the set, if it exists.
func (p *Peers) Get(id PeerID) *Peer { return nil }

// GetPending returns a peer from the list that is pending connection,
// or nil if no suitable peer is found. If wait is true, the call will
// block until a new pending peer is found.
func (p *Peers) GetPending(wait bool) *Peer { return nil }

// SetStatus sets the status of a peer.
func (p *Peers) SetStatus(id PeerID, status PeerStatus) {}

// Subscribe subscribes to peer status changes until the given context is
// cancelled. The returned channel is buffered, but the caller must process
// updates in a timely fashion to avoid blocking updates, and must drain the
// channel after cancelling the context.
//
// If currentUp is true, the channel is pre-populated with all peers that have
// PeerStatusUp at the time of the call. This is convenient for reactors that
// need to know about currently connected peers on startup.
func (p *Peers) Subscribe(ctx context.Context, currentUp bool) PeerUpdates { return nil }

// Banner returns a channel that can be used to ban peers.
// FIXME This should probably be moved into Router.
func (p *Peers) Banner() PeerBanner { return nil }
