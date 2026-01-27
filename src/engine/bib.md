# Engine API -- BIB (Block-In-Blobs)

Engine API changes introduced in BIB.

This specification is based on and extends [Engine API - Osaka](./osaka.md) specification.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Constants](#constants)
- [Structures](#structures)
  - [BlobsBundleV3](#blobsbundlev3)
- [Methods](#methods)
  - [engine_newPayloadV5](#engine_newpayloadv5)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)
  - [engine_getPayloadV6](#engine_getpayloadv6)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)
  - [Update the methods of previous forks](#update-the-methods-of-previous-forks)
    - [Osaka API](#osaka-api)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Constants

| Name | Value | Description |
|------|-------|-------------|
| `VERSIONED_HASH_VERSION_PAYLOAD_BLOB` | `Bytes1(0x11)` | Version byte for payload blob versioned hashes. Format: `0xTV` where T=type (1=payload blob), V=version (1=KZG v1) |

*Note*: The first byte of a versioned hash uses the format `0xTV` where the high nibble (`T`) indicates the data type and the low nibble (`V`) indicates the commitment scheme version:
- `0x01` = Blob transaction data (type=0), KZG v1 (existing EIP-4844)
- `0x11` = Payload blob data (type=1), KZG v1 (new in BIB)

## Structures

### BlobsBundleV3

This structure extends [`BlobsBundleV2`](./osaka.md#blobsbundlev2) with an additional field to distinguish payload blob commitments from blob transaction commitments.

- `commitments`: `Array of DATA` - Same as BlobsBundleV2
- `proofs`: `Array of DATA` - Same as BlobsBundleV2
- `blobs`: `Array of DATA` - Same as BlobsBundleV2
- `executionPayloadBlobsCount`: `QUANTITY`, 64 Bits - Number of payload blob commitments. Blobs at indices `[0, executionPayloadBlobsCount)` are for payload blobs (`0x11` versioned hashes); blobs at indices `[executionPayloadBlobsCount, len(blobs))` are for blob transactions (`0x01` versioned hashes).

## Methods

### engine_newPayloadV5

This method introduces payload blob versioned hash validation, enabling data availability verification for execution payload transaction data.

#### Request

* method: `engine_newPayloadV5`
* params:
  1. `executionPayload`: [`ExecutionPayloadV3`](./cancun.md#executionpayloadv3).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected versioned hashes to validate. This array contains **both** payload blob versioned hashes (`0x11` prefix) and blob transaction versioned hashes (`0x01` prefix), concatenated in that order.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
  4. `executionRequests`: `Array of DATA` - List of execution layer triggered requests.

#### Response

Refer to the response for [`engine_newPayloadV4`](./prague.md#engine_newpayloadv4).

#### Specification

This method follows the same specification as [`engine_newPayloadV4`](./prague.md#engine_newpayloadv4) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload does not fall within the time frame of the BIB activation.

2. Given the expected array of versioned hashes, client software **MUST** run its validation by taking the following steps:
    1. Obtain the blob transaction versioned hashes by concatenating blob versioned hashes lists (`tx.blob_versioned_hashes`) of each [blob transaction](https://eips.ethereum.org/EIPS/eip-4844#new-transaction-type) included in the payload, respecting the order of inclusion. If the payload has no blob transactions, this array **MUST** be `[]`.
    2. Compute the payload blob versioned hashes by:
        1. Concatenate all transaction bytes from `executionPayload.transactions` into a single byte array.
        2. Chunk the concatenated bytes into blobs following EIP-4844 blob encoding (TBD). Each 32-byte field element within a blob **MUST** have its high byte set to `0x00` to ensure the value is less than the BLS modulus. This yields 31 usable bytes per field element, for a total of `4096 * 31 = 126976` usable bytes per blob.
        3. Zero-pad the final blob chunk if necessary.
        4. For each blob chunk, compute the KZG commitment using `blob_to_kzg_commitment`.
        5. For each commitment, compute the versioned hash as `VERSIONED_HASH_VERSION_PAYLOAD_BLOB + hash(commitment)[1:]` (i.e., `0x11` prefix).
    3. Construct the actual combined array by concatenating: `payload_blob_versioned_hashes + blob_tx_versioned_hashes`.
    4. Return `{status: INVALID, latestValidHash: null, validationError: errorMessage | null}` if the expected and actual arrays don't match (including length and order).

    This validation **MUST** be instantly run in all cases even during active sync process.

### engine_getPayloadV6

The response of this method uses `BlobsBundleV3` which includes `executionPayloadBlobsCount` to indicate where blob transaction commitments begin within the combined arrays.

#### Request

* method: `engine_getPayloadV6`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process
* timeout: 1s

#### Response

* result: `object`
  - `executionPayload`: [`ExecutionPayloadV3`](./cancun.md#executionpayloadv3)
  - `blockValue` : `QUANTITY`, 256 Bits - The expected value to be received by the `feeRecipient` in wei
  - `blobsBundle`: [`BlobsBundleV3`](#blobsbundlev3) - Bundle with data corresponding to **both** payload blobs and blob transactions. The `commitments`, `proofs`, and `blobs` arrays contain payload blob data first, followed by blob transaction data. The `executionPayloadBlobsCount` field indicates where blob transaction data begins.
  - `shouldOverrideBuilder` : `BOOLEAN` - Suggestion from the execution layer to use this `executionPayload` instead of an externally provided one
  - `executionRequests`: `Array of DATA` - Execution layer triggered requests obtained from the `executionPayload` transaction execution.
* error: code and message set in case an exception happens while getting the payload.

#### Specification

This method follows the same specification as [`engine_getPayloadV5`](./osaka.md#engine_getpayloadv5) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the built payload does not fall within the time frame of the BIB activation.

2. The call **MUST** return `blobsBundle` containing **both** payload blob data and blob transaction data:
    - `blobsBundle.commitments[0:executionPayloadBlobsCount]`: KZG commitments for payload blobs
    - `blobsBundle.commitments[executionPayloadBlobsCount:]`: KZG commitments for blob transactions
    - The `proofs` and `blobs` arrays follow the same partitioning.

3. The call **MUST** return `executionPayloadBlobsCount` within `blobsBundle` indicating where blob transaction data begins:
    - If the payload has no blob transactions, blob transaction commitments **MUST** be `[]` (i.e., `executionPayloadBlobsCount == len(blobsBundle.commitments)`).

4. Payload blob commitments **MUST** be computed as follows:
    1. Concatenate all transaction bytes from `executionPayload.transactions`.
    2. Chunk into blobs with proper EIP-4844 encoding (high byte of each 32-byte segment set to `0x00`).
    3. Compute KZG commitment and proof for each payload blob.

5. The versioned hashes derived from `blobsBundle.commitments[0:executionPayloadBlobsCount]` using the `0x11` prefix **MUST** correspond to the payload blob portion of `expectedBlobVersionedHashes` in the corresponding `engine_newPayloadV5` call.

6. The call **MUST** return `blobs` and `proofs` that match the `commitments` list, following the same verification as [`engine_getPayloadV5`](./osaka.md#specification).
