# Engine API. Interop edition

This document specifies a subset of the Engine API methods that are required to be implemented for the Merge interop.

*Note*: Sync API design is yet a draft and considered optional for the interop.

## Underlying protocol

This specification is based on [Ethereum JSON-RPC API](https://eth.wiki/json-rpc/API) and inherits the following properties of this standard:
* Supported communication protocols (HTTP and WebSocket)
* Message format and encoding notation
* [Error codes improvement proposal](https://eth.wiki/json-rpc/json-rpc-error-codes-improvement-proposal)

Client software **MUST** expose Engine API at a port independent from JSON-RPC API. The default port for the Engine API is 8550 for HTTP and 8551 for WebSocket.

## Error codes

The list of error codes introduced by this specification can be found below.

| Code | Possible Return message | Description |
| - | - | - |
| 4 | Unknown header | Should be used when a call refers to the unknown header |
| 5 | Unknown payload | Should be used when the `payloadId` parameter of `engine_getPayload` call refers to a payload building process that is unavailable |

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

## Core

### engine_preparePayload

#### Parameters
1. `Object` - The payload attributes:
- `parentHash`: `DATA`, 32 Bytes - hash of the parent block
- `timestamp`: `QUANTITY`, 64 Bits - value for the `timestamp` field of the new payload
- `random`: `DATA`, 32 Bytes - value for the `random` field of the new payload
- `feeRecipient`: `DATA`, 20 Bytes - suggested value for the `coinbase` field of the new payload

#### Returns
1. `payloadId|Error`: `QUANTITY`, 64 Bits - Identifier of the payload building process

#### Specification

1. Given provided field value set client software **SHOULD** build the initial version of the payload which has an empty transaction set.

2. Client software **SHOULD** start the process of updating the payload. The strategy of this process is implementation dependent. The default strategy would be to keep the transaction set up-to-date with the state of local mempool.

3. Client software **SHOULD** stop the updating process either by finishing to serve the [**`engine_getPayload`**](#engine_getPayload) call with the same `payloadId` value or when [`SECONDS_PER_SLOT`](https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/beacon-chain.md#time-parameters-1) (currently set to 12 in the Mainnet configuration) seconds have passed since the point in time identified by the `timestamp` parameter.

4. Client software **MUST** set the payload field values according to the set of parameters passed in the call to this method with exception for the `feeRecipient`. The `coinbase` field value **MAY** deviate from what is specified by the `feeRecipient` parameter.

5. Client software **SHOULD** respond with `2: Action not allowed` error if the sync process is in progress.

6. Client software **SHOULD** respond with `4: Unknown block` error if the parent block is unknown.

### engine_getPayload

#### Parameters
1. `payloadId`: `QUANTITY`, 64 Bits - Identifier of the payload building process

#### Returns
`Object|Error` - Either instance of [`ExecutionPayload`](#ExecutionPayload) or an error

#### Specification

1. Given the `payloadId` client software **MUST** respond with the most recent version of the payload that is available in the corresponding building process at the time of receiving the call.

2. The call **MUST** be responded with `5: Unavailable payload` error if the building process identified by the `payloadId` doesn't exist.

3. Client software **MAY** stop the corresponding building process after serving this call.

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

2. Client software **MUST** defer persisting a valid payload until the corresponding [**`engine_consensusValidated`**](#engine_consensusValidated) message deems the payload valid with respect to the proof-of-stake consensus rules.

3. Client software **MUST** discard the payload if it's deemed invalid.

4. The call **MUST** be responded with `SYNCING` status while the sync process is in progress and thus the execution cannot yet be validated.

5. In the case when the parent block is unknown, client software **MUST** pull the block from the network and take one of the following actions depending on the parent block properties:
  - If the parent block is a PoW block as per [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#specification) definition, then all missing dependencies of the payload **MUST** be pulled from the network and validated accordingly. The call **MUST** be responded according to the validity of the payload and the chain of its ancestors.
  - If the parent block is a PoS block as per [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#specification) definition, then the call **MAY** be responded with `SYNCING` status and sync process **SHOULD** be initiated accordingly.

### engine_consensusValidated

#### Parameters
1. `Object` - Payload validity status with respect to the consensus rules:
- `blockHash`: `DATA`, 32 Bytes - block hash value of the payload
- `status`: `String: VALID|INVALID` - result of the payload validation with respect to the proof-of-stake consensus rules

#### Returns
None or an error

#### Specification

1. The call of this method with `VALID` status maps on the `POS_CONSENSUS_VALIDATED` event of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#specification) and **MUST** be processed according to the specification defined in the EIP.

2. If the status in this call is `INVALID` the payload **MUST** be discarded disregarding its validity with respect to the execution environment rules.

3. Client software **MUST** respond with `4: Unknown block` error if the payload identified by the `blockHash` is unknown.

### engine_forkchoiceUpdated

#### Parameters
1. `Object` - The state of the fork choice:
- `headBlockHash`: `DATA`, 32 Bytes - block hash of the head of the canonical chain
- `finalizedBlockHash`: `DATA`, 32 Bytes - block hash of the most recent finalized block

#### Returns
None or an error

#### Specification

1. This method call maps on the `POS_FORKCHOICE_UPDATED` event of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#specification) and **MUST** be processed according to the specification defined in the EIP.

2. Client software **MUST** respond with `4: Unknown block` error if the payload identified by either the `headBlockHash` or the `finalizedBlockHash` is unknown.
