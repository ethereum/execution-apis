# Engine API -- Binary SSZ Transport

This document specifies a binary SSZ transport for Engine API communication between consensus layer (CL) and execution layer (EL) clients. The binary transport replaces JSON-RPC with raw SSZ over HTTP for fast, efficient CL-EL communication.

SSZ container definitions are provided for all Engine API structures and methods across all forks for backwards compatibility.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Motivation](#motivation)
- [Transport](#transport)
  - [Request format](#request-format)
  - [Response format](#response-format)
  - [Negotiation and fallback](#negotiation-and-fallback)
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
  - [PayloadBodiesV1Response](#payloadbodiesv1response)
  - [PayloadBodiesV2Response](#payloadbodiesv2response)
  - [GetBlobsV1Response](#getblobsv1response)
  - [GetBlobsV2Response](#getblobsv2response)
  - [GetBlobsV3Response](#getblobsv3response)
  - [ErrorResponse](#errorresponse)
- [Method definitions](#method-definitions)
  - [Paris methods](#paris-methods)
  - [Shanghai methods](#shanghai-methods)
  - [Cancun methods](#cancun-methods)
  - [Prague methods](#prague-methods)
  - [Osaka methods](#osaka-methods)
  - [Amsterdam methods](#amsterdam-methods)
- [Example](#example)
- [Error handling](#error-handling)
- [Security considerations](#security-considerations)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Motivation

Fast communication between the consensus layer and execution layer is critical for block propagation and validation timing. The JSON-RPC transport introduces unnecessary overhead in this critical path:

- Binary data (hashes, addresses, transactions, blobs) is hex-encoded, doubling wire size.
- JSON parsing and generation adds CPU overhead on both sides.
- The CL uses SSZ natively, forcing a round-trip conversion (SSZ to JSON, then JSON to internal types) at the Engine API boundary.

Binary SSZ eliminates all of this. The CL sends raw SSZ bytes over HTTP; the EL deserializes directly. No hex encoding, no JSON parsing, no intermediate representations. Payload sizes are reduced by 50% or more compared to JSON-RPC, and serialization is no longer a bottleneck in the critical path between CL and EL.

## Transport

Binary SSZ uses HTTP with path-based method routing. Each Engine API method has a dedicated URL path. Request and response bodies are raw SSZ bytes.

### Request format

```
POST /engine/<methodName> HTTP/1.1
Content-Type: application/ssz

<raw SSZ bytes of the method's request container>
```

The URL path is `/engine/<methodName>` where `<methodName>` corresponds to the JSON-RPC method name with the `engine_` prefix removed. For example, `engine_newPayloadV5` maps to `POST /engine/newPayloadV5`.

The request body is the SSZ serialization of the method's request container. Each method defines a request container that wraps all parameters into a single SSZ object.

### Response format

**Success with data:**

```
HTTP/1.1 200 OK
Content-Type: application/ssz

<raw SSZ bytes of the method's response type>
```

**Null result** (e.g., syncing):

```
HTTP/1.1 204 No Content
```

Methods that can return `null` at the JSON-RPC level use HTTP `204 No Content` with an empty body.

**Error:**

```
HTTP/1.1 <status> <reason>
Content-Type: application/ssz

<raw SSZ bytes of ErrorResponse>
```

### Negotiation and fallback

1. The CL sends a request to the method's URL path with `Content-Type: application/ssz` and a raw SSZ request body.

2. If the EL supports the binary SSZ transport, it **MUST** respond with `Content-Type: application/ssz` and a raw SSZ response body.

3. If the EL does not support the binary SSZ transport, it **MUST** respond with HTTP status `404 Not Found` or `415 Unsupported Media Type`. The CL **MUST** then fall back to JSON-RPC (`POST /`) for subsequent requests.

4. Clients **MUST** continue to support JSON-RPC encoding as a fallback. Both the binary SSZ endpoint and the JSON-RPC endpoint coexist on the same port.

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
| `T or null` | `Optional[T]` (encoded as `List[T, 1]`) |

`Optional[T]` is represented as `List[T, 1]` in SSZ encoding. An empty list (0 elements) denotes absence (`null`). A list with one element denotes presence.

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

### PayloadBodiesV1Response

Response container for `engine_getPayloadBodiesByHashV1` and `engine_getPayloadBodiesByRangeV1`.

```python
class PayloadBodiesV1Response(Container):
    payload_bodies: List[List[ExecutionPayloadBodyV1, 1], MAX_PAYLOAD_BODIES_REQUEST]
```

*Note:* Each inner list has 0 elements for unknown blocks and 1 element for known blocks.

### PayloadBodiesV2Response

Response container for `engine_getPayloadBodiesByHashV2` and `engine_getPayloadBodiesByRangeV2`.

```python
class PayloadBodiesV2Response(Container):
    payload_bodies: List[List[ExecutionPayloadBodyV2, 1], MAX_PAYLOAD_BODIES_REQUEST]
```

*Note:* Each inner list has 0 elements for unknown blocks and 1 element for known blocks.

### GetBlobsV1Response

Response container for `engine_getBlobsV1`.

```python
class GetBlobsV1Response(Container):
    blobs_and_proofs: List[BlobAndProofV1, MAX_BLOB_HASHES_REQUEST]
```

### GetBlobsV2Response

Response container for `engine_getBlobsV2`.

```python
class GetBlobsV2Response(Container):
    blobs_and_proofs: List[BlobAndProofV2, MAX_BLOB_HASHES_REQUEST]
```

### GetBlobsV3Response

Response container for `engine_getBlobsV3`.

```python
class GetBlobsV3Response(Container):
    blobs_and_proofs: List[List[BlobAndProofV2, 1], MAX_BLOB_HASHES_REQUEST]
```

*Note:* Each inner list has 0 elements for a missing blob and 1 element for a present blob.

### ErrorResponse

Used for error responses across all methods.

```python
class ErrorResponse(Container):
    code: uint64
    message: ByteList[MAX_ERROR_MESSAGE_LENGTH]
```

*Note:* Engine API error codes are negative integers in JSON-RPC. The `code` field stores the absolute value. For example, JSON-RPC error code `-38005` is encoded as `38005`.

## Method definitions

Each Engine API method has a dedicated URL path, a request container, and a response type. The request body is the SSZ serialization of the request container. The response body is the SSZ serialization of the response type.

### Paris methods

#### engine_newPayloadV1

`POST /engine/newPayloadV1`

```python
class NewPayloadV1Request(Container):
    execution_payload: ExecutionPayloadV1
```

**Response:** [`PayloadStatusV1`](#payloadstatusv1)

#### engine_forkchoiceUpdatedV1

`POST /engine/forkchoiceUpdatedV1`

```python
class ForkchoiceUpdatedV1Request(Container):
    forkchoice_state: ForkchoiceStateV1
    payload_attributes: Optional[PayloadAttributesV1]
```

**Response:** [`ForkchoiceUpdatedResponseV1`](#forkchoiceupdatedresponsev1)

#### engine_getPayloadV1

`POST /engine/getPayloadV1`

```python
class GetPayloadV1Request(Container):
    payload_id: Bytes8
```

**Response:** [`ExecutionPayloadV1`](#executionpayloadv1)

#### engine_exchangeTransitionConfigurationV1

`POST /engine/exchangeTransitionConfigurationV1`

Deprecated in Cancun.

```python
class ExchangeTransitionConfigurationV1Request(Container):
    transition_configuration: TransitionConfigurationV1
```

**Response:** [`TransitionConfigurationV1`](#transitionconfigurationv1)

### Shanghai methods

#### engine_newPayloadV2

`POST /engine/newPayloadV2`

```python
class NewPayloadV2Request(Container):
    execution_payload: ExecutionPayloadV2
```

*Note:* Always uses `ExecutionPayloadV2`. Pre-Shanghai payloads have an empty `withdrawals` list.

**Response:** [`PayloadStatusV1`](#payloadstatusv1)

#### engine_forkchoiceUpdatedV2

`POST /engine/forkchoiceUpdatedV2`

```python
class ForkchoiceUpdatedV2Request(Container):
    forkchoice_state: ForkchoiceStateV1
    payload_attributes: Optional[PayloadAttributesV2]
```

*Note:* Always uses `PayloadAttributesV2`. Pre-Shanghai attributes have an empty `withdrawals` list.

**Response:** [`ForkchoiceUpdatedResponseV1`](#forkchoiceupdatedresponsev1)

#### engine_getPayloadV2

`POST /engine/getPayloadV2`

```python
class GetPayloadV2Request(Container):
    payload_id: Bytes8
```

**Response:** [`GetPayloadResponseV2`](#getpayloadresponsev2)

#### engine_getPayloadBodiesByHashV1

`POST /engine/getPayloadBodiesByHashV1`

```python
class GetPayloadBodiesByHashV1Request(Container):
    block_hashes: List[Bytes32, MAX_PAYLOAD_BODIES_REQUEST]
```

**Response:** [`PayloadBodiesV1Response`](#payloadbodiesv1response)

#### engine_getPayloadBodiesByRangeV1

`POST /engine/getPayloadBodiesByRangeV1`

```python
class GetPayloadBodiesByRangeV1Request(Container):
    start: uint64
    count: uint64
```

**Response:** [`PayloadBodiesV1Response`](#payloadbodiesv1response)

### Cancun methods

#### engine_newPayloadV3

`POST /engine/newPayloadV3`

```python
class NewPayloadV3Request(Container):
    execution_payload: ExecutionPayloadV3
    expected_blob_versioned_hashes: List[Bytes32, MAX_BLOB_COMMITMENTS_PER_BLOCK]
    parent_beacon_block_root: Bytes32
```

**Response:** [`PayloadStatusV1`](#payloadstatusv1)

#### engine_forkchoiceUpdatedV3

`POST /engine/forkchoiceUpdatedV3`

```python
class ForkchoiceUpdatedV3Request(Container):
    forkchoice_state: ForkchoiceStateV1
    payload_attributes: Optional[PayloadAttributesV3]
```

**Response:** [`ForkchoiceUpdatedResponseV1`](#forkchoiceupdatedresponsev1)

#### engine_getPayloadV3

`POST /engine/getPayloadV3`

```python
class GetPayloadV3Request(Container):
    payload_id: Bytes8
```

**Response:** [`GetPayloadResponseV3`](#getpayloadresponsev3)

#### engine_getBlobsV1

`POST /engine/getBlobsV1`

Deprecated in Osaka.

```python
class GetBlobsV1Request(Container):
    blob_versioned_hashes: List[Bytes32, MAX_BLOB_HASHES_REQUEST]
```

**Response:** [`GetBlobsV1Response`](#getblobsv1response) or HTTP `204 No Content` when syncing.

### Prague methods

#### engine_newPayloadV4

`POST /engine/newPayloadV4`

```python
class NewPayloadV4Request(Container):
    execution_payload: ExecutionPayloadV3
    expected_blob_versioned_hashes: List[Bytes32, MAX_BLOB_COMMITMENTS_PER_BLOCK]
    parent_beacon_block_root: Bytes32
    execution_requests: List[ByteList[MAX_BYTES_PER_TRANSACTION], MAX_EXECUTION_REQUESTS]
```

**Response:** [`PayloadStatusV1`](#payloadstatusv1)

#### engine_getPayloadV4

`POST /engine/getPayloadV4`

```python
class GetPayloadV4Request(Container):
    payload_id: Bytes8
```

**Response:** [`GetPayloadResponseV4`](#getpayloadresponsev4)

### Osaka methods

#### engine_getPayloadV5

`POST /engine/getPayloadV5`

```python
class GetPayloadV5Request(Container):
    payload_id: Bytes8
```

**Response:** [`GetPayloadResponseV5`](#getpayloadresponsev5)

#### engine_getBlobsV2

`POST /engine/getBlobsV2`

```python
class GetBlobsV2Request(Container):
    blob_versioned_hashes: List[Bytes32, MAX_BLOB_HASHES_REQUEST]
```

**Response:** [`GetBlobsV2Response`](#getblobsv2response) or HTTP `204 No Content` when syncing or any blob is missing.

#### engine_getBlobsV3

`POST /engine/getBlobsV3`

```python
class GetBlobsV3Request(Container):
    blob_versioned_hashes: List[Bytes32, MAX_BLOB_HASHES_REQUEST]
```

**Response:** [`GetBlobsV3Response`](#getblobsv3response) or HTTP `204 No Content` when syncing.

### Amsterdam methods

#### engine_newPayloadV5

`POST /engine/newPayloadV5`

```python
class NewPayloadV5Request(Container):
    execution_payload: ExecutionPayloadV4
    expected_blob_versioned_hashes: List[Bytes32, MAX_BLOB_COMMITMENTS_PER_BLOCK]
    parent_beacon_block_root: Bytes32
    execution_requests: List[ByteList[MAX_BYTES_PER_TRANSACTION], MAX_EXECUTION_REQUESTS]
```

**Response:** [`PayloadStatusV1`](#payloadstatusv1)

#### engine_getPayloadV6

`POST /engine/getPayloadV6`

```python
class GetPayloadV6Request(Container):
    payload_id: Bytes8
```

**Response:** [`GetPayloadResponseV6`](#getpayloadresponsev6)

#### engine_forkchoiceUpdatedV4

`POST /engine/forkchoiceUpdatedV4`

```python
class ForkchoiceUpdatedV4Request(Container):
    forkchoice_state: ForkchoiceStateV1
    payload_attributes: Optional[PayloadAttributesV4]
```

**Response:** [`ForkchoiceUpdatedResponseV1`](#forkchoiceupdatedresponsev1)

#### engine_getPayloadBodiesByHashV2

`POST /engine/getPayloadBodiesByHashV2`

```python
class GetPayloadBodiesByHashV2Request(Container):
    block_hashes: List[Bytes32, MAX_PAYLOAD_BODIES_REQUEST]
```

**Response:** [`PayloadBodiesV2Response`](#payloadbodiesv2response)

#### engine_getPayloadBodiesByRangeV2

`POST /engine/getPayloadBodiesByRangeV2`

```python
class GetPayloadBodiesByRangeV2Request(Container):
    start: uint64
    count: uint64
```

**Response:** [`PayloadBodiesV2Response`](#payloadbodiesv2response)

## Example

The following example shows an `engine_newPayloadV5` call using the binary SSZ transport.

### Request

```
POST /engine/newPayloadV5 HTTP/1.1
Host: localhost:8551
Content-Type: application/ssz
Content-Length: 604

<604 bytes: SSZ(NewPayloadV5Request)>
```

The request body is the SSZ serialization of `NewPayloadV5Request` containing:
- `execution_payload`: an `ExecutionPayloadV4` with empty transactions, withdrawals, and block access list
- `expected_blob_versioned_hashes`: empty list
- `parent_beacon_block_root`: `0x0000000000000000000000000000000000000000000000000000000000000000`
- `execution_requests`: empty list

### Response (success)

```
HTTP/1.1 200 OK
Content-Type: application/ssz
Content-Length: 69

<69 bytes: SSZ(PayloadStatusV1)>
```

The response body is the SSZ serialization of `PayloadStatusV1` containing:
- `status`: `0` (VALID)
- `latest_valid_hash`: `0x3559e851470f6e7bbed1db474980683e8c315bfce99b2a6ef47c057c04de7858`
- `validation_error`: empty

### Response (error)

```
HTTP/1.1 400 Bad Request
Content-Type: application/ssz
Content-Length: 48

<48 bytes: SSZ(ErrorResponse)>
```

The response body is the SSZ serialization of `ErrorResponse` containing:
- `code`: `32602` (absolute value of `-32602`, invalid params)
- `message`: `"Invalid execution payload"`

### Fallback behavior

If the EL does not support the binary SSZ transport, a request to `/engine/newPayloadV5` returns HTTP `404 Not Found` or `415 Unsupported Media Type`. The CL detects this and falls back to JSON-RPC at `POST /` for subsequent requests.

## Error handling

Binary SSZ does not change the error semantics of the Engine API. All error codes defined in the [Errors](./common.md#errors) section apply equally.

Error responses use the [`ErrorResponse`](#errorresponse) container. The HTTP status code reflects the error category:

| HTTP Status | Meaning | Engine API Errors |
| - | - | - |
| `400` | Client error | `-32700` (parse error), `-32600` (invalid request), `-32602` (invalid params) |
| `404` | Method not found | `-32601` (method not found) |
| `415` | Unsupported media type | Binary SSZ not supported |
| `500` | Server error | `-32603` (internal error), `-38001` to `-38005` (engine-specific errors) |

Clients **MUST** validate SSZ payloads against the expected schema before processing. Payloads that do not conform to the expected SSZ schema **MUST** be rejected with a `400` response containing an `ErrorResponse` with code `32700`.

## Security considerations

- SSZ deserialization **MUST** enforce the same size limits as JSON deserialization. Implementations **MUST** reject SSZ payloads exceeding defined maximum sizes before attempting full deserialization.
- Implementations **SHOULD** use well-tested SSZ libraries and fuzz test SSZ parsing extensively.
- The binary transport uses the same JWT authentication as the JSON-RPC endpoint. All existing authentication requirements apply.
