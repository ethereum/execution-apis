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
- `blockHash`: `DATA`, 32 Bytes
- `transactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718)
- `withdrawals`: `Array of WithdrawalV1` - Array of withdrawals, each object is an `OBJECT` containing the fields of a `WithdrawalV1` structure.
- `excessDataGas`: `QUANTITY`, 256 bits

### BlobsBundleV1

The fields are encoded as follows:

- `kzgs`: `Array of DATA` - Array of `KZGCommitment` as defined in [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844), 48 bytes each (`DATA`).
- `proofs`: `Array of DATA` - Array of `KZGProof` as defined in [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844), 48 bytes each (`DATA`).
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

The response of this method is extended with [`BlobsBundleV1`](#blobsbundlev1) containing the blobs, their respective KZG commitments
and proofs corresponding to the `versioned_hashes` included in the blob transactions of the execution payload.

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
  - `blobsBundle`: [`BlobsBundleV1`](#BlobsBundleV1) - Bundle with data corresponding to blob transactions included into `executionPayload`
* error: code and message set in case an exception happens while getting the payload.

#### Specification

Refer to the specification for `engine_getPayloadV2` with addition of the following:

1. The call **MUST** return empty `blobs`, `kzgs` and `proofs` if the paylaod doesn't contain any blob transactions.

2. The call **MUST** return `kzgs` matching the versioned hashes of the transactions list of the execution payload, in the same order,
   i.e. `assert verify_kzgs_against_transactions(payload.transactions, bundle.kzgs)` (see EIP-4844 consensus-specs).

3. The call **MUST** return `blobs` and `proofs` that match the `kzgs` list, i.e. `assert len(kzgs) == len(blobs) == len(proofs)` and `assert verify_blob_kzg_proof_batch(bundle.blobs, bundle.kzgs, bundle.proofs)`.
