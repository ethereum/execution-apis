# Engine API -- Paris

Engine API structures and methods specified for Paris.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

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
- [Methods](#methods)
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
## Structures

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
  * If validation fails, the response **MUST** contain `{status: INVALID, latestValidHash: validHash}` where `validHash` **MUST** be:
    - The block hash of the ancestor of the invalid payload satisfying the following two conditions:
      - It is fully validated and deemed `VALID`
      - Any other ancestor of the invalid payload with a higher `blockNumber` is `INVALID`
    - `0x0000000000000000000000000000000000000000000000000000000000000000` if the above conditions are satisfied by a PoW block.
    - `null` if client software cannot determine the ancestor of the invalid
      payload satisfying the above conditions.
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

5. Client software **MUST** begin a new build process if given `PayloadAttributes` doesn't match payload attributes of an existing build process.
   Every new build process **MUST** be uniquely identified by the returned `payloadId` value.

6. If a build process with given `PayloadAttributes` already exists, client software **SHOULD NOT** restart it.

## Methods

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
    - ancestors of a payload are known and comprise a well-formed chain.

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

2. Client software **MAY** skip an update of the forkchoice state and **MUST NOT** begin a payload build process if `forkchoiceState.headBlockHash` references a `VALID` ancestor of the head of canonical chain, i.e. the ancestor passed [payload validation](#payload-validation) process and deemed `VALID`. In the case of such an event, client software **MUST** return `{payloadStatus: {status: VALID, latestValidHash: forkchoiceState.headBlockHash, validationError: null}, payloadId: null}`.

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

5. Execution Layer client software **SHOULD** surface an error to the user if it does not receive a request on this endpoint at least once every 120 seconds.

6. Considering the absence of the `TERMINAL_BLOCK_NUMBER` setting, Consensus Layer client software **MAY** use `0` value for the `terminalBlockNumber` field in the input parameters of this call.

7. Considering the absence of the `TERMINAL_TOTAL_DIFFICULTY` value (i.e. when a value has not been decided), Consensus Layer and Execution Layer client software **MUST** use `115792089237316195423570985008687907853269984665640564039457584007913129638912` value (equal to`2**256-2**10`) for the `terminalTotalDifficulty` input parameter of this call.
