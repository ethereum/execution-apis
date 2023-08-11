# Engine API -- ePBS

Engine API changes introduced in ePBS, based on [Cancun](../cancun.md).

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Structures](#structures)
  - [InclusionListV1](#inclusionlistv1)
- [Methods](#methods)
  - [`engine_newPayloadVePBS`](#engine_newpayloadvepbs)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)
  - [`engine_getInclusionListVePBS`](#engine_getinclusionlistvepbs)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Structures

### InclusionListV1
This structure contains a list of transactions that are on the inclusion list. The fields are encoded as follows:
- `transactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718)

### PayloadStatusVePBS
This structure has the syntax of [`PayloadStatusV1`](../paris.md#payloadstatusv1) and appends a single field: `inclusionList`.

- `status`: `enum` - `"VALID" | "INVALID" | "SYNCING" | "ACCEPTED" | "INVALID_BLOCK_HASH"`
- `latestValidHash`: `DATA|null`, 32 Bytes - the hash of the most recent *valid* block in the branch defined by payload and its ancestors
- `validationError`: `String|null` - a message providing additional details on the validation error if the payload is classified as `INVALID` or `INVALID_BLOCK_HASH`.
- `inclusionList`: [`InclusionListV1`](#inclusionlistv1) // TODO: Should this be `InclusionListV1` or just a simple array of tranasactions? I haven't seen a nested structure in previous specs.

### InclusionListStatusV1
This structure contains the result of processing an inclusion list. The field is encoded as follow:
- `status`: `enum` - `"VALID" | "INVALID"`


## Methods

### `engine_newPayloadVePBS`

#### Request

* method: `engine_newPayloadVePBS`
* params:
  1. `executionPayload`: [`ExecutionPayloadV3`](../cancun.md#executionpayloadv3).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected blob versioned hashes to validate.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
  4. `parentBlockHash`: `DATA`, 32 Bytes - Hash of parent  block. // TODO: Maybe a better word for it to not be confused with `executionPayload.parentHash`?
* timeout: 1s

#### Response
* result: [`PayloadStatusVePBS`](#payloadstatusvepbs)
* error: code and message set in case an exception happens while processing the payload.

#### Specification

This method follows the same specification as [`engine_newPayloadV3`](../cancun.md#engine_newpayloadv3) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload is less than the timestamp of ePBS activation.
2. Client software **MUST** validate the payload if it complies with the inclusion list from previous slot. Return `{status: INVALID, latestValidHash: null, validationError: errorMessage | null, inclusionList: []}` if payload does not comply. TODO: Do we need to return another status or anything to distinguish this error from invalid block hash?
3. If `status` is anything other than `VALID`, `inclusionList` should be blank.
4. If `status` is `VALID`, client software **MUST** return `inclusionList` with a list of transactions in the previous inclusion list that remain valid after executing the payload.


### `engine_getInclusionListV1`

#### Request

* method: `engine_getInclusionListV1`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process
* timeout: 1s

#### Response

* result: [`InclusionListV1`](#inclusionlistv1)
* error: code and message set in case an exception happens while getting the inclusion list.

#### Specification
1. Given the `payloadId` client software **MUST** return a valid inclusion list with no more than `MAX_TRANSACTIONS_PER_INCLUSION_LIST` number of transactions. and total gas limit of the transactions must not exceed `INCLUSION_LIST_MAX_GAS`.
2. The call **MUST** return `-38001: Unknown payload` error if the build process identified by the `payloadId` does not exist.

### `engine_newInclusionListV1`
#### Request

* method: `engine_newInclusionListV1`
* params:
  1. `inclusionList`: [`InclusionListV1`](#inclusionlistv1) - The inclusion list to be processed.
* timeout: 1s

#### Response

* result: [`InclusionListStatusV1`](#inclsionliststatusv1) 
* error: code and message set in case an exception happens while processing the inclusion list.

#### Specification
1. Client software **MUST** validate the length of `transactions` in `inclusionList` to be less than or equal `MAX_TRANSACTIONS_PER_INCLUSION_LIST`.
2. Client software **MUST** validate the total gas limit of `transactions` in `inclusionList` to be less than or equal to `INCLUSION_LIST_MAX_GAS`.
3. Client software **MUST** respond to this method call in the following way:
    * `{status: INVALID}` if `inclusionList` validation has failed.
    * `{status: VALID}` if `inclusionList` validation has succeeded.
4. If any of the above fails due to errors unrelated to the normal processing flow of the method, client software **MUST** respond with an error object.
