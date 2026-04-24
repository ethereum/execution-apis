# Engine API -- Bogota

Engine API changes introduced in Bogota.

This specification is based on and extends [Engine API - Amsterdam](./amsterdam.md) specification.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Constants](#constants)
- [Structures](#structures)
  - [PayloadAttributesV5](#payloadattributesv5)
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
  - [Update the methods of previous forks](#update-the-methods-of-previous-forks)
    - [Amsterdam API](#amsterdam-api)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Constants

| Name | Value |
| - | - |
| `MAX_BYTES_PER_INCLUSION_LIST` |  `uint64(8192) = 2**13` |

## Structures

### PayloadAttributesV5

This structure has the syntax of [`PayloadAttributesV4`](./amsterdam.md#payloadattributesv4) and appends a single field: `inclusionListTransactions`.

- `timestamp`: `QUANTITY`, 64 Bits - value for the `timestamp` field of the new payload
- `prevRandao`: `DATA`, 32 Bytes - value for the `prevRandao` field of the new payload
- `suggestedFeeRecipient`: `DATA`, 20 Bytes - suggested value for the `feeRecipient` field of the new payload
- `withdrawals`: `Array of WithdrawalV1` - Array of withdrawals, each object is an `OBJECT` containing the fields of a `WithdrawalV1` structure.
- `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
- `slotNumber`: `QUANTITY`, 64 Bits - value for the `slotNumber` field of the new payload
- `inclusionListTransactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718).

## Routines

### Payload building

This routine follows the same specification as [Payload building](./paris.md#payload-building) with the following changes to the processing flow:

1. Client software **SHOULD** take `inclusionListTransactions` into account during the payload build process. The built `ExecutionPayload` **MUST** satisfy the inclusion list constraints with respect to `inclusionListTransactions` as defined in [EIP-7805](https://eips.ethereum.org/EIPS/eip-7805).

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

#### Response

* result: [`PayloadStatusV1`](./paris.md#payloadstatusv1), values of the `status` field are modified in the following way:
  - `INVALID_BLOCK_HASH` status value is supplanted by `INVALID`.
  - `INCLUSION_LIST_UNSATISFIED` status value is appended.
* error: code and message set in case an exception happens while processing the payload.

#### Specification

This method follows the same specification as [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload does not fall within the time frame of the Bogota fork.

2. Client software **MUST** return `{status: INCLUSION_LIST_UNSATISFIED, latestValidHash: null, validationError: null}` if `executionPayload` fails to satisfy the inclusion list constraints with respect to `inclusionListTransactions` as defined in [EIP-7805](https://eips.ethereum.org/EIPS/eip-7805).

### engine_getInclusionListV1

#### Request

* method: `engine_getInclusionListV1`
* params:
  1. `parentHash`: `DATA`, 32 Bytes - block hash of the parent block upon which the inclusion list should be built.

* timeout: 1s

#### Response

* result: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718).
* error: code and message set in case an exception happens while getting the inclusion list.

#### Specification

1. Client software **MUST** return `-38007: Unknown parent` error if a block with the given `parentHash` does not exist.

2. Client software **MUST** provide a list of transactions for the inclusion list based on the local view of the mempool. The strategy for selecting which transactions to include is implementation dependent.

3. Client software **MUST** ensure the byte length of the RLP encoding of the returned transaction list does not exceed `MAX_BYTES_PER_INCLUSION_LIST`.

4. Client software **MUST NOT** include any [blob transaction](https://eips.ethereum.org/EIPS/eip-4844#blob-transaction) in the returned transaction list.
 
### engine_forkchoiceUpdatedV5

#### Request

* method: `engine_forkchoiceUpdatedV5`
* params:
  1. `forkchoiceState`: [`ForkchoiceStateV1`](./paris.md#forkchoicestatev1).
  2. `payloadAttributes`: `Object|null` - Instance of [`PayloadAttributesV5`](#payloadattributesv5) or `null`.
* timeout: 8s

#### Response

Refer to the response for [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4).

#### Specification

This method follows the same specification as [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4) with the following changes to the processing flow:

1. Extend point (8) of the `engine_forkchoiceUpdatedV1` [specification](./paris.md#specification-1) by defining the following sequence of checks that **MUST** be run over `payloadAttributes`:

    1. `payloadAttributes` matches the [`PayloadAttributesV5`](#payloadattributesv5) structure, return `-38003: Invalid payload attributes` on failure.

    2. `payloadAttributes.timestamp` does not fall within the time frame of the Bogota fork, return `-38005: Unsupported fork` on failure.

    3. `payloadAttributes.timestamp` is greater than `timestamp` of a block referenced by `forkchoiceState.headBlockHash`, return `-38003: Invalid payload attributes` on failure.

    4. If any of the above checks fails, the `forkchoiceState` update **MUST NOT** be rolled back.

### Update the methods of previous forks

#### Amsterdam API

For the following methods:

- [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5)
- [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4)

a validation **MUST** be added:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of payload is greater than or equal to the Bogota activation timestamp.
