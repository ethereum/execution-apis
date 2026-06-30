# Engine API -- Refactor Proposal (REST + SSZ)

> **Status:** Draft / discussion document. This file proposes a REST
> refactor of the Engine API that moves from JSON-RPC over a single
> endpoint to a resource-oriented HTTP/REST API where hot-path
> request and response bodies are SSZ-encoded. The new API ships at
> `/engine/v1/...` alongside the legacy `engine_*` JSON-RPC endpoint;
> the legacy endpoint is retired at a future fork, at which point
> the REST surface becomes the only way to drive an EL.
>
> **Target fork:** Amsterdam. The new API ships *as* the Amsterdam Engine
> API; clients implement it instead of `engine_*` JSON-RPC at the
> Amsterdam activation timestamp.

---

## Table of contents

- [Mapping from old → new](#mapping-from-old--new)
- [Resource model (overview)](#resource-model-overview)
- [Endpoints](#endpoints)
  - [Payload submission](#payload-submission)
  - [Forkchoice update](#forkchoice-update)
  - [Payload retrieval](#payload-retrieval)
  - [Historical bodies](#historical-bodies)
  - [Blob pool](#blob-pool)
  - [Capabilities & identification](#capabilities--identification)
  - [Examples: every fork](#examples-every-fork)
- [Error model](#error-model)
- [Versioning model](#versioning-model)
- [Authentication](#authentication)
- [Transport & framing](#transport--framing)
- [SSZ encoding conventions](#ssz-encoding-conventions)
- [Message ordering & idempotency](#message-ordering--idempotency)
- [Security considerations](#security-considerations)
- [Motivation](#motivation)
  - [Goals & non-goals](#goals--non-goals)
  - [Why move away from JSON-RPC?](#why-move-away-from-json-rpc)
  - [Why SSZ?](#why-ssz)
  - [Simplifications & removed concepts](#simplifications--removed-concepts)
  - [Summary of design decisions](#summary-of-design-decisions)
- [Future evolution](#future-evolution)
  - [Progressive merkleization](#progressive-merkleization)

> **Reading order note.** The endpoint sketches reference SSZ types
> like `Optional[T]`, `BodyEntry`, and `BlobEntry`. If a definition
> isn't immediately clear, jump to
> [SSZ encoding conventions](#ssz-encoding-conventions) and
> [Message ordering & idempotency](#message-ordering--idempotency)
> further down — they fully define the wire-level details.

---

## Mapping from old → new

If you're migrating from the JSON-RPC engine API, this is the lookup
table. Detail on each new endpoint follows in the sections below.

All hot-path endpoints select the fork via the
`Eth-Execution-Version: <fork>` request header.

| Old method | New endpoint | Notes |
| - | - | - |
| `engine_newPayloadV{1..5}` | `POST /payloads` | `parentBeaconBlockRoot` and `executionRequests` folded into the SSZ envelope; `expectedBlobVersionedHashes` removed; `INVALID_BLOCK_HASH` removed from the status enum |
| `engine_forkchoiceUpdatedV{1..4}` | `POST /forkchoice` | one atomic call; carries forkchoice state, optional `payload_attributes`, and (Amsterdam+) optional `custody_columns` |
| `engine_getPayloadV{1..6}` | `GET /payloads/{id}` | poll-style, same semantics as today |
| `engine_getPayloadBodiesByHashV{1,2}` | `POST /bodies/hash` | header selects both the response schema and the era of returned blocks; `POST` because hash lists are too large for URLs |
| `engine_getPayloadBodiesByRangeV{1,2}` | `GET /bodies?from=...&count=...` | header selects both the response schema and the era of returned blocks |
| `engine_getBlobsV1` | `POST /blobs/v1` | independently versioned; legacy version numbers carry forward |
| `engine_getBlobsV2` | `POST /blobs/v2` | all-or-nothing cell proofs |
| `engine_getBlobsV3` | `POST /blobs/v3` | partial-response cell proofs |
| `engine_getBlobsV4` | `POST /blobs/v4` | cell-range selection |
| `engine_getClientVersionV1` | `GET /identity` + `X-Engine-Client-Version` request header | unscoped |
| `engine_exchangeCapabilities` | `GET /capabilities` | unscoped |
| `engine_exchangeTransitionConfigurationV1` | *removed* | already deprecated since Cancun |

---

## Resource model (overview)

All endpoints live under `/engine/v1/...`. Hot-path endpoints
require the `Eth-Execution-Version: <fork>` request header; diagnostic
endpoints ignore it.

| Resource | Endpoint | Purpose |
| - | - | - |
| Payload | `POST /engine/v1/payloads` | Submit a payload received from the CL gossip network for the EL to validate / import. Replaces `engine_newPayload`. Fork-scoped via `Eth-Execution-Version`. |
| Payload | `GET /engine/v1/payloads/{payloadId}` | Retrieve a built payload by id. Replaces `engine_getPayload`. Fork-scoped via `Eth-Execution-Version`. CL polls when it wants a fresher snapshot. |
| Forkchoice | `POST /engine/v1/forkchoice` | Atomic forkchoice update: update head/safe/finalized, optionally start a payload build, optionally update custody set. Replaces `engine_forkchoiceUpdated`. Fork-scoped via `Eth-Execution-Version`. |
| Bodies | `POST /engine/v1/bodies/hash` | Replaces `engine_getPayloadBodiesByHash`. `Eth-Execution-Version` selects both the response schema *and* the era of returned blocks; out-of-era blocks come back as `available=false`. |
| Bodies | `GET /engine/v1/bodies?from=N&count=M` | Replaces `engine_getPayloadBodiesByRange`. Same fork scoping as `/bodies/hash`. |
| Blob pool | `POST /engine/v1/blobs/v{1..4}` | Replaces `engine_getBlobsV{1..4}`. The `vN` segment carries forward the legacy version numbers; `/v4` is the Amsterdam cell-range variant. Independently versioned (not fork-scoped). |
| Capabilities | `GET /engine/v1/capabilities` | Replaces `engine_exchangeCapabilities`. Unscoped; advertises supported forks, `/blobs/vN` revisions, and per-endpoint request-size limits. |
| Identity | `GET /engine/v1/identity` | Replaces `engine_getClientVersion`. Unscoped. |

Every hot-path body uses SSZ; every metadata endpoint uses JSON.

---

## Endpoints

### Payload submission

#### `POST /engine/v1/payloads`

Replaces `engine_newPayloadV{1..5}`.

- **Request body:** SSZ-encoded `ExecutionPayloadEnvelope`

  ```
  ExecutionPayloadEnvelope {
      payload:                  ExecutionPayload          # the fork's payload SSZ container
      parent_beacon_block_root: Root                      # was a separate param since Cancun
      execution_requests:       List[Bytes, MAX_REQUESTS] # was a separate param since Prague
  }
  ```

  `expected_blob_versioned_hashes` is **removed**: it was a
  defense-in-depth cross-check, but the block-hash check already covers
  the transactions, so the EL recomputes the array from
  `payload.transactions` during validation and a mismatch between CL
  and EL views surfaces as `INVALID` exactly as before.

- **Response body:** SSZ-encoded `PayloadStatus`:

  ```
  PayloadStatus {
      status:           uint8        # VALID=0, INVALID=1, SYNCING=2, ACCEPTED=3
      latest_valid_hash: Optional[Hash32]
      validation_error: Optional[String]
  }
  ```

  `INVALID_BLOCK_HASH` is dropped (already supplanted by `INVALID`).

- **HTTP status:** `200 OK` for any of the four validation outcomes.
  Validation results are not transport errors.

### Forkchoice update

#### `POST /engine/v1/forkchoice`

Replaces `engine_forkchoiceUpdatedV{1..4}`.

- **Request body:** SSZ-encoded `ForkchoiceUpdate`:

  ```
  ForkchoiceUpdate {
      forkchoice_state:    ForkchoiceState              # head / safe / finalized
      payload_attributes:  Optional[PayloadAttributes]  # if present, start a build
      custody_columns:     Optional[Bitvector[CELLS_PER_EXT_BLOB]]  # Amsterdam+, optional
  }
  ```

  All three fields are processed in one transaction: the EL MUST apply
  the forkchoice state, then (if `payload_attributes` is present and
  the new head is `VALID`) start the build, then (if `custody_columns`
  is present) update the custody set, all before returning. If the
  forkchoice update fails, no build is started and no custody change
  is applied.

  When building a payload (Amsterdam+), the EL **MUST** use
  `payload_attributes.target_gas_limit` as the target value for the
  built block's `gas_limit`.

- **Response body:** SSZ-encoded `ForkchoiceUpdateResponse`:

  ```
  ForkchoiceUpdateResponse {
      payload_status: PayloadStatus              # VALID | INVALID | SYNCING
      payload_id:     Optional[Bytes8]           # server-assigned opaque token; set iff a build was started
  }
  ```

  The `payload_id` is an **opaque server-assigned token**. The EL
  chooses how to mint it (counter, random, hash-tree-root over the
  attributes — anything). CLs MUST treat it as opaque bytes and MUST
  NOT recompute or validate its contents.

- **HTTP status:** `200 OK` for all three payload-status outcomes.
  `409 Conflict` is returned for an inconsistent forkchoice state
  (today's `-38002`); `422 Unprocessable Entity` for invalid
  `payload_attributes` (today's `-38003`); `409 Conflict` for a too-deep
  reorg (today's `-38006`).

- **Skip-allowed semantics:** the EL MAY skip applying the forkchoice
  state and instead return `{VALID, latest_valid_hash: head}` if the
  new `head` is a `VALID` ancestor of the latest known finalized block.
  This preserves the existing Paris-spec rule (point 2 of the
  `engine_forkchoiceUpdated` specification) and is deliberate: a CL
  that emits a malformed or stale FCU referencing a head behind
  finalization should not be able to roll the EL back. We keep the
  behaviour that has caught buggy CLs in the past.

- **Stale-fork header:** an FCU with `Eth-Execution-Version: <fork>`
  referencing a `head` from an earlier fork is **allowed**, *as long
  as `payload_attributes` is absent*. The CL needs to update head /
  safe / finalized across fork boundaries during sync and reorg
  recovery, and the header fork has no bearing on which historical
  block can be referenced.
  TODO(MariusVanDerWijden) Is that really the case?

  If `payload_attributes` is present, the `Eth-Execution-Version`
  header MUST match the fork that the new payload would belong to
  (i.e. the fork determined by `payload_attributes.timestamp`).
  Mismatch returns `400 unsupported-fork`. Building a payload is
  the only operation where the header fork is load-bearing on shape,
  so it's the only one we strictly police.

- **Custody-set semantics** (Amsterdam+): the custody update runs
  independently of the forkchoice processing flow. An execution-time 
  custody-set error MUST NOT affect the `payload_status` returned for 
  the forkchoice update.
  A `custody_columns` value, once accepted, remains in effect until
  the next `POST /forkchoice` whose body *also* contains a
  `custody_columns` field. FCUs that omit the field leave the
  custody set unchanged.

### Payload retrieval

#### `GET /engine/v1/payloads/{payloadId}`

Replaces `engine_getPayloadV{1..6}`.

- **Response body:** SSZ-encoded `BuiltPayload`:

  ```
  BuiltPayload {
      payload:                 ExecutionPayload
      block_value:             Uint256
      blobs_bundle:            BlobsBundle
      execution_requests:      List[Bytes, MAX_REQUESTS]
      should_override_builder: bool
  }
  ```

  The shape above is the Amsterdam variant. **Field order is
  normative**: `execution_requests` precedes `should_override_builder`.
  This deliberately diverges from the legacy JSON-RPC envelope, which
  appended `executionRequests` last; SSZ fields are positional, so the
  order is part of the wire format and **MUST** be followed exactly.
  Pre-Amsterdam forks have their own `BuiltPayload` shapes (the fields
  a fork doesn't have are absent, and the `blobs_bundle` revision
  tracks the fork) — see the per-fork `BuiltPayload` catalogue in
  [refactor-ssz.md](./refactor-ssz.md) for every variant from Paris up.
- **404** if `payloadId` is unknown or expired.

Polling semantics are unchanged from `engine_getPayload`: the CL calls
`GET /payloads/{payloadId}` whenever it wants the latest
snapshot of the build. Each call returns the most recent version
available at the time of receipt; the EL MAY stop the build process
after serving a call. `payloadId` values are opaque server-assigned
tokens issued by `POST /forkchoice`.

The EL keeps optimising the payload until the slot deadline, so
successive `GET`s against the same `{payloadId}` may return different
bytes. The EL **MUST** include `Cache-Control: no-store` on the
response, and intermediaries **MUST NOT** cache or revalidate this
resource. CLs **MUST NOT** treat the response as cacheable.

**Path validation.** `{payloadId}` is a path segment carrying a hex-
encoded `Bytes8`. The EL **MUST** validate that the path segment is
well-formed (8 bytes, hex) before dispatching to lookup logic; a
malformed segment returns `400 invalid-request`.

**Token TTL.** A `payloadId` is valid until either the payload was
retrieved by `GET /payloads/{payloadId}` or another payload
was built via a forkchoice with payload attributes.

### Historical bodies

These endpoints are **fork-scoped on both the response schema and the
era of the returned blocks.** The `Eth-Execution-Version` header
tells the EL which `ExecutionPayloadBody` schema to use, *and* limits
the response to blocks whose timestamp falls in that fork's active
time range. A CL fetching bodies that span a fork boundary issues
separate requests, one per header value.

The EL **MUST** apply both meanings of the header value: it **MUST**
serialise each entry against the named schema **and MUST** filter the
response to blocks whose timestamp falls in that fork's active time
range. Using the header only for schema selection — and returning
blocks from outside the fork's era — is **non-conformant**. A
requested block that exists but whose timestamp lies outside the
header fork's range **MUST** come back as `available=false` (it is
not omitted, unlike a past-head block in a range query; see below).

Concretely:

- `POST /bodies/hash` with `Eth-Execution-Version: cancun` returns
  bodies *only* for blocks in the Cancun time range. Requesting a
  Shanghai or Amsterdam hash yields `available=false` for that entry.
- `POST /bodies/hash` with `Eth-Execution-Version: amsterdam` returns
  bodies *only* for Amsterdam blocks. All fields (including
  `block_access_list`) are unconditionally present; older blocks the
  CL accidentally requested come back as `available=false`.

#### `POST /engine/v1/bodies/hash`

Replaces `engine_getPayloadBodiesByHashV{1,2}`. Uses `POST` so that
large hash lists travel in the request body rather than the URL.

- **Request body:** SSZ-encoded `BodiesByHashRequest` (a single-field
  container wrapping `block_hashes: List[Hash32, MAX_BODIES_REQUEST]`;
  see [refactor-ssz.md](./refactor-ssz.md)). Top-level bodies are
  wrapped in a container rather than sent as a bare SSZ list, matching
  the beacon-API convention; this costs a 4-byte offset but lets future
  revisions add fields without a breaking wire change.

#### `GET /engine/v1/bodies?from=N&count=M`

Replaces `engine_getPayloadBodiesByRangeV{1,2}`. Range fits comfortably
in the URL. Block numbers whose timestamp falls outside the URL fork's
active range come back as `available=false`; block numbers past the
latest known block are **omitted entirely** (the response is truncated
at head, not padded — see the response-length note below). If the
requested range straddles a fork boundary the CL re-issues with a
different `Eth-Execution-Version` for the unfilled suffix.

- **Response body** (both endpoints): SSZ-encoded `BodiesResponse`
  (a single-field container wrapping
  `entries: List[BodyEntry, MAX_BODIES_REQUEST]`). Each `BodyEntry`
  carries an `available: boolean` flag and an `ExecutionPayloadBody`
  serialised against the **schema named by `Eth-Execution-Version`**.
  `available` is false in either of the following cases:
  - the block is unavailable / pruned, or
  - the block's timestamp falls outside the header fork's active
    range.

  When `available=false`, the `body` field is zero-valued and CLs
  MUST ignore its contents. See
  [SSZ encoding conventions](#ssz-encoding-conventions) for the
  `BodyEntry` wrapper definition.

- **Response length (range queries).** The response carries one
  `BodyEntry` per known block in the requested range; it is
  **truncated at the latest known block** and is **not** padded out to
  `count` entries with `available=false` placeholders. A request whose
  range extends past head therefore returns fewer than `count`
  entries, and a request entirely past head returns an empty
  `entries` list. This carries forward the legacy
  `engine_getPayloadBodiesByRange` "no trailing nulls" rule. The CL
  detects the unfilled suffix from the shortfall and re-issues against
  the next fork URL if the range straddled a fork boundary.

### Blob pool

The blob endpoint is **independently versioned**: legacy
`engine_getBlobsVN` numbers carry forward onto the URL, so
`engine_getBlobsVN` becomes `POST /blobs/vN`. ELs **MUST** serve at
least the revision matching their current fork (`/blobs/v4` for
Amsterdam) and **MAY** serve any subset of older revisions alongside;
`GET /capabilities` advertises the actual list.

All revisions use `POST` so that 128 versioned hashes (8 KiB hex)
don't have to fit in the URL. All revisions take a single-field
request container (`BlobsVNRequest`) and return a single-field
response container (`BlobsVNResponse`) wrapping
`entries: List[BlobVNEntry, MAX_BLOBS_REQUEST]` on `200 OK`, and use
HTTP **`204 No Content`** to signal that the EL cannot serve the
request at all (syncing, blob pool unavailable, V2 all-or-nothing
miss). Wrapping top-level bodies in a container (rather than a bare
SSZ list) costs a 4-byte offset but matches the beacon-API convention
and keeps the revisions extensible. Within a `200` response, per-blob
misses are reported via `BlobEntry.available = false` on revisions
that support partial responses. Revision-specific contents live inside
`BlobEntry.contents`.

#### `POST /engine/v1/blobs/v1`

Replaces `engine_getBlobsV1` (Cancun, single-proof whole-blob).

- **Request body:** SSZ `BlobsV1Request` (wraps
  `versioned_hashes: List[VersionedHash, MAX_BLOBS_REQUEST]`).
- **Response `BlobEntry.contents`:** `BlobAndProofV1 { blob, proof }`
  (one blob, one 48-byte KZG proof).
- Partial responses supported: missing blobs surface as
  `available=false` per entry. `204 No Content` only when the EL
  cannot serve the request at all (e.g. syncing).

#### `POST /engine/v1/blobs/v2`

Replaces `engine_getBlobsV2` (Osaka, all-or-nothing cell proofs).

- **Request body:** SSZ `BlobsV2Request` (wraps
  `versioned_hashes: List[VersionedHash, MAX_BLOBS_REQUEST]`).
- **Response `BlobEntry.contents`:** `BlobAndProofV2 { blob, proofs }`
  (one blob plus `CELLS_PER_EXT_BLOB` cell proofs).
- **All-or-nothing:** if any requested blob is missing, the EL
  returns `204 No Content`. Otherwise `200 OK` and all entries have
  `available=true`. This matches today's V2 semantics.

#### `POST /engine/v1/blobs/v3`

Replaces `engine_getBlobsV3` (Osaka, partial responses with cell
proofs).

- **Request body:** same as `/v2`.
- **Response:** same `BlobEntry.contents` shape as `/v2`, but missing
  blobs surface as `available=false` per entry rather than collapsing
  the whole response. `204 No Content` only when the EL cannot serve
  the request at all.

#### `POST /engine/v1/blobs/v4`

Replaces `engine_getBlobsV4` (Amsterdam, cell-range selection).

- **Request body:** SSZ `BlobsV4Request`:
  ```
  BlobsV4Request {
      versioned_hashes: List[VersionedHash, MAX_BLOBS_REQUEST]
      indices_bitarray: Bitvector[CELLS_PER_EXT_BLOB]   # which cells to return
  }
  ```
- **Response `BlobEntry.contents`:** `BlobCellsAndProofsV1`
  (per-cell `blob_cells` and `proofs` arrays, with `Optional[T]` =
  `[]` at indices where individual cells are unavailable).

### Capabilities & identification

#### `GET /engine/v1/capabilities`

Returns JSON. The advertisement includes per-endpoint maximum request
sizes so the CL knows how many block-bodies / blob-cells / payloads
the server is willing to serve in one request:

```json
{
  "supported_forks":          ["paris", "shanghai", "cancun", "prague", "osaka", "amsterdam"],
  "fork_scoped_endpoints":    ["payloads", "forkchoice", "bodies"],
  "independently_versioned":  { "blobs": ["v1", "v2", "v3", "v4"] },
  "unscoped_endpoints":       ["capabilities", "identity"],
  "limits": {
    "bodies.max_count":           32,
    "blobs.max_versioned_hashes": 128,
    "payload.max_bytes":          67108864
  }
}
```

The `limits.*` values map onto the SSZ `MAX_*` constants where one
exists: `bodies.max_count` is bounded by `MAX_BODIES_REQUEST` (`32`,
inherited from Shanghai's `engine_getPayloadBodiesByHashV1`) and
`blobs.max_versioned_hashes` by `MAX_VERSIONED_HASHES_PER_REQUEST`
(`128`). `payload.max_bytes` is bounded by `MAX_REQUEST_BODY_SIZE`
(`2**26` = `67108864`, 64 MiB); see
[refactor-ssz.md § `MAX_*` constants](./refactor-ssz.md#max-constants).
The advertised numbers are an upper bound the server is willing to
serve; operators MAY advertise lower values, but MUST NOT advertise
higher than the corresponding `MAX_*` constant.

The `independently_versioned` map advertises endpoints whose URL
carries an explicit `/vN` revision. ELs MAY support multiple
revisions concurrently (e.g. `["v1", "v2"]`); CLs pick whichever they
implement.

#### `GET /engine/v1/identity`

Returns JSON `ClientVersion[]` (same shape as today's
`engine_getClientVersionV1`). The CL identifies itself with a
`X-Engine-Client-Version` header on every request, removing the
mutual-exchange handshake.

### Example: submit a payload

```bash
curl -X POST http://localhost:8551/engine/v1/payloads \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Eth-Execution-Version: amsterdam" \
  -H "Content-Type: application/octet-stream" \
  -H "Accept: application/octet-stream" \
  -H "X-Engine-Client-Version: LH/v6.2.1" \
  --data-binary @new_payload.ssz \
  -o payload_status.ssz
```

Request:

```
POST /engine/v1/payloads HTTP/2
Host: localhost:8551
Authorization: Bearer <JWT>
Eth-Execution-Version: amsterdam
Content-Type: application/octet-stream
Content-Length: 584

<584 bytes: SSZ(ExecutionPayloadEnvelope)>
```

Successful response (`status = VALID`):

```
HTTP/2 200
Content-Type: application/octet-stream
Content-Length: 41

<41 bytes: SSZ(PayloadStatus)>
```

The 41 bytes break down as: `status` (1 byte = `0x01`, `VALID`) +
`latest_valid_hash` (4-byte offset + 32-byte hash = 36 bytes)
+ `validation_error` (4-byte offset + 0 bytes empty list).

Error response (malformed body):

```
HTTP/2 400
Content-Type: application/problem+json
Content-Length: 49

{ "type": "/engine-api/errors/ssz-decode-error" }
```

### Example: poll a built payload

```bash
curl http://localhost:8551/engine/v1/payloads/0x1234567890abcdef \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Eth-Execution-Version: amsterdam" \
  -H "Accept: application/octet-stream" \
  -o built_payload.ssz
```

Response carries `Cache-Control: no-store`; intermediaries MUST NOT
cache. See [Payload retrieval](#payload-retrieval).

### Examples: every fork

The examples above use Amsterdam. The shapes below show how the two
main hot-path bodies change fork-to-fork. The URL is the same for
every fork — only the `Eth-Execution-Version` header value changes
along with the body schema. Field names follow the
[per-fork container catalogue](./refactor-ssz.md#per-fork-container-catalogue);
`<…>` denotes nested SSZ. Only the fields that change per fork are
called out; every fork from Paris on is a valid header value.

#### `POST /payloads` request body (`ExecutionPayloadEnvelope{Fork}`)

```
# Eth-Execution-Version: paris
ExecutionPayloadEnvelopeParis {
    payload: <ExecutionPayloadParis>            # base fields only (no withdrawals/blob-gas/BAL)
}

# Eth-Execution-Version: shanghai
ExecutionPayloadEnvelopeShanghai {
    payload: <ExecutionPayloadShanghai>         # + withdrawals
}

# Eth-Execution-Version: cancun
ExecutionPayloadEnvelopeCancun {
    payload:                  <ExecutionPayloadCancun>   # + blob_gas_used, excess_blob_gas
    parent_beacon_block_root: <Root>                     # NEW: was a side param since Cancun
}

# Eth-Execution-Version: prague
ExecutionPayloadEnvelopePrague {
    payload:                  <ExecutionPayloadPrague>   # == Cancun payload shape
    parent_beacon_block_root: <Root>
    execution_requests:       <List[Bytes, MAX_EXECUTION_REQUESTS_PER_PAYLOAD]>  # NEW: was a side param since Prague
}

# Eth-Execution-Version: osaka — identical envelope shape to Prague
ExecutionPayloadEnvelopeOsaka {
    payload:                  <ExecutionPayloadOsaka>    # == Cancun payload shape
    parent_beacon_block_root: <Root>
    execution_requests:       <List[Bytes, MAX_EXECUTION_REQUESTS_PER_PAYLOAD]>
}

# Eth-Execution-Version: amsterdam
ExecutionPayloadEnvelopeAmsterdam {
    payload:                  <ExecutionPayloadAmsterdam># + block_access_list, slot_number
    parent_beacon_block_root: <Root>
    execution_requests:       <List[Bytes, MAX_EXECUTION_REQUESTS_PER_PAYLOAD]>
}
```

The response for **all** forks is `PayloadStatus` (fork-invariant).

#### `GET /payloads/{id}` response body (`BuiltPayload{Fork}`)

```
# Eth-Execution-Version: paris
BuiltPayloadParis {
    payload:     <ExecutionPayloadParis>
    block_value: <Uint256>
}

# Eth-Execution-Version: shanghai
BuiltPayloadShanghai {
    payload:     <ExecutionPayloadShanghai>
    block_value: <Uint256>
}

# Eth-Execution-Version: cancun
BuiltPayloadCancun {
    payload:                 <ExecutionPayloadCancun>
    block_value:             <Uint256>
    blobs_bundle:            <BlobsBundleV1>            # NEW in Cancun (single proof)
    should_override_builder: <Boolean>                  # NEW in Cancun
}

# Eth-Execution-Version: prague
BuiltPayloadPrague {
    payload:                 <ExecutionPayloadPrague>
    block_value:             <Uint256>
    blobs_bundle:            <BlobsBundleV1>
    execution_requests:      <List[Bytes, MAX_EXECUTION_REQUESTS_PER_PAYLOAD]>  # NEW in Prague, before should_override_builder
    should_override_builder: <Boolean>
}

# Eth-Execution-Version: osaka
BuiltPayloadOsaka {
    payload:                 <ExecutionPayloadOsaka>
    block_value:             <Uint256>
    blobs_bundle:            <BlobsBundleV2>            # CHANGED in Osaka (cell proofs)
    execution_requests:      <List[Bytes, MAX_EXECUTION_REQUESTS_PER_PAYLOAD]>
    should_override_builder: <Boolean>
}

# Eth-Execution-Version: amsterdam
BuiltPayloadAmsterdam {
    payload:                 <ExecutionPayloadAmsterdam>
    block_value:             <Uint256>
    blobs_bundle:            <BlobsBundleV2>
    execution_requests:      <List[Bytes, MAX_EXECUTION_REQUESTS_PER_PAYLOAD]>
    should_override_builder: <Boolean>
}
```

#### `POST /forkchoice` request body (`ForkchoiceUpdate{Fork}`)

The wrapper is the same Paris→Osaka (`custody_columns` is Amsterdam+);
only the inner `payload_attributes` shape changes per fork:

```
# Paris .. Osaka
ForkchoiceUpdate{Fork} {
    forkchoice_state:   <ForkchoiceState>               # fork-invariant
    payload_attributes: Optional[<PayloadAttributes{Fork}>]
}

# Amsterdam
ForkchoiceUpdateAmsterdam {
    forkchoice_state:   <ForkchoiceState>
    payload_attributes: Optional[<PayloadAttributesAmsterdam>]   # + slot_number, target_gas_limit
    custody_columns:    Optional[Bitvector[CELLS_PER_EXT_BLOB]]  # NEW in Amsterdam
}
```

where the per-fork `payload_attributes` adds `withdrawals` at Shanghai,
`parent_beacon_block_root` at Cancun, and
`slot_number`/`target_gas_limit` at Amsterdam. The response
(`ForkchoiceUpdateResponse`) is fork-invariant.

For the `/bodies` and `/blobs/vN` per-fork / per-revision bodies, see
the [container catalogue](./refactor-ssz.md#per-fork-container-catalogue);
`/bodies` varies the inner `ExecutionPayloadBody` (`+withdrawals` at
Shanghai, `+block_access_list` at Amsterdam) and `/blobs/vN` is
independently versioned rather than fork-scoped.

---

## Error model

Errors are signalled by HTTP status code and an
`application/problem+json` body (RFC 7807). To keep responses compact,
we use only **two** of the RFC 7807 fields:

- **`type`** (required) — relative URI identifying the problem class.
  Stable across releases. CLs branch on this string.
- **`detail`** (optional) — human-readable, instance-specific message.
  Omitted when the EL has nothing more to say than the `type` already
  conveys (e.g. canned SSZ-decode failures).

Success codes:

| HTTP status | When |
| - | - |
| `200 OK` | SSZ-encoded response body |
| `204 No Content` | Null result (e.g. blob pool syncing on `/blobs/vN`); empty body |

Error codes:

| HTTP status | `type` | Old JSON-RPC code | When |
| - | - | - | - |
| 400 Bad Request | `/engine-api/errors/parse-error` | -32700 | Body is not valid JSON / SSZ |
| 400 Bad Request | `/engine-api/errors/invalid-request` | -32600 | Request shape is wrong (missing required field, etc.) |
| 400 Bad Request | `/engine-api/errors/ssz-decode-error` | (new) | SSZ decode failed; canned error, no `detail` |
| 400 Bad Request | `/engine-api/errors/unsupported-fork` | -38005 | `Eth-Execution-Version` value missing, unknown, or unsupported by this EL |
| 404 Not Found | `/engine-api/errors/method-not-found` | -32601 | URL does not match any endpoint |
| 404 Not Found | `/engine-api/errors/unknown-payload` | -38001 | `payloadId` does not exist |
| 409 Conflict | `/engine-api/errors/invalid-forkchoice` | -38002 | Forkchoice state is inconsistent (e.g. finalized not ancestor of head) |
| 409 Conflict | `/engine-api/errors/reorg-too-deep` | -38006 | Reorg depth exceeds the EL's limit |
| 413 Payload Too Large | `/engine-api/errors/request-too-large` | -38004 | Body exceeds an advertised `limits.*` value |
| 415 Unsupported Media Type | `/engine-api/errors/unsupported-media-type` | (new) | Request `Content-Type` does not match the endpoint's expected encoding (SSZ for hot-path, JSON for diagnostics) |
| 422 Unprocessable Entity | `/engine-api/errors/invalid-body` | -32602 | Body decoded fine but has invalid values |
| 422 Unprocessable Entity | `/engine-api/errors/invalid-attributes` | -38003 | `payload_attributes` validation failed |
| 500 Internal Server Error | `/engine-api/errors/internal` | -32603 / -32000 | Unrecoverable server error; `detail` carries the message |

`type` URIs are written as **relative references** rooted at
`/engine-api/errors/...`.

Example error body:

```json
{
  "type":   "/engine-api/errors/invalid-forkchoice",
  "detail": "finalized 0xab.. is not an ancestor of head 0xcd.."
}
```

Canned error (no `detail`):

```json
{ "type": "/engine-api/errors/ssz-decode-error" }
```

Validation outcomes for a payload (`VALID`, `INVALID`, `SYNCING`,
`ACCEPTED`) are **not** errors — they remain part of the response
body with HTTP `200 OK`. HTTP errors are reserved for transport,
format, and authentication problems.

---

## Versioning model

Three layers:

1. **Major (`/v1`, future `/v2`)** — `/v1` is the first REST version
   of the engine API. A future `/v2` is reserved for breaking
   transport changes (e.g. moving away from REST, swapping SSZ for
   something else).
2. **Per-fork body schema** — selected via the
   `Eth-Execution-Version: <fork>` request header on hot-path
   endpoints (`/payloads`, `/forkchoice`, `/bodies`). Tracks
   consensus-protocol changes that ride along with fork activations.
   Accepted values span **Paris through Amsterdam** (`paris`,
   `shanghai`, `cancun`, `prague`, `osaka`, `amsterdam`); Paris is
   the earliest fork with an Engine API and therefore the lowest
   value an EL accepts. A request with a header value below the EL's
   earliest supported fork, one it doesn't recognise, or a missing
   header on a fork-scoped endpoint returns
   `400 /engine-api/errors/unsupported-fork`.
3. **Per-endpoint revisions** — selected via a `/vN` URL segment on
   endpoints whose protocol evolves independently of the fork
   schedule (currently just `/blobs/vN`). Tracks engine-API protocol
   changes that don't align with fork activations.

**Blob-parameter-only (BPO) forks** do **not** get their own
`Eth-Execution-Version` value. A BPO fork only changes blob-count
parameters, not any Engine API body schema, so a chain in a BPO era
keeps negotiating against the named fork it layers on — e.g.
BPO1–BPO5 all send `Eth-Execution-Version: osaka` and use the Osaka
wire shapes. Only named forks that change a body schema introduce a
new accepted header value. CLs MUST map a BPO era onto its base named
fork when constructing the header.

The server advertises which forks and which `/vN` revisions it
understands via `GET /engine/v1/capabilities`.

`engine_exchangeCapabilities` is **removed**. Instead the server lists
its supported fork schemas and endpoint set in a single JSON document
at `/engine/v1/capabilities`.

### Capabilities format

We considered advertising capabilities as a flat list of per-endpoint
strings (e.g. `"POST /payloads@amsterdam"`, the format used by the
existing `engine_exchangeCapabilities` method). The structured form
in `GET /capabilities` (separate `supported_forks`,
`fork_scoped_endpoints`, `independently_versioned`,
`unscoped_endpoints`, plus per-endpoint `limits`) is preferred
because:

- Adding a fork doesn't multiply the capability list — one entry in
  `supported_forks` covers every fork-scoped endpoint at once.
- The `limits.*` block can carry numeric per-endpoint bounds
  (`bodies.max_count`, `blobs.max_versioned_hashes`,
  `payload.max_bytes`) which a string-list form can't.
- It's easier to evolve: new fields land alongside, old CLs ignore
  them.

### Transition-window behavior

During the rollout window, a CL upgraded to the REST API may
interact with an EL still on the legacy JSON-RPC engine API. Two
cases:

- **EL doesn't expose `/engine/v1/...` at all.** The CL hits any
  REST URL and gets `404 Not Found` from the legacy server. The CL
  falls back to JSON-RPC for the duration of that EL's lifetime — no
  per-method retry dance.
- **EL exposes `/engine/v1/...` but doesn't know the requested fork.**
  The CL hits a fork-scoped endpoint with
  `Eth-Execution-Version: amsterdam` against an EL that only
  advertised `supported_forks: [..., cancun]`. The EL returns
  `400 /engine-api/errors/unsupported-fork`. The CL learns this once
  from `GET /capabilities` and avoids issuing such requests; if it
  doesn't, the per-request error is structured and explicit, not a
  silent downgrade.

There is **no per-method fallback ladder**. A CL either uses the
REST API or JSON-RPC for the lifetime of an EL connection; mixing
transports within a connection is permitted but not required.

---

## Authentication

Unchanged in spirit: JWT (HS256, 256-bit shared secret). Differences:

- The token MUST be presented as `Authorization: Bearer <jwt>` on every
  request. The HTTP/2 connection itself is not authenticated; each
  request stream carries its own bearer token. This means a single
  long-lived h2 connection between CL and EL is fine — token rotation
  happens per-request, not per-connection.
- IPC (UNIX socket) authentication remains optional, as today.
- JWT claims:
  - `iat` (required, unchanged from today: ±60s window)
  - `id` (optional, unchanged)
  - `clv` is **removed** — the CL version travels in the
    `X-Engine-Client-Version` request header instead. Keeping it in
    two places caused drift; the header is structured, cheap, and
    surfaces in normal HTTP logs.
- **Trace propagation:** CLs MAY include a W3C `traceparent` header
  on each request. ELs that record a `traceparent` SHOULD propagate
  it into their own logs / spans so a slot-level trace can cross the
  CL→EL boundary. Not required, not authenticated, purely diagnostic.

---

## Transport & framing

- **Protocol:** both **HTTP/2 and HTTP/1.1 MUST be supported**, with
  HTTP/2 **preferred**. Both TCP and IPC transports use **cleartext**
  (h2c for HTTP/2); JWT-on-every-request provides authentication, so
  TLS termination is left to a reverse proxy if the operator wants
  it. Servers and CLs **SHOULD** negotiate HTTP/2 where available
  (ALPN over TLS, or the HTTP/2 prior-knowledge / `h2c` upgrade over
  cleartext) and fall back to HTTP/1.1 only when the peer does not
  speak h2. HTTP/2 multiplexing lets a single CL→EL connection carry
  the full request mix (forkchoice, payload submission, blob fetches,
  body fetches) without head-of-line blocking; on HTTP/1.1 that
  benefit is lost (requests serialise per connection, or the CL opens
  several connections), but the API is otherwise identical — same
  paths, headers, bodies, and status codes. CLs that fall back to
  HTTP/1.1 SHOULD use connection pooling to recover some concurrency.
- **Default port:** `8551`, shared with the legacy JSON-RPC engine API.
  The two surfaces are distinguished by path: legacy JSON-RPC remains
  at `/` (and accepts JSON-RPC method calls), the new API lives under
  `/engine/v1/...`. The same JWT secret authenticates both.
- **Base path:** `/engine/v1/...`. `/v1` is the first major REST
  version (a future `/v2` is reserved for breaking transport
  changes). The fork-scoped body schema is selected by the
  `Eth-Execution-Version: <fork>` request header rather than a URL
  segment (`paris`, `shanghai`, `cancun`, `prague`, `osaka`,
  `amsterdam`, …). Adding a fork = adding one accepted header value
  and one set of SSZ schemas. See [Versioning](#versioning-model).
- **Fork header:** every hot-path request MUST carry
  `Eth-Execution-Version: <fork>`. Missing or unknown header on a
  fork-scoped endpoint returns
  `400 /engine-api/errors/unsupported-fork`. Unscoped endpoints
  (`/capabilities`, `/identity`, `/blobs/vN`) MUST ignore the header
  if present.
- **Content-Type / Accept matrix:**

  | Channel | Header | Value |
  | - | - | - |
  | Hot-path request body (`/payloads`, `/forkchoice`, `/bodies`, `/blobs/vN`) | `Content-Type` | `application/octet-stream` (SSZ) |
  | Hot-path request | `Accept` | `application/octet-stream` |
  | Hot-path response success body | `Content-Type` | `application/octet-stream` (SSZ) |
  | Diagnostic request / response (`/capabilities`, `/identity`) | `Content-Type` | `application/json` |
  | Error response body (any endpoint) | `Content-Type` | `application/problem+json` |

  ELs MUST reject hot-path requests carrying any other `Content-Type`
  with `415 Unsupported Media Type`. Diagnostic endpoints MUST be
  served as JSON regardless of `Accept`.
- **Compression:** Servers MAY support `Accept-Encoding: zstd, gzip`.
  Not required to implement; CLs MUST tolerate uncompressed responses.
  Blob bundles compress well, so operators are encouraged to enable
  `zstd` where available.
- **Flow-control window:** servers and CLs **SHOULD** set HTTP/2
  `INITIAL_WINDOW_SIZE` to at least 1 MiB. Default 64 KiB causes
  excessive flow-control round-trips for blob bundles and large
  `getPayload` responses. `MAX_FRAME_SIZE` and `MAX_HEADER_LIST_SIZE`
  use HTTP/2 defaults — not pinned by this spec.
- **Connection lifecycle:** CLs MAY open fresh h2 connections per
  request or reuse a long-lived connection. JWT is per-request so
  token rotation works the same way in both patterns.

### Why a fork header instead of method versioning?

Today every change of a single field bumps the method version
(`engine_newPayloadV1..V5`). The new API puts the fork in a request
header:

```
POST /engine/v1/payloads
Eth-Execution-Version: amsterdam
Content-Type: application/octet-stream
Authorization: Bearer <JWT>

<SSZ-encoded ExecutionPayloadEnvelope>
```

The EL routes by header value, parses the body according to that
fork's SSZ schema, and returns a fork-shaped response. Adding a fork
= adding one accepted header value and one set of SSZ schemas. We
considered putting the fork in the URL (`/{fork}/payloads`) but
chose the header because it keeps URLs stable across forks — content
negotiation is the standard HTTP idiom for "same resource, different
schema version" and the Beacon API already uses
`Eth-Consensus-Version` for the same purpose on the CL side.

---

## SSZ encoding conventions

- **`Optional[T]` ≡ `List[T, 1]`.** SSZ has no native optional type;
  we use a length-0-or-1 list as the convention (`[]` = absent,
  `[t]` = present). The notation `Optional[T]` in this document is
  syntactic sugar for `List[T, 1]`. We picked this over
  `Union[None, T]` because `List` is universally supported across
  SSZ libraries.
- **`String` ≡ `List[byte, MAX_ERROR_BYTES]`** (UTF-8). Empty list
  is the empty string; use `Optional[String]` if absence must be
  distinguishable from empty.
- **Endianness:** SSZ uints are **little-endian**. The JSON-RPC API
  encoded `QUANTITY` values as big-endian hex, so anything that
  carries a uint (`block_value`, `gas_used`, `gas_limit`, `timestamp`,
  `base_fee_per_gas`, `excess_blob_gas`, `blob_gas_used`,
  `block_number`, the `index`/`validatorIndex`/`amount` triple in
  `Withdrawal`) flips byte order on the wire.
- **`MAX_*` constants** live in the fork-scoped SSZ schema files
  (e.g. `MAX_TXS_PER_PAYLOAD`, `MAX_WITHDRAWALS_PER_PAYLOAD`,
  `MAX_BAL_BYTES`, `MAX_VERSIONED_HASHES_PER_REQUEST`).
  `MAX_ERROR_BYTES` is global and pinned at `1024` here.

### JSON-RPC type → SSZ type mapping

For implementers porting from the JSON-RPC API, the legacy openrpc
base types map onto SSZ as follows:

| JSON-RPC type | SSZ type |
| - | - |
| `address` (20 bytes) | `Bytes20` |
| `hash32` (32 bytes) | `Bytes32` |
| `bytes8` (8 bytes) | `Bytes8` |
| `bytes32` (32 bytes) | `Bytes32` |
| `bytes48` (48 bytes) | `Bytes48` |
| `bytes256` (256 bytes) | `ByteVector[256]` |
| `bytesMax32` (0–32 bytes) | `ByteList[32]` |
| `bytes` (variable-length) | `ByteList[MAX_*]` (context-dependent) |
| `uint64` | `uint64` |
| `uint256` | `uint256` |
| `BOOLEAN` | `boolean` |
| `Array of T` | `List[T, MAX_*]` (context-dependent) |
| `T \| null` | `Optional[T]` (= `List[T, 1]`) |

### Cross-fork response containers

Endpoints that return data spanning multiple block-eras come in two
flavours:

1. **Fork-scoped** (e.g. `/bodies`): `Eth-Execution-Version` selects
   the container schema *and* limits the response to blocks from
   that fork's time range. Every field in the fork's body container
   is unconditionally present (no `Optional[T]` for cross-fork
   nullability); blocks outside the fork's range come back as
   `available=false` on the outer entry instead of as a
   zero-valued body:

   ```
   # POST /bodies/hash with Eth-Execution-Version: amsterdam
   BodyEntry {
       available: boolean
       body:      ExecutionPayloadBody
   }

   # Amsterdam ExecutionPayloadBody — every field always present
   ExecutionPayloadBody {
       transactions:       List[Transaction, MAX_TXS]
       withdrawals:        List[Withdrawal, MAX_WITHDRAWALS]
       block_access_list:  ByteList[MAX_BAL_BYTES]
   }
   ```

   A CL fetching a Cancun-era block sends
   `Eth-Execution-Version: cancun` and receives the Cancun container
   (no `block_access_list` field at all, and no `Optional` wrapper on
   `withdrawals`). Cross-fork ranges require multiple requests, one
   per header value.

2. **Independently versioned** (e.g. `/blobs/vN`): each revision is
   its own container, no nullable optionals across revisions. Old
   CLs keep using `/blobs/v1`; new shapes ship as `/blobs/vN+1`
   alongside.

---

## Message ordering & idempotency

HTTP/2 multiplexes streams over a single connection and a server
handler may complete in any order. The Engine API is sensitive to
ordering, so we pin two rules explicitly:

- **CL-driven ordering.** The CL is responsible for serialising
  dependent requests. In particular:
  - Only one `POST /forkchoice` may be in flight at a time.
  - If a `POST /payloads` is logically before a `POST /forkchoice`
    (or vice versa), the CL MUST wait for the first response before
    issuing the second.
  - The EL processes streams in receive order. h2 multiplexing
    across independent CL→EL flows is fine; the CL MUST NOT rely on
    the EL to reorder its own dependent requests.

- **Idempotency, narrowly defined.** Today's
  [`paris.md`](./paris.md) #4 specifies idempotency only with respect
  to `VALID | INVALID`: once a payload is decided one way, it cannot
  flip. But `SYNCING → VALID`, `SYNCING → INVALID`, and
  `ACCEPTED → VALID/INVALID` transitions are explicitly allowed —
  the same payload submitted twice can return different statuses if
  the EL has acquired more state in between. The new spec preserves
  this: an EL MUST NOT short-circuit a retry by returning the cached
  status, and a CL MUST NOT assume two responses to the same envelope
  match. The only invariant is the `VALID ↔ INVALID` boundary.

---

## Security considerations

SSZ `MAX_*` constants bound *on-chain validity*, not per-request
resource use. A naive decoder facing a crafted `Content-Length`,
length prefix, or offset can be coerced into large allocations or
scans before any semantic rejection. ELs implementing this API
**MUST**:

- **Cap by `Content-Length`** against an endpoint-specific maximum
  *before* reading the body when the header is present, and cap the
  bytes read from the body in all cases.
- **Validate SSZ length prefixes and offsets** against the remaining
  buffer size *before* allocating backing storage for variable-length
  fields.
- **Apply per-endpoint operational caps** (reverse proxy,
  server config) in addition to library-level checks. The advertised
  `limits.*` values in `GET /capabilities` are an upper bound, not a
  target — operators are encouraged to reject earlier.

ELs **SHOULD** use well-tested SSZ libraries and fuzz-test SSZ
parsing extensively. JWT authentication is unchanged from the legacy
JSON-RPC API; all existing requirements apply.

---

## Motivation

The remainder of this document is rationale and reference material:
why we made the choices the spec encodes above, plus a consolidated
decision log for quick scanning.

### Goals & non-goals

#### Goals

1. **Reduce wire size and parse cost.** SSZ-encoded bodies are 30–50%
   smaller than hex-JSON for payload-shaped data and parse in linear time
   without nibble decoding. This matters most for blob bundles (multi-MB
   per slot) and the new `blockAccessList`.
2. **Stop the version sprawl.** Today every fork bumps every method that
   touches a changed structure (`engine_newPayload` is at V5,
   `engine_getPayload` at V6, etc.). The new API puts the fork in the
   URL (`/engine/v1/...`) so a single endpoint accepts whatever
   schema that fork mandates; adding a fork = adding one path prefix
   plus one set of SSZ schemas, not bumping every method name.
3. **Self-contained requests.** No more side-channel parameters
   (`expectedBlobVersionedHashes`, `parentBeaconBlockRoot`,
   `executionRequests`) that travel beside the payload — they live
   inside the payload envelope or are unnecessary.
4. **Idiomatic HTTP.** Use HTTP status codes for transport-level outcomes,
   `Content-Type` for negotiation, and a small problem-detail JSON body
   for errors.

#### Non-goals

- Dropping the EL/CL split, changing trust boundaries, or moving CL state
  into the EL (or vice versa).
- Removing JWT. Authentication is unchanged; only the *transport* of the
  bearer token differs (HTTP `Authorization` header, same as today).
- Replacing `eth_*` JSON-RPC. The `eth` namespace stays JSON-RPC. This
  document only refactors the `engine_*` namespace.
- Wire-perfect SSZ container definitions. The encoding *conventions*
  are pinned in this document; the concrete field-by-field SSZ
  containers per fork (e.g. the Amsterdam `ExecutionPayload` schema)
  are deferred to a follow-up.

### Why move away from JSON-RPC?

JSON-RPC over HTTP has served the Engine API since Paris. The pain points
that prompt this refactor:

- **Encoding overhead.** Every `DATA` field is a `0x`-prefixed lowercase
  hex string. A 128 KiB blob becomes a 256 KiB+ string. With Osaka /
  Fulu blob counts and the Amsterdam `blockAccessList`, payloads are
  routinely multi-megabyte.
- **No content negotiation.** A new fork structure forces a new method
  name (`engine_newPayloadV5`), even when the only change is one added
  field. With a REST endpoint and a fork header
  (`Eth-Execution-Version: amsterdam`), the body schema is selected
  by routing + content negotiation, not by method-name suffix.
- **Side-channel params.** JSON-RPC's positional params encourage
  bolting on extras like `parentBeaconBlockRoot` and
  `executionRequests` next to the payload, instead of inside it.
- **Errors are non-standard.** `-38001..-38006` are bespoke and require
  client-side mapping. HTTP status codes + a typed problem body are
  universally understood.

JSON-RPC is fine for the casual `eth_*` query API. For the hot path
between CL and EL, we want something denser and more disciplined.

### Why SSZ?

- The CL already speaks SSZ natively for its block, attestation, blobs,
  KZG, and request structures. The CL today **converts SSZ → JSON →
  hex-strings** when it forwards a payload, then the EL parses hex-JSON
  back to bytes. This conversion is pure overhead and has been a
  recurring source of subtle field-encoding bugs (e.g. the
  `withdrawals.amount` LE-vs-BE note in shanghai.md).
- SSZ's fixed/variable-length distinction lets us validate sizes
  cheaply at the transport layer.
- It's already what consensus-specs uses to define `ExecutionPayload`,
  `Withdrawal`, `BlobsBundle`, etc. We'd be aligning, not inventing.

We keep JSON available for **error bodies, capability discovery, and
client identification** because those are ergonomic to debug with `curl`
and not on the hot path.

#### Why not RLP?

RLP is the EL's native encoding, so reusing it would cut one library
dependency on the EL side. We picked SSZ instead because:

- **The CL natively serialises every payload field as SSZ today.** An
  RLP transport would shift the conversion from "EL parses hex-JSON"
  to "CL re-encodes SSZ as RLP" — same total work, just on a
  different host.
- **SSZ pins fixed/variable lengths at the type level.** The
  transport layer can enforce per-field size limits before
  allocation, which RLP's recursive header structure makes harder.
- **`hash_tree_root` for free.** SSZ types come with a deterministic
  Merkle root we can use for future content-addressed extensions
  (e.g. payload identifiers, capability hashes). RLP would need a
  separate hashing convention.
- **Alignment with the rest of the consensus stack.** Beacon API,
  fork-choice store, gossip — all SSZ. Reusing the same encoding at
  the EL/CL boundary keeps one mental model.

### Simplifications & removed concepts

1. **`expectedBlobVersionedHashes`** — **removed**. The block-hash check
   already covers the transactions, so the EL recomputes the array
   from `payload.transactions` during validation and surfaces a
   mismatch as `INVALID`. The CL no longer sends a redundant copy.
2. **`INVALID_BLOCK_HASH`** — **removed** from the enum. Already
   supplanted by `INVALID` since Shanghai.
3. **`ACCEPTED`** — **kept**. CLs use this status during sync to
   acknowledge well-formed side-branch payloads that don't extend the
   canonical chain.
4. **`shouldOverrideBuilder`** — kept, lives inside the SSZ
   `BuiltPayload` body. (Considered moving to a response header but it
   complicates the SSZ canonicalisation; better inside the body.)
5. **`engine_exchangeCapabilities`** as a polling handshake — replaced
   by a single `GET /capabilities`.
6. **`engine_exchangeTransitionConfigurationV1`** — dropped. Already
   deprecated since Cancun.
7. **`payloadId` derivation** — today both sides recompute an 8-byte
   hash over `(headBlockHash, payloadAttributes)`. The new
   `POST /forkchoice` returns `payload_id` directly in the response;
   it is an **opaque server-assigned token**. The EL chooses how to
   mint it; CLs MUST treat it as opaque bytes.
8. **The split between `engine_*` namespace and the `eth_*` subset
   the EL must expose** — out of scope for this refactor; the `eth_*`
   namespace stays JSON-RPC.
9. **Per-method `timeout` SHOULDs** — replaced with HTTP-standard
   request timeouts and `Retry-After` semantics on 503.

### Summary of design decisions

This is the consolidated decision log. Every item below is normative
and is also detailed in the relevant section earlier in the document;
the summary exists for quick scanning.

#### Scope

- **Target fork:** Amsterdam. The new API ships *as* the Amsterdam
  Engine API. Pre-Amsterdam timestamps continue to be served by the
  legacy JSON-RPC API on the same port; clients run both surfaces.
- **Backwards compatibility** is out of scope. The legacy JSON-RPC
  engine API is left in place by clients; this spec does not require
  or forbid sunset.
- **`eth_*` JSON-RPC subset** (`eth_blockNumber`, `eth_call`,
  `eth_chainId`, `eth_getCode`, `eth_getBlockByHash`,
  `eth_getBlockByNumber`, `eth_getLogs`, `eth_sendRawTransaction`,
  `eth_syncing`) is **not** mirrored under `/engine/v1/...`. CLs that
  need state / log access continue to call them via the legacy
  JSON-RPC root.

#### Transport

- **HTTP/2 and HTTP/1.1 both MUST be supported**, HTTP/2 preferred;
  cleartext (h2c for HTTP/2) for both TCP and IPC. Peers negotiate
  HTTP/2 where available and fall back to HTTP/1.1 otherwise. The API
  surface (paths, headers, bodies, status codes) is identical on both;
  only HTTP/2's stream multiplexing is lost on the 1.1 fallback.
  JWT-on-every-request authenticates; TLS termination is left to a
  reverse proxy.
- **IPC** is h2c over UNIX socket — same paths and headers as TCP,
  single code path.
- **Default port `8551`**, shared with the legacy JSON-RPC API
  (distinguished by path).
- **Flow-control:** SHOULD set `INITIAL_WINDOW_SIZE` ≥ 1 MiB.
  `MAX_FRAME_SIZE` and `MAX_HEADER_LIST_SIZE` use HTTP/2 defaults.
- **Connection lifecycle:** CLs MAY open fresh h2 connections per
  request or reuse a long-lived connection.
- **Compression:** `zstd` and `gzip` MAY be implemented. CLs MUST
  tolerate uncompressed responses.

#### Versioning

- **Fork-scoped endpoints:** `/payloads`, `/forkchoice`, `/bodies`.
  Fork in the `Eth-Execution-Version` request header.
- **Independently versioned endpoints:** `/blobs/vN`. Legacy
  `engine_getBlobsVN` numbers carry forward onto the URL. ELs MUST
  serve at least the revision matching their current fork
  (`/blobs/v4` for Amsterdam) and MAY serve older revisions
  alongside. Future blob-shape changes ship as `/blobs/v5`, `/v6`,
  etc.
- **Unscoped endpoints:** `/capabilities`, `/identity`.
- **Major version `/v1`** is the first REST version; future `/v2`
  reserved for breaking transport changes (e.g. dropping REST or
  SSZ).

#### Encoding

- **Hot-path bodies use SSZ.** Diagnostic / metadata endpoints
  (`/capabilities`, `/identity`, error bodies) use JSON. The legacy
  `engine_*` JSON-RPC endpoint at `/` remains available alongside
  this surface during the rollout window and is retired at a future
  fork `F` (TODO: pin), at which point the REST + SSZ surface
  becomes the only way to drive an EL.
- **`Optional[T]` ≡ `List[T, 1]`** (length 0 = absent, length 1 =
  present). Universally supported by SSZ libraries.
- **Strings ≡ `List[byte, MAX_ERROR_BYTES]`**, `MAX_ERROR_BYTES = 1024`.
- **Endianness:** SSZ uints are little-endian. This flips byte order
  vs the JSON-RPC `QUANTITY` type for `block_value`, `gas_used`,
  `timestamp`, `base_fee_per_gas`, `excess_blob_gas`,
  `blob_gas_used`, `block_number`, and the
  `index`/`validatorIndex`/`amount` triple in `Withdrawal`.
- **`MAX_*` constants** are defined in fork-scoped SSZ schema files;
  `MAX_ERROR_BYTES` is global.
- **Cross-fork response containers** come in two flavours:
  fork-scoped (`/bodies`) uses `Eth-Execution-Version` to pick
  *both* the schema and the era of returned blocks (every body field
  always present; out-of-era blocks come back as `available=false`);
  independently versioned (`/blobs/vN`) gives each revision its own
  dedicated container. Both wrap their entries in
  `BodyEntry { available, body }` / `BlobEntry { available, contents }`.
  Whole-response "syncing / all-or-nothing miss" is signalled by
  HTTP `204 No Content`, not an in-band SSZ sentinel. Per-entry fork
  tags were rejected.

#### Error model

- **RFC 7807 with two fields:** `type` (required, relative URI rooted
  at `/engine-api/errors/...`) and `detail` (optional). Drop `title`,
  `status`, `instance`, `engine_code`.
- **SSZ-decode failures** are a canned `400 Bad Request` with
  `type=/engine-api/errors/ssz-decode-error`, no `detail`.

#### Ordering & idempotency

- **CL-driven ordering.** Only one `POST /forkchoice` in flight at a
  time; `POST /payloads` ordered with respect to surrounding FCUs by
  the CL. No sequence number on the wire.
- **Idempotency is narrow.** `VALID ↔ INVALID` cannot flip. All
  other transitions (`SYNCING → VALID/INVALID`,
  `ACCEPTED → VALID/INVALID`) are allowed; ELs MUST NOT short-circuit
  retries.

#### Forkchoice update (`POST /forkchoice`)

- **Single atomic call** carrying forkchoice state, optional
  `payload_attributes`, and optional `custody_columns`.
- **Skip-allowed semantics:** EL MAY skip applying state when the
  new `head` is a `VALID` ancestor of the latest finalized block,
  guarding against malformed CL FCUs.
- **Stale-fork header** is allowed when `payload_attributes` is
  absent; with `payload_attributes` present,
  `Eth-Execution-Version` MUST match the timestamp's fork
  (otherwise `400 unsupported-fork`).
- **No HTTP-layer body cap** beyond SSZ `MAX_*` constants.
- **Custody-set updates** run independently of the forkchoice flow;
  custody errors do not affect `payload_status`.
- **Custody-set lifetime:** set until the next FCU that includes a
  `custody_columns` field. FCUs that omit it leave the set unchanged.

#### Payload submission (`POST /payloads`)

- **`expectedBlobVersionedHashes` removed.** EL recomputes from
  `payload.transactions`; block-hash check covers transactions.
- **`INVALID_BLOCK_HASH` removed** from the status enum.
- **`ACCEPTED` kept** — CLs use it during sync.
- **Transaction min-length** ("at least 1 byte") remains a
  receiver-side validation rule, not an SSZ schema invariant.

#### Payload retrieval (`GET /payloads/{payloadId}`)

- **Poll-only**, same semantics as today's `engine_getPayload`. No
  SSE / long-poll.
- **`payload_id` is an opaque server-assigned token** issued by
  `POST /forkchoice`. CLs MUST NOT recompute or validate it.
- **`payload_id` lifetime is build-bound, not time-bound.** A token
  remains valid until either the payload was retrieved by
  `GET /payloads/{payloadId}` or another payload was built
  via a forkchoice with payload attributes.
- **`shouldOverrideBuilder`** lives inside the SSZ `BuiltPayload`
  body.

#### Authentication & telemetry

- **JWT (HS256, 256-bit secret)** unchanged in spirit, presented as
  `Authorization: Bearer <jwt>` on every request.
- **JWT claims:** `iat` required (±60s), `id` optional, **`clv`
  removed**.
- **`X-Engine-Client-Version`** is the canonical CL version channel.
- **`traceparent`** (W3C trace context) is supported but optional.

#### Operations

- **Multi-CL setups** are operator-managed. The spec does not track
  CL identity or restrict who calls `POST /forkchoice`. Today's "one
  writer, many readers" pattern carries forward unchanged.
- **`GET /capabilities`** advertises supported forks, fork-scoped
  endpoints, independently-versioned endpoints (with the available
  `/vN` list), unscoped endpoints, and per-endpoint maximum request
  sizes.

#### Removed concepts

- `engine_exchangeCapabilities` — replaced by `GET /capabilities`.
- `engine_exchangeTransitionConfigurationV1` — already deprecated
  since Cancun.
- Per-method `timeout` SHOULDs — replaced by HTTP-standard request
  timeouts and `Retry-After` semantics on 503.
- The mutual-exchange handshake of `engine_getClientVersionV1` —
  replaced by one-way `GET /identity` plus the
  `X-Engine-Client-Version` request header.

---

## Future evolution

### Progressive merkleization

Progressive (chunked / streaming) SSZ merkleization is **not used**
in this draft. It would let the EL stream a partially-built
`BuiltPayload` without recomputing the `hash_tree_root` from
scratch on every `GET` poll, which is attractive for the polling
loop on `/payloads/{id}` — but the spec depends on container
shapes that are still in flux pending
[EIP-7688](https://eips.ethereum.org/EIPS/eip-7688). Adopting
progressive merkleization now would freeze container layout
choices we may want to revisit once 7688 is SFI-ed.

Plan: revisit progressive merkleization in a follow-up revision of
this spec once EIP-7688 is SFI-ed. At that point we can redesign
the affected containers (`BuiltPayload`, `BlobsBundle`) to be
progressive-friendly without inheriting any constraints from the
current shapes.
