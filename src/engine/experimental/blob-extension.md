<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Shard Blob Extension](#shard-blob-extension)
  - [Structures](#structures)
    - [ExecutionPayloadV3](#executionpayloadv3)
    - [BlobsBundleV1](#blobsbundlev1)
  - [Methods](#methods)
    - [engine_newPayloadV3](#engine_newpayloadv3)
      - [Request](#request)
      - [Specification](#specification)
      - [Response](#response)
    - [engine_getPayloadV3](#engine_getpayloadv3)
      - [Request](#request-1)
      - [Response](#response-1)
      - [Specification](#specification-1)
    - [engine_getBlobsBundleV1](#engine_getblobsbundlev1)
      - [Request](#request-2)
      - [Response](#response-2)
      - [Specification](#specification-2)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Shard Blob Extension

This is an extension specific to [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844) to the structures and methods as defined in the [Engine API - Paris](../paris.md) and [Engine API - Shanghai](../shanghai.md).
This extension is backwards-compatible, but not part of the initial Engine API.

## Structures

### ExecutionPayloadV3

This structure has the syntax of `ExecutionPayloadV2` and appends a single field: `excessDataGas`.

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
- `excessDataGas`: `QUANTITY`, 256 bits
- `blockHash`: `DATA`, 32 Bytes
- `transactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718)
- `withdrawals`: `Array of WithdrawalV1` - Array of withdrawals, each object is an `OBJECT` containing the fields of a `WithdrawalV1` structure.

### BlobsBundleV1

The fields are encoded as follows:

- `blockHash`: `DATA`, 32 Bytes
- `kzgs`: `Array of DATA` - Array of `KZGCommitment` as defined in [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844), 48 bytes each (`DATA`).
- `blobs`: `Array of DATA` - Array of blobs, each blob is `FIELD_ELEMENTS_PER_BLOB * BYTES_PER_FIELD_ELEMENT = 4096 * 32 = 131072` bytes (`DATA`) representing a SSZ-encoded `Blob` as defined in [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844)

## Methods

### engine_newPayloadV3

#### Request

* method: `engine_newPayloadV3`
* params:
  1. [`ExecutionPayloadV1`](../paris.md#ExecutionPayloadV1) | [`ExecutionPayloadV2`](../shanghai.md#ExecutionPayloadV2) | [`ExecutionPayloadV3`](#ExecutionPayloadV3), where:
      - `ExecutionPayloadV1` **MUST** be used if the `timestamp` value is lower than the Shanghai timestamp,
      - `ExecutionPayloadV2` **MUST** be used if the `timestamp` value is greater or equal to the Shanghai and lower than the EIP-4844 activation timestamp,
      - `ExecutionPayloadV3` **MUST** be used if the `timestamp` value is greater or equal to the EIP-4844 activation timestamp,
      - Client software **MUST** return `-32602: Invalid params` error if the wrong version of the structure is used in the method call.

#### Specification

Refer to the specification for `engine_newPayloadV2`.

#### Response

Refer to the response for `engine_newPayloadV2`.

### engine_getPayloadV3

#### Request

* method: `engine_getPayloadV3`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process
* timeout: 1s

#### Response

* result: `object`
  - `executionPayload`: [`ExecutionPayloadV1`](../paris.md#ExecutionPayloadV1) | [`ExecutionPayloadV2`](../shanghai.md#ExecutionPayloadV2) |  [`ExecutionPayloadV3`](#ExecutionPayloadV3) where:
    - `ExecutionPayloadV1` **MUST** be returned if the payload `timestamp` is lower than the Shanghai timestamp
    - `ExecutionPayloadV2` **MUST** be returned if the payload `timestamp` is greater or equal to the Shanghai timestamp and lower than the EIP-4844 activation timestamp
    - `ExecutionPayloadV3` **MUST** be returned if the payload `timestamp` is greater or equal to the EIP-4844 activation timestamp
  - `blockValue` : `QUANTITY`, 256 Bits - The expected value to be received by the `feeRecipient` in wei
* error: code and message set in case an exception happens while getting the payload.

#### Specification

Refer to the specification for `engine_getPayloadV2`.

### engine_getBlobsBundleV1

This method retrieves the blobs and their respective KZG commitments corresponding to the `versioned_hashes`
included in the blob transactions of the referenced execution payload.

This method may be combined with `engine_getPayloadV2`.
The separation of concerns aims to minimize changes during the testing phase of the EIP.

#### Request

* method: `engine_getBlobsBundleV1`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process
* timeout: 1s

#### Response

* result: [`BlobsBundleV1`](#BlobsBundleV1)
* error: code and message set in case an exception happens while getting the blobs bundle.

#### Specification

1. Given the `payloadId` client software **MUST** return the blobs bundle corresponding to the most recent version of the payload that was served with `engine_getPayload`, if any,
   and halt any further changes to the payload. The `engine_getBlobsBundleV1` and `engine_getPayloadV2` results **MUST** be consistent as outlined in items 3, 4 and 5 below. 

2. The call **MUST** return `-32001: Unknown payload` error if the build process identified by the `payloadId` does not exist. Note that a payload without any blobs **MUST** return an empty `blobs` and `kzgs` list, not an error.

3. The call **MUST** return `kzgs` matching the versioned hashes of the transactions list of the execution payload, in the same order,
   i.e. `assert verify_kzgs_against_transactions(payload.transactions, bundle.kzgs)` (see EIP-4844 consensus-specs).

4. The call **MUST** return `blobs` that match the `kzgs` list, i.e. `assert len(kzgs) == len(blobs) and all(blob_to_kzg(blob) == kzg for kzg, blob in zip(bundle.kzgs, bundle.blobs))`

5. The call **MUST** return `blockHash` to reference the `blockHash` of the corresponding execution payload, intended for the caller to sanity-check the consistency with the `engine_getPayload` call.
