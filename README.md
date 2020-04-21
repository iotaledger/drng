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

## Setting up a committee
+ generate key-pair
+ create group.toml
+ start drand
+ update goshimmer config with identities and distKey

## Installation
### Official release
Please go use the latest drand binary in the
[release page](https://github.com/iotaledger/drng/releases).

### Manual installation
Drand can be installed via [Golang](https://golang.org/). 
By default, drand saves the configuration
files such as the long-term key pair, the group file, and the collective public
key in the directory `$HOME/.drand/`.

#### Via Golang
Make sure that you have a working [Golang
installation](https://golang.org/doc/install) and that your
[GOPATH](https://golang.org/doc/code.html#GOPATH) is set.

Then compile drand via:
```bash
git clone https://github.com/iotaledger/drng
cd drng
go build
```

## Usage
This section explains in details the workflow to have a working group of drand
nodes generate randomness. On a high-level, the workflow looks like this:
+ **Setup**: generation of individual long-term key pair and the group file and
  starting the drand daemon.
+ **Distributed Key Generation**: each drand node collectively participates in
  the DKG.
+ **Randomness Generation**: the randomness beacon automatically starts as soon
  as the DKG protocol is finished.

### Setup
The setup process for a drand node consists of two steps:
1. Generate the long-term key pair for each node
2. Setup the group configuration file

#### Long-Term Key
To generate the long-term key pair `drand_id.{secret,public}` of the drand
daemon, execute
```
drand generate-keypair <address:port>
```
where `<address:port>` is the address from which your drand daemon is reachable. The
address must be reachable over a TLS connection directly or via a reverse proxy
setup. In case you need non-secured channel, you can pass the `--tls-disable`
flag.

For example, by running:
```bash
drand generate-keypair --tls-disable 172.16.222.3:8000
```

You should get something like:
```toml
You can copy paste the following snippet to a common group.toml file:
[[Nodes]]
Address = "172.16.222.3:8000"
Key = "b03293e70589d34341ab9f141e7a57b43441083823fd4fab13d1900047c00c0337d6c51248bd33ab0f844143b469509a"
TLS = false
```

#### Group Configuration
All informations regarding a group of drand nodes necessary for drand to
function properly are located inside a group.toml configuration file. To run a
DKG protocol, one needs to generate this group configuration file from all
individual long-term keys generated in the previous step. One can do so with:
```
drand group <pk1> <pk2> ... <pkn>
```
where `<pki>` is the public key file `drand_id.public` of the i-th participant.
The group file is generated in the current directory under `group.toml`.
**NOTE:** At this stage, this group file MUST be distributed to all
participants!

A `group.toml` file should look like:

```toml
Threshold = 3
Period = "10s"

[[Nodes]]
Address = "172.16.222.3:8000"
Key = "b03293e70589d34341ab9f141e7a57b43441083823fd4fab13d1900047c00c0337d6c51248bd33ab0f844143b469509a"
TLS = false

[[Nodes]]
Address = "172.16.222.2:8004"
Key = "a29f454f40aad47cf5c7f2a42ce3d0fb8765b7bf2c65176aaeee5fa922aa2ba45c959e59afc8af6abf1eb508de44e822"
TLS = false

[[Nodes]]
Address = "172.16.222.4:8001"
Key = "8b66bfaa83e6fee3e1f3c3952ae1b6b16afccc411d5fd7ab31f3e2e3eb67d7eddcaf9f5b5a51dec2ba0f0fd147d636d4"
TLS = false

[[Nodes]]
Address = "172.16.222.5:8000"
Key = "8b11d779c53574b78c0321f63239e973454bc933f8796020a9dfc95bca593dcdfd87f8996cceca5782f2c5f09b5e3e5a"
TLS = false

[[Nodes]]
Address = "172.16.222.6:8003"
Key = "aac5f76016ece3ff96fd356991b3c1a965af05d09bef5c3d3bbd872e71e2155338e6d63f128b1a6378a47b75a0aa2c19"
TLS = false
```

##### Randomness Beacon Period
drand updates the configuration file after the DKG protocol finishes, with the
distributed public key and automatically starts running the randomness beacon.
By default, a randomness beacon has a period of 1mn, I.E. new randomness is
generated every minute. If you wish to change the period, you must include that
information **inside** the group configuration file. You can do by appending a
flag to the command such as :
```
drand group --period 2m <pk1> <pk2> ... <pkn>
```

Or simply by editing manually the group file afterwards: it's a TOML
configuration file. The period must be readable by the
[time](https://golang.org/pkg/time/#ParseDuration) package.

### Starting drand daemon
The daemon does not go automatically in background, so you must run it with ` &
` in your terminal, within a screen / tmux session, or with the `-d` option
enabled for the docker commands. Once the daemon is running, the way to issue
commands to the daemon is to use the control functionalities.  The control
client has to run on the same server as the drand daemon, so only drand
administrators can issue command to their drand daemons.

#### Without TLS
Although we **do not recommend** it, you can always disable TLS in drand via:
```bash
drand start --tls-disable
```

### Distributed Key Generation
After running all drand daemons, each operator needs to issue a command to
start the DKG protocol, using the group file generated before. One can do so
using the control client with:
```
drand share <group-file>  --timeout 10s
```

One of the nodes has to function as the leader to initiate the DKG protocol (no
additional trust assumptions), he can do so with:
```
drand share --leader <group-file>
```

Once running, the leader initiates the distributed key generation protocol to
compute the distributed public key (`dist_key.public`) and the private key
shares (`dist_key.private`) together with the participants specified in
`drand_group.toml`. Once the DKG has finished, the keys are stored as
`$HOME/.drand/groups/dist_key.{public,private}`.

The timeout is an optional parameter indicating the maximum timeout the DKG
protocol will wait. If there are some failed nodes during the DKG, then the DKG
will finish only after the given timeout. The default value is set to 10s (see
[`core/constants.go`](https://github.com/dedis/drand/blob/master/core/constants.go)
file).

**Distributed Public Key**: More generally, for third party implementation of
randomness beacon verification, one only needs the distributed public key. If
you are an administrator of a drand node, you can use the control port as the
following:
```bash
drand show cokey
```

Otherwise, you can use the exposed API: `/api/info/distkey`.

### GoShimmer configuration
To configure the GoShimmer node to use a given dRNG committee, you need to fill the `drng` section of the GoShimmer `config.json` file, where:
+ *instanceId* is the identifier of the committee;
+ *threshold* is the threshold for reconstructing the collective signature;
+ *distributedPubKey* is the distributed public key generated by the committee during the DKG phase ( you can retrieve this by using the API: `/api/info/distkey` on any of the committee address);
+ *commiteeMembers* is the list of the public keys belonging to the GoShimmer nodes of the committee (you can retrieve this by using the API: `/info` on your GoShimmer node).

An example of such a configuration is:

```json
"drng": {
    "instanceId": 1,
    "threshold": 3,
    "distributedPubKey": "8ed27b059bbb314966660d1fb2ce5b146e6af33d729ab434c5024049c7b9f826eb354db991e4e81d4c820d2d024c8c2b",
    "committeeMembers": [
      "CjUsn86jpFHWnSCx3NhWfU4Lk16mDdy1Hr7ERSTv3xn9",
      "C8x1QxsPWtpQ1LLzzQeLBKEpgSLdBfViLDTQXYdtNupB",
      "2BSJEN4dQrQMdZpXQnjEA1GU7cWBYyV1AZJkKZrsSHmT",
      "97qZKq2m6hbbWcoZrnYrP5gdHxoNMwxdjFR5xM826BHP",
      "D4rCPFGG8WzCU3uKDBbqzFz5vZDg6z5QkUbUG3zNkg9V"
    ]
  },
```