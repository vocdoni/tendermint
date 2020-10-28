# ADR 061: P2P Refactor Scope

## Changelog

- 2020-10-28: Initial draft (@erikgrinaker)

## Context

The `p2p` package responsible for peer-to-peer networking is rather old and has a number of weaknesses, including tight coupling, leaky abstractions, lack of tests, DoS vectors, poor performance, custom protocols, and incorrect behavior. A refactor has been discussed for several years ([#2067](https://github.com/tendermint/tendermint/issues/2067)).

Informal Systems are also building a Rust implementation of Tendermint, [Tendermint-rs](https://github.com/informalsystems/tendermint-rs), and plan to implement P2P networking support over the next year. As part of this work, they have requested adopting e.g. [QUIC](https://datatracker.ietf.org/doc/draft-ietf-quic-transport/) as a transport protocol instead of implementing the custom application-level `MConnection` stream multiplexing protocol that Tendermint currently uses.

## Alternative Approaches

There have been recurring proposals to adopt [LibP2P](https://libp2p.io) instead of maintaining our own P2P networking stack (see [#3696](https://github.com/tendermint/tendermint/issues/3696)). However, this would be a highly breaking protocol change, there are indications that we might have to fork and modify LibP2P, and there are concerns about the abstractions used.

In discussions with Informal Systems we decided to begin with incremental improvements to the current P2P stack, add support for pluggable transports, and then gradually start experimenting with LibP2P as a transport layer. If this proves successful, we can consider adopting it for higher-level components at a later time.

## Decision

The P2P stack will be refactored and improved iteratively, in several phases:

* **Phase 1:** code and API refactoring, maintaining protocol compatibility as far as possible.

* **Phase 2:** additional transports and incremental protocol improvements.

* **Phase 3:** disruptive protocol changes and major new features.

The scope of phases 2 and 3 are still uncertain, and will be revisited once the preceding phases have been completed as we'll have a better sense of requirements and challenges.

## Detailed Design

Separate ADRs will be submitted for specific designs and changes in each phase, following research and prototyping. Below are objectives in order of priority.

### Phase 1: Code and API Refactoring

The first phase will focus on improving the internal abstractions and implementations in the `p2p` package. As far as possible, these should not change the P2P protocol.

* Cleaner, decoupled abstractions for e.g. `Reactor`, `Switch`, and `Peer`. #5287 #3833
    * Reactors should receive messages in separate goroutines or via buffered channel. #2888
* Improved peer lifecycle management. #3679 #3719 #3653 #3540 #3183 #3081 #1356
    * Peer prioritization. #2860 #2041
* Pluggable transports, with `MConnection` as one implementation. #5587 #2430 #805
* Improved peer address handling.
    * Address book refactor. #4848 #2661
    * Transport-agnostic peer addressing. #5587 #3782 #3692
    * Improved detection and advertisement of own address. #5588 #4260 #3716 #1727
    * Support multiple IPs per peer. #1521 #2317

The refactor should also attempt to address testability, observability, quality-of-service, backpressure, DoS resilience, performance, and security - at least to some extent. Much of this will be revisited as explicit objectives in phase 2.

### Phase 2: Additional Transports and Protocol Improvements

This phase will focus on protocol improvements and other breaking changes. The following are considered proposals that will need to be evaluated and discussed separately once the refactor is done, and additional proposals are likely to be added during phase 1.

* QUIC transport. [#198](https://github.com/tendermint/spec/issues/198)
* Noise protocol for secret connection handshake. #5589 #3340
* Peer ID in connection handshake. #5590
* Peer and service discovery (e.g. RPC nodes, state sync snapshots). #5481 #4583
* Compression. #2375
* Rate-limiting, backpressure, and QoS scheduling. #4753 #2338
* Improved metrics and tracing. #3849 #2600
* Simplify and clean up P2P configuration options.

### Phase 3: Disruptive Protocol Changes and Major Features

This phase covers speculative, wide-reaching proposals that are poorly defined and highly uncertain. They will be evaluated and discussed once the previous phases are done.

* Adopt LibP2P. #3696
* Dynamic channel advertisment, as reactors are enabled/disabled. #4394 #1148
* Allow cross-reactor communication, possibly without channels.
* Pubsub-style networking topology and pattern.
* Support multiple chain IDs in the same network.

## Status

Proposed

## Consequences

### Positive

### Negative

### Neutral

## References

> Are there any relevant PR comments, issues that led up to this, or articles referenced for why we made the given design choice? If so link them here!

- {reference link}
