# Engine API -- Witness Retrieval (Proposal)

Engine API extensions that allow the consensus layer to opt in to receiving the
execution witness alongside the existing payload validation and payload
production flows.

This specification is based on and extends [Engine API - Amsterdam](./amsterdam.md).

> **Status:** Design proposal. This document is offered as an alternative to
> introducing dedicated `*WithWitness` methods (cf. issue
> [#741](https://github.com/ethereum/execution-apis/issues/741) and PR
> [#773](https://github.com/ethereum/execution-apis/pull/773)). It is
> intended to be evaluated by Engine API maintainers prior to merging.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Motivation](#motivation)
- [Structures](#structures)
  - [ExecutionWitnessV1](#executionwitnessv1)
  - [PayloadStatusV2](#payloadstatusv2)
  - [PayloadAttributesV5](#payloadattributesv5)
- [Methods](#methods)
  - [engine_newPayloadV6](#engine_newpayloadv6)
  - [engine_getPayloadV7](#engine_getpayloadv7)
  - [engine_forkchoiceUpdatedV5](#engine_forkchoiceupdatedv5)
- [Rationale](#rationale)
- [Capabilities and feature gating](#capabilities-and-feature-gating)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Motivation

Stateless validation flows require the consensus layer (CL) — or a prover
acting on its behalf — to obtain the **execution witness** for a block:
the set of state and bytecode accesses produced during execution. Two
paths exist for the witness to leave the execution layer (EL):

1. **The CL receives an inbound payload over gossip.** The CL forwards it to the
   EL via `engine_newPayload*`, and wants the EL to attach the witness to the
   validation response.
2. **The CL drives block production locally.** The CL initiates payload building
   via `engine_forkchoiceUpdated*`, retrieves the built payload via
   `engine_getPayload*`, and wants the EL to attach the witness to the produced
   payload.

The currently-proposed designs (issue [#741][issue-741], PR [#773][pr-773])
introduce parallel `engine_newPayloadV{N}WithWitness` methods that mirror the
existing methods and additionally return the witness. This document proposes an
alternative: **extend the existing methods with an opt-in flag**, and route the
build-side opt-in through `PayloadAttributes`, where similar fork-scoped
extensions (`withdrawals`, `parentBeaconBlockRoot`) have historically been
added.

[issue-741]: https://github.com/ethereum/execution-apis/issues/741
[pr-773]: https://github.com/ethereum/execution-apis/pull/773

## Structures

### ExecutionWitnessV1

`DATA` - Variable-length bytes representing an SSZ-encoded
`ExecutionWitness` container. The container schema is defined by an
accompanying consensus-specs document and is out of scope for this
specification.

This follows the same on-wire convention used elsewhere in the Engine API
for SSZ payloads (cf. the `blob` field in [`BlobV1`](./cancun.md)): a
JSON-RPC `DATA` value whose underlying bytes are SSZ-encoded according to
a consensus-specs definition. The encoding is fixed by this specification,
so no format tag is required; on SSZ-REST transports (cf.
[PR #764](https://github.com/ethereum/execution-apis/pull/764)) the same
bytes are carried directly without hex wrapping.

### PayloadStatusV2

This structure has the syntax of [`PayloadStatusV1`](./paris.md#payloadstatusv1)
and appends a single optional field: `executionWitness`.

- `status`: `String` - One of: `VALID`, `INVALID`, `SYNCING`, `ACCEPTED`,
  `INVALID_BLOCK_HASH`.
- `latestValidHash`: `DATA|null`, 32 Bytes - The hash of the most recent valid
  block in the branch defined by payload and its ancestors.
- `validationError`: `String|null` - A message providing additional details on
  the validation error if the payload is classified as `INVALID`.
- `executionWitness`: [`ExecutionWitnessV1`](#executionwitnessv1)`|null` -
  SSZ-encoded execution witness. Present **only if** `requestWitness` was
  set to `true` in the associated request **and** `status` is `VALID`.
  `null` in all other cases.

### PayloadAttributesV5

This structure has the syntax of [`PayloadAttributesV4`](./amsterdam.md#payloadattributesv4)
and appends a single field: `requestWitness`.

- `timestamp`: `QUANTITY`, 64 Bits - value for the `timestamp` field of the new payload
- `prevRandao`: `DATA`, 32 Bytes - value for the `prevRandao` field of the new payload
- `suggestedFeeRecipient`: `DATA`, 20 Bytes - suggested value for the `feeRecipient` field of the new payload
- `withdrawals`: `Array of WithdrawalV1` - Array of withdrawals, each object is an `OBJECT` containing the fields of a `WithdrawalV1` structure.
- `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
- `slotNumber`: `QUANTITY`, 64 Bits - value for the `slotNumber` field of the new payload
- `requestWitness`: `BOOLEAN` - if `true`, the EL **MUST** collect and persist
  the execution witness alongside the produced payload so that it can be
  returned by a subsequent call to [`engine_getPayloadV7`](#engine_getpayloadv7).
  Defaults to `false` when omitted.

## Methods

### engine_newPayloadV6

This method extends [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5)
with an optional witness-request flag and an optional witness in the response.

#### Request

* method: `engine_newPayloadV6`
* params:
  1. `executionPayload`: [`ExecutionPayloadV4`](./amsterdam.md#executionpayloadv4).
  2. `expectedBlobVersionedHashes`: `Array of DATA`, 32 Bytes - Array of expected blob versioned hashes to validate.
  3. `parentBeaconBlockRoot`: `DATA`, 32 Bytes - Root of the parent beacon block.
  4. `executionRequests`: `Array of DATA` - List of execution layer triggered requests.
  5. `requestWitness`: `BOOLEAN` - if `true`, the EL **MUST** include the
     execution witness in the response when the payload is `VALID`.
     Defaults to `false` when omitted.
* timeout: 8s

#### Response

* result: [`PayloadStatusV2`](#payloadstatusv2)
* error: code and message set in case an exception happens while processing the
  payload.

#### Specification

This method follows the same specification as
[`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5) with the following
additions:

1. If `requestWitness` is `true` **and** the payload validates to `VALID`, the
   EL **MUST** populate `result.executionWitness` with an
   [`ExecutionWitnessV1`](#executionwitnessv1) object reflecting the state and
   bytecode accesses produced during execution.

2. If `requestWitness` is `true` **and** validation does not result in `VALID`
   (i.e. status is `INVALID`, `SYNCING`, `ACCEPTED`, or `INVALID_BLOCK_HASH`),
   the EL **MUST** set `result.executionWitness` to `null`.

3. If `requestWitness` is `false` or omitted, the EL **MUST** set
   `result.executionWitness` to `null` and **MUST NOT** perform additional work
   to collect the witness on the critical path.

4. Witness collection **MUST NOT** affect the validation outcome reported
   in `result.status`.

5. Clients that have not advertised the `engine_newPayloadV6` capability via
   capabilities exchange **MUST** continue to be addressed via
   [`engine_newPayloadV5`](./amsterdam.md#engine_newpayloadv5).

### engine_getPayloadV7

This method extends [`engine_getPayloadV6`](./amsterdam.md#engine_getpayloadv6)
with an optional witness in the response. The witness is returned only when
the corresponding payload-build call (`engine_forkchoiceUpdatedV5`) was issued
with `payloadAttributes.requestWitness = true`.

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
  - `executionWitness`: [`ExecutionWitnessV1`](#executionwitnessv1)`|null` -
    SSZ-encoded execution witness. Present **only if** the originating
    `engine_forkchoiceUpdatedV5` call set
    `payloadAttributes.requestWitness = true`. `null` otherwise.
* error: code and message set in case an exception happens while getting the payload.

#### Specification

This method follows the same specification as
[`engine_getPayloadV6`](./amsterdam.md#engine_getpayloadv6) with the following
additions:

1. If the build was initiated with `payloadAttributes.requestWitness = true`,
   the EL **MUST** populate `result.executionWitness` with an
   [`ExecutionWitnessV1`](#executionwitnessv1) reflecting the witness collected
   during payload building.

2. If the build was initiated with `payloadAttributes.requestWitness = false`
   (or omitted), the EL **MUST** set `result.executionWitness` to `null` and
   **MUST NOT** perform additional work to collect the witness retroactively.

### engine_forkchoiceUpdatedV5

This method extends [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4)
to accept [`PayloadAttributesV5`](#payloadattributesv5).

#### Request

* method: `engine_forkchoiceUpdatedV5`
* params:
  1. `forkchoiceState`: [`ForkchoiceStateV1`](./paris.md#ForkchoiceStateV1).
  2. `payloadAttributes`: `Object|null` - Instance of [`PayloadAttributesV5`](#payloadattributesv5) or `null`.
  3. `custodyColumns`: `DATA|null`, 16 Bytes - Interpreted as a bitarray of length `CELLS_PER_EXT_BLOB` indicating which column indices form the CL's custody set, or `null` if the CL does not provide custody services.
* timeout: 8s

#### Response

Refer to the response for [`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4).

#### Specification

This method follows the same specification as
[`engine_forkchoiceUpdatedV4`](./amsterdam.md#engine_forkchoiceupdatedv4) with
the following changes:

1. Validate `payloadAttributes` against [`PayloadAttributesV5`](#payloadattributesv5)
   in place of `PayloadAttributesV4`. If `payloadAttributes` is provided and
   does not match this structure, return `-38003: Invalid payload attributes`.

2. If `payloadAttributes.requestWitness` is `true`, the EL **MUST** configure
   the payload-build job initiated by this call to collect and persist the
   execution witness, such that a subsequent
   [`engine_getPayloadV7`](#engine_getpayloadv7) call addressing the resulting
   `payloadId` can return it.

3. If `payloadAttributes.requestWitness` is `false` or omitted, the EL
   **MUST NOT** perform additional work to collect the witness for this build
   job.

## Rationale

### Why a flag, not a parallel method

Engine API methods version once per fork (`engine_newPayloadV1` … `V5`). The
`*WithWitness` design proposed in PR [#773][pr-773] doubles this surface,
producing a `V{N}` and a `V{N}WithWitness` per fork going forward. Each
variant must be specified, openrpc-modeled, capabilities-negotiated, and
client-implemented in lockstep with the standard methods.

Routing the opt-in through a single boolean flag on the existing methods:

- Keeps the method count constant across forks.
- Avoids the response-shape drift between `V{N}` and `V{N}WithWitness` that
  recurs every fork.
- Composes naturally with the SSZ-REST transport proposed in
  PR [#764](https://github.com/ethereum/execution-apis/pull/764): the same
  flag governs whether the SSZ response includes a witness field, with no
  additional REST endpoint needed.

### Why route the build-side opt-in through `PayloadAttributes`

Witness collection on the build path **must be decided before execution
begins**, because the EL needs to instrument trie traversal during execution
to record accesses. `engine_getPayloadV{N}` is called *after* execution has
finished and the block is built; adding a flag there would either force the
EL to always collect witnesses (defeating the opt-in), or trigger re-execution
(unacceptable on the critical path).

`PayloadAttributes` is the canonical extension point for build-side knobs and
has accumulated similar fork-scoped fields previously: `withdrawals` (Shapella),
`parentBeaconBlockRoot` (Cancun), `slotNumber` (Amsterdam). Adding
`requestWitness` here is consistent with that precedent and gives the EL the
information it needs at build-job initiation time.

### Why bump `PayloadStatus` to V2

The Engine API convention bumps a structure's `V` whenever its shape changes
(`ExecutionPayloadV1`…`V4`, `PayloadAttributesV1`…`V5`, `BlobsBundleV1`…`V2`).
Adding `executionWitness` to the response shape is a structural change, so
`PayloadStatusV2` follows that pattern. `engine_newPayloadV1`…`V5` continue to
return `PayloadStatusV1` unchanged; only `engine_newPayloadV6` returns the new
type. This localises the change to a single method version and leaves all
existing openrpc schemas untouched.

### Considered alternatives

1. **`engine_newPayloadV{N}WithWitness` (PR #773).** Proven design and already
   has a working prototype, but doubles method surface and creates a
   redundant verb once SSZ-REST lands (acknowledged by the PR author in
   prior discussion).

2. **Always return the witness.** Simplest schema, but forces witness
   collection on every block — a meaningful CPU and I/O cost (~574 ms
   witness-generation step measured by PR #773's benchmarks) that most CLs
   do not need.

3. **Capabilities-only gating with no flag.** Use the engine_exchangeCapabilities
   handshake alone to decide whether to attach the witness. Too coarse: a CL
   may want the witness for some blocks (e.g. when acting as a stateless
   prover) and not others.

4. **Trailing optional param on `engine_newPayloadV5` without a version
   bump.** Backward-compatible at the JSON-RPC level, but breaks the
   convention that every shape change bumps `V`. Rejected for consistency
   with existing fork-aligned versioning.

## Capabilities and feature gating

The `engine_newPayloadV6`, `engine_forkchoiceUpdatedV5`, and
`engine_getPayloadV7` methods **MUST** be advertised through
`engine_exchangeCapabilities`. CLs **MUST NOT** invoke these methods without
prior capability confirmation from the EL. An EL that does not yet support
witness retrieval continues to be addressed via the V5 / V4 / V6 predecessors
unmodified.

Witness-retrieval support is independent of any specific fork activation:
once advertised, an EL may serve witnesses for any post-Amsterdam block
the CL requests them for. Because the flag is opt-in per call, the EL
incurs witness-collection cost only when a CL asks for it.
