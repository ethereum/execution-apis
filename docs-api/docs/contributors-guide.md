# Contributors Guide

This guide will explain for new and experienced contributors alike how to
propose changes to Ethereum JSON-RPC API.

## Introduction

The Ethereum JSON-RPC API is the canonical interface between users and the
Ethereum network. Each execution layer client implements the API as defined by
the spec.

As the main source of chain information, anything that is not provided over via
API will not be easily accessible to users.

Clients may choose to include additional APIs; however, anything defined in this
repo is canonical and required to pass rpc-compat testing in hive.

## Guiding Principles

When considering a change to the API, it's important to keep a few guiding
principles in mind.

### Necessity

The most common path to a newly standardized method is necessity. As the
protocol changes over time, new types of data become available. EIP-2930
necessitated the introduction of `eth_accessList` and EIP-1559 necessitated
`eth_feeHistory`.

Therefore, a good question to ask before making a new API proposal is whether
or not the method is strictly necessary. Sometimes the answer is yes even
without a protocol change. For example, `eth_getProof` has been possible since
the initial version of Ethereum -- yet, it was only standardized in recent years
as demand for the functionality grew. Before `eth_getProof`, there was no
interface for getting intermediary trie nodes over the API. This is a great
example of a method that became more necessary over time.

Sometimes efficiency is the basis of necessity. If certain patterns of requests
become popular, it can be advantageous to enshrine the behavior into the API.

### Implementation Complexity

How a method is implemented should be carefully considered before proposing a
change to the API. Although each client is able to validate the Ethereum chain,
there can be a huge variance in actual design decisions.

As an example, a proposal for a method such as `eth_totalSupply` seems
reasonable. This is a quantity that users are often interested in and it would
be nice to have it available. However, tracking the total supply is tricky. There
are several avenues where ether can enter and leave supply. This method would
need to either i) compute the value on demand or ii) store value for each block
height.

Option i) is out, because it would involve executing each block starting with
genesis. Option ii) is viable, but it starts enforcing certain requirements on
clients beyond being able to simply validate the chain. Now during block
ingestion, each client needs to store in their database the supply for that
height. The chain reorg logic also needs to consider this new data. It is not
trivial.

### Backwards Compatibility

There is currently no accepted path to making backwards incompatible changes to
the API. This means that proposals which change syntax or semantics of existing
methods are unlikely to be accepted. A more viable approach is to propose a new
method be created.

Historically we have allowed breaking changes to method outputs, and we will
likely continue to do so. Moving forward, such changes will be hidden behind
versioning and releases. We generally do not view adding fields or keys as a
breaking change. Removing keys may be viewed as a breaking change. When in
doubt, proposing a new method is often a more viable approach than modifying
an existing one.

## Standardization

There is no formal process for standardization of API changes. However, the
outline below should give proposal authors and champions a rough process to
follow.

### Idea

An often overlooked aspect of the standardization journey is the idea phase.
This is an important period of time, because some focused effort at this point
in time can save time and make the rest of the process much smoother.

During the idea phase, it is recommended to contemplate the proposal idea in
the context of the guiding principles above. It's also good to get feedback on
the idea in the open. Just one or two rough acknowledgements from client
developers that an idea makes sense and is worth pursuing can avoid wasting a
lot of time formalizing a proposal that has little chance of being accepted.

### Proposal

The formal proposal stage is where the bulk of time will be spent. A formal
proposal is a PR to this repository ([ethereum/execution-apis][exec-apis]). A
good proposal will have the following:

* a modification to the specification implementing the proposal
* test cases for proposal ([guide][test-gen])
* motivation for the change
* links to acknowledgements that proposal idea is sound
* clear rationale for non-obvious design decisions

When adding new method names, update [wordlist.txt][wordlist] if the spell
checker flags them (e.g., camelCase or versioned names like `simulateV1`).

### Acquiring Support

