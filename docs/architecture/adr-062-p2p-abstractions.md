# ADR 062: P2P Abstractions

## Changelog

- 2020-11-06: Initial version (@erikgrinaker)

## Context

[ADR 061](adr-061-p2p-refactor-scope.md) decided to refactor the peer-to-peer (P2P) networking stack. The first phase of this is to redesign and refactor the internal P2P architecture and implementation, while retaining protocol compatibility as far as possible.

## Alternative Approaches

> This section contains information around alternative options that are considered before making a decision. It should contain a explanation on why the alternative approach(es) were not chosen.

## Decision

The P2P stack will be redesigned as a message-oriented architecture, primarily relying on Go channels for communication and scheduling. It will use arbitrary stream transports to communicate with peers, peer-addressable channels to pass Protobuf messages between peers, and a router that routes messages between reactors and peers. Message passing is asynchronous with at-most-once delivery.

## Detailed Design

This ADR is primarily concerned with the architecture and interfaces of the P2P stack, not their internal implementation details. Since implementations can be non-trivial, separate ADRs may be submitted for these. The APIs described here should therefore be considered a rough architecture outline, not a complete and final design.

Primary design objectives have been:

* Loose coupling between components, for a simpler architecture with increased testability.
* Better quality-of-service scheduling of messages, with backpressure and increased performance.
* Better peer address detection, advertisement, and exchange.
* Pluggable transports (not necessarily networked).
* Backwards compatibility with the current P2P network protocols.

The main abstractions in the new stack (following Go visibility rules) are:

* `peer`: A node in the network, uniquely identified by a `PeerID`.

* `Transport`: An arbitrary mechanism to exchange raw bytes with a peer.

* `Channel`: A bidirectional channel to exchange Protobuf messages with arbitrary peers, addressed by `PeerID`. There can be any number of channels, each of which can pass one specific message type.

* `Router`: Maintains transport connections to peers and routes channel messages.

* `peerStore`: Stores peer data for the router, in memory and/or on disk.

* Reactor: While this was a first-class concept in the old P2P stack, it is simply a design pattern in the new design, loosely defined as "something which listens on a channel and reacts to messages" (e.g. as simple as a function).

These concepts and related entities are described in detail below, in a bottom-up order.

### Transports

## Status

> A decision may be "proposed" if it hasn't been agreed upon yet, or "accepted" once it is agreed upon. Once the ADR has been implemented mark the ADR as "implemented". If a later ADR changes or reverses a decision, it may be marked as "deprecated" or "superseded" with a reference to its replacement.

{Deprecated|Proposed|Accepted|Declined}

## Consequences

> This section describes the consequences, after applying the decision. All consequences should be summarized here, not just the "positive" ones.

### Positive

### Negative

### Neutral

## References

> Are there any relevant PR comments, issues that led up to this, or articles referenced for why we made the given design choice? If so link them here!

- {reference link}
