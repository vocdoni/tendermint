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

The P2P stack will be refactored and improved in several phases:

* Phase 1: code and API refactoring, maintaining protocol compatibility as far as possible.

* Phase 2: additional transport protocols and incremental protocol improvements.

* Phase 3: broader disruptive protocol changes and major new features.

The scope of phases 2 and 3 are still uncertain, and will be revisited once the preceding phases have been completed and we have a better sense of requirements and challenges.

## Detailed Design

Separate ADRs will be submitted for specific designs and changes in each phase, following research and prototyping.

> This section does not need to be filled in at the start of the ADR, but must be completed prior to the merging of the implementation.
>
> Here are some common questions that get answered as part of the detailed design:
>
> - What are the user requirements?
>
> - What systems will be affected?
>
> - What new data structures are needed, what data structures will be changed?
>
> - What new APIs will be needed, what APIs will be changed?
>
> - What are the efficiency considerations (time/space)?
>
> - What are the expected access patterns (load/throughput)?
>
> - Are there any logging, monitoring or observability needs?
>
> - Are there any security considerations?
>
> - Are there any privacy considerations?
>
> - How will the changes be tested?
>
> - If the change is large, how will the changes be broken up for ease of review?
>
> - Will these changes require a breaking (major) release?
>
> - Does this change require coordination with the SDK or other?

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