Once a formal proposal has been created, formal support of clients can be
acquired. The issue or PR needs review from execution client developers to
achieve rough consensus. There are several ways to bring proposals forward:

* **All Core Devs Testing (ACDT) calls** — Recommended. Post a request on the
  AllCoreDevs agenda (usually in [ethereum/pm][pm]) to discuss the proposal.
* **RPC Standards calls** — Bring the proposal to the RPC Standards meetings.
* **json-rpc-api Discord channel** — Discuss in the json-rpc-api channel within
  the [ETH R&D Discord][discord].

Oftentimes, support will be conditional on certain changes. This means that
proposals will cycle between formal proposal work and earning support from
clients. This should be expected and not discourage authors.

### Accepting the Change

After client teams acknowledge and accept the change, it is usually on them to
implement the method in their client. Due to the lack of versioning of the API,
it is preferable that clients release the method roughly at the same time so
that there is not much time when some clients support certain methods and
others don't.

## Contribution Workflow

The end-to-end process from proposal to deployment:

1. **Open an Issue or PR** — Issues should describe in as much detail as
   possible what you want and why. PRs should update or add specs in a way
   compatible with the repo structure and OpenRPC so that speccheck and CI can
   pass.

2. **Obtain Client Consensus** — Get rough consensus from execution client
   developers via ACDT (recommended), RPC Standards calls, or the
   [json-rpc-api Discord channel][discord].

3. **Implement in go-ethereum** — The spec must be implemented in go-ethereum
   so that [rpctestgen][rpctestgen] can generate test fixtures (`.io` files).
   rpctestgen uses go-ethereum libraries to produce the test artifacts;
   upstream go-ethereum changes may be required for new methods.

4. **Merge and Hive** — Once the spec is merged and tests pass, [hive][hive]'s
   [rpc-compat][rpc-compat] simulator pulls the `main` branch and
   automatically tests execution clients. See [Hive rpc-compat
   Integration](#hive-rpc-compat-integration) below.

5. **Hive Updates** — Occasionally, hive/rpc-compat may need to be updated to
   add CLI flags for new API namespaces or methods.

6. **Versioned Releases** — Versioned releases are planned so hive can target
   specific execution-apis versions instead of always using `main`.

## CI Requirements

* **speccheck** — Must pass. Validates test cases in `tests/` against the
  OpenRPC specification.

* **rpctestgen / test fill** — Required for existing methods. For **new
  methods** that require upstream go-ethereum changes, rpctestgen may not pass
  until go-ethereum implements the method. In such cases, maintainers may merge
  PRs with CI exceptions. The spec and speccheck must still pass.

For detailed instructions on running these tools locally, see the
[Tools README][tools-readme].

## Hive rpc-compat Integration

The [hive][hive] test framework runs conformance tests against execution
clients via the [rpc-compat][rpc-compat] simulator. How it works:

* **Consumption** — During Docker build, rpc-compat clones execution-apis and
  copies the `tests/` directory into the simulator container.
* **Branch targeting** — The `branch` build arg (default: `main`) controls
  which ref is fetched. This allows testing against specific commits or branches.
* **Results** — Test results are published at [hive.ethpandaops.io][hivetests]
  under the `rpc-compat` tag.
* **Future** — Versioned releases will allow hive to pin to specific
  execution-apis versions.

[exec-apis]: https://github.com/ethereum/execution-apis
[pm]: https://github.com/ethereum/pm
[test-gen]: ./tests.md
[wordlist]: https://github.com/ethereum/execution-apis/blob/main/wordlist.txt
[discord]: https://discord.gg/tWQwJSaqEE
[tools-readme]: https://github.com/ethereum/execution-apis/blob/main/tools/README.md
[rpctestgen]: https://github.com/ethereum/execution-apis/tree/main/tools
[hive]: https://github.com/ethereum/hive
[rpc-compat]: https://github.com/ethereum/hive/tree/master/simulators/ethereum/rpc-compat
[hivetests]: https://hive.ethpandaops.io
