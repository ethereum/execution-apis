# Engine API -- Shanghai

Engine API changes introduced in Shanghai.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Structures](#structures)
  - [WithdrawalV1](#withdrawalv1)
  - [ExecutionPayloadV2](#executionpayloadv2)
  - [ExecutionPayloadBodyV1](#executionpayloadbodyv1)
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
  - [engine_getPayloadBodiesByHashV1](#engine_getpayloadbodiesbyhashv1)
    - [Request](#request-3)
    - [Response](#response-3)
    - [Specification](#specification-3)
  - [engine_getPayloadBodiesByRangeV1](#engine_getpayloadbodiesbyrangev1)
    - [Request](#request-4)
    - [Response](#response-4)
    - [Specification](#specification-4)

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

### ExecutionPayloadBodyV1
This structure contains a body of an execution payload. The fields are encoded as follows:
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

#### Response

Refer to the response for [`engine_newPayloadV1`](./paris.md#engine_newpayloadv1).

#### Specification

This method follows the same specification as [`engine_newPayloadV1`](./paris.md#engine_newpayloadv1) with the exception of the following:

1. If withdrawal functionality is activated, client software **MUST** return an `INVALID` status with the appropriate `latestValidHash` if `payload.withdrawals` is `null`.
   Similarly, if the functionality is not activated, client software **MUST** return an `INVALID` status with the appropriate `latestValidHash` if `payloadAttributes.withdrawals` is not `null`.
   Blocks without withdrawals **MUST** be expressed with an explicit empty list `[]` value.
   Refer to the validity conditions for [`engine_newPayloadV1`](./paris.md#engine_newpayloadv1) to specification of the appropriate `latestValidHash` value.

### engine_forkchoiceUpdatedV2

#### Request

* method: "engine_forkchoiceUpdatedV2"
* params:
  1. `forkchoiceState`: `Object` - instance of [`ForkchoiceStateV1`](./paris.md#ForkchoiceStateV1)
  2. `payloadAttributes`: `Object|null` - instance of [`PayloadAttributesV2`](#PayloadAttributesV2) or `null`

#### Response

Refer to the response for [`engine_forkchoiceUpdatedV1`](./paris.md#engine_forkchoiceupdatedv1).

#### Specification

This method follows the same specification as [`engine_forkchoiceUpdatedV1`](./paris.md#engine_forkchoiceupdatedv1) with the exception of the following:

1. If withdrawal functionality is activated, client software **MUST** return error `-38003: Invalid payload attributes` if `payloadAttributes.withdrawals` is `null`.
   Similarly, if the functionality is not activated, client software **MUST** return error `-38003: Invalid payload attributes` if `payloadAttributes.withdrawals` is not `null`.
   Blocks without withdrawals **MUST** be expressed with an explicit empty list `[]` value.

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

### engine_getPayloadBodiesByHashV1

#### Request

* method: `engine_getPayloadBodiesByHashV1`
* params:
  1. `Array of DATA`, 32 Bytes - Array of `block_hash` field values of the `ExecutionPayload` structure
* timeout: 10s

#### Response

* result: `Array of ExecutionPayloadBodyV1` - Array of [`ExecutionPayloadBodyV1`](#ExecutionPayloadBodyV1) objects.
* error: code and message set in case an exception happens while processing the method call.

#### Specification

1. Given array of block hashes client software **MUST** respond with array of `ExecutionPayloadBodyV1` objects with the corresponding hashes respecting the order of block hashes in the input array.

1. Client software **MUST** place responses in the order given in the request, using `null` for any missing blocks. For instance, if the request is `[A.block_hash, B.block_hash, C.block_hash]` and client software has data of payloads `A` and `C`, but doesn't have data of `B`, the response **MUST** be `[A.body, null, C.body]`.

1. Client software **MAY NOT** respond to requests for finalized blocks by hash.

1. Client software **MUST** set `withdrawals` field to `null` for bodies of pre-Shanghai blocks.

1. This request maps to [`BeaconBlocksByRoot`](https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/p2p-interface.md#beaconblocksbyroot) in the consensus layer `p2p` specification. Callers must be careful to use the execution block hash, instead of the beacon block root.

### engine_getPayloadBodiesByRangeV1

#### Request

* method: `engine_getPayloadBodiesByRangeV1`
* params:
  1. `start`: `QUANTITY`, 64 bits - Starting block number
  1. `count`: `QUANITTY`, 64 bits - Number of blocks to return
* timeout: 10s

#### Response

* result: `Array of ExecutionPayloadBodyV1` - Array of [`ExecutionPayloadBodyV1`](#ExecutionPayloadBodyV1) objects.
* error: code and message set in case an exception happens while processing the method call.

#### Specification

1. Given a `start` and a `count`, the client software **MUST** respond with array of `ExecutionPayloadBodyV1` objects with the corresponding execution block number respecting the order of blocks in the canonical chain, as selected by the latest `engine_forkchoiceUpdated` call.

1. Client software **MUST** support `count` values of at least 32 blocks.

1. Client software **MUST**:
    * Place `null` in the response array for pruned blocks.
    * Omit `null` values where the request extends past the current latest known block, i.e. no trailing `nulls` are allowed in this case.
    * Return array of `null` values if all blocks from requested range are pruned and the range is behind the latest known block.

1. Client software **MUST** set `withdrawals` field to `null` for bodies of pre-Shanghai blocks.

1. This request maps to [`BeaconBlocksByRange`](https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/p2p-interface.md#beaconblocksbyrange) in the consensus layer `p2p` specification.

1. Callers must be careful to not confuse `start` with a slot number, instead mapping the slot to a block number. Callers must also be careful to request non-finalized blocks by hash in order to avoid race conditions around the current view of the canonical chain.

1. Callers must be careful to verify the hash of the received blocks when requesting non-finalized parts of the chain since the response is prone to being re-orged.
