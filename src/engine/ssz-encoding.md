# Engine API -- Binary SSZ Transport

This document specifies a binary SSZ transport for Engine API communication between consensus layer (CL) and execution layer (EL) clients. The binary transport replaces JSON-RPC with resource-oriented REST and raw SSZ encoding for fast, efficient CL-EL communication.

SSZ container definitions are provided for all Engine API structures and methods across all forks for backwards compatibility.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Motivation](#motivation)
- [Transport](#transport)
  - [Base URL](#base-url)
  - [Content types](#content-types)
  - [Authentication](#authentication)
  - [Versioning](#versioning)
  - [Negotiation and fallback](#negotiation-and-fallback)
- [HTTP status codes](#http-status-codes)
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
  - [ClientVersionV1](#clientversionv1)
  - [GetClientVersionV1Response](#getclientversionv1response)
  - [ExchangeCapabilitiesResponse](#exchangecapabilitiesresponse)
- [Endpoints](#endpoints)
  - [Payloads](#payloads)
  - [Forkchoice](#forkchoice)
  - [Blobs](#blobs)
  - [Client](#client)
  - [Transition configuration](#transition-configuration)
  - [Endpoint summary](#endpoint-summary)
- [Example](#example)
- [Security considerations](#security-considerations)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Motivation

Fast communication between the consensus layer and execution layer is critical for block propagation and validation timing. The JSON-RPC transport introduces unnecessary overhead in this critical path:

- Binary data (hashes, addresses, transactions, blobs) is hex-encoded, doubling wire size.
- JSON parsing and generation adds CPU overhead on both sides.
- The CL uses SSZ natively, forcing a round-trip conversion (SSZ to JSON, then JSON to internal types) at the Engine API boundary.

Binary SSZ eliminates all of this. The CL sends raw SSZ bytes over HTTP; the EL deserializes directly. No hex encoding, no JSON parsing, no intermediate representations. Payload sizes are reduced by 50% or more compared to JSON-RPC, and serialization is no longer a bottleneck in the critical path between CL and EL.

## Transport

The binary SSZ transport uses resource-oriented REST over HTTP. Endpoints are organized by resource type (payloads, forkchoice, blobs) with per-endpoint versioning, following the same conventions as the [Beacon API](https://github.com/ethereum/beacon-APIs).

### Base URL

```
http://localhost:8551/engine
```

All endpoints are served under the `/engine` prefix on the existing Engine API port (8551).

### Content types

| Header | Value | Description |
| - | - | - |
| `Content-Type` (request) | `application/octet-stream` | SSZ-encoded request container |
| `Content-Type` (response) | `application/octet-stream` | SSZ-encoded response (success) |
| `Content-Type` (response) | `text/plain` | Human-readable error message |
| `Accept` (request) | `application/octet-stream` | Client accepts SSZ-encoded responses |

Request bodies are the SSZ serialization of the endpoint's request container. Response bodies are the SSZ serialization of the endpoint's response type. GET requests with no body **SHOULD** include the `Accept` header to indicate SSZ preference.

### Authentication

The binary transport uses the same JWT authentication as the JSON-RPC endpoint. All requests **MUST** include a valid JWT bearer token in the `Authorization` header:

```
Authorization: Bearer <JWT token>
```

All existing authentication requirements from the [Engine API specification](./common.md#authentication) apply.

### Versioning

Endpoints use path-based versioning following [Beacon API](https://github.com/ethereum/beacon-APIs) conventions. Each endpoint includes a version number in its path (e.g., `/engine/v5/payloads`). The version number corresponds to the JSON-RPC method version it replaces:

| REST Endpoint | JSON-RPC Equivalent |
| - | - |
| `POST /engine/v5/payloads` | `engine_newPayloadV5` |
| `GET /engine/v6/payloads/{payload_id}` | `engine_getPayloadV6` |
| `POST /engine/v4/forkchoice` | `engine_forkchoiceUpdatedV4` |
| `POST /engine/v3/blobs` | `engine_getBlobsV3` |

When a new fork introduces a new method version, a new versioned endpoint is added. Older versioned endpoints **MAY** be deprecated but **SHOULD** remain available for backwards compatibility.

### Negotiation and fallback

1. The CL sends a request to a versioned REST endpoint with `Content-Type: application/octet-stream` and a raw SSZ request body.

2. If the EL supports the binary SSZ transport, it **MUST** respond with `Content-Type: application/octet-stream` and a raw SSZ response body.

3. If the EL does not support the binary SSZ transport, it **MUST** respond with HTTP status `404 Not Found`. The CL **MUST** then fall back to JSON-RPC (`POST /`) for subsequent requests.

4. Clients **MUST** continue to support JSON-RPC encoding as a fallback. Both the REST endpoints and the JSON-RPC endpoint coexist on the same port.

## HTTP status codes

### Success

| Status | Meaning | Usage |
| - | - | - |
| `200` | OK | Request succeeded, response body contains SSZ-encoded result |
| `204` | No Content | Null result (e.g., syncing), empty body |

### Client errors

| Status | Meaning | Usage |
| - | - | - |
| `400` | Bad Request | Malformed SSZ encoding |
| `401` | Unauthorized | Missing or invalid JWT token |
| `404` | Not Found | Unknown payload ID |
| `409` | Conflict | Invalid forkchoice state |
| `413` | Request Too Large | Request exceeds maximum element count |
| `422` | Unprocessable Entity | Invalid payload attributes |

### Server errors

| Status | Meaning | Usage |
| - | - | - |
| `500` | Internal Server Error | Unexpected server error |

Error responses use `Content-Type: text/plain` with a human-readable error message body.

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
| `MAX_CLIENT_CODE_LENGTH` | `2` | This specification |
| `MAX_CLIENT_NAME_LENGTH` | `64` | This specification |
| `MAX_CLIENT_VERSION_LENGTH` | `64` | This specification |
| `MAX_CLIENT_VERSIONS` | `4` | This specification |
| `MAX_CAPABILITY_NAME_LENGTH` | `64` | This specification |
| `MAX_CAPABILITIES` | `64` | This specification |
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
| `T or null` | `List[T, 1]` |

Nullable types are represented as `List[T, 1]` in SSZ encoding. An empty list (0 elements) denotes absence (`null`). A list with one element denotes presence.

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

### ClientVersionV1

Introduced in [Client Version Specification](./identification.md#clientversionv1).

```python
class ClientVersionV1(Container):
    code: ByteList[MAX_CLIENT_CODE_LENGTH]
    name: ByteList[MAX_CLIENT_NAME_LENGTH]
    version: ByteList[MAX_CLIENT_VERSION_LENGTH]
    commit: Bytes4
```

### GetClientVersionV1Response

Response container for `engine_getClientVersionV1`.

```python
class GetClientVersionV1Response(Container):
    versions: List[ClientVersionV1, MAX_CLIENT_VERSIONS]
```

### ExchangeCapabilitiesResponse

Response container for `engine_exchangeCapabilities`.

```python
class ExchangeCapabilitiesResponse(Container):
    capabilities: List[ByteList[MAX_CAPABILITY_NAME_LENGTH], MAX_CAPABILITIES]
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

## Endpoints

All endpoints use `Content-Type: application/octet-stream` for request and response bodies containing SSZ-encoded data. Error responses use `Content-Type: text/plain`.

### Payloads

#### `POST /engine/v{N}/payloads` — Submit execution payload

Submit an execution payload for validation. The EL validates the payload and returns its status.

| Version | Fork | Request Container | JSON-RPC Equivalent |
| - | - | - | - |
| v1 | Paris | `NewPayloadV1Request` | `engine_newPayloadV1` |
| v2 | Shanghai | `NewPayloadV2Request` | `engine_newPayloadV2` |
| v3 | Cancun | `NewPayloadV3Request` | `engine_newPayloadV3` |
| v4 | Prague | `NewPayloadV4Request` | `engine_newPayloadV4` |
| v5 | Amsterdam | `NewPayloadV5Request` | `engine_newPayloadV5` |

**Request containers:**

```python
class NewPayloadV1Request(Container):
    execution_payload: ExecutionPayloadV1

class NewPayloadV2Request(Container):
    execution_payload: ExecutionPayloadV2

class NewPayloadV3Request(Container):
    execution_payload: ExecutionPayloadV3
    expected_blob_versioned_hashes: List[Bytes32, MAX_BLOB_COMMITMENTS_PER_BLOCK]
    parent_beacon_block_root: Bytes32

class NewPayloadV4Request(Container):
    execution_payload: ExecutionPayloadV3
    expected_blob_versioned_hashes: List[Bytes32, MAX_BLOB_COMMITMENTS_PER_BLOCK]
    parent_beacon_block_root: Bytes32
    execution_requests: List[ByteList[MAX_BYTES_PER_TRANSACTION], MAX_EXECUTION_REQUESTS]

class NewPayloadV5Request(Container):
    execution_payload: ExecutionPayloadV4
    expected_blob_versioned_hashes: List[Bytes32, MAX_BLOB_COMMITMENTS_PER_BLOCK]
    parent_beacon_block_root: Bytes32
    execution_requests: List[ByteList[MAX_BYTES_PER_TRANSACTION], MAX_EXECUTION_REQUESTS]
```

*Note:* `NewPayloadV2Request` always uses `ExecutionPayloadV2`. Pre-Shanghai payloads have an empty `withdrawals` list.

**Response:** `200 OK` — [`PayloadStatusV1`](#payloadstatusv1)

**Errors:**

| Status | Condition |
| - | - |
| `400` | Malformed SSZ encoding |

---

#### `GET /engine/v{N}/payloads/{payload_id}` — Retrieve built payload

Retrieve an execution payload previously requested via forkchoice update with payload attributes. The `{payload_id}` path parameter is the hex-encoded `Bytes8` payload identifier (e.g., `0x1234567890abcdef`).

This is a safe, idempotent GET operation. The EL may continue optimizing the payload until the slot deadline.

| Version | Fork | Response Type | JSON-RPC Equivalent |
| - | - | - | - |
| v1 | Paris | `ExecutionPayloadV1` | `engine_getPayloadV1` |
| v2 | Shanghai | `GetPayloadResponseV2` | `engine_getPayloadV2` |
| v3 | Cancun | `GetPayloadResponseV3` | `engine_getPayloadV3` |
| v4 | Prague | `GetPayloadResponseV4` | `engine_getPayloadV4` |
| v5 | Osaka | `GetPayloadResponseV5` | `engine_getPayloadV5` |
| v6 | Amsterdam | `GetPayloadResponseV6` | `engine_getPayloadV6` |

**Request:** No body. The payload ID is in the URL path.

**Response:** `200 OK` — SSZ-encoded response type from the table above.

**Errors:**

| Status | Condition |
| - | - |
| `400` | Invalid payload ID format |
| `404` | Unknown payload ID |

---

#### `POST /engine/v{N}/payloads/bodies/by-hash` — Get payload bodies by hash

Retrieve execution payload bodies for a list of block hashes. Used for historical sync and block reconstruction.

| Version | Fork | Request Container | Response Type | JSON-RPC Equivalent |
| - | - | - | - | - |
| v1 | Shanghai | `GetPayloadBodiesByHashV1Request` | `PayloadBodiesV1Response` | `engine_getPayloadBodiesByHashV1` |
| v2 | Amsterdam | `GetPayloadBodiesByHashV2Request` | `PayloadBodiesV2Response` | `engine_getPayloadBodiesByHashV2` |

**Request containers:**

```python
class GetPayloadBodiesByHashV1Request(Container):
    block_hashes: List[Bytes32, MAX_PAYLOAD_BODIES_REQUEST]

class GetPayloadBodiesByHashV2Request(Container):
    block_hashes: List[Bytes32, MAX_PAYLOAD_BODIES_REQUEST]
```

**Response:** `200 OK` — [`PayloadBodiesV1Response`](#payloadbodiesv1response) or [`PayloadBodiesV2Response`](#payloadbodiesv2response)

**Errors:**

| Status | Condition |
| - | - |
| `400` | Malformed SSZ encoding |
| `413` | Request exceeds `MAX_PAYLOAD_BODIES_REQUEST` hashes |

---

#### `POST /engine/v{N}/payloads/bodies/by-range` — Get payload bodies by range

Retrieve execution payload bodies for a contiguous range of block numbers.

| Version | Fork | Request Container | Response Type | JSON-RPC Equivalent |
| - | - | - | - | - |
| v1 | Shanghai | `GetPayloadBodiesByRangeV1Request` | `PayloadBodiesV1Response` | `engine_getPayloadBodiesByRangeV1` |
| v2 | Amsterdam | `GetPayloadBodiesByRangeV2Request` | `PayloadBodiesV2Response` | `engine_getPayloadBodiesByRangeV2` |

**Request containers:**

```python
class GetPayloadBodiesByRangeV1Request(Container):
    start: uint64
    count: uint64

class GetPayloadBodiesByRangeV2Request(Container):
    start: uint64
    count: uint64
```

**Response:** `200 OK` — [`PayloadBodiesV1Response`](#payloadbodiesv1response) or [`PayloadBodiesV2Response`](#payloadbodiesv2response)

**Errors:**

| Status | Condition |
| - | - |
| `400` | Malformed SSZ encoding |
| `413` | `count` exceeds `MAX_PAYLOAD_BODIES_REQUEST` |

### Forkchoice

#### `POST /engine/v{N}/forkchoice` — Update fork choice

Update the EL's fork choice state and optionally start building a new payload. The EL updates its canonical chain view and prunes blocks no longer reachable from the head.

When `payload_attributes` is present (list with 1 element), the EL begins building a new block. The returned `payload_id` can be used with `GET /engine/v{N}/payloads/{payload_id}` to retrieve the built payload.

| Version | Fork | Request Container | JSON-RPC Equivalent |
| - | - | - | - |
| v1 | Paris | `ForkchoiceUpdatedV1Request` | `engine_forkchoiceUpdatedV1` |
| v2 | Shanghai | `ForkchoiceUpdatedV2Request` | `engine_forkchoiceUpdatedV2` |
| v3 | Cancun | `ForkchoiceUpdatedV3Request` | `engine_forkchoiceUpdatedV3` |
| v4 | Amsterdam | `ForkchoiceUpdatedV4Request` | `engine_forkchoiceUpdatedV4` |

**Request containers:**

```python
class ForkchoiceUpdatedV1Request(Container):
    forkchoice_state: ForkchoiceStateV1
    payload_attributes: List[PayloadAttributesV1, 1]

class ForkchoiceUpdatedV2Request(Container):
    forkchoice_state: ForkchoiceStateV1
    payload_attributes: List[PayloadAttributesV2, 1]

class ForkchoiceUpdatedV3Request(Container):
    forkchoice_state: ForkchoiceStateV1
    payload_attributes: List[PayloadAttributesV3, 1]

class ForkchoiceUpdatedV4Request(Container):
    forkchoice_state: ForkchoiceStateV1
    payload_attributes: List[PayloadAttributesV4, 1]
```

*Note:* `ForkchoiceUpdatedV2Request` always uses `PayloadAttributesV2`. Pre-Shanghai attributes have an empty `withdrawals` list.

**Response:** `200 OK` — [`ForkchoiceUpdatedResponseV1`](#forkchoiceupdatedresponsev1)

**Errors:**

| Status | Condition |
| - | - |
| `400` | Malformed SSZ encoding |
| `409` | Invalid forkchoice state |
| `422` | Invalid payload attributes |

### Blobs

#### `POST /engine/v{N}/blobs` — Get blobs by versioned hash

Retrieve blobs from the EL's blob pool by their versioned hashes.

| Version | Fork | Request Container | Response Type | JSON-RPC Equivalent |
| - | - | - | - | - |
| v1 | Cancun | `GetBlobsV1Request` | `GetBlobsV1Response` | `engine_getBlobsV1` |
| v2 | Osaka | `GetBlobsV2Request` | `GetBlobsV2Response` | `engine_getBlobsV2` |
| v3 | Osaka | `GetBlobsV3Request` | `GetBlobsV3Response` | `engine_getBlobsV3` |

**Request containers:**

```python
class GetBlobsV1Request(Container):
    blob_versioned_hashes: List[Bytes32, MAX_BLOB_HASHES_REQUEST]

class GetBlobsV2Request(Container):
    blob_versioned_hashes: List[Bytes32, MAX_BLOB_HASHES_REQUEST]

class GetBlobsV3Request(Container):
    blob_versioned_hashes: List[Bytes32, MAX_BLOB_HASHES_REQUEST]
```

**Response:** `200 OK` — SSZ-encoded response type from the table above, or `204 No Content` when the EL is syncing (or for v2, when any blob is missing).

*Note:* `GetBlobsV3Response` uses `List[BlobAndProofV2, 1]` inner lists for per-element nullability (0 elements = missing, 1 element = present). The whole-result null (syncing) uses HTTP `204`.

**Errors:**

| Status | Condition |
| - | - |
| `400` | Malformed SSZ encoding |
| `413` | Request exceeds `MAX_BLOB_HASHES_REQUEST` hashes |

### Client

#### `POST /engine/v1/client/version` — Exchange client version

Exchange client version information between CL and EL. The CL identifies itself in the request; the EL returns its own version(s) in the response. See the [Client Version Specification](./identification.md) for details.

**Request container:**

```python
class GetClientVersionV1Request(Container):
    client_version: ClientVersionV1
```

**Response:** `200 OK` — [`GetClientVersionV1Response`](#getclientversionv1response)

**Errors:**

| Status | Condition |
| - | - |
| `400` | Malformed SSZ encoding |

---

#### `POST /engine/v1/capabilities` — Exchange capabilities

Exchange the list of supported Engine API endpoints between CL and EL. Capability names use the format `"METHOD /path"` (e.g., `"POST /engine/v5/payloads"`). See the [Capabilities specification](./common.md#capabilities) for details.

**Request container:**

```python
class ExchangeCapabilitiesRequest(Container):
    capabilities: List[ByteList[MAX_CAPABILITY_NAME_LENGTH], MAX_CAPABILITIES]
```

**Response:** `200 OK` — [`ExchangeCapabilitiesResponse`](#exchangecapabilitiesresponse)

**Errors:**

| Status | Condition |
| - | - |
| `400` | Malformed SSZ encoding |

### Transition configuration

#### `POST /engine/v1/transition-configuration` — Exchange transition configuration

Deprecated in Cancun. Exchange PoW-to-PoS transition configuration between CL and EL.

**Request container:**

```python
class ExchangeTransitionConfigurationV1Request(Container):
    transition_configuration: TransitionConfigurationV1
```

**Response:** `200 OK` — [`TransitionConfigurationV1`](#transitionconfigurationv1)

### Endpoint summary

All endpoints organized by resource and fork:

| HTTP Method | Path | Fork | JSON-RPC Equivalent |
| - | - | - | - |
| `POST` | `/engine/v1/payloads` | Paris | `engine_newPayloadV1` |
| `POST` | `/engine/v2/payloads` | Shanghai | `engine_newPayloadV2` |
| `POST` | `/engine/v3/payloads` | Cancun | `engine_newPayloadV3` |
| `POST` | `/engine/v4/payloads` | Prague | `engine_newPayloadV4` |
| `POST` | `/engine/v5/payloads` | Amsterdam | `engine_newPayloadV5` |
| `GET` | `/engine/v1/payloads/{payload_id}` | Paris | `engine_getPayloadV1` |
| `GET` | `/engine/v2/payloads/{payload_id}` | Shanghai | `engine_getPayloadV2` |
| `GET` | `/engine/v3/payloads/{payload_id}` | Cancun | `engine_getPayloadV3` |
| `GET` | `/engine/v4/payloads/{payload_id}` | Prague | `engine_getPayloadV4` |
| `GET` | `/engine/v5/payloads/{payload_id}` | Osaka | `engine_getPayloadV5` |
| `GET` | `/engine/v6/payloads/{payload_id}` | Amsterdam | `engine_getPayloadV6` |
| `POST` | `/engine/v1/payloads/bodies/by-hash` | Shanghai | `engine_getPayloadBodiesByHashV1` |
| `POST` | `/engine/v2/payloads/bodies/by-hash` | Amsterdam | `engine_getPayloadBodiesByHashV2` |
| `POST` | `/engine/v1/payloads/bodies/by-range` | Shanghai | `engine_getPayloadBodiesByRangeV1` |
| `POST` | `/engine/v2/payloads/bodies/by-range` | Amsterdam | `engine_getPayloadBodiesByRangeV2` |
| `POST` | `/engine/v1/forkchoice` | Paris | `engine_forkchoiceUpdatedV1` |
| `POST` | `/engine/v2/forkchoice` | Shanghai | `engine_forkchoiceUpdatedV2` |
| `POST` | `/engine/v3/forkchoice` | Cancun | `engine_forkchoiceUpdatedV3` |
| `POST` | `/engine/v4/forkchoice` | Amsterdam | `engine_forkchoiceUpdatedV4` |
| `POST` | `/engine/v1/blobs` | Cancun | `engine_getBlobsV1` |
| `POST` | `/engine/v2/blobs` | Osaka | `engine_getBlobsV2` |
| `POST` | `/engine/v3/blobs` | Osaka | `engine_getBlobsV3` |
| `POST` | `/engine/v1/client/version` | All | `engine_getClientVersionV1` |
| `POST` | `/engine/v1/capabilities` | All | `engine_exchangeCapabilities` |
| `POST` | `/engine/v1/transition-configuration` | Paris | `engine_exchangeTransitionConfigurationV1` |

## Example

The following example shows an `engine_newPayloadV5` call using the binary SSZ transport.

### Submit payload

```bash
curl -X POST http://localhost:8551/engine/v5/payloads \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/octet-stream" \
  -H "Accept: application/octet-stream" \
  --data-binary @new_payload_request.ssz \
  -o payload_status.ssz
```

**Request:**

```
POST /engine/v5/payloads HTTP/1.1
Host: localhost:8551
Authorization: Bearer $JWT_TOKEN
Content-Type: application/octet-stream
Content-Length: 584

<584 bytes: SSZ(NewPayloadV5Request)>
```

The request body is the SSZ serialization of `NewPayloadV5Request` containing:
- `execution_payload`: an `ExecutionPayloadV4` with empty transactions, withdrawals, and block access list
- `expected_blob_versioned_hashes`: empty list
- `parent_beacon_block_root`: `0x0000000000000000000000000000000000000000000000000000000000000000`
- `execution_requests`: empty list

**Response (success):**

```
HTTP/1.1 200 OK
Content-Type: application/octet-stream
Content-Length: 37

<37 bytes: SSZ(PayloadStatusV1)>
```

The response body is the SSZ serialization of `PayloadStatusV1` containing:
- `status`: `0` (VALID)
- `latest_valid_hash`: `0x3559e851470f6e7bbed1db474980683e8c315bfce99b2a6ef47c057c04de7858`
- `validation_error`: empty

**Response (error):**

```
HTTP/1.1 400 Bad Request
Content-Type: text/plain

Invalid SSZ: unexpected end of input at offset 128
```

### Retrieve built payload

```bash
curl -X GET http://localhost:8551/engine/v6/payloads/0x1234567890abcdef \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Accept: application/octet-stream" \
  -o get_payload_response.ssz
```

### Update fork choice

```bash
curl -X POST http://localhost:8551/engine/v4/forkchoice \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/octet-stream" \
  --data-binary @forkchoice_request.ssz \
  -o forkchoice_response.ssz
```

### Fallback behavior

If the EL does not support the binary SSZ transport, a request to `/engine/v5/payloads` returns HTTP `404 Not Found`. The CL detects this and falls back to JSON-RPC at `POST /` for subsequent requests.

## Security considerations

- SSZ deserialization **MUST** enforce the same size limits as JSON deserialization. Implementations **MUST** reject SSZ payloads exceeding defined maximum sizes before attempting full deserialization.
- Implementations **SHOULD** use well-tested SSZ libraries and fuzz test SSZ parsing extensively.
- The binary transport uses the same JWT authentication as the JSON-RPC endpoint. All existing authentication requirements apply.
- The `{payload_id}` path parameter **MUST** be validated as a well-formed hex-encoded `Bytes8` before processing.
