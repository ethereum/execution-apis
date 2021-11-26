# Engine API

This document specifies the Engine API methods that the Consensus Layer uses to interact with the Execution Layer.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Underlying protocol](#underlying-protocol)
- [Versioning](#versioning)
- [Message ordering](#message-ordering)
- [Load-balancing and advanced configurations](#load-balancing-and-advanced-configurations)
- [Errors](#errors)
- [Structures](#structures)
  - [ExecutionPayloadV1](#executionpayloadv1)
  - [ForkchoiceStateV1](#forkchoicestatev1)
  - [PayloadAttributesV1](#payloadattributesv1)
- [Core](#core)
  - [engine_executePayloadV1](#engine_executepayloadv1)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)
  - [engine_forkchoiceUpdatedV1](#engine_forkchoiceupdatedv1)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)
      - [Payload build process](#payload-build-process)
  - [engine_getPayloadV1](#engine_getpayloadv1)
    - [Request](#request-2)
    - [Response](#response-2)
    - [Specification](#specification-2)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Underlying protocol

This specification is based on [Ethereum JSON-RPC API](https://eth.wiki/json-rpc/API) and inherits the following properties of this standard:
* Supported communication protocols (HTTP and WebSocket)
* Message format and encoding notation
* [Error codes improvement proposal](https://eth.wiki/json-rpc/json-rpc-error-codes-improvement-proposal)

Client software **MUST** expose Engine API at a port independent from JSON-RPC API.
The default port for the Engine API is 8550.
The Engine API is exposed under the `engine` namespace.

To facilitate an Engine API consumer to access state and logs (e.g. proof-of-stake deposits) through the same connection,
the client **MUST** also expose the `eth` namespace. 

## Versioning

The versioning of the Engine API is defined as follows:
* The version of each method and structure is independent from versions of other methods and structures.
* The `VX`, where the `X` is the number of the version, is suffixed to the name of each method and structure.
* The version of a method or a structure **MUST** be incremented by one if any of the following is changed:
  * a set of method parameters
  * a method response value
  * a method behavior
  * a set of structure fields
* The specification **MAY** reference a method or a structure without the version suffix e.g. `engine_executePayload`. These statements should be read as related to all versions of the referenced method or structure.

## Message ordering

Consensus Layer client software **MUST** utilize JSON-RPC request IDs that are strictly
increasing.

Execution Layer client software **MUST** execute calls strictly in the order of request IDs
to avoid degenerate race conditions.

## Load-balancing and advanced configurations

The Engine API supports a one-to-many Consensus Layer to Execution Layer configuration.
Intuitively this is because the Consensus Layer drives the Execution Layer and thus can drive many of them independently.

On the other hand, generic many-to-one Consensus Layer to Execution Layer configurations are not supported out-of-the-box.
The Execution Layer, by default, only supports one chain head at a time and thus has undefined behavior when multiple Consensus Layers simultaneously control the head.
The Engine API does work properly, if in such a many-to-one configuration, only one Consensus Layer instantiation is able to *write* to the Execution Layer's chain head and initiate the payload build process (i.e. call `engine_forkchoiceUpdated` ),
while other Consensus Layers can only safely insert payloads (i.e. `engine_executePayload`) and read from the Execution Layer.

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
| -32001 | Unknown payload | Payload does not exist / is not available. |
| -32002 | Invalid terminal block | Terminal block doesn't satisfy terminal block conditions. |

Each error returns a `null` `data` value, except `-32000` which returns the `data` object with a `err` member that explains the error encountered.

For example:

```console
$ curl https://localhost:8550 \
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

## Structures

Fields having `DATA` and `QUANTITY` types **MUST** be encoded according to the [HEX value encoding](https://eth.wiki/json-rpc/API#hex-value-encoding) section of Ethereum JSON-RPC API.

*Note:* Byte order of encoded value having `QUANTITY` type is big-endian.

### ExecutionPayloadV1

This structure maps on the [`ExecutionPayload`](https://github.com/ethereum/consensus-specs/blob/dev/specs/merge/beacon-chain.md#ExecutionPayload) structure of the beacon chain spec. The fields are encoded as follows:
- `parentHash`: `DATA`, 32 Bytes
- `feeRecipient`:  `DATA`, 20 Bytes
- `stateRoot`: `DATA`, 32 Bytes
- `receiptRoot`: `DATA`, 32 Bytes
- `logsBloom`: `DATA`, 256 Bytes
- `random`: `DATA`, 32 Bytes
- `blockNumber`: `QUANTITY`, 64 Bits
- `gasLimit`: `QUANTITY`, 64 Bits
- `gasUsed`: `QUANTITY`, 64 Bits
- `timestamp`: `QUANTITY`, 64 Bits
- `extraData`: `DATA`, 0 to 32 Bytes
- `baseFeePerGas`: `QUANTITY`, 256 Bits
- `blockHash`: `DATA`, 32 Bytes
- `transactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718)

### ForkchoiceStateV1

This structure encapsulates the fork choice state. The fields are encoded as follows:
- `headBlockHash`: `DATA`, 32 Bytes - block hash of the head of the canonical chain
- `safeBlockHash`: `DATA`, 32 Bytes - the "safe" block hash of the canonical chain under certain synchrony and honesty assumptions. This value **MUST** be either equal to or an ancestor of `headBlockHash`
- `finalizedBlockHash`: `DATA`, 32 Bytes - block hash of the most recent finalized block

### PayloadAttributesV1

This structure contains the attributes required to initiate a payload build process in the context of an `engine_forkchoiceUpdated` call. The fields are encoded as follows:
- `timestamp`: `QUANTITY`, 64 Bits - value for the `timestamp` field of the new payload
- `random`: `DATA`, 32 Bytes - value for the `random` field of the new payload
- `suggestedFeeRecipient`: `DATA`, 20 Bytes - suggested value for the `feeRecipient` field of the new payload

## Core

### engine_executePayloadV1

#### Request

* method: `engine_executePayloadV1`
* params: 
  1. [`ExecutionPayloadV1`](#ExecutionPayloadV1)

#### Response

* result: `object`
    - `status`: `enum` - `"VALID" | "INVALID" | "SYNCING"`
    - `latestValidHash`: `DATA|null`, 32 Bytes - the hash of the most recent *valid* block in the branch defined by payload and its ancestors
    - `message`: `STRING|null` - the message providing additional details on the response to the method call if needed
* error: code and message set in case an exception happens while executing the payload.

#### Specification

1. Client software **MUST** validate the payload according to the execution environment rule set with modifications to this rule set defined in the [Block Validity](https://eips.ethereum.org/EIPS/eip-3675#block-validity) section of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#specification) and return the validation result.
    * If validation succeeds, return `{status: VALID, latestValidHash: payload.blockHash}`
    * If validation fails, return `{status: INVALID, latestValidHash: validHash}` where `validHash` is the block hash of the most recent *valid* ancestor of the invalid payload. That is, the valid ancestor of the payload with the highest `blockNumber`.

2. Client software **MUST** discard the payload if it's deemed invalid.

3. Client software **MUST** return `{status: SYNCING, latestValidHash: null}` if the sync process is already in progress or if requisite data for payload validation is missing. In the event that requisite data to validate the payload is missing (e.g. does not have payload identified by `parentHash`), the client software **SHOULD** initiate the sync process.

4. Client software **MAY** provide additional details on the payload validation by utilizing `message` field in the response object. For example, particular error message occurred during the payload execution may accompany a response with `INVALID` status.

5. Client software **MUST** return `-32002: Invalid terminal block` error if the parent block is a PoW block that doesn't satisfy terminal block conditions according to [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#definitions). This check maps on the Transition block validity section of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#transition-block-validity).

### engine_forkchoiceUpdatedV1

#### Request

* method: "engine_forkchoiceUpdatedV1"
* params: 
  1. `forkchoiceState`: `Object` - instance of [`ForkchoiceStateV1`](#ForkchoiceStateV1)
  2. `payloadAttributes`: `Object|null` - instance of [`PayloadAttributesV1`](#PayloadAttributesV1) or `null`

#### Response

* result: `object`
    - `status`: `enum` - `"SUCCESS" | "SYNCING"`
    - `payloadId`: `DATA|null`, 8 Bytes - identifier of the payload build process or `null`
* error: code and message set in case an exception happens while updating the forkchoice or initiating the payload build process.

#### Specification

1. The values `(forkchoiceState.headBlockHash, forkchoiceState.finalizedBlockHash)` of this method call map on the `POS_FORKCHOICE_UPDATED` event of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#specification) and **MUST** be processed according to the specification defined in the EIP.

2. All updates to the forkchoice state resulting from this call **MUST** be made atomically.

3. Client software **MUST** return `{status: SUCCESS, payloadId: null}` if `payloadAttributes` is `null` and the client is not `SYNCING`.

4. Client software **MUST** return `{status: SYNCING, payloadId: null}` if the payload identified by either the `forkchoiceState.headBlockHash` or the `forkchoiceState.finalizedBlockHash` is unknown or if the sync process is in progress. In the event that either the `forkchoiceState.headBlockHash` or the `forkchoiceState.finalizedBlockHash` is unknown, the client software **SHOULD** initiate the sync process.

5. Client software **MUST** return `{status: SUCCESS, payloadId: buildProcessId}` if `payloadAttributes` is not `null` and the client is not `SYNCING`, and **MUST** begin a payload build process building on top of `forkchoiceState.headBlockHash` and identified via `buildProcessId` value. The build process is specified in the [Payload build process](#payload-build-process) section.

6. Client software **MUST** return `-32002: Invalid terminal block` error if a block referenced by `forkchoiceState.headBlockHash` is a PoW block that doesn't satisfy terminal block conditions according to [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#definitions).

7. If any of the above fails due to errors unrelated to the client software's normal `SYNCING` status, the client software **MUST** return an error.

##### Payload build process
The payload build process is specified as follows:
* Client software **MUST** set the payload field values according to the set of parameters passed into this method with exception of the `suggestedFeeRecipient`. The built `ExecutionPayload` **MAY** deviate the `feeRecipient` field value from what is specified by the `suggestedFeeRecipient` parameter.
* Client software **SHOULD** build the initial version of the payload which has an empty transaction set.
* Client software **SHOULD** start the process of updating the payload. The strategy of this process is implementation dependent. The default strategy is to keep the transaction set up-to-date with the state of local mempool.
* Client software **SHOULD** stop the updating process when either a call to `engine_getPayload` with the build process's `payloadId` is made or [`SECONDS_PER_SLOT`](https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/beacon-chain.md#time-parameters-1) (12s in the Mainnet configuration) have passed since the point in time identified by the `timestamp` parameter.

### engine_getPayloadV1

#### Request

* method: `engine_getPayloadV1`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process

#### Response

* result: [`ExecutionPayloadV1`](#ExecutionPayloadV1)
* error: code and message set in case an exception happens while getting the payload.

#### Specification

1. Given the `payloadId` client software **MUST** return the most recent version of the payload that is available in the corresponding build process at the time of receiving the call.

2. The call **MUST** return `-32001: Unknown payload` error if the build process identified by the `payloadId` does not exist.

3. Client software **MAY** stop the corresponding build process after serving this call.
