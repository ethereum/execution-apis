# Engine API -- ePBS

Engine API changes introduced in ePBS, based on [Cancun](../cancun.md).

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Structures](#structures)
  - [InclusionListV1](#inclusionlistv1)
  - [ExecutionPayloadVePBS](#executionpayloadvepbs)
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


## Methods

### `engine_newPayloadVePBS`

#### Request

* method: `engine_newPayloadVePBS`
* params:
  1. `executionPayload`: [`ExecutionPayloadV3`](../cancun.md#executionpayloadv3).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected blob versioned hashes to validate.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
  4. `parentBlockHash`: `DATA`, 32 Bytes - Hash of parent  block. // TODO: Maybe a better word for it to not be confused with `executionPayload.parentHash`?
  5. `inclusionList`: [`InclusionListV1`][#inclusionlistv1]
* timeout: 1s

#### Response

Refer to the response for [`engine_newPayloadV3`](../cancun.md#engine_newpayloadv3).

#### Specification

This method follows the same specification as [`engine_newPayloadV3`](../cancun.md#engine_newpayloadv3) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload is less than the timestamp of ePBS activation.
2. Client software **MUST** return ? if the length of `transactions` in `inclusionList` exceeds `MAX_TRANSACTIONS_PER_INCLUSION_LIST`, or total gas of `transactions` in `inclusionList` exceeds `INCLUSION_LIST_MAX_GAS`. TODO: Possibly introduce a new error code for invalid inclusion list
3. Client software **MUST** validate the payload if it complies with the inclusion list from previous slot. Return `{status: INVALID, latestValidHash: null, validationError: errorMessage | null}` if payload does not comply. TODO: Do we need to return another status or anything to distinguish this error from invalid block hash?


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
1. Given the `payloadId` client software **MUST** return a valid inclusion list with no more than `MAX_TRANSACTIONS_PER_INCLUSION_LIST` number of transactions.
2. The call **MUST** return `-38001: Unknown payload` error if the build process identified by the `payloadId` does not exist.
