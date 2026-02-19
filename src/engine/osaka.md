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
  - [engine_getBlobsV3](#engine_getblobsv3)
    - [Request](#request-2)
    - [Response](#response-2)
    - [Specification](#specification-2)
  - [Update the methods of previous forks](#update-the-methods-of-previous-forks)
    - [Cancun API](#cancun-api)
    - [Prague API](#prague-api)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Structures

### BlobsBundleV2

The fields are encoded as follows:

- `commitments`: `Array of DATA` - Array of `KZGCommitment` as defined in [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844), 48 bytes each (`DATA`).
- `proofs`: `Array of DATA` - Array of `KZGProof` (48 bytes each, type defined in [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844), semantics defined in [EIP-7594](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-7594.md)).
- `blobs`: `Array of DATA` - Array of blobs, each blob is `FIELD_ELEMENTS_PER_BLOB * BYTES_PER_FIELD_ELEMENT = 4096 * 32 = 131072` bytes (`DATA`) representing a SSZ-encoded `Blob` as defined in [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844)

`blobs` and `commitments` arrays **MUST** be of same length and `proofs` **MUST** contain exactly `CELLS_PER_EXT_BLOB` * `len(blobs)` cell proofs.

### BlobAndProofV2

The fields are encoded as follows:

- `blob`: `DATA` - `FIELD_ELEMENTS_PER_BLOB * BYTES_PER_FIELD_ELEMENT = 4096 * 32 = 131072` bytes (`DATA`) representing a SSZ-encoded `Blob` as defined in [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844).
- `proofs`: `Array of DATA` - Array of `KZGProof` as defined in [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844), 48 bytes each (`DATA`).

`proofs` **MUST** contain exactly `CELLS_PER_EXT_BLOB` cell proofs.

## Methods

### engine_getPayloadV5

This method is updated in a backward incompatible way. Instead of returning `blobsBundle` as [`BlobsBundleV1`](./cancun.md#blobsbundlev1), it returns as [`BlobsBundleV2`](#BlobsBundleV2).

#### Request

* method: `engine_getPayloadV5`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process
* timeout: 1s

#### Response

* result: `object`
  - `executionPayload`: [`ExecutionPayloadV3`](./cancun.md#executionpayloadv3)
  - `blockValue` : `QUANTITY`, 256 Bits - The expected value to be received by the `feeRecipient` in wei
  - `blobsBundle`: [`BlobsBundleV2`](#BlobsBundleV2) - Bundle with data corresponding to blob transactions included into `executionPayload`
  - `shouldOverrideBuilder` : `BOOLEAN` - Suggestion from the execution layer to use this `executionPayload` instead of an externally provided one
  - `executionRequests`: `Array of DATA` - Execution layer triggered requests obtained from the `executionPayload` transaction execution.
* error: code and message set in case an exception happens while getting the payload.

#### Specification

This method follows the same specification as [`engine_getPayloadV4`](./prague.md#engine_getpayloadv4) with changes of the following:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the built payload does not fall within the time frame of the Osaka fork.

2. The call **MUST** return `BlobsBundleV2` with empty `blobs`, `commitments` and `proofs` if the payload doesn't contain any blob transactions.

3. The call **MUST** return `blobs` and `proofs` that match the `commitments` list, i.e. 
   1. `assert len(blobsBundle.commitments) == len(blobsBundle.blobs)` and
   2. `assert len(blobsBundle.proofs) == len(blobsBundle.blobs) * CELLS_PER_EXT_BLOB` and
   3. `assert verify_cell_kzg_proof_batch(commitments, cell_indices, cells, blobsBundle.proofs)` (see [EIP-7594 consensus-specs](https://github.com/ethereum/consensus-specs/blob/master/specs/fulu/polynomial-commitments-sampling.md#verify_cell_kzg_proof_batch_impl))
      1. `commitments` should list each commitment `CELLS_PER_EXT_BLOB` times, repeating it for every cell. In python, `[blobsBundle.commitments[i] for i in range(len(blobsBundle.blobs)) for _ in range(CELLS_PER_EXT_BLOB)]`
      2. `cell_indices` should be `[0, ..., CELLS_PER_EXT_BLOB, 0, ..., CELLS_PER_EXT_BLOB, ...]`. In python, `list(range(CELLS_PER_EXT_BLOB)) * len(blobsBundle.blobs)`
      3. `cells` is the list of cells for an extended blob. In python, `[cell for blob in blobsBundle.blobs for cell in compute_cells(blob)]` (see [compute_cells](https://github.com/ethereum/consensus-specs/blob/master/specs/fulu/polynomial-commitments-sampling.md#compute_cells) in consensus-specs)
      4. All of the inputs to `verify_cell_kzg_proof_batch` have the same length, `CELLS_PER_EXT_BLOB * len(blobsBundle.blobs)`

### engine_getBlobsV2

Consensus layer clients **MAY** use this method to fetch blobs from the execution layer blob pool.

#### Request

* method: `engine_getBlobsV2`
* params:
  1. `Array of DATA`, 32 Bytes - Array of blob versioned hashes.
* timeout: 1s

#### Response

* result: `Array of BlobAndProofV2` - Array of [`BlobAndProofV2`](#BlobAndProofV2) or `null` in case of any missing blobs.
* error: code and message set in case an error occurs during processing of the request.

#### Specification

Refer to the specification for [`engine_getBlobsV1`](./cancun.md#engine_getblobsv1) with changes of the following:

1. Given an array of blob versioned hashes client software **MUST** respond with an array of `BlobAndProofV2` objects with matching versioned hashes, respecting the order of versioned hashes in the input array.
2. Client software **MUST** return `null` in case of any missing or older version blobs. For instance, 
   1. if the request is `[A_versioned_hash, B_versioned_hash, C_versioned_hash]` and client software has data for blobs `A` and `C`, but doesn't have data for `B`, the response **MUST** be `null`.
   2. if the request is `[A_versioned_hash_for_blob_with_blob_proof]`, the response **MUST** be `null` as well.
3. Client software **MUST** support request sizes of at least 128 blob versioned hashes. The client **MUST** return `-38004: Too large request` error if the number of requested blobs is too large.
4. Client software **MUST** return `null` if syncing or otherwise unable to serve blob pool data.
5. Callers **MUST** consider that execution layer clients may prune old blobs from their pool, and will respond with `null` if a blob has been pruned.

### engine_getBlobsV3

Consensus layer clients **MAY** use this method to fetch blobs from the execution layer blob pool, with support for partial responses. For an all-or-nothing query style, refer to `engine_getBlobsV2`.

#### Request

* method: `engine_getBlobsV3`
* params:
  1. `Array of DATA`, 32 Bytes - Array of blob versioned hashes.
* timeout: 1s

#### Response

* result: `Array of BlobAndProofV2` - Array of [`BlobAndProofV2`](#BlobAndProofV2), inserting `null` only at the positions of the missing blobs, or a `null` literal in the designated cases specified below.
* error: code and message set in case an error occurs during processing of the request.

#### Specification

> To assist the reader, we highlight differences against `engine_getBlobsV2` using italic.

1. Given an array of blob versioned hashes client software **MUST** respond with an array of `BlobAndProofV2` objects with matching versioned hashes, respecting the order of versioned hashes in the input array.
2. Given an array of blob versioned hashes, if client software has every one of the requested blobs, it **MUST** return an array of _`BlobAndProofV2`_ objects whose order exactly matches the input array. For instance, if the request is `[A_versioned_hash, B_versioned_hash, C_versioned_hash]` and client software has `A`, `B` and `C` available, the response **MUST** be `[A, B, C]`.
3. If one or more of the requested blobs are unavailable, _the client **MUST** return an array of the same length and order, inserting `null` only at the positions of the missing blobs._ For instance, if the request is `[A_versioned_hash, B_versioned_hash, C_versioned_hash]` and client software has data for blobs `A` and `C`, but doesn't have data for `B`, _the response **MUST** be `[A, null, C]`. If all blobs are missing, the client software must return an array of matching length, filled with `null` at all positions._
4. Client software **MUST** support request sizes of at least 128 blob versioned hashes. The client **MUST** return `-38004: Too large request` error if the number of requested blobs is too large.
5. Client software **MUST** return `null` if syncing or otherwise unable to generally serve blob pool data.
6. Callers **MUST** consider that execution layer clients may prune old blobs from their pool, and will respond with `null` at the corresponding position if a blob has been pruned.


### Update the methods of previous forks

#### Cancun API

This section defines how Osaka payload should be handled by the ['Cancun API'](./cancun.md).

For the following methods:

- [`engine_getBlobsV1`](./cancun.md#engine_getblobsv1)

a validation **MUST** be added:

1. Client software **MUST** return `-38005: Unsupported fork` error if the Osaka fork has been activated.

#### Prague API

This section defines how Osaka payload should be handled by the ['Prague API'](./prague.md).

For the following methods:

- [`engine_getPayloadV4`](./prague.md#engine_getpayloadv4)

a validation **MUST** be added:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of payload greater or equal to the Osaka activation timestamp.
