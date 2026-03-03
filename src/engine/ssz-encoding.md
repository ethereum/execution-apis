# Engine API -- SSZ Encoding

This document specifies an optional SSZ encoding for Engine API payloads as an alternative to the default JSON encoding. SSZ encoding reduces serialization overhead and aligns the Engine API with the native encoding format used by the consensus layer.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Motivation](#motivation)
- [Encoding negotiation](#encoding-negotiation)
- [SSZ type mappings](#ssz-type-mappings)
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

## SSZ type mappings

Each JSON-encoded base type used in the Engine API maps to a specific SSZ type. The mappings below correspond to the types defined in the [base types schema](../schemas/base-types.yaml).

### Fixed-size types

| JSON Type | Size | SSZ Type |
| - | - | - |
| `address` | 20 bytes | `Bytes20` |
| `hash32` | 32 bytes | `Bytes32` |
| `bytes8` | 8 bytes | `Bytes8` |
| `bytes32` | 32 bytes | `Bytes32` |
| `bytes48` | 48 bytes | `Bytes48` |
| `bytes65` | 65 bytes | `Bytes65` |
| `bytes96` | 96 bytes | `Bytes96` |
| `bytes256` | 256 bytes | `ByteVector[256]` |
| `uint64` | 8 bytes | `uint64` |
| `uint256` | 32 bytes | `uint256` |
| `BOOLEAN` | 1 byte | `boolean` |

### Variable-size types

| JSON Type | SSZ Type |
| - | - |
| `bytes` (variable-length hex data) | `ByteList[MAX_LENGTH]` where `MAX_LENGTH` is context-dependent |
| `bytesMax32` (up to 32 bytes hex data) | `ByteList[32]` |

### Composite types

| JSON Type | SSZ Type |
| - | - |
| `Array of T` | `List[T, MAX_LENGTH]` where `MAX_LENGTH` is context-dependent |
| Object (e.g. `ExecutionPayloadV1`) | `Container` with fields mapped per this table |

### Nullable fields

Fields that may be `null` in the JSON encoding (e.g. `latestValidHash` in `PayloadStatusV1`, `withdrawals` in `ExecutionPayloadBodyV1`) are represented using `Optional[T]` in SSZ.

## Request and response format

SSZ-encoded Engine API requests and responses follow the existing JSON-RPC method semantics. The SSZ encoding applies to the method parameters and result values — the JSON-RPC envelope (`jsonrpc`, `id`, `method`) remains JSON-encoded.

Specifically, when SSZ encoding is in use:

1. The HTTP request body is a JSON-RPC request where each element of `params` is replaced with its SSZ-encoded hexadecimal representation (a `DATA` string).

2. The HTTP response body is a JSON-RPC response where the `result` field is replaced with the SSZ-encoded hexadecimal representation of the result value.

This approach preserves compatibility with JSON-RPC tooling while encoding the payload data in SSZ.

*Note:* Future versions of this specification may define a fully binary request/response format that replaces the JSON-RPC envelope.

## Example

The following example shows an `engine_newPayloadV1` call with a minimal payload, first using JSON encoding and then using SSZ encoding.

### JSON-encoded request (current behavior)

```console
$ curl https://localhost:8551 \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "engine_newPayloadV1",
  "params": [{
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
    "transactions": []
  }]
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

The consensus layer client sends the same `engine_newPayloadV1` call, but with `Content-Type: application/ssz` and `Accept: application/ssz`. The `params` array contains the SSZ-serialized `ExecutionPayloadV1` as a hex-encoded `DATA` string:

```console
$ curl https://localhost:8551 \
    -X POST \
    -H "Content-Type: application/ssz" \
    -H "Accept: application/ssz" \
    -d '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "engine_newPayloadV1",
  "params": ["0x3b8fb240d288781d4aac94d3fd16809ee413bc99294a085798a589dae51ddd4a..."]
}'
```

The single hex string in `params` is the SSZ serialization of the `ExecutionPayloadV1` container, encoding all fields (`parentHash`, `feeRecipient`, `stateRoot`, etc.) in their binary SSZ representations concatenated per the SSZ specification.

### SSZ-encoded response

The execution layer responds with `Content-Type: application/ssz`. The `result` field contains the SSZ-serialized `PayloadStatusV1` as a hex-encoded `DATA` string:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "0x00355...858"
}
```

Where the binary data encodes:
- `status`: `0x00` (VALID)
- `latestValidHash`: `0x3559e851470f6e7bbed1db474980683e8c315bfce99b2a6ef47c057c04de7858`
- `validationError`: absent (Optional not present)

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
