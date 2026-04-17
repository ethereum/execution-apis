# Engine API -- HTTP REST bindings

This document specifies optional **HTTP REST** endpoints for a subset of the Engine API. They are exposed on the same authenticated Engine HTTP server and port as JSON-RPC (see [Common Definitions](./common.md)).

The Engine API's primary interface is JSON-RPC. This document introduces an **optional** REST endpoint, `POST /new-payload-with-witness`, that performs the same payload validation and execution as [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5) but additionally returns an execution witness. The endpoint accepts the same JSON parameters as `engine_newPayloadV5` and returns the response as SSZ-encoded bytes. All validation rules, processing logic, and error semantics are inherited from `engine_newPayloadV5` unless explicitly stated otherwise below.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Structures](#structures)
  - [SSZ PayloadStatusV1](#ssz-payloadstatusv1)
  - [SSZ ExecutionWitnessV1](#ssz-executionwitnessv1)
  - [SSZ NewPayloadWithWitnessResponseV1](#ssz-newpayloadwithwitnessresponsev1)
- [Endpoints](#endpoints)
  - [new-payload-with-witness](#new-payload-with-witness)
    - [Request](#request)
    - [Successful response](#successful-response)
    - [Error responses](#error-responses)
    - [Specification](#specification)
- [Capabilities](#capabilities)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Structures

### SSZ PayloadStatusV1

This type is an SSZ `Container` encoding the same logical object as JSON-RPC [`PayloadStatusV1`](./paris.md#payloadstatusv1).

| Index | Field name | SSZ type |
| ----- | ---------- | -------- |
| 0 | `status` | `uint8` |
| 1 | `latest_valid_hash` | `Union[None, ByteVector[32]]` |
| 2 | `validation_error` | `Union[None, List[uint8, VALIDATION_ERROR_MAX]]` |

Constants:

* `VALIDATION_ERROR_MAX`: `8192` (maximum number of `uint8` elements in the list when the `validation_error` union variant is present).

`uint8` values for `status` **MUST** be:

| Value | JSON-RPC `status` string |
| ----- | ------------------------ |
| `0` | `VALID` |
| `1` | `INVALID` |
| `2` | `SYNCING` |
| `3` | `ACCEPTED` |
| `4` | `INVALID_BLOCK_HASH` |

Mapping from JSON-RPC field values to SSZ:

* `latestValidHash: null` **MUST** be encoded as the `None` variant of `latest_valid_hash`.
* `latestValidHash: <32-byte DATA>` **MUST** be encoded as the `ByteVector[32]` variant (raw bytes, **not** hex).
* `validationError: null` **MUST** be encoded as the `None` variant of `validation_error`.
* `validationError: <string>` **MUST** be encoded as UTF-8 bytes in the `List[uint8, VALIDATION_ERROR_MAX]` variant. If the UTF-8 encoding exceeds `VALIDATION_ERROR_MAX` bytes, client software **MUST** truncate to `VALIDATION_ERROR_MAX` bytes without splitting a multi-byte UTF-8 code point (i.e. truncate to the longest prefix that is valid UTF-8 and fits the limit).

### SSZ ExecutionWitnessV1

This type is an SSZ `Container` encoding the execution witness produced during block processing. It contains the raw Merkle trie nodes, contract bytecodes, and block headers required to statelessly verify the block's execution.

| Index | Field name | SSZ type |
| ----- | ---------- | -------- |
| 0 | `state` | `List[List[uint8, MAX_WITNESS_ITEM_BYTES], MAX_WITNESS_ITEMS]` |
| 1 | `codes` | `List[List[uint8, MAX_WITNESS_ITEM_BYTES], MAX_WITNESS_ITEMS]` |
| 2 | `headers` | `List[List[uint8, MAX_WITNESS_ITEM_BYTES], MAX_WITNESS_ITEMS]` |

Constants:

* `MAX_WITNESS_ITEMS`: `1048576` — maximum number of items per witness field.
* `MAX_WITNESS_ITEM_BYTES`: `1048576` — maximum byte length of a single witness item.

For detailed field semantics and specific encoding requirements, refer to the [Execution Witness specification](https://github.com/ethereum/execution-specs/blob/e2ff4cc00ca94cb5872090e0f813894e231f6be3/src/ethereum/forks/amsterdam/stateless.py#L27-L47).

An empty list (`[]`) for any field indicates no data of that category was accessed during execution.

### SSZ NewPayloadWithWitnessResponseV1

This type is an SSZ `Container` combining the payload validation result and the execution witness.

| Index | Field name | SSZ type |
| ----- | ---------- | -------- |
| 0 | `status` | `uint8` |
| 1 | `latest_valid_hash` | `Union[None, ByteVector[32]]` |
| 2 | `validation_error` | `Union[None, List[uint8, VALIDATION_ERROR_MAX]]` |
| 3 | `witness` | `Union[None, ExecutionWitnessV1]` |

Constants:

* `VALIDATION_ERROR_MAX`: as defined in [SSZ PayloadStatusV1](#ssz-payloadstatusv1).

Fields `status`, `latest_valid_hash`, and `validation_error` carry the same semantics as [SSZ PayloadStatusV1](#ssz-payloadstatusv1).

The `witness` field is a `Union` type: the `None` variant **MUST** be used when the `status` is not `VALID` or no witness was produced; the `ExecutionWitnessV1` variant **MUST** be used when the `status` is `VALID` and a witness was generated.

Serialization of `Container`, `Union`, `List`, and `ByteVector` types **MUST** follow the Ethereum consensus [Simple Serialize (SSZ) specification][ssz].

[ssz]: https://github.com/ethereum/consensus-specs/blob/master/ssz/simple-serialize.md

## Endpoints

### new-payload-with-witness

This endpoint **MUST** implement the same payload validation and execution logic as JSON-RPC [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5). In addition, when execution succeeds with `VALID` status, the response **MUST** include the execution witness.

#### Request

* HTTP method: `POST`
* Path: `/new-payload-with-witness`
* Header `Content-Type` **MUST** be `application/json`.
* Body: a JSON array of exactly **four** elements, in order:
  1. `executionPayload`: [`ExecutionPayloadV4`](./amsterdam.md#executionpayloadv4)
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes each
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes
  4. `executionRequests`: `Array of DATA`

  These parameters are as defined for [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5).
* timeout: 8s

#### Successful response

* HTTP status: `200 OK`
* Header `Content-Type` **MUST** be `application/octet-stream`
* Body: the SSZ serialization of [`NewPayloadWithWitnessResponseV1`](#ssz-newpayloadwithwitnessresponsev1) as defined above.

#### Error responses

Unless otherwise specified, error bodies **MUST** use `Content-Type: application/json` and **MUST** be a JSON object with at least:

* `code`: `integer` — same semantics as JSON-RPC `error.code` for Engine API errors (see [Errors](./common.md#errors)).
* `message`: `string` — same semantics as JSON-RPC `error.message`.

| Condition | HTTP status | Notes |
| --------- | ----------- | ----- |
| Missing or invalid JWT | `401` | Per [Authentication](./authentication.md). |
| Malformed JSON body, or body is not a JSON array | `400` | e.g. `-32700` Parse error where applicable. |
| Wrong number of parameters or invalid parameter shapes | `400` | `-32602` Invalid params |
| Unsupported fork or other Engine errors that JSON-RPC surfaces as `error` | `400` or `500` as appropriate | Same `code` as JSON-RPC (e.g. `-38005` Unsupported fork). |
| HTTP method other than `POST` | `405` | |

When JSON-RPC would return a **result** `PayloadStatusV1` (including `INVALID`, `SYNCING`, etc.), this endpoint **MUST** respond with `200 OK` and an SSZ body — **not** an HTTP error — because those outcomes are part of normal payload processing.

#### Specification

This endpoint follows the same specification as [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5) with the following additions:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload does not fall within the time frame of the Amsterdam activation.

2. When the payload status is `VALID`, the EL **MUST** include the execution witness in the `witness` field of the response. The witness **MUST** be the SSZ serialization of an `ExecutionWitnessV1` containing all state data accessed during block execution.

3. When the payload status is not `VALID` (e.g. `INVALID`, `SYNCING`, `ACCEPTED`), the `witness` field **MUST** be empty (`[]`).


## Capabilities

Execution layer client software that implements the endpoint in this document **SHOULD** include the following string in the response array of [`engine_exchangeCapabilities`](./common.md#engine_exchangecapabilities) when the corresponding route is supported:

* `rest_engine_newPayloadWithWitness` — `POST /new-payload-with-witness` is available

Consensus layer client software **MAY** use this to discover REST support. This capability name is **not** a JSON-RPC method name; it identifies an optional REST feature.
