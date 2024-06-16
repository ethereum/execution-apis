# Engine API -- Prague

Engine API changes introduced in Prague.

This specification is based on and extends [Engine API - Cancun](./cancun.md) specification.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Structures](#structures)
  - [DepositRequestV1](#depositrequestv1)
  - [WithdrawalRequestV1](#withdrawalrequestv1)
  - [ConsolidationRequestV1](#consolidationrequestv1)
  - [ExecutionPayloadV4](#executionpayloadv4)
  - [ExecutionPayloadBodyV2](#executionpayloadbodyv2)
- [Methods](#methods)
  - [engine_newPayloadV4](#engine_newpayloadv4)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)
  - [engine_getPayloadV4](#engine_getpayloadv4)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)
  - [engine_getPayloadBodiesByHashV2](#engine_getpayloadbodiesbyhashv2)
    - [Request](#request-2)
    - [Response](#response-2)
    - [Specification](#specification-2)
  - [engine_getPayloadBodiesByRangeV2](#engine_getpayloadbodiesbyrangev2)
    - [Request](#request-3)
    - [Response](#response-3)
    - [Specification](#specification-3)
  - [Update the methods of previous forks](#update-the-methods-of-previous-forks)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Structures

### DepositRequestV1
This structure maps onto the deposit object from [EIP-6110](https://eips.ethereum.org/EIPS/eip-6110).
The fields are encoded as follows:

- `pubkey`: `DATA`, 48 Bytes
- `withdrawalCredentials`: `DATA`, 32 Bytes
- `amount`: `QUANTITY`, 64 Bits
- `signature`: `DATA`, 96 Bytes
- `index`: `QUANTITY`, 64 Bits

*Note:* The `amount` value is represented in Gwei.

### WithdrawalRequestV1
This structure maps onto the withdrawal request from [EIP-7002](https://eips.ethereum.org/EIPS/eip-7002).
The fields are encoded as follows:

- `sourceAddress`: `DATA`, 20 Bytes
- `validatorPubkey`: `DATA`, 48 Bytes
- `amount`: `QUANTITY`, 64 Bits

*Note:* The `amount` value is represented in Gwei.

### ConsolidationRequestV1
This structure maps onto the consolidation request from [EIP-7251](https://eips.ethereum.org/EIPS/eip-7251).
The fields are encoded as follows:

- `sourceAddress`: `DATA`, 20 Bytes
- `sourcePubkey`: `DATA`, 48 Bytes
- `targetPubkey`: `DATA`, 48 Bytes

### ExecutionPayloadV4

This structure has the syntax of [`ExecutionPayloadV3`](./cancun.md#executionpayloadv3) and appends the new fields: `depositRequests` and `withdrawalRequests`.

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
- `depositRequests`: `Array of DepositRequestV1` - Array of deposits, each object is an `OBJECT` containing the fields of a `DepositRequestV1` structure.
- `withdrawalRequests`: `Array of WithdrawalRequestV1` - Array of withdrawal requests, each object is an `OBJECT` containing the fields of a `WithdrawalRequestV1` structure.
- `consolidationRequests`: `Array of ConsolidationRequestV1` - Array of consolidation requests, each object is an `OBJECT` containing the fields of a `ConsolidationRequestV1` structure.

### ExecutionPayloadBodyV2

This structure has the syntax of [`ExecutionPayloadBodyV1`](./shanghai.md#executionpayloadv1) and appends the new fields: `depositRequests` and `withdrawalRequests`.

- `transactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718)
- `withdrawals`: `Array of WithdrawalV1` - Array of withdrawals, each object is an `OBJECT` containing the fields of a `WithdrawalV1` structure.
- `depositRequests`: `Array of DepositRequestV1` - Array of deposits, each object is an `OBJECT` containing the fields of a `DepositRequestV1` structure.
- `withdrawalRequests`: `Array of WithdrawalRequestV1` - Array of withdrawal requests, each object is an `OBJECT` containing the fields of a `WithdrawalRequestV1` structure.
- `consolidationRequests`: `Array of ConsolidationRequestV1` - Array of consolidation requests, each object is an `OBJECT` containing the fields of a `ConsolidationRequestV1` structure.

## Methods

### engine_newPayloadV4

The request of this method is updated with [`ExecutionPayloadV4`](#ExecutionPayloadV4).

#### Request

* method: `engine_newPayloadV4`
* params:
  1. `executionPayload`: [`ExecutionPayloadV4`](#ExecutionPayloadV4).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected blob versioned hashes to validate.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.

#### Response

Refer to the response for [`engine_newPayloadV3`](./cancun.md#engine_newpayloadv3).

#### Specification

This method follows the same specification as [`engine_newPayloadV3`](./cancun.md#engine_newpayloadv3) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload does not fall within the time frame of the Prague fork.

### engine_getPayloadV4

The response of this method is updated with [`ExecutionPayloadV4`](#ExecutionPayloadV4).

#### Request

* method: `engine_getPayloadV4`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process
* timeout: 1s

#### Response

* result: `object`
  - `executionPayload`: [`ExecutionPayloadV4`](#ExecutionPayloadV4)
  - `blockValue` : `QUANTITY`, 256 Bits - The expected value to be received by the `feeRecipient` in wei
  - `blobsBundle`: [`BlobsBundleV1`](#BlobsBundleV1) - Bundle with data corresponding to blob transactions included into `executionPayload`
  - `shouldOverrideBuilder` : `BOOLEAN` - Suggestion from the execution layer to use this `executionPayload` instead of an externally provided one
* error: code and message set in case an exception happens while getting the payload.

#### Specification

This method follows the same specification as [`engine_getPayloadV3`](./cancun.md#engine_getpayloadv3) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the built payload does not fall within the time frame of the Prague fork.

### engine_getPayloadBodiesByHashV2

The response of this method is updated with [`ExecutionPayloadBodyV2`](#executionpayloadbodyv2).

#### Request

* method: `engine_getPayloadBodiesByHashV2`
* params:
  1. `Array of DATA`, 32 Bytes - Array of `block_hash` field values of the `ExecutionPayload` structure
* timeout: 10s

#### Response

* result: `Array of ExecutionPayloadBodyV2` - Array of [`ExecutionPayloadBodyV2`](#executionpayloadbodyv2) objects.
* error: code and message set in case an exception happens while processing the method call.

#### Specification

This method follows the same specification as [`engine_getPayloadBodiesByHashV1`](./shanghai.md#engine_getpayloadbodiesbyhashv1) with the addition of the following:

1. Client software **MUST** set `depositRequests` and `withdrawalRequests` fields to `null` for bodies of pre-Prague blocks.

### engine_getPayloadBodiesByRangeV2

The response of this method is updated with [`ExecutionPayloadBodyV2`](#executionpayloadbodyv2).

#### Request

* method: `engine_getPayloadBodiesByRangeV2`
* params:
  1. `start`: `QUANTITY`, 64 bits - Starting block number
  1. `count`: `QUANITTY`, 64 bits - Number of blocks to return
* timeout: 10s

#### Response

* result: `Array of ExecutionPayloadBodyV2` - Array of [`ExecutionPayloadBodyV2`](#executionpayloadbodyv2) objects.
* error: code and message set in case an exception happens while processing the method call.

#### Specification

This method follows the same specification as [`engine_getPayloadBodiesByRangeV2`](./shanghai.md#engine_getpayloadbodiesbyrangev1) with the addition of the following:

1. Client software **MUST** set `depositRequests` and `withdrawalRequests` fields to `null` for bodies of pre-Prague blocks.

### Update the methods of previous forks

This document defines how Prague payload should be handled by the [`Cancun API`](./cancun.md).

For the following methods:

- [`engine_newPayloadV3`](./cancun.md#engine_newpayloadV3)
- [`engine_getPayloadV3`](./cancun.md#engine_getpayloadv3)

a validation **MUST** be added:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of payload or payloadAttributes greater or equal to the Prague activation timestamp.
