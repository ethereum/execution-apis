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
  - [InclusionListClaimV1](#inclusionlistclaimv1)
- [Routines](#routines)
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
| `MAX_INCLUSION_LIST_CLAIMS` | `uint64(4096) = 2**12` |
| `MAX_VERIFY_GAS_PER_IL` | `uint64(1048576) = 2**20` |

The VERIFY budget and claim cap are defined in [EIP-7805](https://eips.ethereum.org/EIPS/eip-7805#constants), including their activation conditions.

## Structures

### PayloadAttributesV5

This structure has the syntax of [`PayloadAttributesV4`](./amsterdam.md#payloadattributesv4) and appends the fields `inclusionListTransactions` and `inclusionListMembership`.

- `timestamp`: `QUANTITY`, 64 Bits - value for the `timestamp` field of the new payload
- `prevRandao`: `DATA`, 32 Bytes - value for the `prevRandao` field of the new payload
- `suggestedFeeRecipient`: `DATA`, 20 Bytes - suggested value for the `feeRecipient` field of the new payload
- `withdrawals`: `Array of WithdrawalV1` - Array of withdrawals, each object is an `OBJECT` containing the fields of a `WithdrawalV1` structure.
- `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
- `slotNumber`: `QUANTITY`, 64 Bits - value for the `slotNumber` field of the new payload
- `targetGasLimit`: `QUANTITY`, 64 Bits - target value for the `gasLimit` field of the new payload
- `inclusionListTransactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718).
- `inclusionListMembership`: `Array of DATA`, 2 Bytes - Array of inclusion list memberships, one per element of `inclusionListTransactions` at the same position. Each element is a bitvector of length `INCLUSION_LIST_COMMITTEE_SIZE`, encoded as SSZ `Bitvector[INCLUSION_LIST_COMMITTEE_SIZE]`, in which bit `i` is set if and only if the inclusion list of the inclusion list committee member at committee position `i` carries the transaction.

The transaction list contains distinct, nonempty byte strings; its order is unrestricted. Membership uses only the inclusion lists selected by the consensus layer under EIP-7805. Bit `i` is bit `i % 8` of byte `i // 8`, with bit zero the least significant bit: committee positions `{0, 9}` encode as `0x0102`, and `{15}` as `0x0080`. The same rules apply to `engine_newPayloadV6`.

### PayloadStatusV2

This structure has the syntax of `PayloadStatusV1` and appends a single field: `inclusionListSatisfied`

- `status`: `enum` - `"VALID" | "INVALID" | "SYNCING" | "ACCEPTED"`
- `latestValidHash`: `DATA|null`, 32 Bytes - the hash of the most recent *valid* block in the branch defined by payload and its ancestors
- `validationError`: `String|null` - a message providing additional details on the validation error if the payload is classified as `INVALID`.
- `inclusionListSatisfied`: `BOOLEAN|null` - whether the payload satisfied the inclusion list constraints if it is deemed `VALID`; `null` otherwise.

### InclusionListClaimV1

This structure maps onto the inclusion list claim defined in [EIP-7805](https://eips.ethereum.org/EIPS/eip-7805). The fields are encoded as follows:

- `transactionHash`: `DATA`, 32 Bytes - `keccak256` of the claimed inclusion list transaction, `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718).
- `transactionIndex`: `QUANTITY`, 64 Bits - index in `executionPayload.transactions` at which the omission of the claimed transaction is evaluated.

Index zero is before the first transaction; `len(executionPayload.transactions)` is the end of the payload. The claim list retains its committed order and duplicates. The consensus layer checks its commitment; the execution layer applies EIP-7805's evaluation rules.

## Routines

### Payload building

This routine follows the same specification as [Payload building](./paris.md#payload-building) with the following changes to the processing flow:

1. Client software **MUST** take `inclusionListTransactions` into account during the payload build process. The built `ExecutionPayload`, together with the `inclusionListClaims` returned by [`engine_getPayloadV7`](#engine_getpayloadv7), **MUST** satisfy the inclusion list constraints with respect to `inclusionListTransactions` and `inclusionListMembership` as defined in [EIP-7805](https://eips.ethereum.org/EIPS/eip-7805).

## Methods

### engine_newPayloadV6

Method parameter list is extended with `inclusionListTransactions`, `inclusionListMembership` and `inclusionListClaims`.

#### Request

* method: `engine_newPayloadV6`
* params:
  1. `executionPayload`: [`ExecutionPayloadV4`](./amsterdam.md#executionpayloadv4).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected blob versioned hashes to validate.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
  4. `executionRequests`: `Array of DATA` - List of execution layer triggered requests. Each list element is a `requests` byte array as defined by [EIP-7685](https://eips.ethereum.org/EIPS/eip-7685). The first byte of each element is the `request_type` and the remaining bytes are the `request_data`. Elements of the list **MUST** be ordered by `request_type` in ascending order. Elements with empty `request_data` **MUST** be excluded from the list.
  5. `inclusionListTransactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718).
  6. `inclusionListMembership`: `Array of DATA`, 2 Bytes - Array of inclusion list memberships, one per element of `inclusionListTransactions` at the same position. Each element is a bitvector of length `INCLUSION_LIST_COMMITTEE_SIZE`, encoded as SSZ `Bitvector[INCLUSION_LIST_COMMITTEE_SIZE]`, in which bit `i` is set if and only if the inclusion list of the inclusion list committee member at committee position `i` carries the transaction.
  7. `inclusionListClaims`: `Array of InclusionListClaimV1` - Array of inclusion list claims committed to by the builder of the payload, as defined in [EIP-7805](https://eips.ethereum.org/EIPS/eip-7805).
* timeout: 6s

#### Response

* result: [`PayloadStatusV2`](#payloadstatusv2)
* error: code and message set in case an exception happens while processing the payload.

#### Specification

This method follows the same specification as [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload does not fall within the time frame of the Bogota fork.

2. Client software **MUST** return `-32602: Invalid params` error if any of the following holds:

    1. `inclusionListTransactions` contains an empty or duplicate byte string, or `inclusionListMembership` does not have the same number of elements as `inclusionListTransactions`.

    2. An element of `inclusionListMembership` is not exactly `INCLUSION_LIST_COMMITTEE_SIZE / 8` bytes, or has no bit set.

    3. For some committee position `i`, the sum of the byte lengths of the transactions whose membership has bit `i` set exceeds `MAX_TRANSACTIONS_BYTES_PER_INCLUSION_LIST`.

    4. `inclusionListClaims` has more than `MAX_INCLUSION_LIST_CLAIMS` elements or an element does not match `InclusionListClaimV1`, including a malformed hash or a noncanonical or overflowing `uint64` index.

   Otherwise well-formed transaction byte strings **MUST NOT** cause an RPC error solely because their transaction encoding or validity checks fail. Well-formed duplicate or unmatched claims, and indices above the payload transaction count, **MUST NOT** cause an RPC error: EIP-7805 ignores repeated transaction hashes and irrelevant claims, and clamps remaining indices to the end of the payload. Exceeding the VERIFY budget is likewise handled by EIP-7805's per-member fill, not as an RPC error.

3. Client software **MUST** set `inclusionListSatisfied` in the following way:

    1. If the payload is deemed `VALID`, `inclusionListSatisfied` **MUST** be set to whether the payload satisfied the inclusion list constraints with respect to this call's `inclusionListTransactions`, `inclusionListMembership` and `inclusionListClaims`, even if execution validity was already known. A failed inclusion list check **MUST NOT** change execution validity to `INVALID`.

    2. Otherwise, `inclusionListSatisfied` **MUST** be `null`.

4. An inclusion list verdict is specific to the supplied transactions, membership and claims, not just the execution block hash. After an `ACCEPTED` or `SYNCING` response, the consensus layer **MUST** retain these inputs for the relevant beacon block and repeat this method with them once execution validation can complete. `engine_forkchoiceUpdatedV5` does not return an inclusion list verdict.

### engine_getPayloadV7

This method is updated to return `inclusionListClaims`.

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
  - `inclusionListClaims`: `Array of InclusionListClaimV1` - Inclusion list claims for transactions omitted from `executionPayload`.
* error: code and message set in case an exception happens while getting the payload.

#### Specification

This method follows the same specification as [`engine_getPayloadV6`](./amsterdam.md#engine_getpayloadv6) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the built payload does not fall within the time frame of the Bogota fork.

2. `inclusionListClaims` **MUST NOT** have more than `MAX_INCLUSION_LIST_CLAIMS` elements. Each claim **MUST** identify a distinct omitted Profile 2 candidate from the `inclusionListTransactions` used to build the payload, with an index no greater than `len(executionPayload.transactions)` at which its omission is excused under EIP-7805. When Profile 2 is inactive under EIP-7805, the list **MUST** be empty.

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

    4. When Profile 2 is active under EIP-7805, the total VERIFY budget cost of the distinct Profile 2 candidates in the transaction list **SHOULD NOT** exceed `MAX_VERIFY_GAS_PER_IL`, using EIP-7805's candidate and cost definitions. This is a selection recommendation, not an inclusion list validity rule.

2. The strategy for selecting transactions is implementation dependent.

### engine_forkchoiceUpdatedV5

#### Request

* method: `engine_forkchoiceUpdatedV5`
* params:
  1. `forkchoiceState`: [`ForkchoiceStateV1`](./paris.md#forkchoicestatev1).
  2. `payloadAttributes`: `Object|null` - Instance of [`PayloadAttributesV5`](#payloadattributesv5) or `null`.
  3. `custodyColumns`: `DATA|null`, 16 Bytes - Interpreted as a bitarray of length `CELLS_PER_EXT_BLOB` indicating which column indices form the CL's custody set, or `null` if the CL does not provide custody services.
* timeout: 8s

#### Response

* result: `object`
  - `payloadStatus`: `PayloadStatusV1`, with the same `status` value restrictions as [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4).
  - `payloadId`: `DATA|null`, 8 Bytes - identifier of the payload build process or `null`
* error: code and message set in case an exception happens while the validating payload, updating the forkchoice or initiating the payload build process.

#### Specification

This method follows the same specification as [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4) with the following changes to the processing flow:

1. Extend point (8) of the `engine_forkchoiceUpdatedV1` [specification](./paris.md#specification-1) by defining the following sequence of checks that **MUST** be run over `payloadAttributes`:

    1. `payloadAttributes` matches the [`PayloadAttributesV5`](#payloadattributesv5) structure and `payloadAttributes.inclusionListTransactions` and `payloadAttributes.inclusionListMembership` pass checks (2.1) through (2.3) of [`engine_newPayloadV6`](#engine_newpayloadv6), return `-38003: Invalid payload attributes` on failure. Invalid transaction encodings or transaction validity failures are handled by EIP-7805, not rejected as invalid attributes.

    2. `payloadAttributes.timestamp` falls within the time frame of the Bogota fork, return `-38005: Unsupported fork` on failure.

### Update the methods of previous forks

#### Amsterdam API

For the following methods:

- [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5)
- [`engine_getPayloadV6`](./amsterdam.md#engine_getpayloadv6)
- [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4)

a validation **MUST** be added:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of payload is greater than or equal to the Bogota activation timestamp.
