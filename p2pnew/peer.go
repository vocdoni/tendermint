package p2p

import (
	"context"
	"net/url"
)

// URL represents an externally provided peer address, which is resolved into a
// set of specific transport endpoints. For example, the URL may contain a DNS
// hostname which is resolved into a set of IP addresses. The URL scheme
// specifies which transport to use.
type URL url.URL

// Resolve resolves the URL into a set of endpoints (e.g. resolving a DNS name
// into a list of IP addresses).
func (u URL) Resolve(ctx context.Context) ([]Endpoint, error) {
	return nil, nil
}

// Peer contains information about a peer.
type Peer struct {
	// ID contains the peer's node ID.
	ID []byte
	// URLs is a list of locally provided peer URLs (e.g. given in config file).
	URLs []URL
	// Endpoints contains resolved network endpoints for the node, e.g. IP/port pairs.
	Endpoints []Endpoint
}
