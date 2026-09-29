# Engine API -- EIP-XXXX

Engine API changes introduced in EIP-XXXX.

This specification is based on and extends [Engine API - Bogota](./bogota.md) specification.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Constants](#constants)
- [Structures](#structures)
  - [InclusionListClaimV1](#inclusionlistclaimv1)
- [Methods](#methods)
  - [engine_newPayloadV7](#engine_newpayloadv7)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)
  - [engine_getPayloadV7](#engine_getpayloadv7)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)
  - [Update the methods of previous forks](#update-the-methods-of-previous-forks)
    - [Bogota API](#bogota-api)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Constants

| Name | Value |
| - | - |
| `INCLUSION_LIST_COMMITTEE_SIZE` | `uint64(16) = 2**4` |
| `MAX_INCLUSION_LIST_CLAIMS` | `uint64(1024) = 2**10` |

## Structures

### InclusionListClaimV1

This structure maps onto the inclusion list claim defined in EIP-XXXX. The fields are encoded as follows:

- `transactionHash`: `DATA`, 32 Bytes - `keccak256` of the claimed inclusion list transaction, `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718).
- `transactionIndex`: `QUANTITY`, 64 Bits - index in `executionPayload.transactions` at which the omission of the claimed transaction is evaluated.

## Methods

### engine_newPayloadV7

Method parameter list is extended with `inclusionListClaims`, and inclusion list transactions are grouped per inclusion list.

#### Request

* method: `engine_newPayloadV7`
* params:
  1. `executionPayload`: [`ExecutionPayloadV4`](./amsterdam.md#executionpayloadv4).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected blob versioned hashes to validate.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
  4. `executionRequests`: `Array of DATA` - List of execution layer triggered requests. Each list element is a `requests` byte array as defined by [EIP-7685](https://eips.ethereum.org/EIPS/eip-7685). The first byte of each element is the `request_type` and the remaining bytes are the `request_data`. Elements of the list **MUST** be ordered by `request_type` in ascending order. Elements with empty `request_data` **MUST** be excluded from the list.
  5. `inclusionLists`: `Array of Array of DATA` - Array of inclusion lists, one per inclusion list committee member, each an array of transaction objects in the order of that inclusion list. Each transaction object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718).
  6. `inclusionListClaims`: `Array of InclusionListClaimV1` - Array of inclusion list claims committed to by the builder of the payload.
* timeout: 6s

#### Response

* result: [`PayloadStatusV2`](./bogota.md#payloadstatusv2)
* error: code and message set in case an exception happens while processing the payload.

#### Specification

This method follows the same specification as [`engine_newPayloadV6`](./bogota.md#engine_newpayloadv6) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload is less than the EIP-XXXX activation timestamp.

2. Client software **MUST** return `-32602: Invalid params` error if any of the following holds:

    1. `inclusionLists` has more than `INCLUSION_LIST_COMMITTEE_SIZE` elements.

    2. The sum of the byte lengths of the transactions in an element of `inclusionLists` exceeds `MAX_TRANSACTIONS_BYTES_PER_INCLUSION_LIST`. The total byte length of all inclusion list transactions is therefore at most `INCLUSION_LIST_COMMITTEE_SIZE * MAX_TRANSACTIONS_BYTES_PER_INCLUSION_LIST`.

    3. `inclusionListClaims` has more than `MAX_INCLUSION_LIST_CLAIMS` elements.

3. Client software **MUST** determine whether the payload satisfied the inclusion list constraints with respect to `inclusionLists` and `inclusionListClaims` as defined in EIP-XXXX. Duplicate, unmatched and out-of-range claims **MUST NOT** cause an error; they are resolved as defined in EIP-XXXX.

4. Client software **MUST** retain `inclusionLists` and `inclusionListClaims` for a payload with `ACCEPTED` status. Client software **MAY** discard them once the payload is no longer the tip of a branch.

### engine_getPayloadV7

This method is updated to return `inclusionListClaims`.

#### Request

* method: `engine_getPayloadV7`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process
* timeout: 1s

#### Response

* result: `object`
  - `executionPayload`: [`ExecutionPayloadV4`](./amsterdam.md#executionpayloadv4)
  - `blockValue` : `QUANTITY`, 256 Bits - The expected value to be received by the `feeRecipient` in wei
  - `blobsBundle`: [`BlobsBundleV2`](./osaka.md#blobsbundlev2) - Bundle with data corresponding to blob transactions included into `executionPayload`
  - `shouldOverrideBuilder` : `BOOLEAN` - Suggestion from the execution layer to use this `executionPayload` instead of an externally provided one
  - `executionRequests`: `Array of DATA` - Execution layer triggered requests obtained from the `executionPayload` transaction execution.
  - `inclusionListClaims`: `Array of InclusionListClaimV1` - Inclusion list claims for transactions omitted from `executionPayload`.
* error: code and message set in case an exception happens while getting the payload.

#### Specification

This method follows the same specification as [`engine_getPayloadV6`](./amsterdam.md#engine_getpayloadv6) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the built payload is less than the EIP-XXXX activation timestamp.

2. `inclusionListClaims` **MUST NOT** have more than `MAX_INCLUSION_LIST_CLAIMS` elements. Each claim **MUST** refer to a transaction of the `inclusionListTransactions` used to build the payload that is omitted from `executionPayload`, and **SHOULD** have an index at which that omission is excused as defined in EIP-XXXX.

### Update the methods of previous forks

#### Bogota API

For the following methods:

- [`engine_newPayloadV6`](./bogota.md#engine_newpayloadv6)
- [`engine_getPayloadV6`](./amsterdam.md#engine_getpayloadv6)

a validation **MUST** be added:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload is greater than or equal to the EIP-XXXX activation timestamp.

For [`engine_forkchoiceUpdatedV5`](./bogota.md#engine_forkchoiceupdatedv5), if the payload referenced by `forkchoiceState.headBlockHash` was received with [`engine_newPayloadV7`](#engine_newpayloadv7), client software **MUST** use the retained `inclusionLists` and `inclusionListClaims` if it checks whether the payload satisfies the inclusion list constraints while processing the call.
