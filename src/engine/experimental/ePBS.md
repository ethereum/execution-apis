# Engine API -- ePBS

Engine API changes introduced in ePBS, based on [Cancun](../cancun.md).

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Structures](#structures)
  - [InclusionListV1](#inclusionlistv1)
  - [InclusionListStatusV1](#inclusionliststatusv1)

- [Methods](#methods)
  - [`engine_getInclusionListV1`](#engine_getinclusionlistv1)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Structures

### InclusionListV1
This structure contains a list of transactions that are on the inclusion list. The fields are encoded as follows:
- `transactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718)

### InclusionListStatusV1
This structure contains the result of processing an inclusion list. The field is encoded as follow:
- `status`: `enum` - `"VALID" | "INVALID"`


## Methods

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

* result: [`InclusionListStatusV1`](#inclusionliststatusv1) 
* error: code and message set in case an exception happens while processing the inclusion list.

#### Specification
1. Client software **MUST** validate the length of `transactions` in `inclusionList` to be less than or equal `MAX_TRANSACTIONS_PER_INCLUSION_LIST`.
2. Client software **MUST** validate the total gas limit of `transactions` in `inclusionList` to be less than or equal to `INCLUSION_LIST_MAX_GAS`.
3. Client software **MUST** respond to this method call in the following way:
    * `{status: INVALID}` if `inclusionList` validation has failed.
    * `{status: VALID}` if `inclusionList` validation has succeeded.
4. If any of the above fails due to errors unrelated to the normal processing flow of the method, client software **MUST** respond with an error object.
