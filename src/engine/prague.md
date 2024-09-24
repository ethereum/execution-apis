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
  - [ExecutionRequestsV1](#executionrequestsv1)
- [Methods](#methods)
  - [engine_newPayloadV4](#engine_newpayloadv4)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)
  - [engine_getPayloadV4](#engine_getpayloadv4)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)
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

### ExecutionRequestsV1

This container holds execution layer triggered requests.

- `deposits`: `Array of DepositRequestV1` - Array of deposits, each object is an `OBJECT` containing the fields of a `DepositRequestV1` structure.
- `withdrawals`: `Array of WithdrawalRequestV1` - Array of withdrawal requests, each object is an `OBJECT` containing the fields of a `WithdrawalRequestV1` structure.
- `consolidations`: `Array of ConsolidationRequestV1` - Array of consolidation requests, each object is an `OBJECT` containing the fields of a `ConsolidationRequestV1` structure.

*Note*: The order of items within `deposits`, `withdrawals` and `consolidations` lists is defined by
[EIP-6110](https://eips.ethereum.org/EIPS/eip-6110), [EIP-7002](https://eips.ethereum.org/EIPS/eip-7002) and [EIP-7251](https://eips.ethereum.org/EIPS/eip-7251) respectively.

## Methods

### engine_newPayloadV4

Method parameter list is extended with `executionRequests`.

#### Request

* method: `engine_newPayloadV4`
* params:
  1. `executionPayload`: [`ExecutionPayloadV3`](./cancun.md#executionpayloadv3).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected blob versioned hashes to validate.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
  4. `executionRequests`: [`ExecutionRequestsV1`](#ExecutionRequestsV1)

#### Response

Refer to the response for [`engine_newPayloadV3`](./cancun.md#engine_newpayloadv3).

#### Specification

This method follows the same specification as [`engine_newPayloadV3`](./cancun.md#engine_newpayloadv3) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload does not fall within the time frame of the Prague fork.

2. Client software **MUST** incorporate `executionRequests` into the `blockHash` validation process.
   That is, if `executionRequests` does not match the execution requests commitment in the execution layer block header
   the call **MUST** be responded with `{status: INVALID, latestValidHash: null, validationError: errorMessage | null}`.

### engine_getPayloadV4

The response of this method is updated with [`ExecutionPayloadV4`](#ExecutionPayloadV4) and new [`ExecutionRequestsV1`](#ExecutionRequestsV1).

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
  - `executionRequests`: [`ExecutionRequestsV1`](#ExecutionRequestsV1) - Execution layer trigerred requests obtained from the `executionPayload` transaction execution
* error: code and message set in case an exception happens while getting the payload.

#### Specification

This method follows the same specification as [`engine_getPayloadV3`](./cancun.md#engine_getpayloadv3) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the built payload does not fall within the time frame of the Prague fork.

2. The call **MUST** return `executionRequests` object containing deposit, withdrawal and consolidation requests obtained from transaction execution of the `executionPayload`.
   The way the requests are obtained from the payload execution is defined by the [EIP-6110](https://eips.ethereum.org/EIPS/eip-6110),
   [EIP-7002](https://eips.ethereum.org/EIPS/eip-7002) and [EIP-7251](https://eips.ethereum.org/EIPS/eip-7251) respectively.

### Update the methods of previous forks

This document defines how Prague payload should be handled by the [`Cancun API`](./cancun.md).

For the following methods:

- [`engine_newPayloadV3`](./cancun.md#engine_newpayloadV3)
- [`engine_getPayloadV3`](./cancun.md#engine_getpayloadv3)

a validation **MUST** be added:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of payload or payloadAttributes greater or equal to the Prague activation timestamp.
