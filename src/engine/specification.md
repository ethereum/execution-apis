# Engine API

This document specifies the Engine API methods that the Consensus Layer uses to interact with the Execution Layer.

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
- [Structures](#structures)
  - [ExecutionPayloadV1](#executionpayloadv1)
  - [ForkchoiceStateV1](#forkchoicestatev1)
  - [PayloadAttributesV1](#payloadattributesv1)
  - [PayloadStatusV1](#payloadstatusv1)
  - [TransitionConfigurationV1](#transitionconfigurationv1)
- [Routines](#routines)
  - [Payload validation](#payload-validation)
  - [Sync](#sync)
  - [Payload building](#payload-building)
- [Core](#core)
  - [engine_newPayloadV1](#engine_newpayloadv1)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)
  - [engine_forkchoiceUpdatedV1](#engine_forkchoiceupdatedv1)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)
  - [engine_getPayloadV1](#engine_getpayloadv1)
    - [Request](#request-2)
    - [Response](#response-2)
    - [Specification](#specification-2)
  - [engine_exchangeTransitionConfigurationV1](#engine_exchangetransitionconfigurationv1)
    - [Request](#request-3)
    - [Response](#response-3)
    - [Specification](#specification-3)

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

Consensus Layer client software **MUST** wait for a specified `timeout` before aborting the call. In such an event, the Consensus Layer client software **MAY** retry the call.

Consensus Layer client software **MAY** wait for response longer than it is specified by the `timeout` parameter.

## Structures

Values of a field of `DATA` type **MUST** be encoded as a hexadecimal string with a `0x` prefix matching the regular expression `^0x(?:[a-fA-F0-9]{2})*$`.

Values of a field of `QUANTITY` type **MUST** be encoded as a hexadecimal string with a `0x` prefix and the leading 0s stripped (except for the case of encoding the value `0`) matching the regular expression `^0x(?:0|(?:[a-fA-F1-9][a-fA-F0-9]*))$`.

*Note:* Byte order of encoded value having `QUANTITY` type is big-endian.

### ExecutionPayloadV1

This structure maps on the [`ExecutionPayload`](https://github.com/ethereum/consensus-specs/blob/dev/specs/bellatrix/beacon-chain.md#ExecutionPayload) structure of the beacon chain spec. The fields are encoded as follows:

- `parentHash`: `DATA`, 32 Bytes
- `feeRecipient`:  `DATA`, 20 Bytes
- `stateRoot`: `DATA`, 32 Bytes
- `receiptsRoot`: `DATA`, 32 Bytes
- `logsBloom`: `DATA`, 256 Bytes
- `prevRandao`: `DATA`, 32 Bytes
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

*Note:* `safeBlockHash` and `finalizedBlockHash` fields are allowed to have `0x0000000000000000000000000000000000000000000000000000000000000000` value unless transition block is finalized.

### PayloadAttributesV1

This structure contains the attributes required to initiate a payload build process in the context of an `engine_forkchoiceUpdated` call. The fields are encoded as follows:

- `timestamp`: `QUANTITY`, 64 Bits - value for the `timestamp` field of the new payload
- `prevRandao`: `DATA`, 32 Bytes - value for the `prevRandao` field of the new payload
- `suggestedFeeRecipient`: `DATA`, 20 Bytes - suggested value for the `feeRecipient` field of the new payload

### PayloadStatusV1

This structure contains the result of processing a payload. The fields are encoded as follows:

- `status`: `enum` - `"VALID" | "INVALID" | "SYNCING" | "ACCEPTED" | "INVALID_BLOCK_HASH"`
- `latestValidHash`: `DATA|null`, 32 Bytes - the hash of the most recent *valid* block in the branch defined by payload and its ancestors
- `validationError`: `String|null` - a message providing additional details on the validation error if the payload is classified as `INVALID` or `INVALID_BLOCK_HASH`.

### TransitionConfigurationV1

This structure contains configurable settings of the transition process. The fields are encoded as follows:
- `terminalTotalDifficulty`: `QUANTITY`, 256 Bits - maps on the `TERMINAL_TOTAL_DIFFICULTY` parameter of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#client-software-configuration)
- `terminalBlockHash`: `DATA`, 32 Bytes - maps on `TERMINAL_BLOCK_HASH` parameter of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#client-software-configuration)
- `terminalBlockNumber`: `QUANTITY`, 64 Bits - maps on `TERMINAL_BLOCK_NUMBER` parameter of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#client-software-configuration)

## Routines

### Payload validation

Payload validation process consists of validating a payload with respect to the block header and execution environment rule sets. The process is specified as follows:

1. Client software **MAY** obtain a parent state by executing ancestors of a payload as a part of the validation process. In this case each ancestor **MUST** also pass payload validation process.

2. Client software **MUST** validate that the most recent PoW block in the chain of a payload ancestors satisfies terminal block conditions according to [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#transition-block-validity). This check maps to the transition block validity section of the EIP. If this validation fails, the response **MUST** contain `{status: INVALID, latestValidHash: 0x0000000000000000000000000000000000000000000000000000000000000000}`. Additionally, each block in a tree of descendants of an invalid terminal block **MUST** be deemed `INVALID`.

3. Client software **MUST** validate a payload according to the block header and execution environment rule set with modifications to these rule sets defined in the [Block Validity](https://eips.ethereum.org/EIPS/eip-3675#block-validity) section of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#specification):
  * If validation succeeds, the response **MUST** contain `{status: VALID, latestValidHash: payload.blockHash}`
  * If validation fails, the response **MUST** contain `{status: INVALID, latestValidHash: validHash}` where `validHash` is the block hash of the most recent *valid* ancestor of the invalid payload. That is, the valid ancestor of the payload with the highest `blockNumber`
  * If the most recent valid ancestor is a PoW block, `latestValidHash` **MUST** be set to `0x0000000000000000000000000000000000000000000000000000000000000000`
  * Client software **MUST NOT** surface an `INVALID` payload over any API endpoint and p2p interface.

4. Client software **MAY** provide additional details on the validation error if a payload is deemed `INVALID` by assigning the corresponding message to the `validationError` field.

5. The process of validating a payload on the canonical chain **MUST NOT** be affected by an active sync process on a side branch of the block tree. For example, if side branch `B` is `SYNCING` but the requisite data for validating a payload from canonical branch `A` is available, client software **MUST** run full validation of the payload and respond accordingly.

### Sync

In the context of this specification, the sync is understood as the process of obtaining data required to validate a payload. The sync process may consist of the following stages:

1. Pulling data from remote peers in the network.
2. Passing ancestors of a payload through the [Payload validation](#payload-validation) and obtaining a parent state.

*Note:* Each of these stages is optional. Exact behavior of client software during the sync process is implementation dependent.

### Payload building

The payload build process is specified as follows:

1. Client software **MUST** set the payload field values according to the set of parameters passed into this method with exception of the `suggestedFeeRecipient`. The built `ExecutionPayload` **MAY** deviate the `feeRecipient` field value from what is specified by the `suggestedFeeRecipient` parameter.

2. Client software **SHOULD** build the initial version of the payload which has an empty transaction set.

3. Client software **SHOULD** start the process of updating the payload. The strategy of this process is implementation dependent. The default strategy is to keep the transaction set up-to-date with the state of local mempool.

4. Client software **SHOULD** stop the updating process when either a call to `engine_getPayload` with the build process's `payloadId` is made or [`SECONDS_PER_SLOT`](https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/beacon-chain.md#time-parameters-1) (12s in the Mainnet configuration) have passed since the point in time identified by the `timestamp` parameter.

## Core

### engine_newPayloadV1

#### Request

* method: `engine_newPayloadV1`
* params: 
  1. [`ExecutionPayloadV1`](#ExecutionPayloadV1)
* timeout: 8s

#### Response

* result: [`PayloadStatusV1`](#PayloadStatusV1)
* error: code and message set in case an exception happens while processing the payload.

#### Specification

1. Client software **MUST** validate `blockHash` value as being equivalent to `Keccak256(RLP(ExecutionBlockHeader))`, where `ExecutionBlockHeader` is the execution layer block header (the former PoW block header structure). Fields of this object are set to the corresponding payload values and constant values according to the Block structure section of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#block-structure), extended with the corresponding section of [EIP-4399](https://eips.ethereum.org/EIPS/eip-4399#block-structure). Client software **MUST** run this validation in all cases even if this branch or any other branches of the block tree are in an active sync process.

2. Client software **MAY** initiate a sync process if requisite data for payload validation is missing. Sync process is specified in the [Sync](#sync) section.

3. Client software **MUST** validate the payload if it extends the canonical chain and requisite data for the validation is locally available. The validation process is specified in the [Payload validation](#payload-validation) section.

4. Client software **MAY NOT** validate the payload if the payload doesn't belong to the canonical chain.

5. Client software **MUST** respond to this method call in the following way:
  * `{status: INVALID_BLOCK_HASH, latestValidHash: null, validationError: errorMessage | null}` if the `blockHash` validation has failed
  * `{status: INVALID, latestValidHash: 0x0000000000000000000000000000000000000000000000000000000000000000, validationError: errorMessage | null}` if terminal block conditions are not satisfied
  * `{status: SYNCING, latestValidHash: null, validationError: null}` if requisite data for the payload's acceptance or validation is missing
  * with the payload status obtained from the [Payload validation](#payload-validation) process if the payload has been fully validated while processing the call
  * `{status: ACCEPTED, latestValidHash: null, validationError: null}` if the following conditions are met:
    - the `blockHash` of the payload is valid
    - the payload doesn't extend the canonical chain
    - the payload hasn't been fully validated
    - ancestors of a payload are know and comprise a well-formed chain.

6. If any of the above fails due to errors unrelated to the normal processing flow of the method, client software **MUST** respond with an error object.

### engine_forkchoiceUpdatedV1

#### Request

* method: "engine_forkchoiceUpdatedV1"
* params: 
  1. `forkchoiceState`: `Object` - instance of [`ForkchoiceStateV1`](#ForkchoiceStateV1)
  2. `payloadAttributes`: `Object|null` - instance of [`PayloadAttributesV1`](#PayloadAttributesV1) or `null`
* timeout: 8s

#### Response

* result: `object`
  - `payloadStatus`: [`PayloadStatusV1`](#PayloadStatusV1); values of the `status` field in the context of this method are restricted to the following subset:
    * `"VALID"`
    * `"INVALID"`
    * `"SYNCING"`
  - `payloadId`: `DATA|null`, 8 Bytes - identifier of the payload build process or `null`
* error: code and message set in case an exception happens while the validating payload, updating the forkchoice or initiating the payload build process.

#### Specification

1. Client software **MAY** initiate a sync process if `forkchoiceState.headBlockHash` references an unknown payload or a payload that can't be validated because data that are requisite for the validation is missing. The sync process is specified in the [Sync](#sync) section.

2. Client software **MAY** skip an update of the forkchoice state and **MUST NOT** begin a payload build process if `forkchoiceState.headBlockHash` references an ancestor of the head of canonical chain. In the case of such an event, client software **MUST** return `{payloadStatus: {status: VALID, latestValidHash: forkchoiceState.headBlockHash, validationError: null}, payloadId: null}`.

3. If `forkchoiceState.headBlockHash` references a PoW block, client software **MUST** validate this block with respect to terminal block conditions according to [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#transition-block-validity). This check maps to the transition block validity section of the EIP. Additionally, if this validation fails, client software **MUST NOT** update the forkchoice state and **MUST NOT** begin a payload build process.

4. Before updating the forkchoice state, client software **MUST** ensure the validity of the payload referenced by `forkchoiceState.headBlockHash`, and **MAY** validate the payload while processing the call. The validation process is specified in the [Payload validation](#payload-validation) section. If the validation process fails, client software **MUST NOT** update the forkchoice state and **MUST NOT** begin a payload build process.

5. Client software **MUST** update its forkchoice state if payloads referenced by `forkchoiceState.headBlockHash` and `forkchoiceState.finalizedBlockHash` are `VALID`. The update is specified as follows:
  * The values `(forkchoiceState.headBlockHash, forkchoiceState.finalizedBlockHash)` of this method call map on the `POS_FORKCHOICE_UPDATED` event of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#block-validity) and **MUST** be processed according to the specification defined in the EIP
  * All updates to the forkchoice state resulting from this call **MUST** be made atomically.

6. Client software **MUST** return `-38002: Invalid forkchoice state` error if the payload referenced by `forkchoiceState.headBlockHash` is `VALID` and a payload referenced by either `forkchoiceState.finalizedBlockHash` or `forkchoiceState.safeBlockHash` does not belong to the chain defined by `forkchoiceState.headBlockHash`.

7. Client software **MUST** ensure that `payloadAttributes.timestamp` is greater than `timestamp` of a block referenced by `forkchoiceState.headBlockHash`. If this condition isn't held client software **MUST** respond with `-38003: Invalid payload attributes` and **MUST NOT** begin a payload build process. In such an event, the `forkchoiceState` update **MUST NOT** be rolled back.

8. Client software **MUST** begin a payload build process building on top of `forkchoiceState.headBlockHash` and identified via `buildProcessId` value if `payloadAttributes` is not `null` and the forkchoice state has been updated successfully. The build process is specified in the [Payload building](#payload-building) section.

9. Client software **MUST** respond to this method call in the following way:
  * `{payloadStatus: {status: SYNCING, latestValidHash: null, validationError: null}, payloadId: null}` if `forkchoiceState.headBlockHash` references an unknown payload or a payload that can't be validated because requisite data for the validation is missing
  * `{payloadStatus: {status: INVALID, latestValidHash: validHash, validationError: errorMessage | null}, payloadId: null}` obtained from the [Payload validation](#payload-validation) process if the payload is deemed `INVALID`
  * `{payloadStatus: {status: INVALID, latestValidHash: 0x0000000000000000000000000000000000000000000000000000000000000000, validationError: errorMessage | null}, payloadId: null}` obtained either from the [Payload validation](#payload-validation) process or as a result of validating a terminal PoW block referenced by `forkchoiceState.headBlockHash`
  * `{payloadStatus: {status: VALID, latestValidHash: forkchoiceState.headBlockHash, validationError: null}, payloadId: null}` if the payload is deemed `VALID` and a build process hasn't been started
  * `{payloadStatus: {status: VALID, latestValidHash: forkchoiceState.headBlockHash, validationError: null}, payloadId: buildProcessId}` if the payload is deemed `VALID` and the build process has begun
  * `{error: {code: -38002, message: "Invalid forkchoice state"}}` if `forkchoiceState` is either invalid or inconsistent
  * `{error: {code: -38003, message: "Invalid payload attributes"}}` if the payload is deemed `VALID` and `forkchoiceState` has been applied successfully, but no build process has been started due to invalid `payloadAttributes`.

10. If any of the above fails due to errors unrelated to the normal processing flow of the method, client software **MUST** respond with an error object.

### engine_getPayloadV1

#### Request

* method: `engine_getPayloadV1`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process
* timeout: 1s

#### Response

* result: [`ExecutionPayloadV1`](#ExecutionPayloadV1)
* error: code and message set in case an exception happens while getting the payload.

#### Specification

1. Given the `payloadId` client software **MUST** return the most recent version of the payload that is available in the corresponding build process at the time of receiving the call.

2. The call **MUST** return `-38001: Unknown payload` error if the build process identified by the `payloadId` does not exist.

3. Client software **MAY** stop the corresponding build process after serving this call.

### engine_exchangeTransitionConfigurationV1

#### Request

* method: `engine_exchangeTransitionConfigurationV1`
* params:
  1. `transitionConfiguration`: `Object` - instance of [`TransitionConfigurationV1`](#TransitionConfigurationV1)
* timeout: 1s

#### Response

* result: [`TransitionConfigurationV1`](#TransitionConfigurationV1)
* error: code and message set in case an exception happens while getting a transition configuration.

#### Specification

1. Execution Layer client software **MUST** respond with configurable setting values that are set according to the Client software configuration section of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#client-software-configuration).

2. Execution Layer client software **SHOULD** surface an error to the user if local configuration settings mismatch corresponding values received in the call of this method, with exception for `terminalBlockNumber` value.

3. Consensus Layer client software **SHOULD** surface an error to the user if local configuration settings mismatch corresponding values obtained from the response to the call of this method.

4. Consensus Layer client software **SHOULD** poll this endpoint every 60 seconds.

5. Execution Layer client software **SHOULD** surface an error to the user if it does not recieve a request on this endpoint at least once every 120 seconds.

6. Considering the absence of the `TERMINAL_BLOCK_NUMBER` setting, Consensus Layer client software **MAY** use `0` value for the `terminalBlockNumber` field in the input parameters of this call.

[json-rpc-spec]: https://playground.open-rpc.org/?schemaUrl=https://raw.githubusercontent.com/ethereum/execution-apis/assembled-spec/openrpc.json&uiSchema[appBar][ui:splitView]=false&uiSchema[appBar][ui:input]=false&uiSchema[appBar][ui:examplesDropdown]=false
