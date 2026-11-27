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
  - [ExecutionPayloadV5](#executionpayloadv5)
- [Routines](#routines)
  - [Payload building](#payload-building)
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
  - [engine_getPayloadV7](#engine_getpayloadv7)
    - [Request](#request-4)
    - [Response](#response-4)
    - [Specification](#specification-4)
  - [Update the methods of previous forks](#update-the-methods-of-previous-forks)
    - [Amsterdam API](#amsterdam-api)
    - [Bogota API](#bogota-api)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Constants

| Name | Value |
| - | - |
| `MAX_TRANSACTIONS_BYTES_PER_INCLUSION_LIST` |  `uint64(8192) = 2**13` |
| `MAX_DEPENDENCY_PROOF_BYTES` | `uint64(8388608) = 2**23` |

`MAX_DEPENDENCY_PROOF_BYTES` is the [EIP-8141](https://eips.ethereum.org/EIPS/eip-8141) constant of the same name. The *EIP-8141 fork* is the fork activating EIP-8141.

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
- `inclusionListTransactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`).

### PayloadStatusV2

This structure has the syntax of `PayloadStatusV1` and appends a single field: `inclusionListSatisfied`

- `status`: `enum` - `"VALID" | "INVALID" | "SYNCING" | "ACCEPTED"`
- `latestValidHash`: `DATA|null`, 32 Bytes - the hash of the most recent *valid* block in the branch defined by payload and its ancestors
- `validationError`: `String|null` - a message providing additional details on the validation error if the payload is classified as `INVALID`.
- `inclusionListSatisfied`: `BOOLEAN|null` - whether the payload satisfied the inclusion list constraints if it is deemed `VALID`; `null` otherwise.

### ExecutionPayloadV5

This structure has the syntax of [`ExecutionPayloadV4`](./amsterdam.md#executionpayloadv4) and appends a single field: `dependencyProof`.

- `parentHash`: `DATA`, 32 Bytes
- `feeRecipient`:  `DATA`, 20 Bytes
- `stateRoot`: `DATA`, 32 Bytes
- `receiptsRoot`: `DATA`, 32 Bytes
- `logsBloom`: `DATA`, 256 Bytes
- `prevRandao`: `DATA`, 32 Bytes
- `blockNumber`: `QUANTITY`, 64 Bits
- `gasLimit`: `QUANTITY`, 64 Bits
- `gasUsed`: `QUANTITY`, 64 Bits
- `timestamp`: `QUANTITY`, 64 Bits
- `extraData`: `DATA`, 0 to 32 Bytes
- `baseFeePerGas`: `QUANTITY`, 256 Bits
- `blockHash`: `DATA`, 32 Bytes
- `transactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718)
- `withdrawals`: `Array of WithdrawalV1` - Array of withdrawals, each object is an `OBJECT` containing the fields of a `WithdrawalV1` structure.
- `blobGasUsed`: `QUANTITY`, 64 Bits
- `excessBlobGas`: `QUANTITY`, 64 Bits
- `blockAccessList`: `DATA` - RLP-encoded block access list as defined in [EIP-7928](https://eips.ethereum.org/EIPS/eip-7928)
- `slotNumber`: `QUANTITY`, 64 Bits
- `dependencyProof`: `DATA` - the block's `dependency_proof` as defined in [EIP-8141](https://eips.ethereum.org/EIPS/eip-8141), at most `MAX_DEPENDENCY_PROOF_BYTES` bytes, `"0x"` if the payload has no dependencies

The dependency proof is part of neither the block header nor the block body. The header field `dependency_proof_hash` of EIP-8141 is not sent. Client software derives it as `keccak256(dependencyProof)` and includes it when it computes `blockHash`, so a `dependencyProof` other than the one the block was built with makes `blockHash` mismatch.

## Routines

### Payload building

This routine follows the same specification as [Payload building](./paris.md#payload-building) with the following changes to the processing flow:

1. Client software **MUST** take `inclusionListTransactions` into account during the payload build process. The built `ExecutionPayload` **MUST** satisfy the inclusion list constraints with respect to `inclusionListTransactions` as defined in [EIP-7805](https://eips.ethereum.org/EIPS/eip-7805).

2. Client software **MUST NOT** reject a payload solely because `inclusionListTransactions` contains a malformed transaction
(e.g., an arbitrary byte string). Such transactions must be tolerated in accordance with [EIP-7805](https://eips.ethereum.org/EIPS/eip-7805).

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
  5. `inclusionListTransactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`).
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

4. Client software **MUST NOT** reject a payload solely because `inclusionListTransactions` contains a malformed transaction
(e.g., an arbitrary byte string). Such transactions must be tolerated in accordance with [EIP-7805](https://eips.ethereum.org/EIPS/eip-7805).

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

    2. `payloadAttributes.timestamp` falls within the time frame of the Bogota fork, return `-38005: Unsupported fork` on failure.

2. Extend point (9) of the `engine_forkchoiceUpdatedV1` [specification](./paris.md#specification-1) by defining `payloadStatus.inclusionListSatisfied`:

    1. If the payload referenced by `forkchoiceState.headBlockHash` is deemed `VALID`, `payloadStatus.inclusionListSatisfied` **MUST** be set to whether the payload satisfied the inclusion list constraints.

    2. Client software **MUST** use the retained `inclusionListTransactions` if it validates the payload and checks whether the payload satisfies the inclusion list constraints while processing the call.

    3. Otherwise, `payloadStatus.inclusionListSatisfied` **MUST** be `null`.

### engine_newPayloadV7

This method is updated to support the new `ExecutionPayloadV5` structure.

#### Request

* method: `engine_newPayloadV7`
* params:
  1. `executionPayload`: [`ExecutionPayloadV5`](#executionpayloadv5).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected blob versioned hashes to validate.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
  4. `executionRequests`: `Array of DATA` - List of execution layer triggered requests. Each list element is a `requests` byte array as defined by [EIP-7685](https://eips.ethereum.org/EIPS/eip-7685). The first byte of each element is the `request_type` and the remaining bytes are the `request_data`. Elements of the list **MUST** be ordered by `request_type` in ascending order. Elements with empty `request_data` **MUST** be excluded from the list.
  5. `inclusionListTransactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`).
* timeout: 6s

#### Response

* result: [`PayloadStatusV2`](#payloadstatusv2)
* error: code and message set in case an exception happens while processing the payload.

#### Specification

This method follows the same specification as [`engine_newPayloadV6`](#engine_newpayloadv6) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload does not fall within the time frame of the EIP-8141 fork.

2. Client software **MUST** return `-32602: Invalid params` error if the `dependencyProof` field is missing.

3. Client software **MUST** return `INVALID` if `dependencyProof` is longer than `MAX_DEPENDENCY_PROOF_BYTES`, and **SHOULD** do so before hashing it.

4. Client software **MUST** validate `dependencyProof` against the block validity rules of EIP-8141 and return `INVALID` if any dependency of the payload is not satisfied by its witness.

### engine_getPayloadV7

This method is updated to return the new `ExecutionPayloadV5` structure.

#### Request

* method: `engine_getPayloadV7`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process

#### Response

* result: `object`
  - `executionPayload`: [`ExecutionPayloadV5`](#executionpayloadv5)
  - `blockValue` : `QUANTITY`, 256 Bits - The expected value to be received by the `feeRecipient` in wei
  - `blobsBundle`: [`BlobsBundleV2`](./osaka.md#blobsbundlev2) - Bundle with data corresponding to blob transactions included into `executionPayload`
  - `shouldOverrideBuilder` : `BOOLEAN` - Suggestion from the execution layer to use this `executionPayload` instead of an externally provided one
  - `executionRequests`: `Array of DATA` - Execution layer triggered requests obtained from the `executionPayload` transaction execution.
* error: code and message set in case an exception happens while getting the payload.

#### Specification

This method follows the same specification as [`engine_getPayloadV6`](./amsterdam.md#engine_getpayloadv6) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the built payload does not fall within the time frame of the EIP-8141 fork.

2. Client software **MUST** set `dependencyProof` to the `dependency_proof` of the built payload, whose hash the payload's header commits to.

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
- [`engine_getPayloadV6`](./amsterdam.md#engine_getpayloadv6)

a validation **MUST** be added:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of payload is greater than or equal to the EIP-8141 activation timestamp.
