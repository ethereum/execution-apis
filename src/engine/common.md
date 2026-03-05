# Engine API -- Common Definitions

This document specifies common definitions and requirements affecting Engine API specification in general.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Underlying protocol](#underlying-protocol)
  - [Authentication](#authentication)
- [Versioning](#versioning)
- [Message ordering](#message-ordering)
- [Load-balancing and advanced configurations](#load-balancing-and-advanced-configurations)
- [Errors](#errors)
- [Timeouts](#timeouts)
- [Encoding](#encoding)
- [Capabilities](#capabilities)
  - [engine_exchangeCapabilities](#engine_exchangecapabilities)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)
  - [engine_exchangeCapabilitiesV2](#engine_exchangecapabilitiesv2)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Underlying protocol

Message format and encoding notation used by this specification are inherited
from [Ethereum JSON-RPC Specification][json-rpc-spec].

Client software **MUST** expose Engine API at a port independent from JSON-RPC API.
The default port for the Engine API is 8551.
The Engine API is exposed under the `engine` namespace.

To facilitate an Engine API consumer to access state and logs (e.g. proof-of-stake deposits) through the same connection,
the client **MUST** also expose the following subset of `eth` methods:
* `eth_blockNumber`
* `eth_call`
* `eth_chainId`
* `eth_getCode`
* `eth_getBlockByHash`
* `eth_getBlockByNumber`
* `eth_getLogs`
* `eth_sendRawTransaction`
* `eth_syncing`

These methods are described in [Ethereum JSON-RPC Specification][json-rpc-spec].

### Authentication

Engine API uses JWT authentication enabled by default.
JWT authentication is specified in [Authentication](./authentication.md) document.

## Versioning

The versioning of the Engine API is defined as follows:

* The version of each method and structure is independent from versions of other methods and structures.
* The `VX`, where the `X` is the number of the version, is suffixed to the name of each method and structure.
* The version of a method or a structure **MUST** be incremented by one if any of the following is changed:
  * a set of method parameters
  * a method response value
  * a method behavior
  * a set of structure fields
* The specification **MAY** reference a method or a structure without the version suffix e.g. `engine_newPayload`. These statements should be read as related to all versions of the referenced method or structure.

## Message ordering

Consensus Layer client software **MUST** respect the order of the corresponding fork choice update events
when making calls to the `engine_forkchoiceUpdated` method.

Execution Layer client software **MUST** process `engine_forkchoiceUpdated` method calls
in the same order as they have been received.

## Load-balancing and advanced configurations

The Engine API supports a one-to-many Consensus Layer to Execution Layer configuration.
Intuitively this is because the Consensus Layer drives the Execution Layer and thus can drive many of them independently.

On the other hand, generic many-to-one Consensus Layer to Execution Layer configurations are not supported out-of-the-box.
The Execution Layer, by default, only supports one chain head at a time and thus has undefined behavior when multiple Consensus Layers simultaneously control the head.
The Engine API does work properly, if in such a many-to-one configuration, only one Consensus Layer instantiation is able to *write* to the Execution Layer's chain head and initiate the payload build process (i.e. call `engine_forkchoiceUpdated` ),
while other Consensus Layers can only safely insert payloads (i.e. `engine_newPayload`) and read from the Execution Layer.

## Errors

The list of error codes introduced by this specification can be found below.

| Code | Message | Meaning |
| - | - | - |
| -32700 | Parse error | Invalid JSON was received by the server. |
| -32600 | Invalid Request | The JSON sent is not a valid Request object. |
| -32601 | Method not found | The method does not exist / is not available. |
| -32602 | Invalid params | Invalid method parameter(s). |
| -32603 | Internal error | Internal JSON-RPC error. |
| -32000 | Server error | Generic client error while processing request. |
| -38001 | Unknown payload | Payload does not exist / is not available. |
| -38002 | Invalid forkchoice state | Forkchoice state is invalid / inconsistent. |
| -38003 | Invalid payload attributes | Payload attributes are invalid / inconsistent. |
| -38004 | Too large request | Number of requested entities is too large. |
| -38005 | Unsupported fork | Payload belongs to a fork that is not supported. |

Each error returns a `null` `data` value, except `-32000` which returns the `data` object with a `err` member that explains the error encountered.

For example:

```console
$ curl https://localhost:8551 \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","method":"engine_getPayloadV1","params": ["0x1"],"id":1}'
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32000,
    "message": "Server error",
    "data": {
        "err": "Database corrupted"
    }
  }
}
```

## Timeouts

Consensus Layer client software **MUST** wait for a specified `timeout` before aborting the call. In such an event, the Consensus Layer client software **SHOULD** retry the call when it is needed to keep progressing.

Consensus Layer client software **MAY** wait for response longer than it is specified by the `timeout` parameter.

## Encoding

Values of a field of `DATA` type **MUST** be encoded as a hexadecimal string with a `0x` prefix matching the regular expression `^0x(?:[a-fA-F0-9]{2})*$`.

Values of a field of `QUANTITY` type **MUST** be encoded as a hexadecimal string with a `0x` prefix and the leading 0s stripped (except for the case of encoding the value `0`) matching the regular expression `^0x(?:0|(?:[a-fA-F1-9][a-fA-F0-9]*))$`.

*Note:* Byte order of encoded value having `QUANTITY` type is big-endian.

[json-rpc-spec]: https://playground.open-rpc.org/?schemaUrl=https://raw.githubusercontent.com/ethereum/execution-apis/assembled-spec/openrpc.json&uiSchema[appBar][ui:splitView]=false&uiSchema[appBar][ui:input]=false&uiSchema[appBar][ui:examplesDropdown]=false

## Capabilities

Execution and consensus layer client software may exchange with a list of supported Engine API methods by calling `engine_exchangeCapabilities` method.

Execution layer clients **MUST** support `engine_exchangeCapabilities` method, while consensus layer clients are free to choose whether to call it or not.

*Note:* The method itself doesn't have a version suffix.

### engine_exchangeCapabilities

#### Request

* method: `engine_exchangeCapabilities`
* params:
    1. `Array of string` -- Array of strings, each string is a name of a method supported by consensus layer client software.
* timeout: 1s

#### Response

`Array of string` -- Array of strings, each string is a name of a method supported by execution layer client software.

#### Specification

1. Consensus and execution layer client software **MAY** exchange with a list of currently supported Engine API methods. Execution layer client software **MUST NOT** log any error messages if this method has either never been called or hasn't been called for a significant amount of time.

2. Request and response lists **MUST** contain Engine API methods that are currently supported by consensus and execution client software respectively. Name of each method in both lists **MUST** include suffixed version. Consider the following examples:
    * Client software of both layers currently supports `V1` and `V2` versions of `engine_newPayload` method:
        * params: `[["engine_newPayloadV1", "engine_newPayloadV2", ...]]`,
        * response: `["engine_newPayloadV1", "engine_newPayloadV2", ...]`.
    * `V1` method has been deprecated and `V3` method has been introduced on execution layer side since the last call:
        * params: `[["engine_newPayloadV1", "engine_newPayloadV2", ...]]`,
        * response: `["engine_newPayloadV2", "engine_newPayloadV3", ...]`.
    * The same capabilities modification has happened in consensus layer client, so, both clients have the same capability set again:
        * params: `[["engine_newPayloadV2", "engine_newPayloadV3", ...]]`,
        * response: `["engine_newPayloadV2", "engine_newPayloadV3", ...]`.

3. The `engine_exchangeCapabilities` method **MUST NOT** be returned in the response list.

### engine_exchangeCapabilitiesV2

This method extends `engine_exchangeCapabilities` by adding a `supportedProtocols` field to the response. This lets the EL advertise alternative communication protocols (e.g., SSZ-REST, gRPC) and their endpoints alongside the existing JSON-RPC capability exchange.

#### Request

* method: `engine_exchangeCapabilitiesV2`
* params:
    1. `Array of string` -- Array of strings, each string is a name of a method supported by consensus layer client software.
* timeout: 1s

#### Response

* result: `object`
    * `capabilities`: `Array of string` -- Array of strings, each string is a name of a method supported by execution layer client software.
    * `supportedProtocols`: `Array of CommunicationChannelV1` -- List of communication protocols the EL supports.

##### CommunicationChannelV1

This structure contains information about a communication protocol supported by the execution layer.

- `protocol`: `String` - Identifier for the protocol. See [Protocol Identifiers](#protocol-identifiers).
- `url`: `String` - The endpoint where this protocol is available.

##### Protocol Identifiers

This specification defines one protocol identifier:

| Identifier | Description |
| - | - |
| `json_rpc` | JSON-RPC over HTTP, as currently used by the Engine API. |

Follow-up specifications may define additional identifiers. Some examples:

* `ssz_rest` -- SSZ-encoded payloads over REST using `application/octet-stream`.
* `grpc` -- gRPC with Protocol Buffers or SSZ serialization.
* `ssz_websocket` -- SSZ-encoded payloads over WebSocket.

#### Specification

1. The request format is identical to `engine_exchangeCapabilities` -- an array of method names supported by the CL.

2. The `capabilities` field in the response follows the same rules as the response of `engine_exchangeCapabilities`.

3. The EL **MUST** always include at least one entry with `protocol` set to `json_rpc` in `supportedProtocols`.

4. The EL **MAY** include additional entries for other protocols it supports.

5. The CL **SHOULD** call this method on startup to discover available protocols.

6. The CL **MAY** switch to any advertised protocol it supports. If the CL doesn't support any of the alternatives, it falls back to `json_rpc`.

7. All protocols **MUST** use the same JWT authentication as the existing Engine API.

8. The `engine_exchangeCapabilitiesV2` method **MUST NOT** appear in the `capabilities` response list.

9. CL clients that receive a method-not-found error for `engine_exchangeCapabilitiesV2` **SHOULD** fall back to `engine_exchangeCapabilities`.

10. EL clients **MUST** continue supporting `engine_exchangeCapabilities` for backwards compatibility. If called, it behaves exactly as before -- returning a flat array of method names.
