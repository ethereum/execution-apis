# Engine API -- Osaka

Engine API changes introduced in Osaka.

This specification is based on and extends [Engine API - Prague](./prague.md) specification.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Structures](#structures)
  - [BlobsBundleV2](#blobsbundlev2)
  - [BlobAndProofV2](#blobandproofv2)
- [Methods](#methods)
  - [engine_getPayloadV5](#engine_getpayloadv5)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)
  - [engine_getBlobsV2](#engine_getblobsv2)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Structures

### BlobsBundleV2

The fields are encoded as follows:

- `commitments`: `Array of DATA` - Array of `KZGCommitment` as defined in [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844), 48 bytes each (`DATA`).
- `cellProofs`: `Array of DATA` - Array of `KZGProof` (48 bytes each, type defined in [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844), semantics defined in [EIP-7594](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-7594.md)).
- `blobs`: `Array of DATA` - Array of blobs, each blob is `FIELD_ELEMENTS_PER_BLOB * BYTES_PER_FIELD_ELEMENT = 4096 * 32 = 131072` bytes (`DATA`) representing a SSZ-encoded `Blob` as defined in [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844)

`blobs` and `commitments` arrays **MUST** be of same length, `cellProofs` contains exactly `CELLS_PER_EXT_BLOB` * `len(blobs)` cell proofs.

### BlobAndProofV2

The fields are encoded as follows:

- `blob`: `DATA` - `FIELD_ELEMENTS_PER_BLOB * BYTES_PER_FIELD_ELEMENT = 4096 * 32 = 131072` bytes (`DATA`) representing a SSZ-encoded `Blob` as defined in [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844).
- `cellProofs`: `Array of DATA` - Array of `KZGProof` as defined in [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844), 48 bytes each (`DATA`).

`cellProofs` contains exactly `CELLS_PER_EXT_BLOB` cell proofs.

## Methods

### engine_getPayloadV5

This method is updated in a backward incompatible way. Instead of returning `BlobBundleV1`, it returns `BlobsBundleV2`.

#### Request

* method: `engine_getPayloadV5`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process
* timeout: 1s

#### Response

* result: `object`
  - `executionPayload`: [`ExecutionPayloadV3`](#ExecutionPayloadV3)
  - `blockValue` : `QUANTITY`, 256 Bits - The expected value to be received by the `feeRecipient` in wei
  - `blobsBundle`: [`BlobsBundleV2`](#BlobsBundleV2) - Bundle with data corresponding to blob transactions included into `executionPayload`
  - `shouldOverrideBuilder` : `BOOLEAN` - Suggestion from the execution layer to use this `executionPayload` instead of an externally provided one
  - `executionRequests`: `Array of DATA` - Execution layer triggered requests obtained from the `executionPayload` transaction execution.
* error: code and message set in case an exception happens while getting the payload.

#### Specification

This method follows the same specification as [`engine_getPayloadV4`](./prague.md#engine_getpayloadv4) with changes of the following:

1. The call **MUST** return `BlobsBundleV2` with empty `blobs`, `commitments` and `cellProofs` if the payload doesn't contain any blob transactions.

2. The call **MUST** return `blobs` and `cellProofs` that match the `commitments` list, i.e. 
   1. `assert len(blobsBundle.commitments) == len(blobsBundle.blobs)` and
   2. `assert len(blobsBundle.cellProofs) == len(blobsBundle.blobs) * CELLS_PER_EXT_BLOB` and
   3. `assert verify_cell_kzg_proof_batch(commitments, cell_indices, cells, blobsBundle.cellProofs)` (see [EIP-7594 consensus-specs](https://github.com/ethereum/consensus-specs/blob/36d80adb44c21c66379c6207a9578f9b1dcc8a2d/specs/fulu/polynomial-commitments-sampling.md#verify_cell_kzg_proof_batch))
      1. `cell_indices` should be `[0, ..., CELLS_PER_EXT_BLOB, 0, ..., CELLS_PER_EXT_BLOB, ...]`. In python, `list(range(CELLS_PER_EXT_BLOB)) * len(blobsBundle.blobs)`
      2. `commitments` should list each commitment `CELLS_PER_EXT_BLOB` times, repeating it for every cell. In python, `[blobsBundle.commitments[i] for i in range(len(blobsBundle.blobs)) for _ in range(CELLS_PER_EXT_BLOB)]`
      3. All of the inputs to `verify_cell_kzg_proof_batch` have the same length, `CELLS_PER_EXT_BLOB * len(blobsBundle.blobs)`

### engine_getBlobsV2

Consensus layer clients **MAY** use this method to fetch blobs from the execution layer blob pool.

*Note*: This is a new optional method introduced after Pectra.

#### Request

* method: `engine_getBlobsV2`
* params:
  1. `Array of DATA`, 32 Bytes - Array of blob versioned hashes.
* timeout: 1s

#### Response

* result: `Array of BlobAndProofV2` - Array of [`BlobAndProofV2`](#BlobAndProofV2), items of which may be `null`.
* error: code and message set in case an error occurs during processing of the request.

#### Specification

Refer to the specification for [`engine_getBlobsV1`](./cancun.md#engine_getblobsv1) with changes of the following:

1. return `BlobAndProofV2` instead of `BlobAndProofV1`
