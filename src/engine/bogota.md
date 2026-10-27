# Engine API -- Bogota

Engine API changes introduced in Bogota.

This specification is based on and extends [Engine API - Amsterdam](./amsterdam.md) specification.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Constants](#constants)
- [Structures](#structures)
  - [PayloadAttributesV5](#payloadattributesv5)
  - [PayloadStatusV2](#payloadstatusv2)
  - [OmissionClaimV1](#omissionclaimv1)
- [Routines](#routines)
  - [Inclusion checking](#inclusion-checking)
  - [Payload building](#payload-building)
- [Methods](#methods)
  - [engine_newPayloadV6](#engine_newpayloadv6)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)
  - [engine_getPayloadV7](#engine_getpayloadv7)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)
  - [engine_getInclusionListV1](#engine_getinclusionlistv1)
    - [Request](#request-2)
    - [Response](#response-2)
    - [Specification](#specification-2)
  - [engine_forkchoiceUpdatedV5](#engine_forkchoiceupdatedv5)
    - [Request](#request-3)
    - [Response](#response-3)
    - [Specification](#specification-3)
  - [Update the methods of previous forks](#update-the-methods-of-previous-forks)
    - [Amsterdam API](#amsterdam-api)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Constants

| Name | Value |
| - | - |
| `MAX_TRANSACTIONS_BYTES_PER_INCLUSION_LIST` |  `uint64(8192) = 2**13` |
| `INCLUSION_LIST_COMMITTEE_SIZE` | `uint64(16) = 2**4` |
| `MAX_OMISSION_CLAIMS` | `uint64(4096) = 2**12` |
| `MAX_VERIFY_GAS_PER_IL` | `uint64(1048576) = 2**20` |
| `MAX_VERIFY_GAS_PER_TX` | `uint64(262144) = 2**18` |

Both EIP-7805 profiles apply from Bogota activation. The VERIFY budgets and claim cap are defined in [EIP-7805](https://eips.ethereum.org/EIPS/eip-7805#constants).

`engine_newPayloadV6`, `PayloadAttributesV5` and `engine_forkchoiceUpdatedV5` replace the earlier FOCIL devnet shapes in place: the flat `inclusionListTransactions` list becomes separate `inclusionLists` plus `omissionClaims`, and `engine_forkchoiceUpdatedV5` returns `PayloadStatusV1` again. Test fixtures and releases built for the earlier shapes **MUST** be regenerated.

## Structures

### PayloadAttributesV5

This structure has the syntax of [`PayloadAttributesV4`](./amsterdam.md#payloadattributesv4) and appends `inclusionLists`.

- `timestamp`: `QUANTITY`, 64 Bits - value for the `timestamp` field of the new payload
- `prevRandao`: `DATA`, 32 Bytes - value for the `prevRandao` field of the new payload
- `suggestedFeeRecipient`: `DATA`, 20 Bytes - suggested value for the `feeRecipient` field of the new payload
- `withdrawals`: `Array of WithdrawalV1` - Array of withdrawals, each object is an `OBJECT` containing the fields of a `WithdrawalV1` structure.
- `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
- `slotNumber`: `QUANTITY`, 64 Bits - value for the `slotNumber` field of the new payload
- `targetGasLimit`: `QUANTITY`, 64 Bits - target value for the `gasLimit` field of the new payload
- `inclusionLists`: `Array of Array of DATA` - Exactly `INCLUSION_LIST_COMMITTEE_SIZE` transaction lists, in committee-position order. Each transaction is the exact nonempty EIP-2718 envelope or legacy RLP bytes. An empty list denotes no applicable IL at that position.

The consensus layer selects valid, non-equivocating ILs under EIP-7805. A member holding several committee positions supplies the same list at each. Entries retain their order and duplicates; the per-list byte limit applies before deduplication. The execution layer fills each list's budget independently and may share decoding, signature verification and replay work across lists. The same rules apply to `engine_newPayloadV6`.

### PayloadStatusV2

This structure has the syntax of `PayloadStatusV1` and appends `inclusionListSatisfied`.

- `status`: `enum` - `"VALID" | "INVALID" | "SYNCING" | "ACCEPTED"`
- `latestValidHash`: `DATA|null`, 32 Bytes - the hash of the most recent *valid* block in the branch defined by payload and its ancestors
- `validationError`: `String|null` - a message providing additional details on the validation error if the payload is classified as `INVALID`.
- `inclusionListSatisfied`: `BOOLEAN|null` - the completed inclusion verdict when `status` is `VALID`; `null` otherwise.

### OmissionClaimV1

This structure maps onto the omission claim defined in [EIP-7805](https://eips.ethereum.org/EIPS/eip-7805). The fields are encoded as follows:

- `transactionHash`: `DATA`, 32 Bytes - Ordinary transaction hash of the claimed inclusion list transaction.
- `evaluationIndex`: `QUANTITY`, 64 Bits - index in `executionPayload.transactions` at which the omission of the claimed transaction is evaluated.

Index zero is before the first transaction; `len(executionPayload.transactions)` is the end of the payload. The claim list retains its committed order and duplicates. The consensus layer checks its commitment; the execution layer applies EIP-7805's evaluation rules.

## Routines

### Inclusion checking

A `VALID` response to `engine_newPayloadV6` **MUST** include the completed inclusion verdict for that call's separate ILs and omission claims. An unsatisfied inclusion list **MUST NOT** change execution validity or forkchoice processing. Forkchoice responses do not report inclusion satisfaction.

Checking may require re-execution if prior execution did not retain the required insertion states and both gas counters. Authenticated BAL-indexed reconstruction can avoid this only with the corresponding gas counters available. Missing local data **MUST NOT** excuse an omission or produce a `VALID` response before inclusion checking completes; normal execution-sync statuses and RPC timeout/error handling apply.

### Payload building

This routine follows the same specification as [Payload building](./paris.md#payload-building) with the following changes to the processing flow:

1. Client software **MUST** take `inclusionLists` into account during the payload build process. The built `ExecutionPayload`, together with the `omissionClaims` returned by [`engine_getPayloadV7`](#engine_getpayloadv7), **MUST** satisfy the inclusion list constraints with respect to `inclusionLists` as defined in [EIP-7805](https://eips.ethereum.org/EIPS/eip-7805).

## Methods

### engine_newPayloadV6

Method parameter list is extended with `inclusionLists` and `omissionClaims`.

#### Request

* method: `engine_newPayloadV6`
* params:
  1. `executionPayload`: [`ExecutionPayloadV4`](./amsterdam.md#executionpayloadv4).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected blob versioned hashes to validate.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
  4. `executionRequests`: `Array of DATA` - List of execution layer triggered requests. Each list element is a `requests` byte array as defined by [EIP-7685](https://eips.ethereum.org/EIPS/eip-7685). The first byte of each element is the `request_type` and the remaining bytes are the `request_data`. Elements of the list **MUST** be ordered by `request_type` in ascending order. Elements with empty `request_data` **MUST** be excluded from the list.
  5. `inclusionLists`: `Array of Array of DATA` - Separate ILs as defined in [`PayloadAttributesV5`](#payloadattributesv5).
  6. `omissionClaims`: `Array of OmissionClaimV1` - The ordered omission claims committed to by the builder, as defined in EIP-7805.
* timeout: 6s

#### Response

* result: [`PayloadStatusV2`](#payloadstatusv2)
* error: code and message set in case an exception happens while processing the payload.

#### Specification

This method follows the same specification as [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload does not fall within the time frame of the Bogota fork.

2. Client software **MUST** return `-32602: Invalid params` error if any of the following holds:

    1. `inclusionLists` does not have exactly `INCLUSION_LIST_COMMITTEE_SIZE` entries, an entry is not a list of nonempty byte strings, or the total transaction byte length of any entry exceeds `MAX_TRANSACTIONS_BYTES_PER_INCLUSION_LIST`.

    2. `omissionClaims` has more than `MAX_OMISSION_CLAIMS` elements or an element does not match `OmissionClaimV1`, including a `transactionHash` that is not 32 bytes of `DATA`, or an `evaluationIndex` that does not follow the `QUANTITY` [encoding](./common.md#encoding) or exceeds `2**64 - 1`. For example, an `evaluationIndex` of `"0x01"`, `"0x"` or `"0x10000000000000000"` **MUST** return `-32602`.

   Otherwise well-formed transaction byte strings **MUST NOT** cause an RPC error solely because their transaction encoding or validity checks fail. Well-formed duplicate or unmatched claims, and indices above the payload transaction count, **MUST NOT** cause an RPC error: EIP-7805 ignores repeated transaction hashes and irrelevant claims, and clamps remaining indices to the end of the payload. Exceeding the VERIFY budget is likewise handled by EIP-7805's per-member fill, not as an RPC error.

3. Client software **MUST** evaluate inclusion against this call's inputs, even if execution validity is already known. It **MUST** complete inclusion checking before returning `VALID`, with `inclusionListSatisfied` set to the boolean verdict. For any other execution status, `inclusionListSatisfied` **MUST** be `null`.

### engine_getPayloadV7

This method is updated to return `omissionClaims`.

#### Request

* method: `engine_getPayloadV7`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process
* timeout: 1s

#### Response

* result: `object`
  - `executionPayload`: [`ExecutionPayloadV4`](./amsterdam.md#executionpayloadv4)
  - `blockValue` : `QUANTITY`, 256 Bits - The expected value to be received by the `feeRecipient` in wei
  - `blobsBundle`: [`BlobsBundleV2`](./osaka.md#blobsbundlev2) - Bundle with data corresponding to blob transactions included into `executionPayload`
  - `shouldOverrideBuilder` : `BOOLEAN` - Suggestion from the execution layer to use this `executionPayload` instead of an externally provided one
  - `executionRequests`: `Array of DATA` - Execution layer triggered requests obtained from the `executionPayload` transaction execution.
  - `omissionClaims`: `Array of OmissionClaimV1` - Omission claims for transactions omitted from `executionPayload`.
* error: code and message set in case an exception happens while getting the payload.

#### Specification

This method follows the same specification as [`engine_getPayloadV6`](./amsterdam.md#engine_getpayloadv6) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the built payload does not fall within the time frame of the Bogota fork.

2. `omissionClaims` **MUST NOT** have more than `MAX_OMISSION_CLAIMS` elements. The EL **MUST** emit only claims it has verified: each identifies a distinct omitted Profile 2 candidate from the build's `inclusionLists`, with an index no greater than `len(executionPayload.transactions)` at which EIP-7805 excuses its omission.

### engine_getInclusionListV1

#### Request

* method: `engine_getInclusionListV1`
* params: []
* timeout: 1s

#### Response

* result: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718).
* error: code and message set in case an exception happens while getting the inclusion list.

#### Specification

1. Client software **MUST** provide, based on its local view of the mempool, a list of transactions for the inclusion list satisfying the following conditions:

    1. Every transaction **MUST** have non-zero length (at least 1 byte).

    2. The transaction list **MUST NOT** include any [blob transaction](https://eips.ethereum.org/EIPS/eip-4844#blob-transaction).

    3. The total byte length of the transaction list **MUST NOT** exceed `MAX_TRANSACTIONS_BYTES_PER_INCLUSION_LIST`.

    4. The total VERIFY budget cost of the distinct Profile 2 candidates in the transaction list **SHOULD NOT** exceed `MAX_VERIFY_GAS_PER_IL`, using EIP-7805's candidate and cost definitions. This is a selection recommendation, not an inclusion list validity rule.

2. The strategy for selecting transactions is implementation dependent.

### engine_forkchoiceUpdatedV5

Compared with V4, only the payload attributes change to `PayloadAttributesV5`; the response remains `PayloadStatusV1`.

#### Request

* method: `engine_forkchoiceUpdatedV5`
* params:
  1. `forkchoiceState`: [`ForkchoiceStateV1`](./paris.md#forkchoicestatev1).
  2. `payloadAttributes`: `Object|null` - Instance of [`PayloadAttributesV5`](#payloadattributesv5) or `null`.
  3. `custodyColumns`: `DATA|null`, 16 Bytes - Interpreted as a bitarray of length `CELLS_PER_EXT_BLOB` indicating which column indices form the CL's custody set, or `null` if the CL does not provide custody services.
* timeout: 8s

#### Response

* result: `object`
  - `payloadStatus`: [`PayloadStatusV1`](./paris.md#payloadstatusv1), with the same `status` value restrictions as [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4).
  - `payloadId`: `DATA|null`, 8 Bytes - identifier of the payload build process or `null`
* error: code and message set in case an exception happens while the validating payload, updating the forkchoice or initiating the payload build process.

#### Specification

This method follows the same specification as [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4) with the following changes to the processing flow:

This method reports execution validity only. Inclusion results are returned by `engine_newPayloadV6`; `payloadAttributes` describes the child build and cannot supply the head's inclusion inputs.

1. Extend point (8) of the `engine_forkchoiceUpdatedV1` [specification](./paris.md#specification-1) by defining the following sequence of checks that **MUST** be run over `payloadAttributes`:

    1. `payloadAttributes` matches the [`PayloadAttributesV5`](#payloadattributesv5) structure and `payloadAttributes.inclusionLists` passes check (2.1) of [`engine_newPayloadV6`](#engine_newpayloadv6), return `-38003: Invalid payload attributes` on failure. Invalid transaction encodings or transaction validity failures are handled by EIP-7805, not rejected as invalid attributes.

    2. `payloadAttributes.timestamp` falls within the time frame of the Bogota fork, return `-38005: Unsupported fork` on failure.

### Update the methods of previous forks

#### Amsterdam API

For the following methods:

- [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5)
- [`engine_getPayloadV6`](./amsterdam.md#engine_getpayloadv6)
- [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4)

a validation **MUST** be added:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of payload is greater than or equal to the Bogota activation timestamp.
