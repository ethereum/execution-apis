# Engine API -- Shanghai

Engine API changes introduced in Shanghai.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Structures](#structures)
  - [WithdrawalV1](#withdrawalv1)
  - [ExecutionPayloadV2](#executionpayloadv2)
  - [PayloadAttributesV2](#payloadattributesv2)
- [Methods](#methods)
  - [engine_newPayloadV2](#engine_newpayloadv2)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)
  - [engine_forkchoiceUpdatedV2](#engine_forkchoiceupdatedv2)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)
  - [engine_getPayloadV2](#engine_getpayloadv2)
    - [Request](#request-2)
    - [Response](#response-2)
    - [Specification](#specification-2)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Structures

### WithdrawalV1

This structure maps onto the validator withdrawal object from the beacon chain spec.
The fields are encoded as follows:

- `index`: `QUANTITY`, 64 Bits
- `validatorIndex`: `QUANTITY`, 64 Bits
- `address`: `DATA`, 20 Bytes
- `amount`: `QUANTITY`, 256 Bits

*Note*: the `amount` value is represented on the beacon chain as a little-endian value in units of Gwei, whereas the `amount` in this structure *MUST* be converted to a big-endian value in units of Wei.

### ExecutionPayloadV2

This structure has the syntax of `ExecutionPayloadV1` and appends a single field: `withdrawals`.

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

### PayloadAttributesV2

This structure has the syntax of `PayloadAttributesV1` and appends a single field: `withdrawals`.

- `timestamp`: `QUANTITY`, 64 Bits - value for the `timestamp` field of the new payload
- `prevRandao`: `DATA`, 32 Bytes - value for the `prevRandao` field of the new payload
- `suggestedFeeRecipient`: `DATA`, 20 Bytes - suggested value for the `feeRecipient` field of the new payload
- `withdrawals`: `Array of WithdrawalV1` - Array of withdrawals, each object is an `OBJECT` containing the fields of a `WithdrawalV1` structure.

## Methods

### engine_newPayloadV2

#### Request

* method: `engine_newPayloadV2`
* params:
  1. [`ExecutionPayloadV2`](#ExecutionPayloadV2)
* timeout: 8s

#### Response

* result: [`PayloadStatusV1`](./paris.md#payloadstatusv1), values of the `status` field are restricted in the following way:
  - `INVALID_BLOCK_HASH` status value is supplanted by `INVALID`.
* error: code and message set in case an exception happens while processing the payload.

#### Specification

This method follows the same specification as [`engine_newPayloadV1`](./paris.md#engine_newpayloadv1) with the exception of the following:

1. If withdrawal functionality is activated, client software **MUST** return an `INVALID` status with the appropriate `latestValidHash` if `payload.withdrawals` is `null`.
   Similarly, if the functionality is not activated, client software **MUST** return an `INVALID` status with the appropriate `latestValidHash` if `payloadAttributes.withdrawals` is not `null`.
   Blocks without withdrawals **MUST** be expressed with an explicit empty list `[]` value.
   Refer to the validity conditions for [`engine_newPayloadV1`](./paris.md#engine_newpayloadv1) to specification of the appropriate `latestValidHash` value.
2. Client software **MAY NOT** validate terminal PoW block conditions during payload validation (point (2) in the [Payload validation](./paris.md#payload-validation) routine).
3. Client software **MUST** return `{status: INVALID, latestValidHash: null, validationError: errorMessage | null}` if the `blockHash` validation has failed.

### engine_forkchoiceUpdatedV2

#### Request

* method: "engine_forkchoiceUpdatedV2"
* params:
  1. `forkchoiceState`: `Object` - instance of [`ForkchoiceStateV1`](./paris.md#ForkchoiceStateV1)
  2. `payloadAttributes`: `Object|null` - instance of [`PayloadAttributesV2`](#PayloadAttributesV2) or `null`
* timeout: 8s

#### Response

Refer to the response for [`engine_forkchoiceUpdatedV1`](./paris.md#engine_forkchoiceupdatedv1).

#### Specification

This method follows the same specification as [`engine_forkchoiceUpdatedV1`](./paris.md#engine_forkchoiceupdatedv1) with the exception of the following:

1. If withdrawal functionality is activated, client software **MUST** return error `-38003: Invalid payload attributes` if `payloadAttributes.withdrawals` is `null`.
   Similarly, if the functionality is not activated, client software **MUST** return error `-38003: Invalid payload attributes` if `payloadAttributes.withdrawals` is not `null`.
   Blocks without withdrawals **MUST** be expressed with an explicit empty list `[]` value.
2. Client software **MAY NOT** validate terminal PoW block conditions in the following places:
  - during payload validation (point (2) in the [Payload validation](./paris.md#payload-validation) routine specification),
  - when updating the forkchoice state (point (3) in the [`engine_forkchoiceUpdatedV1`](./paris.md#engine_forkchoiceupdatedv1) method specification).

### engine_getPayloadV2

#### Request

* method: `engine_getPayloadV2`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process
* timeout: 1s

#### Response

* result: `object`
  - `executionPayload`: [`ExecutionPayloadV2`](#ExecutionPayloadV2)
  - `blockValue` : `QUANTITY`, 256 Bits - The expected value to be received by the `feeRecipient` in wei
* error: code and message set in case an exception happens while getting the payload.

#### Specification

This method follows the same specification as [`engine_getPayloadV1`](./paris.md#engine_getpayloadv1) with the addition of the following:

  1. Client software **SHOULD** use the sum of the block's priority fees or any other algorithm to determine `blockValue`.