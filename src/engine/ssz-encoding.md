# Engine API -- SSZ Encoding

This document specifies an optional SSZ encoding for Engine API payloads as an alternative to the default JSON encoding. SSZ encoding reduces serialization overhead and aligns the Engine API with the native encoding format used by the consensus layer.

SSZ container definitions are provided for all Engine API structures and methods across all forks for backwards compatibility.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Motivation](#motivation)
- [Encoding negotiation](#encoding-negotiation)
- [Constants](#constants)
- [SSZ type mappings](#ssz-type-mappings)
- [Container definitions](#container-definitions)
  - [WithdrawalV1](#withdrawalv1)
  - [ExecutionPayloadV1](#executionpayloadv1)
  - [ExecutionPayloadV2](#executionpayloadv2)
  - [ExecutionPayloadV3](#executionpayloadv3)
  - [ExecutionPayloadV4](#executionpayloadv4)
  - [PayloadStatusV1](#payloadstatusv1)
  - [ForkchoiceStateV1](#forkchoicestatev1)
  - [PayloadAttributesV1](#payloadattributesv1)
  - [PayloadAttributesV2](#payloadattributesv2)
  - [PayloadAttributesV3](#payloadattributesv3)
  - [PayloadAttributesV4](#payloadattributesv4)
  - [ForkchoiceUpdatedResponseV1](#forkchoiceupdatedresponsev1)
  - [ExecutionPayloadBodyV1](#executionpayloadbodyv1)
  - [ExecutionPayloadBodyV2](#executionpayloadbodyv2)
  - [BlobsBundleV1](#blobsbundlev1)
  - [BlobsBundleV2](#blobsbundlev2)
  - [BlobAndProofV1](#blobandproofv1)
  - [BlobAndProofV2](#blobandproofv2)
  - [TransitionConfigurationV1](#transitionconfigurationv1)
  - [GetPayloadResponseV2](#getpayloadresponsev2)
  - [GetPayloadResponseV3](#getpayloadresponsev3)
  - [GetPayloadResponseV4](#getpayloadresponsev4)
  - [GetPayloadResponseV5](#getpayloadresponsev5)
  - [GetPayloadResponseV6](#getpayloadresponsev6)
- [Method definitions](#method-definitions)
  - [Paris methods](#paris-methods)
  - [Shanghai methods](#shanghai-methods)
  - [Cancun methods](#cancun-methods)
  - [Prague methods](#prague-methods)
  - [Osaka methods](#osaka-methods)
  - [Amsterdam methods](#amsterdam-methods)
- [Request and response format](#request-and-response-format)
- [Example](#example)
- [Error handling](#error-handling)
- [Security considerations](#security-considerations)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Motivation

The current JSON-RPC encoding introduces serialization overhead that grows with payload size. Binary data (hashes, addresses, bytecode) must be hex-encoded, doubling their size. As Ethereum scales through increased gas limits and blob transactions, this overhead becomes a bottleneck for block propagation and validation timing.

The consensus layer already uses SSZ for all internal data structures and network communication. The current architecture requires converting between SSZ and JSON at the Engine API boundary in both directions. SSZ encoding for the Engine API eliminates this double conversion, reduces payload sizes by 40-60%, and provides deterministic encoding.

## Encoding negotiation

SSZ encoding support is negotiated via standard HTTP content negotiation headers. No additional capability exchange is required.

| Header | Value | Meaning |
| - | - | - |
| `Content-Type` | `application/ssz` | The request body is SSZ-encoded |
| `Content-Type` | `application/json` | The request body is JSON-encoded (default) |
| `Accept` | `application/ssz` | The client prefers an SSZ-encoded response |
| `Accept` | `application/json` | The client prefers a JSON-encoded response (default) |

The negotiation works as follows:

1. The consensus layer client sends a request with `Accept: application/ssz` to indicate it can handle SSZ-encoded responses.

2. If the execution layer client supports SSZ encoding, it **SHOULD** respond with `Content-Type: application/ssz` and an SSZ-encoded body.

3. If the execution layer client does not support SSZ encoding, it **MUST** respond with `Content-Type: application/json` and a JSON-encoded body as usual. The `Accept` header is silently ignored.

4. A client receiving a request with `Content-Type: application/ssz` that does not support SSZ encoding **MUST** respond with HTTP status `415 Unsupported Media Type`. The requesting client **MUST** then fall back to JSON encoding for subsequent requests.

5. Clients **MUST** continue to support JSON encoding regardless of SSZ support. SSZ encoding is an optimization, not a replacement.

6. If no `Content-Type` header is present, the request **MUST** be parsed as JSON. If no `Accept` header is present, the response **SHOULD** use the same encoding as the request.

## Constants

| Name | Value | Source |
| - | - | - |
| `MAX_BYTES_PER_TRANSACTION` | `2**30` (1,073,741,824) | [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844) |
| `MAX_TRANSACTIONS_PER_PAYLOAD` | `2**20` (1,048,576) | [Bellatrix](https://github.com/ethereum/consensus-specs/blob/dev/specs/bellatrix/beacon-chain.md) |
| `MAX_WITHDRAWALS_PER_PAYLOAD` | `2**4` (16) | [Capella](https://github.com/ethereum/consensus-specs/blob/dev/specs/capella/beacon-chain.md) |
| `BYTES_PER_LOGS_BLOOM` | `256` | [Bellatrix](https://github.com/ethereum/consensus-specs/blob/dev/specs/bellatrix/beacon-chain.md) |
| `MAX_EXTRA_DATA_BYTES` | `2**5` (32) | [Bellatrix](https://github.com/ethereum/consensus-specs/blob/dev/specs/bellatrix/beacon-chain.md) |
| `MAX_BLOB_COMMITMENTS_PER_BLOCK` | `2**12` (4,096) | [Deneb](https://github.com/ethereum/consensus-specs/blob/dev/specs/deneb/beacon-chain.md) |
| `FIELD_ELEMENTS_PER_BLOB` | `4096` | [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844) |
| `BYTES_PER_FIELD_ELEMENT` | `32` | [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844) |
| `CELLS_PER_EXT_BLOB` | `128` | [EIP-7594](https://eips.ethereum.org/EIPS/eip-7594) |
| `MAX_PAYLOAD_BODIES_REQUEST` | `2**5` (32) | [Shanghai](./shanghai.md#engine_getpayloadbodiesbyhashv1) |
| `MAX_BLOB_HASHES_REQUEST` | `128` | [Osaka](./osaka.md#engine_getblobsv2) |
| `MAX_EXECUTION_REQUESTS` | `2**8` (256) | [EIP-7685](https://eips.ethereum.org/EIPS/eip-7685) |
| `MAX_ERROR_MESSAGE_LENGTH` | `1024` | This specification |
| `BLOB_SIZE` | `FIELD_ELEMENTS_PER_BLOB * BYTES_PER_FIELD_ELEMENT` (131,072) | Derived |

## SSZ type mappings

Each JSON-encoded base type used in the Engine API maps to a specific SSZ type. The mappings below correspond to the types defined in the [base types schema](../schemas/base-types.yaml).

| JSON Type | SSZ Type |
| - | - |
| `address` (20 bytes) | `Bytes20` |
| `hash32` (32 bytes) | `Bytes32` |
| `bytes8` (8 bytes) | `Bytes8` |
| `bytes32` (32 bytes) | `Bytes32` |
| `bytes48` (48 bytes) | `Bytes48` |
| `bytes256` (256 bytes) | `ByteVector[256]` |
| `uint64` | `uint64` |
| `uint256` | `uint256` |
| `BOOLEAN` | `boolean` |
| `bytes` (variable-length) | `ByteList[MAX_LENGTH]` (context-dependent) |
| `bytesMax32` (0 to 32 bytes) | `ByteList[32]` |
| `Array of T` | `List[T, MAX_LENGTH]` (context-dependent) |

## Container definitions

### WithdrawalV1

Introduced in [Shanghai](./shanghai.md#withdrawalv1).

```python
class WithdrawalV1(Container):
    index: uint64
    validator_index: uint64
    address: Bytes20
    amount: uint64
```

### ExecutionPayloadV1

Introduced in [Paris](./paris.md#executionpayloadv1).

```python
class ExecutionPayloadV1(Container):
    parent_hash: Bytes32
    fee_recipient: Bytes20
    state_root: Bytes32
    receipts_root: Bytes32
    logs_bloom: ByteVector[BYTES_PER_LOGS_BLOOM]
    prev_randao: Bytes32
    block_number: uint64
    gas_limit: uint64
    gas_used: uint64
    timestamp: uint64
    extra_data: ByteList[MAX_EXTRA_DATA_BYTES]
    base_fee_per_gas: uint256
    block_hash: Bytes32
    transactions: List[ByteList[MAX_BYTES_PER_TRANSACTION], MAX_TRANSACTIONS_PER_PAYLOAD]
```

### ExecutionPayloadV2

Introduced in [Shanghai](./shanghai.md#executionpayloadv2). Extends `ExecutionPayloadV1` with `withdrawals`.

```python
class ExecutionPayloadV2(Container):
    parent_hash: Bytes32
    fee_recipient: Bytes20
    state_root: Bytes32
    receipts_root: Bytes32
    logs_bloom: ByteVector[BYTES_PER_LOGS_BLOOM]
    prev_randao: Bytes32
    block_number: uint64
    gas_limit: uint64
    gas_used: uint64
    timestamp: uint64
    extra_data: ByteList[MAX_EXTRA_DATA_BYTES]
    base_fee_per_gas: uint256
    block_hash: Bytes32
    transactions: List[ByteList[MAX_BYTES_PER_TRANSACTION], MAX_TRANSACTIONS_PER_PAYLOAD]
    withdrawals: List[WithdrawalV1, MAX_WITHDRAWALS_PER_PAYLOAD]
```

### ExecutionPayloadV3

Introduced in [Cancun](./cancun.md#executionpayloadv3). Extends `ExecutionPayloadV2` with `blob_gas_used` and `excess_blob_gas`.

```python
class ExecutionPayloadV3(Container):
    parent_hash: Bytes32
    fee_recipient: Bytes20
    state_root: Bytes32
    receipts_root: Bytes32
    logs_bloom: ByteVector[BYTES_PER_LOGS_BLOOM]
    prev_randao: Bytes32
    block_number: uint64
    gas_limit: uint64
    gas_used: uint64
    timestamp: uint64
    extra_data: ByteList[MAX_EXTRA_DATA_BYTES]
    base_fee_per_gas: uint256
    block_hash: Bytes32
    transactions: List[ByteList[MAX_BYTES_PER_TRANSACTION], MAX_TRANSACTIONS_PER_PAYLOAD]
    withdrawals: List[WithdrawalV1, MAX_WITHDRAWALS_PER_PAYLOAD]
    blob_gas_used: uint64
    excess_blob_gas: uint64
```

### ExecutionPayloadV4

Introduced in [Amsterdam](./amsterdam.md#executionpayloadv4). Extends `ExecutionPayloadV3` with `block_access_list` and `slot_number`.

```python
class ExecutionPayloadV4(Container):
    parent_hash: Bytes32
    fee_recipient: Bytes20
    state_root: Bytes32
    receipts_root: Bytes32
    logs_bloom: ByteVector[BYTES_PER_LOGS_BLOOM]
    prev_randao: Bytes32
    block_number: uint64
    gas_limit: uint64
    gas_used: uint64
    timestamp: uint64
    extra_data: ByteList[MAX_EXTRA_DATA_BYTES]
    base_fee_per_gas: uint256
    block_hash: Bytes32
    transactions: List[ByteList[MAX_BYTES_PER_TRANSACTION], MAX_TRANSACTIONS_PER_PAYLOAD]
    withdrawals: List[WithdrawalV1, MAX_WITHDRAWALS_PER_PAYLOAD]
    blob_gas_used: uint64
    excess_blob_gas: uint64
    block_access_list: ByteList[MAX_BYTES_PER_TRANSACTION]
    slot_number: uint64
```

### PayloadStatusV1

Introduced in [Paris](./paris.md#payloadstatusv1). The `status` field is encoded as a `uint8` enum.

```python
class PayloadStatusV1(Container):
    status: uint8
    latest_valid_hash: Bytes32
    validation_error: ByteList[MAX_ERROR_MESSAGE_LENGTH]
```

*Note:* `latest_valid_hash` is all zeros when absent (e.g. when `status` is `SYNCING` or `ACCEPTED`). `validation_error` is empty when absent.

| `status` value | Meaning |
| - | - |
| `0` | VALID |
| `1` | INVALID |
| `2` | SYNCING |
| `3` | ACCEPTED |
| `4` | INVALID_BLOCK_HASH |

### ForkchoiceStateV1

Introduced in [Paris](./paris.md#forkchoicestatev1).

```python
class ForkchoiceStateV1(Container):
    head_block_hash: Bytes32
    safe_block_hash: Bytes32
    finalized_block_hash: Bytes32
```

### PayloadAttributesV1

Introduced in [Paris](./paris.md#payloadattributesv1).

```python
class PayloadAttributesV1(Container):
    timestamp: uint64
    prev_randao: Bytes32
    suggested_fee_recipient: Bytes20
```

### PayloadAttributesV2

Introduced in [Shanghai](./shanghai.md#payloadattributesv2). Extends `PayloadAttributesV1` with `withdrawals`.

```python
class PayloadAttributesV2(Container):
    timestamp: uint64
    prev_randao: Bytes32
    suggested_fee_recipient: Bytes20
    withdrawals: List[WithdrawalV1, MAX_WITHDRAWALS_PER_PAYLOAD]
```

### PayloadAttributesV3

Introduced in [Cancun](./cancun.md#payloadattributesv3). Extends `PayloadAttributesV2` with `parent_beacon_block_root`.

```python
class PayloadAttributesV3(Container):
    timestamp: uint64
    prev_randao: Bytes32
    suggested_fee_recipient: Bytes20
    withdrawals: List[WithdrawalV1, MAX_WITHDRAWALS_PER_PAYLOAD]
    parent_beacon_block_root: Bytes32
```

### PayloadAttributesV4

Introduced in [Amsterdam](./amsterdam.md#payloadattributesv4). Extends `PayloadAttributesV3` with `slot_number`.

```python
class PayloadAttributesV4(Container):
    timestamp: uint64
    prev_randao: Bytes32
    suggested_fee_recipient: Bytes20
    withdrawals: List[WithdrawalV1, MAX_WITHDRAWALS_PER_PAYLOAD]
    parent_beacon_block_root: Bytes32
    slot_number: uint64
```

### ForkchoiceUpdatedResponseV1

Used by all versions of `engine_forkchoiceUpdated`.

```python
class ForkchoiceUpdatedResponseV1(Container):
    payload_status: PayloadStatusV1
    payload_id: Bytes8
```

*Note:* `payload_id` is all zeros when no payload building was initiated.

### ExecutionPayloadBodyV1

Introduced in [Shanghai](./shanghai.md#executionpayloadbodyv1).

```python
class ExecutionPayloadBodyV1(Container):
    transactions: List[ByteList[MAX_BYTES_PER_TRANSACTION], MAX_TRANSACTIONS_PER_PAYLOAD]
    withdrawals: List[WithdrawalV1, MAX_WITHDRAWALS_PER_PAYLOAD]
```

*Note:* `withdrawals` is empty for pre-Shanghai blocks.

### ExecutionPayloadBodyV2

Introduced in [Amsterdam](./amsterdam.md#executionpayloadbodyv2). Extends `ExecutionPayloadBodyV1` with `block_access_list`.

```python
class ExecutionPayloadBodyV2(Container):
    transactions: List[ByteList[MAX_BYTES_PER_TRANSACTION], MAX_TRANSACTIONS_PER_PAYLOAD]
    withdrawals: List[WithdrawalV1, MAX_WITHDRAWALS_PER_PAYLOAD]
    block_access_list: ByteList[MAX_BYTES_PER_TRANSACTION]
```

*Note:* `withdrawals` is empty for pre-Shanghai blocks. `block_access_list` is empty for pre-Amsterdam blocks.

### BlobsBundleV1

Introduced in [Cancun](./cancun.md#blobsbundlev1).

```python
class BlobsBundleV1(Container):
    commitments: List[Bytes48, MAX_BLOB_COMMITMENTS_PER_BLOCK]
    proofs: List[Bytes48, MAX_BLOB_COMMITMENTS_PER_BLOCK]
    blobs: List[ByteVector[BLOB_SIZE], MAX_BLOB_COMMITMENTS_PER_BLOCK]
```

### BlobsBundleV2

Introduced in [Osaka](./osaka.md#blobsbundlev2). Proofs are cell proofs with `CELLS_PER_EXT_BLOB` proofs per blob.

```python
class BlobsBundleV2(Container):
    commitments: List[Bytes48, MAX_BLOB_COMMITMENTS_PER_BLOCK]
    proofs: List[Bytes48, MAX_BLOB_COMMITMENTS_PER_BLOCK * CELLS_PER_EXT_BLOB]
    blobs: List[ByteVector[BLOB_SIZE], MAX_BLOB_COMMITMENTS_PER_BLOCK]
```

### BlobAndProofV1

Introduced in [Cancun](./cancun.md#blobandproofv1).

```python
class BlobAndProofV1(Container):
    blob: ByteVector[BLOB_SIZE]
    proof: Bytes48
```

### BlobAndProofV2

Introduced in [Osaka](./osaka.md#blobandproofv2).

```python
class BlobAndProofV2(Container):
    blob: ByteVector[BLOB_SIZE]
    proofs: List[Bytes48, CELLS_PER_EXT_BLOB]
```

### TransitionConfigurationV1

Introduced in [Paris](./paris.md#transitionconfigurationv1). Deprecated in Cancun.

```python
class TransitionConfigurationV1(Container):
    terminal_total_difficulty: uint256
    terminal_block_hash: Bytes32
    terminal_block_number: uint64
```

### GetPayloadResponseV2

Response container for [`engine_getPayloadV2`](./shanghai.md#engine_getpayloadv2).

```python
class GetPayloadResponseV2(Container):
    execution_payload: ExecutionPayloadV2
    block_value: uint256
```

*Note:* `engine_getPayloadV2` may return `ExecutionPayloadV1` for pre-Shanghai timestamps. The SSZ encoding uses `ExecutionPayloadV2` in all cases; pre-Shanghai payloads have an empty `withdrawals` list.

### GetPayloadResponseV3

Response container for [`engine_getPayloadV3`](./cancun.md#engine_getpayloadv3).

```python
class GetPayloadResponseV3(Container):
    execution_payload: ExecutionPayloadV3
    block_value: uint256
    blobs_bundle: BlobsBundleV1
    should_override_builder: boolean
```

### GetPayloadResponseV4

Response container for [`engine_getPayloadV4`](./prague.md#engine_getpayloadv4).

```python
class GetPayloadResponseV4(Container):
    execution_payload: ExecutionPayloadV3
    block_value: uint256
    blobs_bundle: BlobsBundleV1
    should_override_builder: boolean
    execution_requests: List[ByteList[MAX_BYTES_PER_TRANSACTION], MAX_EXECUTION_REQUESTS]
```

### GetPayloadResponseV5

Response container for [`engine_getPayloadV5`](./osaka.md#engine_getpayloadv5).

```python
class GetPayloadResponseV5(Container):
    execution_payload: ExecutionPayloadV3
    block_value: uint256
    blobs_bundle: BlobsBundleV2
    should_override_builder: boolean
    execution_requests: List[ByteList[MAX_BYTES_PER_TRANSACTION], MAX_EXECUTION_REQUESTS]
```

### GetPayloadResponseV6

Response container for [`engine_getPayloadV6`](./amsterdam.md#engine_getpayloadv6).

```python
class GetPayloadResponseV6(Container):
    execution_payload: ExecutionPayloadV4
    block_value: uint256
    blobs_bundle: BlobsBundleV2
    should_override_builder: boolean
    execution_requests: List[ByteList[MAX_BYTES_PER_TRANSACTION], MAX_EXECUTION_REQUESTS]
```

## Method definitions

This section defines the SSZ types for each method's parameters and result, organized by fork. Each parameter is individually SSZ-encoded in the JSON-RPC `params` array. Nullable parameters remain `null` when absent.

### Paris methods

#### engine_newPayloadV1

| Parameter | SSZ Type |
| - | - |
| `executionPayload` | [`ExecutionPayloadV1`](#executionpayloadv1) |

| Result | SSZ Type |
| - | - |
| Payload status | [`PayloadStatusV1`](#payloadstatusv1) |

#### engine_forkchoiceUpdatedV1

| Parameter | SSZ Type |
| - | - |
| `forkchoiceState` | [`ForkchoiceStateV1`](#forkchoicestatev1) |
| `payloadAttributes` | [`PayloadAttributesV1`](#payloadattributesv1) or `null` |

| Result | SSZ Type |
| - | - |
| Forkchoice updated response | [`ForkchoiceUpdatedResponseV1`](#forkchoiceupdatedresponsev1) |

#### engine_getPayloadV1

| Parameter | SSZ Type |
| - | - |
| `payloadId` | `Bytes8` |

| Result | SSZ Type |
| - | - |
| Execution payload | [`ExecutionPayloadV1`](#executionpayloadv1) |

#### engine_exchangeTransitionConfigurationV1

Deprecated in Cancun.

| Parameter | SSZ Type |
| - | - |
| `transitionConfiguration` | [`TransitionConfigurationV1`](#transitionconfigurationv1) |

| Result | SSZ Type |
| - | - |
| Transition configuration | [`TransitionConfigurationV1`](#transitionconfigurationv1) |

### Shanghai methods

#### engine_newPayloadV2

| Parameter | SSZ Type |
| - | - |
| `executionPayload` | [`ExecutionPayloadV1`](#executionpayloadv1) or [`ExecutionPayloadV2`](#executionpayloadv2) (by timestamp) |

| Result | SSZ Type |
| - | - |
| Payload status | [`PayloadStatusV1`](#payloadstatusv1) |

#### engine_forkchoiceUpdatedV2

| Parameter | SSZ Type |
| - | - |
| `forkchoiceState` | [`ForkchoiceStateV1`](#forkchoicestatev1) |
| `payloadAttributes` | [`PayloadAttributesV1`](#payloadattributesv1), [`PayloadAttributesV2`](#payloadattributesv2), or `null` |

| Result | SSZ Type |
| - | - |
| Forkchoice updated response | [`ForkchoiceUpdatedResponseV1`](#forkchoiceupdatedresponsev1) |

#### engine_getPayloadV2

| Parameter | SSZ Type |
| - | - |
| `payloadId` | `Bytes8` |

| Result | SSZ Type |
| - | - |
| Get payload response | [`GetPayloadResponseV2`](#getpayloadresponsev2) |

#### engine_getPayloadBodiesByHashV1

| Parameter | SSZ Type |
| - | - |
| `blockHashes` | `List[Bytes32, MAX_PAYLOAD_BODIES_REQUEST]` |

| Result | SSZ Type |
| - | - |
| Payload bodies | `List[List[`[`ExecutionPayloadBodyV1`](#executionpayloadbodyv1)`, 1], MAX_PAYLOAD_BODIES_REQUEST]` |

*Note:* Each inner list has 0 elements for unknown blocks and 1 element for known blocks.

#### engine_getPayloadBodiesByRangeV1

| Parameter | SSZ Type |
| - | - |
| `start` | `uint64` |
| `count` | `uint64` |

| Result | SSZ Type |
| - | - |
| Payload bodies | `List[List[`[`ExecutionPayloadBodyV1`](#executionpayloadbodyv1)`, 1], MAX_PAYLOAD_BODIES_REQUEST]` |

*Note:* Each inner list has 0 elements for unknown blocks and 1 element for known blocks.

### Cancun methods

#### engine_newPayloadV3

| Parameter | SSZ Type |
| - | - |
| `executionPayload` | [`ExecutionPayloadV3`](#executionpayloadv3) |
| `expectedBlobVersionedHashes` | `List[Bytes32, MAX_BLOB_COMMITMENTS_PER_BLOCK]` |
| `parentBeaconBlockRoot` | `Bytes32` |

| Result | SSZ Type |
| - | - |
| Payload status | [`PayloadStatusV1`](#payloadstatusv1) |

#### engine_forkchoiceUpdatedV3

| Parameter | SSZ Type |
| - | - |
| `forkchoiceState` | [`ForkchoiceStateV1`](#forkchoicestatev1) |
| `payloadAttributes` | [`PayloadAttributesV3`](#payloadattributesv3) or `null` |

| Result | SSZ Type |
| - | - |
| Forkchoice updated response | [`ForkchoiceUpdatedResponseV1`](#forkchoiceupdatedresponsev1) |

#### engine_getPayloadV3

| Parameter | SSZ Type |
| - | - |
| `payloadId` | `Bytes8` |

| Result | SSZ Type |
| - | - |
| Get payload response | [`GetPayloadResponseV3`](#getpayloadresponsev3) |

#### engine_getBlobsV1

Deprecated in Osaka.

| Parameter | SSZ Type |
| - | - |
| `blobVersionedHashes` | `List[Bytes32, MAX_BLOB_HASHES_REQUEST]` |

| Result | SSZ Type |
| - | - |
| Blobs and proofs | `List[`[`BlobAndProofV1`](#blobandproofv1)`, MAX_BLOB_HASHES_REQUEST]` |

*Note:* Returns `null` at the JSON-RPC level when syncing (SSZ encoding only applies to non-null results).

### Prague methods

#### engine_newPayloadV4

| Parameter | SSZ Type |
| - | - |
| `executionPayload` | [`ExecutionPayloadV3`](#executionpayloadv3) |
| `expectedBlobVersionedHashes` | `List[Bytes32, MAX_BLOB_COMMITMENTS_PER_BLOCK]` |
| `parentBeaconBlockRoot` | `Bytes32` |
| `executionRequests` | `List[ByteList[MAX_BYTES_PER_TRANSACTION], MAX_EXECUTION_REQUESTS]` |

| Result | SSZ Type |
| - | - |
| Payload status | [`PayloadStatusV1`](#payloadstatusv1) |

#### engine_getPayloadV4

| Parameter | SSZ Type |
| - | - |
| `payloadId` | `Bytes8` |

| Result | SSZ Type |
| - | - |
| Get payload response | [`GetPayloadResponseV4`](#getpayloadresponsev4) |

### Osaka methods

#### engine_getPayloadV5

| Parameter | SSZ Type |
| - | - |
| `payloadId` | `Bytes8` |

| Result | SSZ Type |
| - | - |
| Get payload response | [`GetPayloadResponseV5`](#getpayloadresponsev5) |

#### engine_getBlobsV2

Returns `null` for the entire result if any blob is missing or if syncing.

| Parameter | SSZ Type |
| - | - |
| `blobVersionedHashes` | `List[Bytes32, MAX_BLOB_HASHES_REQUEST]` |

| Result | SSZ Type |
| - | - |
| Blobs and proofs | `List[`[`BlobAndProofV2`](#blobandproofv2)`, MAX_BLOB_HASHES_REQUEST]` |

*Note:* Returns `null` at the JSON-RPC level when syncing or any blob is missing (SSZ encoding only applies to non-null results).

#### engine_getBlobsV3

Returns per-element `null` for missing blobs, or `null` for the entire result if syncing.

| Parameter | SSZ Type |
| - | - |
| `blobVersionedHashes` | `List[Bytes32, MAX_BLOB_HASHES_REQUEST]` |

| Result | SSZ Type |
| - | - |
| Blobs and proofs | `List[List[`[`BlobAndProofV2`](#blobandproofv2)`, 1], MAX_BLOB_HASHES_REQUEST]` |

*Note:* Returns `null` at the JSON-RPC level when syncing. Each inner list has 0 elements for a missing blob and 1 element for a present blob.

### Amsterdam methods

#### engine_newPayloadV5

| Parameter | SSZ Type |
| - | - |
| `executionPayload` | [`ExecutionPayloadV4`](#executionpayloadv4) |
| `expectedBlobVersionedHashes` | `List[Bytes32, MAX_BLOB_COMMITMENTS_PER_BLOCK]` |
| `parentBeaconBlockRoot` | `Bytes32` |
| `executionRequests` | `List[ByteList[MAX_BYTES_PER_TRANSACTION], MAX_EXECUTION_REQUESTS]` |

| Result | SSZ Type |
| - | - |
| Payload status | [`PayloadStatusV1`](#payloadstatusv1) |

#### engine_getPayloadV6

| Parameter | SSZ Type |
| - | - |
| `payloadId` | `Bytes8` |

| Result | SSZ Type |
| - | - |
| Get payload response | [`GetPayloadResponseV6`](#getpayloadresponsev6) |

#### engine_forkchoiceUpdatedV4

| Parameter | SSZ Type |
| - | - |
| `forkchoiceState` | [`ForkchoiceStateV1`](#forkchoicestatev1) |
| `payloadAttributes` | [`PayloadAttributesV4`](#payloadattributesv4) or `null` |

| Result | SSZ Type |
| - | - |
| Forkchoice updated response | [`ForkchoiceUpdatedResponseV1`](#forkchoiceupdatedresponsev1) |

#### engine_getPayloadBodiesByHashV2

| Parameter | SSZ Type |
| - | - |
| `blockHashes` | `List[Bytes32, MAX_PAYLOAD_BODIES_REQUEST]` |

| Result | SSZ Type |
| - | - |
| Payload bodies | `List[List[`[`ExecutionPayloadBodyV2`](#executionpayloadbodyv2)`, 1], MAX_PAYLOAD_BODIES_REQUEST]` |

*Note:* Each inner list has 0 elements for unknown blocks and 1 element for known blocks.

#### engine_getPayloadBodiesByRangeV2

| Parameter | SSZ Type |
| - | - |
| `start` | `uint64` |
| `count` | `uint64` |

| Result | SSZ Type |
| - | - |
| Payload bodies | `List[List[`[`ExecutionPayloadBodyV2`](#executionpayloadbodyv2)`, 1], MAX_PAYLOAD_BODIES_REQUEST]` |

*Note:* Each inner list has 0 elements for unknown blocks and 1 element for known blocks.

## Request and response format

SSZ-encoded Engine API requests and responses follow the existing JSON-RPC method semantics. The SSZ encoding applies to the method parameters and result values — the JSON-RPC envelope (`jsonrpc`, `id`, `method`) remains JSON-encoded.

Specifically, when SSZ encoding is in use:

1. The HTTP request body is a JSON-RPC request where each element of `params` is replaced with its SSZ-encoded hexadecimal representation (a `DATA` string). Parameters that are `null` remain `null`.

2. The HTTP response body is a JSON-RPC response where the `result` field is replaced with the SSZ-encoded hexadecimal representation of the result value. A `null` result remains `null`.

This approach preserves compatibility with JSON-RPC tooling while encoding the payload data in SSZ.

*Note:* Future versions of this specification may define a fully binary request/response format that replaces the JSON-RPC envelope.

## Example

The following example shows an `engine_newPayloadV5` call, first using JSON encoding and then using SSZ encoding.

### JSON-encoded request (current behavior)

```console
$ curl https://localhost:8551 \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "engine_newPayloadV5",
  "params": [
    {
      "parentHash": "0x3b8fb240d288781d4aac94d3fd16809ee413bc99294a085798a589dae51ddd4a",
      "feeRecipient": "0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b",
      "stateRoot": "0xca3149fa9e37db08d1cd49c9061db1002ef1cd58db2210f2115c8c989b2bdf45",
      "receiptsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
      "logsBloom": "0x0000...0000",
      "prevRandao": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "blockNumber": "0x1",
      "gasLimit": "0x1c9c380",
      "gasUsed": "0x0",
      "timestamp": "0x5",
      "extraData": "0x",
      "baseFeePerGas": "0x7",
      "blockHash": "0x3559e851470f6e7bbed1db474980683e8c315bfce99b2a6ef47c057c04de7858",
      "transactions": [],
      "withdrawals": [],
      "blobGasUsed": "0x0",
      "excessBlobGas": "0x0",
      "blockAccessList": "0x",
      "slotNumber": "0x1"
    },
    [],
    "0x0000000000000000000000000000000000000000000000000000000000000000",
    []
  ]
}'
```

### JSON-encoded response

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "status": "VALID",
    "latestValidHash": "0x3559e851470f6e7bbed1db474980683e8c315bfce99b2a6ef47c057c04de7858",
    "validationError": null
  }
}
```

### SSZ-encoded request

The consensus layer client sends the same call with `Content-Type: application/ssz` and `Accept: application/ssz`. Each element of the `params` array is individually SSZ-encoded as a hex `DATA` string:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "engine_newPayloadV5",
  "params": [
    "0x3b8fb240d288781d4aac94d3fd16809ee413bc99...",
    "0x",
    "0x0000000000000000000000000000000000000000000000000000000000000000",
    "0x"
  ]
}
```

- `params[0]`: SSZ-serialized `ExecutionPayloadV4` container
- `params[1]`: SSZ-serialized `List[Bytes32, MAX_BLOB_COMMITMENTS_PER_BLOCK]` (empty list)
- `params[2]`: SSZ-serialized `Bytes32` (parent beacon block root)
- `params[3]`: SSZ-serialized `List[ByteList, MAX_EXECUTION_REQUESTS]` (empty list)

### SSZ-encoded response

The execution layer responds with `Content-Type: application/ssz`. The `result` field contains the SSZ-serialized `PayloadStatusV1`:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "0x00013559e851470f6e7bbed1db474980683e8c315bfce99b2a6ef47c057c04de7858"
}
```

Where the binary data encodes:
- `status`: `0x00` (VALID)
- `latest_valid_hash`: present, `0x3559e851470f6e7bbed1db474980683e8c315bfce99b2a6ef47c057c04de7858`
- `validation_error`: absent

### Fallback behavior

If the execution layer does not support SSZ, the same request with `Accept: application/ssz` returns a standard JSON response with `Content-Type: application/json`. The consensus layer detects this and continues using JSON for subsequent requests.

If the consensus layer sends `Content-Type: application/ssz` to an execution layer that does not support it, the execution layer responds with HTTP `415 Unsupported Media Type`. The consensus layer **MUST** retry the request using JSON encoding.

## Error handling

SSZ encoding does not change the error semantics of the Engine API. All error codes defined in the [Errors](./common.md#errors) section apply equally to SSZ-encoded requests.

Additionally:

| Code | Message | Meaning |
| - | - | - |
| -32700 | Parse error | Invalid SSZ data was received by the server. |

Clients **MUST** validate SSZ payloads against the expected schema before processing. Payloads that do not conform to the expected SSZ schema **MUST** be rejected with a `-32700` error.

## Security considerations

- SSZ deserialization **MUST** enforce the same size limits as JSON deserialization. Implementations **MUST** reject SSZ payloads exceeding defined maximum sizes before attempting full deserialization.
- Implementations **SHOULD** use well-tested SSZ libraries and fuzz test SSZ parsing extensively.
