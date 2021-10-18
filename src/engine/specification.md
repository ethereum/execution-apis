# Engine API

This document specifies the Engine API methods that the Consensus Layer uses to interact with the Execution Layer.

## Underlying protocol

This specification is based on [Ethereum JSON-RPC API](https://eth.wiki/json-rpc/API) and inherits the following properties of this standard:
* Supported communication protocols (HTTP and WebSocket)
* Message format and encoding notation
* [Error codes improvement proposal](https://eth.wiki/json-rpc/json-rpc-error-codes-improvement-proposal)

Client software **MUST** expose Engine API at a port independent from JSON-RPC API. The default port for the Engine API is 8550 for HTTP and 8551 for WebSocket.

## Load-balancing and advanced configurations

The Engine API supports a 1:many Consensus Layer to Execution Layer configuration.
Intuitively this is because the Consensus Layer drives the Execution Layer and thus can drive many of them independently.

On the other hand, generic many:1 Consensus Layer to Execution Layer configurations are not supported out-of-the-box.
The Execution Layer, by default, only supports one chain head at a time and thus has undefined behavior when multiple Consensus Layers simultaneously control the head.
The Engine API does work properly, if in such a many:1 configuration, only one Consensus Layer instantiation is able to *write* to the Execution Layer's chain head and initiate the payload build process (i.e. call `engine_forkchoiceUpdated` ),
while other Consensus Layers can only safely insert payloads (i.e. `engine_executePayload`) and read from the Execution Layer.

## Error codes

The list of error codes introduced by this specification can be found below.

| Code | Possible Return message | Description |
| - | - | - |
| 5 | Unknown payload | Should be used when the `payloadId` parameter of `engine_getPayload` call refers to a payload build process that is unavailable |

## Structures

Fields having `DATA` and `QUANTITY` types **MUST** be encoded according to the [HEX value encoding](https://eth.wiki/json-rpc/API#hex-value-encoding) section of Ethereum JSON-RPC API.

*Note:* Byte order of encoded value having `QUANTITY` type is big-endian.

### ExecutionPayload

This structure maps on the [`ExecutionPayload`](https://github.com/ethereum/consensus-specs/blob/dev/specs/merge/beacon-chain.md#ExecutionPayload) structure of the beacon chain spec. The fields are encoded as follows:
- `parentHash`: `DATA`, 32 Bytes
- `coinbase`:  `DATA`, 20 Bytes
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

### PayloadAttributes

This structure contains the attributes required to initiate a payload build process in the context of an `engine_forkchoiceUpdated` call. The fields are encoded as follows:
- `timestamp`: `QUANTITY`, 64 Bits - value for the `timestamp` field of the new payload
- `random`: `DATA`, 32 Bytes - value for the `random` field of the new payload
- `feeRecipient`: `DATA`, 20 Bytes - suggested value for the `coinbase` field of the new payload

## Core

### engine_executePayload

#### Parameters
1. `Object` - Instance of [`ExecutionPayload`](#ExecutionPayload)

#### Returns
`Object` - Response object:
1. `status`: `String` - the result of the payload execution:
  - `VALID` - given payload is valid
  - `INVALID` - given payload is invalid
  - `SYNCING` - sync process is in progress

#### Specification

1. Client software **MUST** validate the payload according to the execution environment rule set with modifications to this rule set defined in the [Block Validity](https://eips.ethereum.org/EIPS/eip-3675#block-validity) section of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#specification) and respond with the validation result.

2. Client software **MUST** discard the payload if it's deemed invalid.

3. The call **MUST** return `SYNCING` status while the sync process is in progress and thus the execution cannot yet be validated.

4. In the case when the parent block is unknown, client software **MUST** pull the block from the network and take one of the following actions depending on the parent block properties:
  - If the parent block is a PoW block as per [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#specification) definition, then all missing dependencies of the payload **MUST** be pulled from the network and validated accordingly. The call **MUST** be responded according to the validity of the payload and the chain of its ancestors.
  - If the parent block is a PoS block as per [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#specification) definition, then the call **MAY** be responded with `SYNCING` status and sync process **SHOULD** be initiated accordingly.

### engine_forkchoiceUpdated

#### Parameters
1. `Object` - The state of the fork choice:
- `headBlockHash`: `DATA`, 32 Bytes - block hash of the head of the canonical chain
- `finalizedBlockHash`: `DATA`, 32 Bytes - block hash of the most recent finalized block
- `payloadAttributes`: `Object|None` - instance of [`PayloadAttributes`](#PayloadAttributes) or `None`

#### Returns

`Object|Error` - Either instance of response object or an error. Response object:
1. `status`: `String` - The result of updating the fork choice
  - `SUCCESS` - The fork choice is successfully updated, and if requested, the payload build process has successfully started
  - `SYNCING` - Sync process is in progress

#### Specification

1. This method call maps on the `POS_FORKCHOICE_UPDATED` event of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#specification) and **MUST** be processed according to the specification defined in the EIP.

2. Client software **MUST** return `SYNCING` status if the payload identified by either the `headBlockHash` or the `finalizedBlockHash` is unknown or if the sync process is in progress. In the event that either the `headBlockHash` or the `finalizedBlockHash` is unknown, the client software **SHOULD** initiate the sync process.

3. Client software **MUST** begin a payload build process building on top of `headBlockHash` if `payloadAttributes` is not `None` and the client is not `SYNCING`. The build process is specified as:
  * The payload build process **MUST** be identifid via `payloadId` where `payloadId` is defined as the first `8` bytes of the `sha256` hash of concatenation of `headBlockHash`, `payloadAttributes.timestamp`, `payloadAttributes.random`, and `payloadAttributes.feeRecipient` where `payloadAttributes.timestamp` is encoded as big-endian and padded fully to 8 bytes -- i.e. `sha256(headBlockHash + timestamp + random + feeRecipient)[0:8]`.
  * Client software **MUST** set the payload field values according to the set of parameters passed into this method with exception of the `feeRecipient`. The prepared `ExecutionPayload` **MAY** deviate the `coinbase` field value from what is specified by the `feeRecipient` parameter.
  * Client software **SHOULD** build the initial version of the payload which has an empty transaction set.
  * Client software **SHOULD** start the process of updating the payload. The strategy of this process is implementation dependent. The default strategy is to keep the transaction set up-to-date with the state of local mempool.
  * Client software **SHOULD** stop the updating process when either a call to [**`engine_getPayload`**](#engine_getPayload) with the build process's `payloadId` is made or [`SECONDS_PER_SLOT`](https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/beacon-chain.md#time-parameters-1) (currently set to 12 in the Mainnet configuration) seconds have passed since the point in time identified by the `timestamp` parameter.

4. If any of the above fails due to errors unrelated to the client software's normal `SYNCING` status, the client software **MUST** return an error.

### engine_getPayload

#### Parameters
1. `payloadId`: `DATA`, 8 bytes - Identifier of the payload build process

#### Returns
`Object|Error` - Either instance of [`ExecutionPayload`](#ExecutionPayload) or an error

#### Specification

1. Given the `payloadId` client software **MUST** return the most recent version of the payload that is available in the corresponding build process at the time of receiving the call.

2. The call **MUST** return `5: Unknown payload` error if the build process identified by the `payloadId` does not exist.

3. Client software **MAY** stop the corresponding build process after serving this call.
