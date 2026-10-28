# Engine API -- Bogota

Engine API changes introduced in Bogota.

This specification is based on and extends [Engine API - Amsterdam](./amsterdam.md) specification.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Constants](#constants)
- [Structures](#structures)
  - [PayloadAttributesV5](#payloadattributesv5)
  - [PayloadStatusV2](#payloadstatusv2)
  - [RecursiveStarkV1](#recursivestarkv1)
  - [InclusionListV1](#inclusionlistv1)
  - [PayloadAttributesV6](#payloadattributesv6)
- [Routines](#routines)
  - [Payload building](#payload-building)
  - [Inclusion list proof obligations](#inclusion-list-proof-obligations)
- [Methods](#methods)
  - [engine_newPayloadV6](#engine_newpayloadv6)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)
  - [engine_getInclusionListV1](#engine_getinclusionlistv1)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)
  - [engine_forkchoiceUpdatedV5](#engine_forkchoiceupdatedv5)
    - [Request](#request-2)
    - [Response](#response-2)
    - [Specification](#specification-2)
  - [engine_newPayloadV7](#engine_newpayloadv7)
    - [Request](#request-3)
    - [Response](#response-3)
    - [Specification](#specification-3)
  - [engine_getInclusionListV2](#engine_getinclusionlistv2)
    - [Request](#request-4)
    - [Response](#response-4)
    - [Specification](#specification-4)
  - [engine_forkchoiceUpdatedV6](#engine_forkchoiceupdatedv6)
    - [Request](#request-5)
    - [Response](#response-5)
    - [Specification](#specification-5)
  - [Update the methods of previous forks](#update-the-methods-of-previous-forks)
    - [Amsterdam API](#amsterdam-api)
    - [Bogota API](#bogota-api)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Constants

| Name | Value |
| - | - |
| `MAX_TRANSACTIONS_BYTES_PER_INCLUSION_LIST` |  `uint64(8192) = 2**13` |
| `MAX_IL_PROOF_BYTES` | `uint64(1310720) = 5 * 2**18` |

`MAX_IL_PROOF_BYTES` is the [EIP-8288](https://eips.ethereum.org/EIPS/eip-8288) constant of the same name and is provisional there. The *EIP-8288 fork* is the fork activating EIP-8288.

## Structures

### PayloadAttributesV5

This structure has the syntax of [`PayloadAttributesV4`](./amsterdam.md#payloadattributesv4) and appends a single field: `inclusionListTransactions`.

- `timestamp`: `QUANTITY`, 64 Bits - value for the `timestamp` field of the new payload
- `prevRandao`: `DATA`, 32 Bytes - value for the `prevRandao` field of the new payload
- `suggestedFeeRecipient`: `DATA`, 20 Bytes - suggested value for the `feeRecipient` field of the new payload
- `withdrawals`: `Array of WithdrawalV1` - Array of withdrawals, each object is an `OBJECT` containing the fields of a `WithdrawalV1` structure.
- `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
- `slotNumber`: `QUANTITY`, 64 Bits - value for the `slotNumber` field of the new payload
- `targetGasLimit`: `QUANTITY`, 64 Bits - target value for the `gasLimit` field of the new payload
- `inclusionListTransactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718).

### PayloadStatusV2

This structure has the syntax of `PayloadStatusV1` and appends a single field: `inclusionListSatisfied`

- `status`: `enum` - `"VALID" | "INVALID" | "SYNCING" | "ACCEPTED"`
- `latestValidHash`: `DATA|null`, 32 Bytes - the hash of the most recent *valid* block in the branch defined by payload and its ancestors
- `validationError`: `String|null` - a message providing additional details on the validation error if the payload is classified as `INVALID`.
- `inclusionListSatisfied`: `BOOLEAN|null` - whether the payload satisfied the inclusion list constraints if it is deemed `VALID`; `null` otherwise.

### RecursiveStarkV1

This structure is the `InclusionListProof` that [EIP-8288](https://eips.ethereum.org/EIPS/eip-8288) appends to a signed inclusion list.

- `starkProof`: `DATA` - the serialized recursive aggregate proof, or `"0x"` for an empty proof
- `proofDeps`: `Array of DATA`, 96 Bytes each - the dependencies the proof claims, each in the EIP-8288 96-byte dependency encoding, or `[]` for an empty proof

The SSZ serialization of the corresponding `InclusionListProof`, `8 + len(starkProof) + 96 * len(proofDeps)` bytes, **MUST NOT** exceed `MAX_IL_PROOF_BYTES`.

### InclusionListV1

This structure is one inclusion list counted by [EIP-7805](https://eips.ethereum.org/EIPS/eip-7805), with its EIP-8288 proof.

- `transactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718).
- `recursiveStark`: [`RecursiveStarkV1`](#recursivestarkv1) - the proof of the signed inclusion list, passed unchanged even if it does not verify.

### PayloadAttributesV6

This structure has the syntax of [`PayloadAttributesV4`](./amsterdam.md#payloadattributesv4) and appends a single field: `inclusionLists`.

- `timestamp`: `QUANTITY`, 64 Bits - value for the `timestamp` field of the new payload
- `prevRandao`: `DATA`, 32 Bytes - value for the `prevRandao` field of the new payload
- `suggestedFeeRecipient`: `DATA`, 20 Bytes - suggested value for the `feeRecipient` field of the new payload
- `withdrawals`: `Array of WithdrawalV1` - Array of withdrawals, each object is an `OBJECT` containing the fields of a `WithdrawalV1` structure.
- `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
- `slotNumber`: `QUANTITY`, 64 Bits - value for the `slotNumber` field of the new payload
- `targetGasLimit`: `QUANTITY`, 64 Bits - target value for the `gasLimit` field of the new payload
- `inclusionLists`: `Array of InclusionListV1` - the counted inclusion lists in committee-position order, each an `OBJECT` containing the fields of an [`InclusionListV1`](#inclusionlistv1) structure.

## Routines

### Payload building

This routine follows the same specification as [Payload building](./paris.md#payload-building) with the following changes to the processing flow:

1. Client software **MUST** take `inclusionListTransactions` into account during the payload build process. The built `ExecutionPayload` **MUST** satisfy the inclusion list constraints with respect to `inclusionListTransactions` as defined in [EIP-7805](https://eips.ethereum.org/EIPS/eip-7805).

2. With [`PayloadAttributesV6`](#payloadattributesv6), the constraints apply to the transactions that `inclusionLists` obligate, as defined in [Inclusion list proof obligations](#inclusion-list-proof-obligations). Client software **MAY** aggregate the proofs that verify there into the payload's recursive STARK, discarding the dependencies absent from the payload.

### Inclusion list proof obligations

Client software evaluates each [`InclusionListV1`](#inclusionlistv1) independently under the fork at the payload's `timestamp`, which fixes the static validity rules and the `AGGREGATED_VK`.

1. `recursiveStark` verifies if `starkProof` is nonempty, `proofDeps` is sorted in ascending byte order without duplicates, every scheme in `proofDeps` is in `ENABLED_SCHEMES`, and `starkProof` verifies against `get_deps_hash(proofDeps)` under the `AGGREGATED_VK`, as defined in EIP-8288.

2. A transaction without a dependency verification frame is obligated. A transaction with one is obligated if it passes the static validity rules of EIP-8141 and EIP-8288, `recursiveStark` verifies, and every dependency of the transaction is in `proofDeps`.

3. The inclusion list constraints of EIP-7805 apply to the union of the obligated transactions of all lists, assuming the dependencies of an obligated transaction are satisfied. A transaction that does not decode and a `recursiveStark` that does not verify affect only their own list, and are not an error.

## Methods

### engine_newPayloadV6

Method parameter list is extended with `inclusionListTransactions`.

#### Request

* method: `engine_newPayloadV6`
* params:
  1. `executionPayload`: [`ExecutionPayloadV4`](./amsterdam.md#executionpayloadv4).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected blob versioned hashes to validate.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
  4. `executionRequests`: `Array of DATA` - List of execution layer triggered requests. Each list element is a `requests` byte array as defined by [EIP-7685](https://eips.ethereum.org/EIPS/eip-7685). The first byte of each element is the `request_type` and the remaining bytes are the `request_data`. Elements of the list **MUST** be ordered by `request_type` in ascending order. Elements with empty `request_data` **MUST** be excluded from the list.
  5. `inclusionListTransactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718).
* timeout: 6s

#### Response

* result: [`PayloadStatusV2`](#payloadstatusv2)
* error: code and message set in case an exception happens while processing the payload.

#### Specification

This method follows the same specification as [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload does not fall within the time frame of the Bogota fork.

2. Client software **MUST** set `inclusionListSatisfied` in the following way:

    1. If the payload is deemed `VALID`, `inclusionListSatisfied` **MUST** be set to whether the payload satisfied the inclusion list constraints.

    2. Otherwise, `inclusionListSatisfied` **MUST** be `null`.

3. Client software **MUST** retain `inclusionListTransactions` for a payload with `ACCEPTED` status. Client software **MAY** discard them once the payload is no longer the tip of a branch.

### engine_getInclusionListV1

#### Request

* method: `engine_getInclusionListV1`
* params: []
* timeout: 1s

#### Response

* result: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718).
* error: code and message set in case an exception happens while getting the inclusion list.

#### Specification

1. Client software **MUST** provide, based on its local view of the mempool, a list of transactions for the inclusion list satisfying the following conditions:

    1. Every transaction **MUST** have non-zero length (at least 1 byte).

    2. The transaction list **SHOULD NOT** include any [blob transaction](https://eips.ethereum.org/EIPS/eip-4844#blob-transaction).

    3. The total byte length of the transaction list **MUST NOT** exceed `MAX_TRANSACTIONS_BYTES_PER_INCLUSION_LIST`.

2. The strategy for selecting transactions is implementation dependent.

### engine_forkchoiceUpdatedV5

#### Request

* method: `engine_forkchoiceUpdatedV5`
* params:
  1. `forkchoiceState`: [`ForkchoiceStateV1`](./paris.md#forkchoicestatev1).
  2. `payloadAttributes`: `Object|null` - Instance of [`PayloadAttributesV5`](#payloadattributesv5) or `null`.
  3. `custodyColumns`: `DATA|null`, 16 Bytes - Interpreted as a bitarray of length `CELLS_PER_EXT_BLOB` indicating which column indices form the CL's custody set, or `null` if the CL does not provide custody services.
* timeout: 8s

#### Response

* result: `object`
  - `payloadStatus`: [`PayloadStatusV2`](#payloadstatusv2), with the same `status` value restrictions as [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4).
  - `payloadId`: `DATA|null`, 8 Bytes - identifier of the payload build process or `null`
* error: code and message set in case an exception happens while the validating payload, updating the forkchoice or initiating the payload build process.

#### Specification

This method follows the same specification as [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4) with the following changes to the processing flow:

1. Extend point (8) of the `engine_forkchoiceUpdatedV1` [specification](./paris.md#specification-1) by defining the following sequence of checks that **MUST** be run over `payloadAttributes`:

    1. `payloadAttributes` matches the [`PayloadAttributesV5`](#payloadattributesv5) structure, return `-38003: Invalid payload attributes` on failure.

    2. `payloadAttributes.timestamp` does not fall within the time frame of the Bogota fork, return `-38005: Unsupported fork` on failure.

2. Extend point (9) of the `engine_forkchoiceUpdatedV1` [specification](./paris.md#specification-1) by defining `payloadStatus.inclusionListSatisfied`:

    1. If the payload referenced by `forkchoiceState.headBlockHash` is deemed `VALID`, `payloadStatus.inclusionListSatisfied` **MUST** be set to whether the payload satisfied the inclusion list constraints.

    2. Client software **MUST** use the retained `inclusionListTransactions` if it validates the payload and checks whether the payload satisfies the inclusion list constraints while processing the call.

    3. Otherwise, `payloadStatus.inclusionListSatisfied` **MUST** be `null`.

### engine_newPayloadV7

Method parameter list replaces `inclusionListTransactions` with `inclusionLists`.

#### Request

* method: `engine_newPayloadV7`
* params:
  1. `executionPayload`: [`ExecutionPayloadV4`](./amsterdam.md#executionpayloadv4).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected blob versioned hashes to validate.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
  4. `executionRequests`: `Array of DATA` - List of execution layer triggered requests, as in [`engine_newPayloadV6`](#engine_newpayloadv6).
  5. `inclusionLists`: `Array of InclusionListV1` - the counted inclusion lists in committee-position order, each an `OBJECT` containing the fields of an [`InclusionListV1`](#inclusionlistv1) structure.
* timeout: 6s

#### Response

* result: [`PayloadStatusV2`](#payloadstatusv2)
* error: code and message set in case an exception happens while processing the payload.

#### Specification

This method follows the same specification as [`engine_newPayloadV6`](#engine_newpayloadv6) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload does not fall within the time frame of the EIP-8288 fork.

2. Client software **MUST** return `-32602: Invalid params` error if `inclusionLists` has more than the EIP-7805 `IL_COMMITTEE_SIZE` entries, or if an entry's `transactions` exceed `MAX_TRANSACTIONS_BYTES_PER_INCLUSION_LIST` bytes or its `recursiveStark` exceeds `MAX_IL_PROOF_BYTES`.

3. Client software **MUST** set `inclusionListSatisfied` with respect to the transactions that `inclusionLists` obligate, as defined in [Inclusion list proof obligations](#inclusion-list-proof-obligations).

4. Client software **MUST** retain `inclusionLists` for a payload with `ACCEPTED` status. Client software **MAY** discard them once the payload is no longer the tip of a branch.

### engine_getInclusionListV2

#### Request

* method: `engine_getInclusionListV2`
* params: []
* timeout: 1s

#### Response

* result: [`InclusionListV1`](#inclusionlistv1)
* error: code and message set in case an exception happens while getting the inclusion list.

#### Specification

This method follows the same specification as [`engine_getInclusionListV1`](#engine_getinclusionlistv1) for `transactions`, with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the EIP-8288 fork is not scheduled.

2. `recursiveStark` **MUST** be an empty proof, `starkProof` of `"0x"` and `proofDeps` of `[]`, or a proof that verifies, as defined in [Inclusion list proof obligations](#inclusion-list-proof-obligations) under the fork of the payload the list constrains, and whose `proofDeps` contain the dependencies of every transaction in `transactions`.

3. Client software **MAY** reuse a proof it holds, such as a verified mempool wrapper proof, whose `proofDeps` also contain dependencies of no transaction in `transactions`. Client software **SHOULD NOT** delay the response beyond the timeout to produce a proof, and **SHOULD** leave out a transaction with a dependency verification frame that `recursiveStark` does not cover, as it is not obligated.

### engine_forkchoiceUpdatedV6

#### Request

* method: `engine_forkchoiceUpdatedV6`
* params:
  1. `forkchoiceState`: [`ForkchoiceStateV1`](./paris.md#forkchoicestatev1).
  2. `payloadAttributes`: `Object|null` - Instance of [`PayloadAttributesV6`](#payloadattributesv6) or `null`.
  3. `custodyColumns`: `DATA|null`, 16 Bytes - as in [`engine_forkchoiceUpdatedV5`](#engine_forkchoiceupdatedv5).
* timeout: 8s

#### Response

Refer to the response for [`engine_forkchoiceUpdatedV5`](#engine_forkchoiceupdatedv5).

#### Specification

This method follows the same specification as [`engine_forkchoiceUpdatedV5`](#engine_forkchoiceupdatedv5) with the following changes to the processing flow:

1. Replace the checks of point (1) of [`engine_forkchoiceUpdatedV5`](#engine_forkchoiceupdatedv5) with:

    1. `payloadAttributes` matches the [`PayloadAttributesV6`](#payloadattributesv6) structure and the bounds of point (2) of [`engine_newPayloadV7`](#engine_newpayloadv7), return `-38003: Invalid payload attributes` on failure.

    2. `payloadAttributes.timestamp` does not fall within the time frame of the EIP-8288 fork, return `-38005: Unsupported fork` on failure.

2. Client software **MUST** use the retained `inclusionLists` in place of `inclusionListTransactions` when it checks whether the payload referenced by `forkchoiceState.headBlockHash` satisfies the inclusion list constraints.

### Update the methods of previous forks

#### Amsterdam API

For the following methods:

- [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5)
- [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4)

a validation **MUST** be added:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of payload is greater than or equal to the Bogota activation timestamp.

#### Bogota API

For the following methods:

- [`engine_newPayloadV6`](#engine_newpayloadv6)
- [`engine_forkchoiceUpdatedV5`](#engine_forkchoiceupdatedv5)

a validation **MUST** be added:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of payload is greater than or equal to the EIP-8288 fork activation timestamp.
