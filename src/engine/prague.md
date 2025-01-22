# Engine API -- Prague

Engine API changes introduced in Prague.

This specification is based on and extends [Engine API - Cancun](./cancun.md) specification.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

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

## Methods

### engine_newPayloadV4

Method parameter list is extended with `executionRequests`.

#### Request

* method: `engine_newPayloadV4`
* params:
  1. `executionPayload`: [`ExecutionPayloadV3`](./cancun.md#executionpayloadv3).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected blob versioned hashes to validate.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
  4. `executionRequests`: `Array of DATA` - List of execution layer triggered requests. Each list element is a `requests` byte array as defined by [EIP-7685](https://eips.ethereum.org/EIPS/eip-7685). The first byte of each element is the `request_type` and the remaining bytes are the `request_data`. Elements of the list **MUST** be ordered by `request_type` in ascending order. Elements with empty `request_data` **MUST** be excluded from the list. If any element is out of order, has a length of 1-byte or shorter, or more than one element has the same type byte, client software **MUST** return `-32602: Invalid params` error.

#### Response

Refer to the response for [`engine_newPayloadV3`](./cancun.md#engine_newpayloadv3).

#### Specification

This method follows the same specification as [`engine_newPayloadV3`](./cancun.md#engine_newpayloadv3) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload does not fall within the time frame of the Prague fork.

2. Given the `executionRequests`, client software **MUST** compute the execution requests commitment
and incorporate it into the `blockHash` validation process.
That is, if the computed commitment does not match the corresponding commitment in the execution layer block header,
the call **MUST** return `{status: INVALID, latestValidHash: null, validationError: errorMessage | null}`.

### engine_getPayloadV4

The response of this method is extended with the `executionRequests` field.

#### Request

* method: `engine_getPayloadV4`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process
* timeout: 1s

#### Response

* result: `object`
  - `executionPayload`: [`ExecutionPayloadV3`](./cancun.md#executionpayloadv3)
  - `blockValue` : `QUANTITY`, 256 Bits - The expected value to be received by the `feeRecipient` in wei
  - `blobsBundle`: [`BlobsBundleV1`](#BlobsBundleV1) - Bundle with data corresponding to blob transactions included into `executionPayload`
  - `shouldOverrideBuilder` : `BOOLEAN` - Suggestion from the execution layer to use this `executionPayload` instead of an externally provided one
  - `executionRequests`: `Array of DATA` - Execution layer triggered requests obtained from the `executionPayload` transaction execution.
* error: code and message set in case an exception happens while getting the payload.

#### Specification

This method follows the same specification as [`engine_getPayloadV3`](./cancun.md#engine_getpayloadv3) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the built payload does not fall within the time frame of the Prague fork.

2. The call **MUST** return `executionRequests` list representing execution layer triggered requests. Each list element is a `requests` byte array as defined by [EIP-7685](https://eips.ethereum.org/EIPS/eip-7685). The first byte of each element is the `request_type` and the remaining bytes are the `request_data`. Elements of the list **MUST** be ordered by `request_type` in ascending order. Elements with empty `request_data` **MUST** be excluded from the list. Elements **MUST** be longer than 1-byte, and each `request_type` **MUST** be unique within the list.

### Update the methods of previous forks

This document defines how Prague payload should be handled by the [`Cancun API`](./cancun.md).

For the following methods:

- [`engine_newPayloadV3`](./cancun.md#engine_newpayloadV3)
- [`engine_getPayloadV3`](./cancun.md#engine_getpayloadv3)

a validation **MUST** be added:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of payload or payloadAttributes greater or equal to the Prague activation timestamp.
