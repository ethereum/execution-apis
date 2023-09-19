# Engine API -- ePBS

Engine API changes introduced in ePBS, based on [Cancun](../cancun.md).

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Structures](#structures)
  - [InclusionListV1](#inclusionlistv1)
  - [InclusionListSummaryEntryV1](#inclusionlistsummaryentryv1)
  - [InclusionListStatusV1](#inclusionliststatusv1)
  - [PayloadAttributesVePBS](#payloadattributesvepbs)
- [Methods](#methods)
  - [`engine_getInclusionListV1`](#engine_getinclusionlistv1)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)
  - [`engine_newInclusionListV1`](#engine_newinclusionlistv1)
    - [Request](#request-2)
    - [Response](#response-2)
    - [Specification](#specification-2)
  - [`engine_forkchoiceUpdatedVePBS`](#engine_forkchoiceupdatedvepbs)
    - [Request](#request-3)
    - [Response](#response-3)
    - [Specification](#specification-3)


<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Structures

### InclusionListV1
This structure maps onto inclusion list object from [{pending EIP}](). The fields are encoded as follows:
- `transactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718)
- `summary`: `Array of InclusionListSummaryEntryV1` - Array of summary entries. Each object is an `OBJECT` containing the fields of a `InclusionListSummaryEntryV1` structure.

### InclusionListSummaryEntryV1
This structure maps onto inclusion list summary entry object from [{pending EIP}](). The fields are encoded as follows:
- `address`: `DATA`, 20 Bytes
- `gasLimit`: `QUANTITY`, 64 Bits

### InclusionListStatusV1
This structure contains the result of processing an inclusion list. The field is encoded as follow:
- `status`: `enum` - `"VALID" | "INVALID"`
- `validationError`: `String|null` - a message providing additional details on the validation error if the payload is classified as `INVALID`.

### PayloadAttributesVePBS

This structure has the syntax of [`PayloadAttributesV3`](./cancun.md#payloadattributesv3) and appends a single field: `inclusionListSummary`.

- `timestamp`: `QUANTITY`, 64 Bits - value for the `timestamp` field of the new payload
- `prevRandao`: `DATA`, 32 Bytes - value for the `prevRandao` field of the new payload
- `suggestedFeeRecipient`: `DATA`, 20 Bytes - suggested value for the `feeRecipient` field of the new payload
- `withdrawals`: `Array of WithdrawalV1` - Array of withdrawals, each object is an `OBJECT` containing the fields of a `WithdrawalV1` structure.
- `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
- `inclusionListSummary`: `Array of InclusionListSummaryEntryV1` - Array of summary entries. Each object is an `OBJECT` containing the fields of a `InclusionListSummaryEntryV1` structure.


## Methods

### `engine_getInclusionListV1`

#### Request

* method: `engine_getInclusionListV1`
* params:
  1. `parentHash`: `DATA`, 32 Bytes - hash of the block which the returning inclusion list bases on
* timeout: 1s

#### Response

* result: [`InclusionListV1`](#inclusionlistv1)
* error: code and message set in case an exception happens while getting the inclusion list.

#### Specification
1. Client software **MUST** return the most recent version of the inclusion list based on `parentHash`.
2. The call **MUST** return `-38001: Unknown payload` error if the build process identified by the `parent_block_hash` does not exist.

### `engine_newInclusionListV1`
#### Request

* method: `engine_newInclusionListV1`
* params:
  1. `inclusionList`: [`InclusionListV1`](#inclusionlistv1) - The inclusion list to be processed.
  2. `parentHash`: `DATA`, 32 Bytes - hash of the block whose corresponding execution state will be used to validate against the inclusion list.
* timeout: 1s

#### Response

* result: [`InclusionListStatusV1`](#inclusionliststatusv1) 
* error: code and message set in case an exception happens while processing the inclusion list.

#### Specification
1. Client software **MUST** respond to this method call in the following way:
    * `{status: INVALID}` if `inclusionList` validation has failed.
    * `{status: VALID}` if `inclusionList` validation has succeeded.
2. If any of the above fails due to errors unrelated to the normal processing flow of the method, client software **MUST** respond with an error object.

### `engine_forkchoiceUpdatedVePBS`
#### Request

* method: `engine_forkchoiceUpdatedVePBS`
* params:
  1. `forkchoiceState`: [`ForkchoiceStateV1`](./paris.md#ForkchoiceStateV1).
  2. `payloadAttributes`: `Object|null` - Instance of [`PayloadAttributesVePBS`](#payloadattributesvepbs) or `null`.
* timeout: 8s

#### Response

Refer to the response for [`engine_forkchoiceUpdatedV3`](./cancun.md#engine_forkchoiceupdatedv3).

#### Specification

This method follows the same specification as [`engine_forkchoiceUpdatedV3`](./cancun.md#engine_forkchoiceupdatedv3) with addition of the following:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `payloadAttributes` is set and the `payloadAttributes.timestamp` does not fall within the time frame of the ePBS fork.

