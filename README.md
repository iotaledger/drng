This repo is a based on [drand](https://github.com/drand/drand) and just adds 
the API-dRNG integration of [GoShimmer](https://github.com/iotaledger/goshimmer).

### Disclaimer
**This software is considered experimental and has NOT received a third-party
audit yet. Therefore, DO NOT USE it in production or for anything security
critical at this point.**

# dRNG - Distributed Random Number Generator
This repository provides a first dRNG implementation that requires a prefixed committee. Such a committee can be selected, for instance, by the IOTA community voting on which node should be part of the committee. Ideally, IOTA will have different committees, each of one with a recommended priority, so that the network can freely decide which one to follow.

**Note:** future iteration of the dRNG will allow for a more dynamic committee. One approach could be to select the IOTA nodes with highest **mana** and refresh the committee selection every now and then. Since there is no perfect consensus on mana and different nodes can have different mana values, we require all of the nodes interested in the committee participation to prepare a special *application* message which determines the value of mana of a given node. Then the committee is formed from the top *n* highest mana holders candidates. Such a committee would be updated periodically, to account for nodes going offline and changes in mana.

## Motivation
At its core, the Fast Probabilistic Consensus (FPC) runs to resolve potential conflicting transactions by voting on them. 
FPC requires a random number generator (RNG) to be more resilient to an attack aiming at creating a meta-stable state, 
where nodes in the network are constantly toggling their opinion on a given transaction and thus are unable to finalize it. 
Such a RNG can be provided by either a trusted and centralized entity or be decentralized and distributed. 
Clearly, the fully decentralized nature of coordicide mandates the latter option, and this option is referred to a distributed RNG (dRNG).

A dRNG can be implemented in very different ways, for instance by leveraging on cryptographic primitives such as verifiable secret sharing and threshold signatures, 
by using cryptographic sortition or also with verifiable delay functions. 
After reviewing some existing solutions, we decided to use a variant of the [drand](https://github.com/drand/drand) protocol, 
originally developed within the [DEDIS organization](https://github.com/dedis), and as of December 2019, is now under the drand organization.
This protocol has been already used by other projects such as [The League of Entropy](https://www.cloudflare.com/leagueofentropy/).

## Drand - A Distributed Randomness Beacon Daemon
Drand (pronounced "dee-rand") is a distributed randomness beacon daemon written
in [Golang](https://golang.org/). Servers running drand can be linked with each
other to produce collective, publicly verifiable, unbiased, unpredictable
random values at fixed intervals using bilinear pairings and threshold
cryptography. Drand nodes can also serve locally-generated private randomness
to clients.

In a nutshell, drand works in two phases: **setup** and **generation**.
In the setup phase, a set of nodes (hereafter referred as “committee”) run a distributed key generation (DKG) protocol 
to create a collective private and public key pair shared among the members of the committee. 
More specifically, at the end of this phase, each member obtains a copy of the public key as well as a private key share of the collective private key, 
such that no individual member knows the entire collective private key. 
These private key shares will then be used by the committee members to sign their contributions during the next phase.
The generation phase works in discrete rounds. 
In every round, the committee produces a new random value by leveraging on a deterministic threshold signature scheme such as BLS. 
Each member of the committee creates in round *r* the partial BLS signature *σ_r* on the message *m=H(r || ς_r-1)* 
where *ς_r-1* denotes the full BLS threshold signature from the previous round *r−1* and *H* is a cryptographic hash function. 
Once at least *t* members have broadcasted their partial signatures *σ_r* on *m*, 
anyone can recover the full BLS threshold signature *ς_r* (via Lagrange interpolation) which corresponds to the random value of round *r*. 
Then, the committee moves to the next round and reiterates the above process. For the first round, each member signs a seed fixed during the setup phase. 
This process ensures that every new random value depends on all previously generated signatures. 
If you are interested in knowing more about drand, we recommend you to check out their [Github repository](https://github.com/drand/drand).

## Installation

You can check out our [wiki](https://github.com/iotaledger/drng/wiki) to know more about installing, using drand and seeting up a committee.