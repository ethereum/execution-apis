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
  - [InclusionListProofCommitmentV1](#inclusionlistproofcommitmentv1)
  - [InclusionListProofV1](#inclusionlistproofv1)
  - [InclusionListProofStatusV1](#inclusionlistproofstatusv1)
- [Routines](#routines)
  - [Payload building](#payload-building)
  - [Inclusion list proof verdicts and retention](#inclusion-list-proof-verdicts-and-retention)
- [Methods](#methods)
  - [engine_newPayloadV6](#engine_newpayloadv6)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)
  - [engine_getInclusionListV1](#engine_getinclusionlistv1)
    - [Request](#request-1)
    - [Response](#response-1)
    - [Specification](#specification-1)
  - [engine_forkchoiceUpdatedV5](#engine_forkchoiceupdatedv5)
    - [Request](#request-2)
    - [Response](#response-2)
    - [Specification](#specification-2)
  - [engine_verifyInclusionListProofV1](#engine_verifyinclusionlistproofv1)
    - [Request](#request-3)
    - [Response](#response-3)
    - [Specification](#specification-3)
  - [engine_getInclusionListProofV1](#engine_getinclusionlistproofv1)
    - [Request](#request-4)
    - [Response](#response-4)
    - [Specification](#specification-4)
  - [engine_submitInclusionListProofV1](#engine_submitinclusionlistproofv1)
    - [Request](#request-5)
    - [Response](#response-5)
    - [Specification](#specification-5)
  - [Update the methods of previous forks](#update-the-methods-of-previous-forks)
    - [Amsterdam API](#amsterdam-api)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Constants

| Name | Value |
| - | - |
| `MAX_TRANSACTIONS_BYTES_PER_INCLUSION_LIST` |  `uint64(8192) = 2**13` |
| `MAX_AGGREGATE_PROOF_BYTES` | `uint64(1310720) = 5 * 2**18` |
| `MAX_DEPS_PER_AGGREGATE_WRAPPER` | `uint64(4096) = 2**12` |

`MAX_AGGREGATE_PROOF_BYTES` and `MAX_DEPS_PER_AGGREGATE_WRAPPER` are the [EIP-8288](https://eips.ethereum.org/EIPS/eip-8288) constants of the same names and are provisional there.

## Structures

### PayloadAttributesV5

This structure has the syntax of [`PayloadAttributesV4`](./amsterdam.md#payloadattributesv4) and appends a single field: `inclusionListTransactions`.

- `timestamp`: `QUANTITY`, 64 Bits - value for the `timestamp` field of the new payload
- `prevRandao`: `DATA`, 32 Bytes - value for the `prevRandao` field of the new payload
- `suggestedFeeRecipient`: `DATA`, 20 Bytes - suggested value for the `feeRecipient` field of the new payload
- `withdrawals`: `Array of WithdrawalV1` - Array of withdrawals, each object is an `OBJECT` containing the fields of a `WithdrawalV1` structure.
- `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
- `slotNumber`: `QUANTITY`, 64 Bits - value for the `slotNumber` field of the new payload
- `targetGasLimit`: `QUANTITY`, 64 Bits - target value for the `gasLimit` field of the new payload
- `inclusionListTransactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718).

### PayloadStatusV2

This structure has the syntax of `PayloadStatusV1` and appends a single field: `inclusionListSatisfied`

- `status`: `enum` - `"VALID" | "INVALID" | "SYNCING" | "ACCEPTED"`
- `latestValidHash`: `DATA|null`, 32 Bytes - the hash of the most recent *valid* block in the branch defined by payload and its ancestors
- `validationError`: `String|null` - a message providing additional details on the validation error if the payload is classified as `INVALID`.
- `inclusionListSatisfied`: `BOOLEAN|null` - whether the payload satisfied the inclusion list constraints if it is deemed `VALID`; `null` otherwise.

### InclusionListProofCommitmentV1

This structure is an inclusion list's commitment to its aggregate proof, as defined in [EIP-8288](https://eips.ethereum.org/EIPS/eip-8288).

- `starkProofHash`: `DATA`, 32 Bytes - `keccak256` of the committed aggregate proof
- `depsHash`: `DATA`, 32 Bytes - `get_deps_hash` of the committed proof's dependency list

### InclusionListProofV1

This structure is an inclusion list proof, the `il_proof` of [EIP-8288](https://eips.ethereum.org/EIPS/eip-8288): an aggregate proof with its dependency list.

- `starkProof`: `DATA`, at most `MAX_AGGREGATE_PROOF_BYTES` bytes - the serialized recursive aggregate proof
- `proofDeps`: `Array of DATA`, 96 Bytes each - the proof's dependency list, each entry in the EIP-8288 96-byte dependency encoding, sorted in ascending byte order without duplicates, at most `MAX_DEPS_PER_AGGREGATE_WRAPPER` entries

The size bounds are checked by the methods after parameter decoding, not by the schema. A `proofDeps` entry that is not exactly 96 bytes is a parameter decoding error. An `InclusionListProofV1` matches an `InclusionListProofCommitmentV1` if `keccak256(starkProof) == starkProofHash` and `get_deps_hash(proofDeps) == depsHash`.

### InclusionListProofStatusV1

- `status`: `enum` - `"VALID" | "INVALID" | "PROOF_UNKNOWN"`
- `validationError`: `String|null` - a message providing additional details on the validation error if the status is `INVALID`.

`PROOF_UNKNOWN` means the execution layer does not hold the committed proof's bytes and was given no matching proof. It is distinct from `SYNCING`: the execution layer lacks one proof, not the chain.

## Routines

### Payload building

This routine follows the same specification as [Payload building](./paris.md#payload-building) with the following changes to the processing flow:

1. Client software **MUST** take `inclusionListTransactions` into account during the payload build process. The built `ExecutionPayload` **MUST** satisfy the inclusion list constraints with respect to `inclusionListTransactions` as defined in [EIP-7805](https://eips.ethereum.org/EIPS/eip-7805).

### Inclusion list proof verdicts and retention

An aggregate-proof verdict is the result, `VALID` or `INVALID`, of verifying an aggregate proof against a dependency digest under an `AGGREGATED_VK`, keyed `(keccak256(starkProof), depsHash, AGGREGATED_VK)`. The *target fork* of a call is the fork at its `timestamp`, the timestamp of the block the inclusion list constrains, and fixes the `AGGREGATED_VK` and static validity rules the call applies.

Client software *holds a verified proof* under `(starkProofHash, depsHash)` for a target fork if it has the proof bytes hashing to `starkProofHash` with a `VALID` verdict keyed with that fork's `AGGREGATED_VK`, together with a `proofDeps` such that `get_deps_hash(proofDeps) == depsHash`. It obtains them from a fully validated EIP-8288 mode-1 mempool wrapper, whose `deps` is `proofDeps`, or from an `InclusionListProofV1`. A verdict without the proof bytes is not a held proof: the verdict cache only avoids verifying the same proof again.

Client software **MUST** retain proofs as specified in the Proof Availability section of EIP-8288, until deadlines the consensus layer supplies as Unix time in milliseconds:

1. The proof bytes, `proofDeps` and verdict of a proof submitted through [`engine_submitInclusionListProofV1`](#engine_submitinclusionlistproofv1) until that call's `viewFreezeDeadlineMs`.

2. The proof bytes and `proofDeps` of a proof for which [`engine_verifyInclusionListProofV1`](#engine_verifyinclusionlistproofv1) returned `VALID` until that call's `attestationDeadlineMs`.

A retention requirement whose deadline has already passed imposes no retention. Client software does not derive these deadlines from slot lengths. It **MAY** keep a verdict after releasing the proof bytes, to avoid verifying the proof again if it is supplied again.

## Methods

### engine_newPayloadV6

Method parameter list is extended with `inclusionListTransactions`.

#### Request

* method: `engine_newPayloadV6`
* params:
  1. `executionPayload`: [`ExecutionPayloadV4`](./amsterdam.md#executionpayloadv4).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected blob versioned hashes to validate.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
  4. `executionRequests`: `Array of DATA` - List of execution layer triggered requests. Each list element is a `requests` byte array as defined by [EIP-7685](https://eips.ethereum.org/EIPS/eip-7685). The first byte of each element is the `request_type` and the remaining bytes are the `request_data`. Elements of the list **MUST** be ordered by `request_type` in ascending order. Elements with empty `request_data` **MUST** be excluded from the list.
  5. `inclusionListTransactions`: `Array of DATA` - Array of transaction objects, each object is a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718).
* timeout: 6s

#### Response

* result: [`PayloadStatusV2`](#payloadstatusv2)
* error: code and message set in case an exception happens while processing the payload.

#### Specification

This method follows the same specification as [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5) with the following changes:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of the payload does not fall within the time frame of the Bogota fork.

2. Client software **MUST** set `inclusionListSatisfied` in the following way:

    1. If the payload is deemed `VALID`, `inclusionListSatisfied` **MUST** be set to whether the payload satisfied the inclusion list constraints.

    2. Otherwise, `inclusionListSatisfied` **MUST** be `null`.

3. Client software **MUST** retain `inclusionListTransactions` for a payload with `ACCEPTED` status. Client software **MAY** discard them once the payload is no longer the tip of a branch.

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
  - `payloadStatus`: [`PayloadStatusV2`](#payloadstatusv2), with the same `status` value restrictions as [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4).
  - `payloadId`: `DATA|null`, 8 Bytes - identifier of the payload build process or `null`
* error: code and message set in case an exception happens while the validating payload, updating the forkchoice or initiating the payload build process.

#### Specification

This method follows the same specification as [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4) with the following changes to the processing flow:

1. Extend point (8) of the `engine_forkchoiceUpdatedV1` [specification](./paris.md#specification-1) by defining the following sequence of checks that **MUST** be run over `payloadAttributes`:

    1. `payloadAttributes` matches the [`PayloadAttributesV5`](#payloadattributesv5) structure, return `-38003: Invalid payload attributes` on failure.

    2. `payloadAttributes.timestamp` does not fall within the time frame of the Bogota fork, return `-38005: Unsupported fork` on failure.

2. Extend point (9) of the `engine_forkchoiceUpdatedV1` [specification](./paris.md#specification-1) by defining `payloadStatus.inclusionListSatisfied`:

    1. If the payload referenced by `forkchoiceState.headBlockHash` is deemed `VALID`, `payloadStatus.inclusionListSatisfied` **MUST** be set to whether the payload satisfied the inclusion list constraints.

    2. Client software **MUST** use the retained `inclusionListTransactions` if it validates the payload and checks whether the payload satisfies the inclusion list constraints while processing the call.

    3. Otherwise, `payloadStatus.inclusionListSatisfied` **MUST** be `null`.

### engine_verifyInclusionListProofV1

#### Request

* method: `engine_verifyInclusionListProofV1`
* params:
  1. `timestamp`: `QUANTITY`, 64 Bits - the `timestamp` of the block the inclusion list constrains.
  2. `transactions`: `Array of DATA` - the inclusion list's transactions, each a byte list (`DATA`) representing `TransactionType || TransactionPayload` or `LegacyTransaction` as defined in [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718).
  3. `proofCommitment`: [`InclusionListProofCommitmentV1`](#inclusionlistproofcommitmentv1) - the inclusion list's non-null proof commitment.
  4. `ilProof`: `Object|null` - an instance of [`InclusionListProofV1`](#inclusionlistproofv1), or `null`.
  5. `attestationDeadlineMs`: `QUANTITY`, 64 Bits - the EIP-7805 attestation deadline of the target block's slot, as Unix time in milliseconds.
* timeout: 1s

#### Response

* result: [`InclusionListProofStatusV1`](#inclusionlistproofstatusv1)
* error: code and message set in case an exception happens while verifying the inclusion list proof.

#### Specification

1. Client software **MUST** return `-38005: Unsupported fork` error if the target fork does not activate EIP-8288.

2. Client software **MUST** derive `deps`, the sorted, deduplicated union of the dependencies of the `transactions` that decode and pass static validity for the target fork, as specified by EIP-8288. Other transactions contribute no dependencies.

3. Client software **MUST** ignore an `ilProof` that does not match `proofCommitment`.

4. Client software **MUST** return `{status: INVALID, validationError: errorMessage | null}` if any of the following holds:

    1. `deps` is empty.

    2. A `proofDeps` known for `proofCommitment.depsHash` omits a dependency in `deps`. A `proofDeps` is known for a `depsHash` only if it is the `proofDeps` of a matching `ilProof`, of a held verified proof, or of a proof client software has verified as `VALID` against that `depsHash`.

    3. It holds an `INVALID` verdict keyed `(proofCommitment.starkProofHash, proofCommitment.depsHash, AGGREGATED_VK)` with the target fork's `AGGREGATED_VK`, whether or not `ilProof` is supplied.

    4. A matching `ilProof` fails verification under the target fork's `AGGREGATED_VK`.

    5. A matching `ilProof` fails, without verification, the checks of step 2 of [`engine_submitInclusionListProofV1`](#engine_submitinclusionlistproofv1).

5. Client software **MUST** return `{status: VALID, validationError: null}` if `deps` is nonempty, it holds a verified proof under `(proofCommitment.starkProofHash, proofCommitment.depsHash)` for the target fork or a matching `ilProof` verifies under the target fork's `AGGREGATED_VK`, and every dependency in `deps` is in that proof's `proofDeps`. On `VALID`, it **MUST** retain the proof bytes and `proofDeps` until `attestationDeadlineMs`, as specified in [Inclusion list proof verdicts and retention](#inclusion-list-proof-verdicts-and-retention). When it already holds the proof, this is a lookup with no proof verification. A supplied `ilProof` matching a cached `VALID` verdict is checked against `proofCommitment` and not verified again.

6. Client software **MUST** return `{status: PROOF_UNKNOWN, validationError: null}` if `deps` is nonempty, it does not hold the proof under the pair for the target fork, including when it holds a `VALID` verdict but no longer the proof bytes, and no matching `ilProof` was supplied.

7. Client software **MUST** record the verdict of a matching `ilProof` it verifies or rejects under step 4, keyed with the target fork's `AGGREGATED_VK`, and on `VALID` keep the proof bytes and `proofDeps` with it.

8. After successful parameter decoding, a supplied `ilProof` that matches `proofCommitment` but fails the proof-shape checks of step 4, item 5, returns `INVALID`; a nonmatching `ilProof` is ignored. Malformed parameters, such as missing fields, wrong types or invalid hex, still return `-32602: Invalid params`.

### engine_getInclusionListProofV1

#### Request

* method: `engine_getInclusionListProofV1`
* params:
  1. `timestamp`: `QUANTITY`, 64 Bits - the `timestamp` of the block the inclusion list constrains.
  2. `starkProofHash`: `DATA`, 32 Bytes - the `starkProofHash` of the inclusion list's commitment.
* timeout: 1s

#### Response

* result: `Object|null` - an instance of [`InclusionListProofV1`](#inclusionlistproofv1), or `null`.
* error: code and message set in case an exception happens while getting the inclusion list proof.

#### Specification

1. Client software **MUST** return `-38005: Unsupported fork` error if the target fork does not activate EIP-8288.

2. If client software holds the bytes of a proof hashing to `starkProofHash` with a `VALID` verdict under the target fork's `AGGREGATED_VK`, from a fully validated mode-1 wrapper or a held `InclusionListProofV1`, it **MUST** return that proof as `starkProof` with its `proofDeps`, taken from the wrapper's `deps` or the `InclusionListProofV1`.

3. Otherwise client software **MUST** return `null`, including when it holds the verdict but no longer the proof bytes.

### engine_submitInclusionListProofV1

#### Request

* method: `engine_submitInclusionListProofV1`
* params:
  1. `timestamp`: `QUANTITY`, 64 Bits - the `timestamp` of the block constrained by the inclusion lists of the proof's slot.
  2. `ilProof`: [`InclusionListProofV1`](#inclusionlistproofv1) - the `il_proof` of a `SignedInclusionListProof` received on the EIP-8288 proof topic.
  3. `viewFreezeDeadlineMs`: `QUANTITY`, 64 Bits - the EIP-7805 view freeze deadline of the proof's slot, as Unix time in milliseconds.
* timeout: 1s

#### Response

* result: [`InclusionListProofStatusV1`](#inclusionlistproofstatusv1), with `status` restricted to `VALID` or `INVALID`.
* error: code and message set in case an exception happens while verifying the inclusion list proof.

#### Specification

1. Client software **MUST** return `-38005: Unsupported fork` error if the target fork does not activate EIP-8288.

2. Client software **MUST** return `{status: INVALID, validationError: errorMessage | null}`, without verifying the proof, if `starkProof` is longer than `MAX_AGGREGATE_PROOF_BYTES`, or if `proofDeps` is not sorted without duplicates, has more than `MAX_DEPS_PER_AGGREGATE_WRAPPER` entries or has an entry whose scheme is outside the target fork's `ENABLED_SCHEMES`.

3. Otherwise client software **MUST** verify `starkProof` against `get_deps_hash(proofDeps)` under the target fork's `AGGREGATED_VK`, return `VALID` or `INVALID`, and record the verdict keyed `(keccak256(starkProof), get_deps_hash(proofDeps), AGGREGATED_VK)`. On `VALID` it **MUST** keep the proof bytes and `proofDeps` with the verdict until at least `viewFreezeDeadlineMs`, so that a later [`engine_verifyInclusionListProofV1`](#engine_verifyinclusionlistproofv1) call for a list committing to the pair under that fork is a lookup.

4. If client software already holds a verdict under the same key, it **MAY** return it without verifying the proof again.

### Update the methods of previous forks

#### Amsterdam API

For the following methods:

- [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5)
- [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4)

a validation **MUST** be added:

1. Client software **MUST** return `-38005: Unsupported fork` error if the `timestamp` of payload is greater than or equal to the Bogota activation timestamp.
