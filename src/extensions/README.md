# Extensions overview

## Proposal
####  Goal
A standard for JSON-RPC error codes & ship a shared catalog of JSON-RPC error codes and messages for EVM clients to unlock consistent tooling and developer ergonomics.

#### Motivation
Client implementations and EVM-compatible chains currently reuse codes or return generic error messages, making cross-client debugging brittle.

#### Solution
The solution incorporates [OpenRPC's extension schemas](https://github.com/open-rpc/specification-extension-spec) feature, specifically `x-error-group` [extension](https://github.com/open-rpc/tools/blob/main/packages/extensions/src/x-error-groups/x-error-groups.json), so common scenarios can be bundled into reusable categories, each backed by a reserved range of **200 codes** between **-31999 and -30000** (outside the JSON-RPC 2.0 reserved bands).
With the error grouping and inline provisioning offered by the extension schemas, we could onboard methods over time with granular control over the errors or groups each method would need to handle without copy pasting in the final spec.

The corresponding PR definition frames these groups as the canonical vocabulary for wallets, infra providers, and execution clients.

## Solution Layout
- `components/` – YAML fragments exposing each error family as an OpenRPC `x-error-group` definition.
- `schemas/x-error-category-ranges.json` – Extension to official `x-error-groups` that enforces the reserved integer windows per category during validation.
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
| Gas errors | `GAS_ERRORS` | $-31999$ to $-31800$ | `gas-error-groups.yaml` |
| Execution errors | `EXECUTION_ERRORS` | $-31799$ to $-31600$ | `execution-errors.yaml` |
| (Future) consensus | — | $-31599$ to $-31400$ |
| (Future) networking | — | $-31399$ to $-31200$ |
| TxPool errors | `TXPOOL_ERRORS` | $-31199$ to $-31000$ | `txpool-errors.yaml` |


**Validation** of these bands happens through `XErrorGroupsJSON.schema` in `scripts/build.js`, so build failures flag any out-of-range additions early.


## Extending the catalog
1. Pick or create a YAML fragment under `components/` and add the new entry with `code`, `message`, `data`, and `x-error-category` per the proposal.
2. Stay within the reserved window; the JSON Schema guard in `schemas/x-error-category-ranges.json` will break the build if you drift.
3. Reference the group from the relevant method spec via `$ref: '#/components/x-error-group/<GroupName>'` and layer any bespoke errors inline.
4. Run the documentation build (e.g. `node scripts/build.js`) to regenerate `refs-openrpc.json` / `openrpc.json` and confirm validation passes.

Following this flow keeps the execution client surface aligned with the standard and preserves interoperability for downstream consumers.


## Catalog reference

### [JSON-RPC standard errors](https://www.jsonrpc.org/specification) (`rpc-standard-errors.yaml`)
| Code | Message | Notes |
| --- | --- | --- |
| $-32700$ | Parse error | An error occurred on the server while parsing the JSON text |
| $-32600$ | Invalid request | The JSON sent is not a valid request object |
| $-32601$ | Method not found | The method does not exist / is not available |
| $-32602$ | Invalid params | Invalid method parameter(s) |
| $-32603$ | Internal error | Internal JSON-RPC error |

### [JSON-RPC non-standard errors](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1474.md) (`rpc-non-standard-errors.yaml`)
| Code | Message | Notes |
| --- | --- | --- |
| $-32000$ | Invalid input | Missing or invalid parameters |
| $-32001$ | Resource not found | Requested resource not found |
| $-32002$ | Resource unavailable | Requested resource not available |
| $-32003$ | Transaction rejected | Transaction creation failed |
| $-32004$ | Method not supported | Method is not implemented |
| $-32005$ | Limit exceeded | Request exceeds defined limit |
| $-32006$ | JSON-RPC version not supported | Version of JSON-RPC protocol is not supported |

### Gas errors (`gas-error-groups.yaml`)
| Code | Message | Data |
| --- | --- | --- |
| $-31800$ | GAS_TOO_LOW | Transaction gas is too low / intrinsic gas too low |
| $-31801$ | OUT_OF_GAS | The transaction ran out of gas |
| $-31802$ | GAS_PRICE_TOO_LOW | Gas price too low / gas price below configured minimum gas price |
| $-31803$ | BLOCK_GAS_LIMIT_EXCEEDED | Tx gas limit exceeds max block gas limit / intrinsic gas exceeds gas limit |
| $-31804$ | FEE_CAP_EXCEEDED | Tx fee exceeds cap / max priority fee per gas higher than max fee per gas |
| $-31805$ | GAS_OVERFLOW | Gas overflow error |
| $-31806$ | INSUFFICIENT_TRANSACTION_PRICE | Transaction price must be greater than base fee / max fee per gas less than block base fee |
| $-31807$ | INVALID_MAX_PRIORITY_FEE_PER_GAS | Max priority fee per gas higher than $2^{256}-1$ |
| $-31808$ | INVALID_MAX_FEE_PER_GAS | Max fee per gas higher than $2^{256}-1$ |
| $-31809$ | INSUFFICIENT_FUNDS | Insufficient funds for gas * price + value |
| $-31810$ | TRANSACTION_UNDERPRICED | Transaction's gas price is below the minimum for txpool |
| $-31811$ | REPLACEMENT_TRANSACTION_UNDERPRICED | Replacement transaction is sent without the required price bump |

### Execution errors (`execution-errors.yaml`)
| Code | Message | Data |
| --- | --- | --- |
| $-31600$ | NONCE_TOO_LOW | Transaction nonce is lower than the sender account's current nonce |
| $-31601$ | NONCE_TOO_HIGH | Transaction nonce is higher than the sender account's current nonce |
| $-31602$ | EXECUTION_REVERTED | Execution is reverted by REVERT Opcode |
| $-31603$ | INVALID_OPCODE | An invalid opcode was encountered during execution |
| $-31604$ | OUT_OF_COUNTERS | Not enough step counters to continue the execution |

### TxPool errors (`txpool-errors.yaml`)
| Code | Message | Data |
| --- | --- | --- |
| $-31000$ | ALREADY_KNOWN | Transaction is already known to the transaction pool |
| $-31001$ | INVALID_SENDER | Transaction sender is invalid |
| $-31002$ | INVALID_SIGNATURE | Transaction signature is invalid |
| $-31003$ | TXPOOL_FULL | Transaction pool is full |
| $-31004$ | NEGATIVE_VALUE | Transaction with negative value |
| $-31005$ | OVERSIZED_DATA | Transaction input data exceeds the allowed limit |
| $-31006$ | SENDER_BLACKLISTED | Transaction sender is blacklisted |
| $-31007$ | RECEIVER_BLACKLISTED | Transaction receiver is blacklisted |
| $-31008$ | CHAIN_ID_MISMATCH | Transaction chain ID does not match the expected chain ID |
| $-31009$ | TX_NOT_PERMITTED | Transaction is protected and cannot be permitted for unauthorized users |
| $-31010$ | INVALID_RLP_DATA | Transaction Data contains invalid RLP encoding |
