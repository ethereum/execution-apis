# Engine API -- Prague

Engine API changes introduced in Prague.

This specification is based on and extends [Engine API - Cancun](./cancun.md) specification.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Structures](#structures)
  - [DepositRequestV1](#depositrequestv1)
  - [WithdrawalRequestV1](#withdrawalrequestv1)
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
- `validatorPublicKey`: `DATA`, 48 Bytes
- `amount`: `QUANTITY`, 64 Bits

*Note:* The `amount` value is represented in Gwei.

## Methods

### engine_newPayloadV4

The request of this method is updated with [`ExecutionPayloadV3`](./cancun.md#ExecutionPayloadV4).

#### Request

* method: `engine_newPayloadV4`
* params:
  1. `executionPayload`: [`ExecutionPayloadV4`](#ExecutionPayloadV4).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected blob versioned hashes to validate.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
  4. `expectedDepositRequests`: `Array of DepositRequestV1` - Array of expected deposit requests to validate.
  5. `expectedWithdrawalRequests`: `Array of WithdrawalRequestV1` - Array of expected withdrawal requests to validate.

#### Response

Refer to the response for [`engine_newPayloadV3`](./cancun.md#engine_newpayloadv3).

#### Specification

This method follows the same specification as [`engine_newPayloadV3`](./cancun.md#engine_newpayloadv3) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload does not fall within the time frame of the Prague fork.

2. Given the expected array of deposit requests, client software **MUST** run its validation by taking the following steps:
    1. Obtain the actual deposit requests array as it is specified by the [EIP-6110](https://eips.ethereum.org/EIPS/eip-6110).
    2. Return `{status: INVALID, latestValidHash: validHash, validationError: errorMessage | null}` if the expected and the actual arrays don't match.

3. Given the expected array of withdrawal requests, client software **MUST** run its validation by taking the following steps:
    1. Obtain the actual withdrawal requests array as it is specified by the [EIP-7002](https://eips.ethereum.org/EIPS/eip-7002).
    2. Return `{status: INVALID, latestValidHash: validHash, validationError: errorMessage | null}` if the expected and the actual arrays don't match.

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
  - `blobsBundle`: [`BlobsBundleV1`](./cancun.md#BlobsBundleV1) - Bundle with data corresponding to blob transactions included into `executionPayload`
  - `shouldOverrideBuilder` : `BOOLEAN` - Suggestion from the execution layer to use this `executionPayload` instead of an externally provided one
  - `depositRequests`: `Array of DepositRequestV1` - Array of deposits, each object is an `OBJECT` containing the fields of a `DepositRequestV1` structure.
  - `withdrawalRequests`: `Array of WithdrawalRequestV1` - Array of withdrawal requests, each object is an `OBJECT` containing the fields of a `WithdrawalRequestV1` structure.
* error: code and message set in case an exception happens while getting the payload.

#### Specification

This method follows the same specification as [`engine_getPayloadV3`](./cancun.md#engine_getpayloadv3) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the built payload does not fall within the time frame of the Prague fork.

2. Client software **MUST** return `depositRequests` array obtained from the block execution according to the [EIP-6110](https://eips.ethereum.org/EIPS/eip-6110) specification.

3. Client software **MUST** return `withdrawalRequests` array obtained from the block execution according to the [EIP-7002](https://eips.ethereum.org/EIPS/eip-7002) specification.

### Update the methods of previous forks

This document defines how Prague payload should be handled by the [`Cancun API`](./cancun.md).

For the following methods:

- [`engine_newPayloadV3`](./cancun.md#engine_newpayloadV3)
- [`engine_getPayloadV3`](./cancun.md#engine_getpayloadv3)

a validation **MUST** be added:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of payload or payloadAttributes greater or equal to the Prague activation timestamp.
