# Extensions overview

## Proposal
####  Goal
A standard for JSON-RPC error codes & ship a shared catalog of JSON-RPC error codes and messages for EVM clients to unlock consistent tooling and developer ergonomics.

#### Motivation
Client implementations and EVM-compatible chains currently reuse codes or return generic error messages, making cross-client debugging brittle.

#### Solution
The solution incorporates [OpenRPC's extension schemas](https://github.com/open-rpc/specification-extension-spec) feature, specifically `x-error-group` [extension](https://github.com/open-rpc/tools/blob/main/packages/extensions/src/x-error-groups/x-error-groups.json), so common scenarios can be bundled into reusable categories, each backed by a reserved range of **200 codes** outside the JSON-RPC 2.0 reserved bands.
With the error grouping and inline provisioning offered by the extension schemas, we could onboard methods over time with granular control over the errors or groups each method would need to handle without copy pasting in the final spec.

The corresponding PR definition frames these groups as the canonical vocabulary for wallets, infra providers, and execution clients.

## Solution Layout
- `components/` – YAML fragments exposing each error family as an OpenRPC `x-error-group` definition.
- `schemas/x-error-category-ranges.json` – Extension to official `x-error-groups` that enforces the reserved integer windows per category during validation.
    - This is to achieve inbuild validation of the reserved ranges per category using native `minimum` & `maximum` properties of the extended schema.
    - Validation happens while running  `scripts/validate.js` after building the final `refs-openrpc.json` / `openrpc.json`.
- `scripts/build.js` – Loads the schema above, augments the `XErrorGroupsJSON` extension, and merges the groups into `refs-openrpc.json` / `openrpc.json`.

## Implemented Methods
Currently, only below methods import all the error groups via `$ref` and may include inline method-specific codes while still inheriting the standard set.
- `eth_sendTransaction` in `src/eth/submit.yaml`
- `eth_sendRawTransaction` in `src/eth/submit.yaml`


## Reserved ranges at a glance
| Extension group | Category label | Reserved range | Source |
| --- | --- | --- | --- |
| JSON-RPC standard | — | $-32768$ to $-32000$ | JSON-RPC 2.0 spec |
| JSON-RPC non-standard | Client-specific | $-32099$ to $-32000$ | JSON-RPC 2.0 addendum |
| Gas errors | `GAS_ERRORS` | $800$ to $999$ | `gas-error-groups.yaml` |
| Execution errors | `EXECUTION_ERRORS` | $1$ to $999$ | `execution-errors.yaml` |
| TxPool errors | `TXPOOL_ERRORS` | $1000$ to $1199$ | `txpool-errors.yaml` |
| ZK execution errors | `ZK_EXECUTION_ERRORS` | $2000$ to $2199$ | `zk-execution-errors.yaml` |


**Validation** of these bands happens through `XErrorGroupsJSON.schema` in `scripts/build.js`, so build failures flag any out-of-range additions early.


## Extending the catalog
1. Pick or create a YAML fragment under `components/` and add the new entry with `code`, `message`, `data`, and `x-error-category` per the proposal.
2. Stay within the reserved window; the JSON Schema guard in `schemas/x-error-category-ranges.json` will break the build if you drift.
3. Reference the group from the relevant method spec via `$ref: '#/components/x-error-group/<GroupName>'` and layer any bespoke errors inline.
4. Run the documentation build (e.g. `node scripts/build.js`) to regenerate `refs-openrpc.json` / `openrpc.json` and confirm validation passes.

Following this flow keeps the execution client surface aligned with the standard and preserves interoperability for downstream consumers.