# Engine API -- Forkchoice

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Structures](#structures)
  - [ForkchoiceStateV1](#forkchoicestatev1)
  - [PayloadAttributesV1](#payloadattributesv1)
  - [PayloadAttributesV2](#payloadattributesv2)
- [Routines](#routines)
  - [Payload building](#payload-building)
- [Methods](#methods)
  - [engine_forkchoiceUpdatedV1](#engine_forkchoiceupdatedv1)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)
  - [engine_forkchoiceUpdatedV2](#engine_forkchoiceupdatedv2)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Structures

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

### PayloadAttributesV2

This structure has the syntax of `PayloadAttributesV1` and appends a single field: `withdrawals`.

- `timestamp`: `QUANTITY`, 64 Bits - value for the `timestamp` field of the new payload
- `prevRandao`: `DATA`, 32 Bytes - value for the `prevRandao` field of the new payload
- `suggestedFeeRecipient`: `DATA`, 20 Bytes - suggested value for the `feeRecipient` field of the new payload
- `withdrawals`: `Array of WithdrawalV1` - Array of withdrawals, each object is an `OBJECT` containing the fields of a [`WithdrawalV1`](./payload.md#withdrawalv1) structure.

## Routines

### Payload building

The payload build process is specified as follows:

1. Client software **MUST** set the payload field values according to the set of parameters passed into this method with exception of the `suggestedFeeRecipient`. The built `ExecutionPayload` **MAY** deviate the `feeRecipient` field value from what is specified by the `suggestedFeeRecipient` parameter.

2. Client software **SHOULD** build the initial version of the payload which has an empty transaction set.

3. Client software **SHOULD** start the process of updating the payload. The strategy of this process is implementation dependent. The default strategy is to keep the transaction set up-to-date with the state of local mempool.

4. Client software **SHOULD** stop the updating process when either a call to `engine_getPayload` with the build process's `payloadId` is made or [`SECONDS_PER_SLOT`](https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/beacon-chain.md#time-parameters-1) (12s in the Mainnet configuration) have passed since the point in time identified by the `timestamp` parameter.

## Methods

### engine_forkchoiceUpdatedV1

* status: **`Final`**

#### Request

* method: "engine_forkchoiceUpdatedV1"
* params:
  1. `forkchoiceState`: `Object` - instance of [`ForkchoiceStateV1`](#ForkchoiceStateV1)
  2. `payloadAttributes`: `Object|null` - instance of [`PayloadAttributesV1`](#PayloadAttributesV1) or `null`
* timeout: 8s

#### Response

* result: `object`
  - `payloadStatus`: [`PayloadStatusV1`](./payload.md#PayloadStatusV1); values of the `status` field in the context of this method are restricted to the following subset:
    * `"VALID"`
    * `"INVALID"`
    * `"SYNCING"`
  - `payloadId`: `DATA|null`, 8 Bytes - identifier of the payload build process or `null`
* error: code and message set in case an exception happens while the validating payload, updating the forkchoice or initiating the payload build process.

#### Specification

1. Client software **MAY** initiate a sync process if `forkchoiceState.headBlockHash` references an unknown payload or a payload that can't be validated because data that are requisite for the validation is missing. The sync process is specified in the [Sync](./payload.md#sync) section.

2. Client software **MAY** skip an update of the forkchoice state and **MUST NOT** begin a payload build process if `forkchoiceState.headBlockHash` references an ancestor of the head of canonical chain. In the case of such an event, client software **MUST** return `{payloadStatus: {status: VALID, latestValidHash: forkchoiceState.headBlockHash, validationError: null}, payloadId: null}`.

3. If `forkchoiceState.headBlockHash` references a PoW block, client software **MUST** validate this block with respect to terminal block conditions according to [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#transition-block-validity). This check maps to the transition block validity section of the EIP. Additionally, if this validation fails, client software **MUST NOT** update the forkchoice state and **MUST NOT** begin a payload build process.

4. Before updating the forkchoice state, client software **MUST** ensure the validity of the payload referenced by `forkchoiceState.headBlockHash`, and **MAY** validate the payload while processing the call. The validation process is specified in the [Payload validation](./payload.md#payload-validation) section. If the validation process fails, client software **MUST NOT** update the forkchoice state and **MUST NOT** begin a payload build process.

5. Client software **MUST** update its forkchoice state if payloads referenced by `forkchoiceState.headBlockHash` and `forkchoiceState.finalizedBlockHash` are `VALID`. The update is specified as follows:
  * The values `(forkchoiceState.headBlockHash, forkchoiceState.finalizedBlockHash)` of this method call map on the `POS_FORKCHOICE_UPDATED` event of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#block-validity) and **MUST** be processed according to the specification defined in the EIP
  * All updates to the forkchoice state resulting from this call **MUST** be made atomically.

6. Client software **MUST** return `-38002: Invalid forkchoice state` error if the payload referenced by `forkchoiceState.headBlockHash` is `VALID` and a payload referenced by either `forkchoiceState.finalizedBlockHash` or `forkchoiceState.safeBlockHash` does not belong to the chain defined by `forkchoiceState.headBlockHash`.

7. Client software **MUST** ensure that `payloadAttributes.timestamp` is greater than `timestamp` of a block referenced by `forkchoiceState.headBlockHash`. If this condition isn't held client software **MUST** respond with `-38003: Invalid payload attributes` and **MUST NOT** begin a payload build process. In such an event, the `forkchoiceState` update **MUST NOT** be rolled back.

8. Client software **MUST** begin a payload build process building on top of `forkchoiceState.headBlockHash` and identified via `buildProcessId` value if `payloadAttributes` is not `null` and the forkchoice state has been updated successfully. The build process is specified in the [Payload building](#payload-building) section.

9. Client software **MUST** respond to this method call in the following way:
  * `{payloadStatus: {status: SYNCING, latestValidHash: null, validationError: null}, payloadId: null}` if `forkchoiceState.headBlockHash` references an unknown payload or a payload that can't be validated because requisite data for the validation is missing
  * `{payloadStatus: {status: INVALID, latestValidHash: validHash, validationError: errorMessage | null}, payloadId: null}` obtained from the [Payload validation](./payload.md#payload-validation) process if the payload is deemed `INVALID`
  * `{payloadStatus: {status: INVALID, latestValidHash: 0x0000000000000000000000000000000000000000000000000000000000000000, validationError: errorMessage | null}, payloadId: null}` obtained either from the [Payload validation](./payload.md#payload-validation) process or as a result of validating a terminal PoW block referenced by `forkchoiceState.headBlockHash`
  * `{payloadStatus: {status: VALID, latestValidHash: forkchoiceState.headBlockHash, validationError: null}, payloadId: null}` if the payload is deemed `VALID` and a build process hasn't been started
  * `{payloadStatus: {status: VALID, latestValidHash: forkchoiceState.headBlockHash, validationError: null}, payloadId: buildProcessId}` if the payload is deemed `VALID` and the build process has begun
  * `{error: {code: -38002, message: "Invalid forkchoice state"}}` if `forkchoiceState` is either invalid or inconsistent
  * `{error: {code: -38003, message: "Invalid payload attributes"}}` if the payload is deemed `VALID` and `forkchoiceState` has been applied successfully, but no build process has been started due to invalid `payloadAttributes`.

10. If any of the above fails due to errors unrelated to the normal processing flow of the method, client software **MUST** respond with an error object.

### engine_forkchoiceUpdatedV2

* status: **`Draft`**

#### Request

* method: "engine_forkchoiceUpdatedV2"
* params:
  1. `forkchoiceState`: `Object` - instance of [`ForkchoiceStateV1`](#ForkchoiceStateV1)
  2. `payloadAttributes`: `Object|null` - instance of [`PayloadAttributesV2`](#PayloadAttributesV2) or `null`

#### Response

Refer to the response for [`engine_forkchoiceUpdatedV1`](#engine_forkchoiceupdatedv1).

#### Specification

This method follows the same specification as [`engine_forkchoiceUpdatedV1`](#engine_forkchoiceupdatedv1) with the exception of the following:

1. If withdrawal functionality is activated, client software **MUST** return error `-38003: Invalid payload attributes` if `payloadAttributes.withdrawals` is `null`.
   Similarly, if the functionality is not activated, client software **MUST** return error `-38003: Invalid payload attributes` if `payloadAttributes.withdrawals` is not `null`.
   Blocks without withdrawals **MUST** be expressed with an explicit empty list `[]` value.

