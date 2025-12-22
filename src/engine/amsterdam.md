# Engine API -- Amsterdam

Engine API changes introduced in Amsterdam.

This specification is based on and extends [Engine API - Osaka](./osaka.md) specification.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Structures](#structures)
  - [ExecutionPayloadV4](#executionpayloadv4)
- [Methods](#methods)
  - [engine_newPayloadV5](#engine_newpayloadv5)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)
  - [engine_getPayloadV6](#engine_getpayloadv6)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)
  - [engine_getBlockAccessListsByHashV1](#engine_getbalsbyhashv1)
    - [Request](#request-2)
    - [Response](#response-2)
    - [Specification](#specification-2)
  - [engine_getBlockAccessListsByRangeV1](#engine_getbalsbyrangev1)
    - [Request](#request-3)
    - [Response](#response-3)
    - [Specification](#specification-3)
  - [Update the methods of previous forks](#update-the-methods-of-previous-forks)
    - [Osaka API](#osaka-api)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Structures

### ExecutionPayloadV4

This structure has the syntax of [`ExecutionPayloadV3`](./cancun.md#executionpayloadv3) and appends the new field: `blockAccessList`.

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

## Methods

### engine_newPayloadV5

This method is updated to support the new `ExecutionPayloadV4` structure.

#### Request

* method: `engine_newPayloadV5`
* params:
  1. `executionPayload`: [`ExecutionPayloadV4`](#executionpayloadv4).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected blob versioned hashes to validate.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
  4. `executionRequests`: `Array of DATA` - List of execution layer triggered requests.

#### Response

Refer to the response for [`engine_newPayloadV4`](./prague.md#engine_newpayloadv4).

#### Specification

This method follows the same specification as [`engine_newPayloadV4`](./prague.md#engine_newpayloadv4) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload does not fall within the time frame of the Amsterdam activation.

2. Client software **MUST** return `-32602: Invalid params` error if the `blockAccessList` field is missing.

3. Client software **MUST** validate the `blockAccessList` field by executing the payload's transactions and verifying that the computed access list matches the provided one.
If this validation fails, the call **MUST** return `{status: INVALID, latestValidHash: null, validationError: errorMessage | null}`.

### engine_getPayloadV6

This method is updated to return the new `ExecutionPayloadV4` structure.

#### Request

* method: `engine_getPayloadV6`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process
* timeout: 1s

#### Response

* result: `object`
  - `executionPayload`: [`ExecutionPayloadV4`](#executionpayloadv4)
  - `blockValue` : `QUANTITY`, 256 Bits - The expected value to be received by the `feeRecipient` in wei
  - `blobsBundle`: [`BlobsBundleV2`](./osaka.md#blobsbundlev2) - Bundle with data corresponding to blob transactions included into `executionPayload`
  - `shouldOverrideBuilder` : `BOOLEAN` - Suggestion from the execution layer to use this `executionPayload` instead of an externally provided one
  - `executionRequests`: `Array of DATA` - Execution layer triggered requests obtained from the `executionPayload` transaction execution.
* error: code and message set in case an exception happens while getting the payload.

#### Specification

This method follows the same specification as [`engine_getPayloadV5`](./osaka.md#engine_getpayloadv5) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the built payload does not fall within the time frame of the Amsterdam activation.

2. When building the block, client software **MUST** collect all account accesses and state changes during transaction execution and populate the `blockAccessList` field in the returned `ExecutionPayloadV4` with the RLP-encoded access list.

### engine_getBlockAccessListsByHashV1

This method retrieves RLP-encoded block access lists for specified blocks.

#### Request

* method: `engine_getBlockAccessListsByHashV1`
* params:
  1. `blockHashes`: `Array of DATA`, 32 Bytes - Array of block hashes to retrieve block access lists for
* timeout: 10s

#### Response

* result: `Array` - Array of block access list body objects or `null` for blocks without block access lists
  - `blockAccessList`: `DATA` - RLP-encoded block access list as defined in EIP-7928, or `null` if unavailable
* error: code and message set in case an exception happens while getting the block access lists.

#### Specification

1. Client software **MUST** return an array of the same length as the input array.

2. Client software **MUST** place responses in the same order as the corresponding block hashes in the input array.

3. Client software **MUST** return `null` for any block hash that:
   - Is unknown or unavailable
   - Predates the Amsterdam fork activation
   - Has been pruned from storage

4. Client software **MUST** return `-38004: Too large request` error if the requested range is too large.

5. Client software **MUST** retain block access lists for at least 3533 epochs (the weak subjectivity period) to support synchronization with re-execution.

### engine_getBlockAccessListsByRangeV1

This method retrieves RLP-encoded block access lists for a range of blocks.

#### Request

* method: `engine_getBlockAccessListsByRangeV1`
* params:
  1. `start`: `QUANTITY`, 64 Bits - Starting block number
  2. `count`: `QUANTITY`, 64 Bits - Number of blocks to retrieve block access lists for
* timeout: 10s

#### Response

* result: `Array` - Array of block access list body objects or `null` for blocks without block access lists
  - `blockAccessList`: `DATA` - RLP-encoded block access list as defined in EIP-7928, or `null` if unavailable
* error: code and message set in case an exception happens while getting the block access lists.

#### Specification

1. Client software **MUST** return block access lists for the range `[start, start + count)`.

2. Client software **MUST** return an array of length equal to `count` or less if the range extends beyond the current head.

3. Client software **MUST** return `null` for any block in the range that:
   - Has not been processed yet
   - Predates the Amsterdam fork activation
   - Has been pruned from storage

4. Client software **MUST** return `-38004: Too large request` error if the requested range is too large.

5. Client software **MUST** retain block access lists for at least 3533 epochs (the weak subjectivity period) to support synchronization with re-execution.

### Update the methods of previous forks

#### Osaka API

For the following methods:

- [`engine_newPayloadV4`](./prague.md#engine_newpayloadv4)
- [`engine_getPayloadV5`](./osaka.md#engine_getpayloadv5)

a validation **MUST** be added:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of payload greater or equal to the Amsterdam activation timestamp.

For the [`engine_forkchoiceUpdatedV3`](./cancun.md#engine_forkchoiceupdatedv3) the following modification **MUST** be made:
1. Return `-38005: Unsupported fork` if `payloadAttributes.timestamp` doesn't fall within the time frame of the Cancun, Prague, Osaka *or Amsterdam* forks.
