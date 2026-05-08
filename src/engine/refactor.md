# Engine API -- Refactor Proposal (REST + SSZ)

> **Status:** Draft / discussion document. This file proposes a v2 of the
> Engine API that moves from JSON-RPC over a single endpoint to a
> resource-oriented HTTP/REST API where request and response bodies are
> SSZ-encoded. It also takes the opportunity to simplify the surface that
> has accumulated since Paris.
>
> **Target fork:** Amsterdam. The new API ships *as* the Amsterdam Engine
> API; clients implement it instead of `engine_*` JSON-RPC at the
> Amsterdam activation timestamp.

This document is meant to be read alongside the existing fork-scoped specs
([Paris](./paris.md), [Shanghai](./shanghai.md), [Cancun](./cancun.md),
[Prague](./prague.md), [Osaka](./osaka.md), [Amsterdam](./amsterdam.md)).
Concrete byte-level structures are deferred to a later iteration; the goal
here is to align on the *shape* of the new API.

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
- [Error model](#error-model)
- [Versioning model](#versioning-model)
- [Authentication](#authentication)
- [Transport & framing](#transport--framing)
- [SSZ encoding conventions](#ssz-encoding-conventions)
- [Message ordering & idempotency](#message-ordering--idempotency)
- [Motivation](#motivation)
  - [Goals & non-goals](#goals--non-goals)
  - [Why move away from JSON-RPC?](#why-move-away-from-json-rpc)
  - [Why SSZ?](#why-ssz)
  - [Simplifications & removed concepts](#simplifications--removed-concepts)
  - [Summary of design decisions](#summary-of-design-decisions)

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

| Old method | New endpoint | Notes |
| - | - | - |
| `engine_newPayloadV{1..5}` | `POST /{fork}/payloads` | `parentBeaconBlockRoot` and `executionRequests` folded into the SSZ envelope; `expectedBlobVersionedHashes` removed; `INVALID_BLOCK_HASH` removed from the status enum |
| `engine_forkchoiceUpdatedV{1..4}` | `POST /{fork}/forkchoice` | one atomic call; carries forkchoice state, optional `payload_attributes`, and (Amsterdam+) optional `custody_columns` |
| `engine_getPayloadV{1..6}` | `GET /{fork}/payloads/{id}` | poll-style, same semantics as today |
| `engine_getPayloadBodiesByHashV{1,2}` | `POST /{fork}/bodies/hash` | `{fork}` selects the response *schema* (not the era of requested blocks); `POST` because hash lists are too large for URLs |
| `engine_getPayloadBodiesByRangeV{1,2}` | `GET /{fork}/bodies?from=...&count=...` | `{fork}` selects the response schema |
| `engine_getBlobsV1` | `POST /blobs/v1` | independently versioned; legacy version numbers carry forward |
| `engine_getBlobsV2` | `POST /blobs/v2` | all-or-nothing cell proofs |
| `engine_getBlobsV3` | `POST /blobs/v3` | partial-response cell proofs |
| `engine_getBlobsV4` | `POST /blobs/v4` | cell-range selection |
| `engine_getClientVersionV1` | `GET /identity` + `X-Engine-Client-Version` request header | unscoped |
| `engine_exchangeCapabilities` | `GET /capabilities` | unscoped |
| `engine_exchangeTransitionConfigurationV1` | *removed* | already deprecated since Cancun |

---

## Resource model (overview)

Hot-path endpoints are scoped under `/engine/v2/{fork}/...`. Diagnostic
endpoints are unscoped.

| Resource | Endpoint | Purpose |
| - | - | - |
| Payload | `POST /engine/v2/{fork}/payloads` | Submit a payload received from the CL gossip network for the EL to validate / import. Replaces `engine_newPayload`. |
| Payload | `GET /engine/v2/{fork}/payloads/{payloadId}` | Retrieve a built payload by id. Replaces `engine_getPayload`. CL polls when it wants a fresher snapshot. |
| Forkchoice | `POST /engine/v2/{fork}/forkchoice` | Atomic forkchoice update: update head/safe/finalized, optionally start a payload build, optionally update custody set. Replaces `engine_forkchoiceUpdated`. |
| Bodies | `POST /engine/v2/{fork}/bodies/hash` | Replaces `engine_getPayloadBodiesByHash`. Fork-scoped: `{fork}` selects the *response schema*, not the fork of the requested blocks. |
| Bodies | `GET /engine/v2/{fork}/bodies?from=N&count=M` | Replaces `engine_getPayloadBodiesByRange`. Fork-scoped on response shape. |
| Blob pool | `POST /engine/v2/blobs/v{1..4}` | Replaces `engine_getBlobsV{1..4}`. The `vN` segment carries forward the legacy version numbers; `/v4` is the Amsterdam cell-range variant, `/v1` is the original Cancun whole-blob variant, and intermediate revisions live alongside. ELs MUST serve at least the current-fork revision (`/v4` for Amsterdam) and MAY serve older revisions alongside. |
| Capabilities | `GET /engine/v2/capabilities` | Replaces `engine_exchangeCapabilities`. Unscoped; advertises supported forks, `/blobs/vN` revisions, and per-endpoint request-size limits. |
| Identity | `GET /engine/v2/identity` | Replaces `engine_getClientVersion`. Unscoped. |

Every hot-path body uses SSZ; every metadata endpoint uses JSON.

---

## Endpoints

### Payload submission

#### `POST /engine/v2/{fork}/payloads`

Replaces `engine_newPayloadV{1..5}`.

- **Request body:** SSZ-encoded `ExecutionPayloadEnvelope`, a container
  that bundles together everything that today travels alongside the
  payload as separate JSON-RPC params:

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
      status:           uint8        # VALID=1, INVALID=2, SYNCING=3, ACCEPTED=4
      latest_valid_hash: Optional[Hash32]
      validation_error: Optional[String]
  }
  ```

  `INVALID_BLOCK_HASH` is dropped (already supplanted by `INVALID`).
  `ACCEPTED` is **kept** — CLs rely on it during sync to acknowledge
  side-branch payloads that are well-formed but don't extend the
  canonical chain.

- **HTTP status:** `200 OK` for any of the four validation outcomes.
  Validation results are not transport errors.

### Forkchoice update

#### `POST /engine/v2/{fork}/forkchoice`

Replaces `engine_forkchoiceUpdatedV{1..4}`. This is the **single
atomic** call that updates the EL's forkchoice state, optionally
triggers a payload build, and (post-Amsterdam) optionally updates the
CL's custody set. Atomicity matters: the CL relies on the EL having
applied the new head before — and only if — the build is started, and
on the build being keyed against the freshly-applied head.

- **Request body:** SSZ-encoded `ForkchoiceUpdate`:

  ```
  ForkchoiceUpdate {
      forkchoice_state:    ForkchoiceState              # head / safe / finalized
      payload_attributes:  Optional[PayloadAttributes]  # if present, start a build on top of head
      custody_columns:     Optional[Bitvector[CELLS_PER_EXT_BLOB]]  # Amsterdam+, optional
  }
  ```

  All three fields are processed in one transaction: the EL MUST apply
  the forkchoice state, then (if `payload_attributes` is present and
  the new head is `VALID`) start the build, then (if `custody_columns`
  is present) update the custody set, all before returning. If the
  forkchoice update fails, no build is started and no custody change
  is applied.

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
  NOT recompute or validate its contents. This is a change from
  today's behavior where both sides derived an 8-byte hash over
  `(headBlockHash, payloadAttributes)`.

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

- **Stale-fork URL:** an FCU at `/engine/v2/{fork}/forkchoice`
  referencing a `head` from an earlier fork is **allowed**, *as long
  as `payload_attributes` is absent*. The CL needs to update head /
  safe / finalized across fork boundaries during sync and reorg
  recovery, and the URL fork has no bearing on which historical
  block can be referenced.

  If `payload_attributes` is present, the URL `{fork}` MUST match
  the fork that the new payload would belong to (i.e. the fork
  determined by `payload_attributes.timestamp`). Mismatch returns
  `400 unsupported-fork`. Building a payload is the only operation
  where the URL fork is load-bearing on shape, so it's the only one
  we strictly police.

- **Custody-set semantics** (Amsterdam+): the custody update runs
  independently of the forkchoice processing flow, matching the
  Amsterdam spec's "MUST run custody set update independently to the
  fork choice update". An execution-time custody-set error MUST NOT
  affect the `payload_status` returned for the forkchoice update.
  A `custody_columns` value, once accepted, remains in effect until
  the next `POST /forkchoice` whose body *also* contains a
  `custody_columns` field. FCUs that omit the field leave the
  custody set unchanged.

- **No body cap.** `POST /forkchoice` bodies are bounded by the SSZ
  schema's `MAX_*` constants (small for `ForkchoiceState` and
  `PayloadAttributes`, fixed for `custody_columns`). No additional
  HTTP-layer cap is imposed.

### Payload retrieval

#### `GET /engine/v2/{fork}/payloads/{payloadId}`

Replaces `engine_getPayloadV{1..6}`.

- **Response body:** SSZ-encoded `BuiltPayload`:

  ```
  BuiltPayload {
      payload:                ExecutionPayload
      block_value:            Uint256
      blobs_bundle:           BlobsBundle
      execution_requests:     List[Bytes, MAX_REQUESTS]
      should_override_builder: bool
  }
  ```
- **404** if `payloadId` is unknown or expired.

Polling semantics are unchanged from `engine_getPayload`: the CL calls
`GET /{fork}/payloads/{payloadId}` whenever it wants the latest
snapshot of the build. Each call returns the most recent version
available at the time of receipt; the EL MAY stop the build process
after serving a call. `payloadId` values are opaque server-assigned
tokens issued by `POST /forkchoice`.

**Token TTL.** A `payloadId` is valid for **at least 10 minutes**
after its issuing `POST /forkchoice` returns. After 10 minutes the
EL MAY garbage-collect the token and respond `404 unknown-payload`
to subsequent `GET`s. ELs MUST NOT recycle a token within its TTL
(no collisions); after expiry the token namespace is free to reuse.
A CL that needs a fresh `payloadId` after expiry simply issues a new
`POST /forkchoice` with the same attributes.

### Historical bodies

These endpoints are **fork-scoped on the response schema, not on the
era of the requested blocks**. The `{fork}` segment tells the EL which
`ExecutionPayloadBody` shape to use when serialising the response.
A CL that has just upgraded to the Amsterdam schema can ask for
`/amsterdam/bodies/hash` and receive `block_access_list` populated
for Amsterdam blocks and `[]` (the SSZ optional sentinel — see
[SSZ encoding conventions](#ssz-encoding-conventions)) for older
blocks; a CL still on Cancun asks `/cancun/bodies/hash` and
gets responses serialised against the Cancun container, never seeing
the trailing `block_access_list` field at all.

This is different from the `/payloads` and `/forkchoice` `{fork}`
segments, where the URL fork *must* match the timestamp of the
referenced block. For `/bodies` the URL fork is purely a schema
selector and the requester chooses freely.

The blob endpoint takes yet another approach: it carries a `/vN`
revision instead of a `{fork}` segment, because blob protocol
evolution has historically not aligned with fork activations. See
the [Blob pool](#blob-pool) section.

#### `POST /engine/v2/{fork}/bodies/hash`

Replaces `engine_getPayloadBodiesByHashV{1,2}`. Uses `POST` so that
large hash lists travel in the request body rather than the URL.

- **Request body:** SSZ-encoded `List[Hash32, MAX_BODIES_REQUEST]`.

#### `GET /engine/v2/{fork}/bodies?from=N&count=M`

Replaces `engine_getPayloadBodiesByRangeV{1,2}`. Range fits comfortably
in the URL.

- **Response body** (both endpoints): SSZ-encoded
  `List[BodyEntry, MAX_BODIES_REQUEST]`. Each `BodyEntry` carries an
  `available: boolean` flag (false for unavailable / pruned blocks,
  matching today's `null` semantics) and an `ExecutionPayloadBody`
  serialised against the **`{fork}` schema from the URL**. Fields
  introduced in `{fork}` or earlier are present (with `Optional[T]`
  set to `None` for blocks predating the field's introduction); fields
  introduced in forks newer than `{fork}` are absent from the
  container entirely. See
  [SSZ encoding conventions](#ssz-encoding-conventions).

### Blob pool

The blob endpoint is **independently versioned**: blobs are looked up
by versioned hash (not by fork), so the `{fork}` URL segment doesn't
help. But the blob protocol *has* evolved on its own clock — four
distinct semantics across two forks (V1 single proof in Cancun, V2
cell proofs in Osaka, V3 partial responses, V4 cell-range selection
in Amsterdam). The new spec carries those legacy version numbers
forward onto the URL: `engine_getBlobsVN` becomes `POST /blobs/vN`.
ELs **MUST** serve at least the revision matching their current fork
(`/blobs/v4` for Amsterdam) and **MAY** serve any subset of older
revisions alongside; `GET /capabilities` advertises the actual list.

This is a different versioning axis from the fork-scoped endpoints
(`/{fork}/payloads`, `/{fork}/forkchoice`, `/{fork}/bodies`). Those
track *consensus protocol* changes coupled to fork activations.
`/blobs/vN` tracks *engine-API blob protocol* changes that have
historically not aligned with fork activations.

All revisions use `POST` so that 128 versioned hashes (8 KiB hex)
don't have to fit in the URL. All revisions return SSZ
`Optional[List[BlobEntry, MAX_BLOBS_REQUEST]]`, where the outer
`Optional` is the "all-or-nothing"/syncing channel (`None` =
"cannot serve this request, retry later or fall back") and each
`BlobEntry` carries an `available: boolean` per-entry flag for
per-blob misses on revisions that support partial responses.
Revision-specific contents live inside `BlobEntry.contents`.

#### `POST /engine/v2/blobs/v1`

Replaces `engine_getBlobsV1` (Cancun, single-proof whole-blob).

- **Request body:** SSZ `List[VersionedHash, MAX_BLOBS_REQUEST]`.
- **Response `BlobEntry.contents`:** `BlobAndProofV1 { blob, proof }`
  (one blob, one 48-byte KZG proof).
- Partial responses supported: missing blobs surface as
  `available=false` per entry. The outer `Optional` returns `None`
  only if the EL cannot serve the request at all (e.g. syncing).

#### `POST /engine/v2/blobs/v2`

Replaces `engine_getBlobsV2` (Osaka, all-or-nothing cell proofs).

- **Request body:** SSZ `List[VersionedHash, MAX_BLOBS_REQUEST]`.
- **Response `BlobEntry.contents`:** `BlobAndProofV2 { blob, proofs }`
  (one blob plus `CELLS_PER_EXT_BLOB` cell proofs).
- **All-or-nothing:** if any requested blob is missing, the outer
  `Optional[List[...]]` returns `None`. Otherwise all entries have
  `available=true`. This matches today's V2 semantics.

#### `POST /engine/v2/blobs/v3`

Replaces `engine_getBlobsV3` (Osaka, partial responses with cell
proofs).

- **Request body:** same as `/v2`.
- **Response:** same `BlobEntry.contents` shape as `/v2`, but missing
  blobs surface as `available=false` per entry rather than collapsing
  the whole response to `None`. The outer `Optional` returns `None`
  only when the EL cannot serve the request at all.

#### `POST /engine/v2/blobs/v4`

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

#### `GET /engine/v2/capabilities`

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
    "bodies.max_count":           128,
    "blobs.max_versioned_hashes": 128,
    "payload.max_bytes":          67108864
  }
}
```

The `independently_versioned` map advertises endpoints whose URL
carries an explicit `/vN` revision. ELs MAY support multiple
revisions concurrently (e.g. `["v1", "v2"]`); CLs pick whichever they
implement.

#### `GET /engine/v2/identity`

Returns JSON `ClientVersion[]` (same shape as today's
`engine_getClientVersionV1`). The CL identifies itself with a
`X-Engine-Client-Version` header on every request, removing the
mutual-exchange handshake.

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

We deliberately drop the other RFC 7807 fields:

- `title` would just duplicate `type` (RFC 7807 says it SHOULD NOT
  vary between occurrences of the same `type`); CLs can render their
  own from a static `type → title` map.
- `status` duplicates the HTTP status line.
- `instance` adds a per-request URI; operators get correlation from
  logs already.

There is **no** legacy `engine_code` extension. CLs migrating from
the JSON-RPC API map old codes to new `type` strings via the table
below; after migration the codes are gone.

| HTTP status | `type` | Old JSON-RPC code | When |
| - | - | - | - |
| 400 Bad Request | `/engine-api/errors/parse-error` | -32700 | Body is not valid JSON / SSZ |
| 400 Bad Request | `/engine-api/errors/invalid-request` | -32600 | Request shape is wrong (missing required field, etc.) |
| 400 Bad Request | `/engine-api/errors/ssz-decode-error` | (new) | SSZ decode failed; canned error, no `detail` |
| 400 Bad Request | `/engine-api/errors/unsupported-fork` | -38005 | URL `{fork}` is not supported by this EL |
| 404 Not Found | `/engine-api/errors/method-not-found` | -32601 | URL does not match any endpoint |
| 404 Not Found | `/engine-api/errors/unknown-payload` | -38001 | `payloadId` does not exist |
| 409 Conflict | `/engine-api/errors/invalid-forkchoice` | -38002 | Forkchoice state is inconsistent (e.g. finalized not ancestor of head) |
| 409 Conflict | `/engine-api/errors/reorg-too-deep` | -38006 | Reorg depth exceeds the EL's limit |
| 413 Payload Too Large | `/engine-api/errors/request-too-large` | -38004 | Body exceeds an advertised `limits.*` value |
| 422 Unprocessable Entity | `/engine-api/errors/invalid-body` | -32602 | Body decoded fine but has invalid values |
| 422 Unprocessable Entity | `/engine-api/errors/invalid-attributes` | -38003 | `payload_attributes` validation failed |
| 500 Internal Server Error | `/engine-api/errors/internal` | -32603 / -32000 | Unrecoverable server error; `detail` carries the message |

`type` URIs are written as **relative references** rooted at
`/engine-api/errors/...`. RFC 7807 allows relative URIs, and the
short form keeps error bodies small without losing identifier
stability. CLs MUST treat them as opaque strings — they MUST NOT
attempt to dereference them.

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

1. **Major (`/v2`)** — bumped only for breaking transport changes
   (e.g. moving away from REST, swapping SSZ for something else).
2. **Per-fork body schema** — selected via the `{fork}` URL segment
   on hot-path endpoints (`/{fork}/payloads`, `/{fork}/forkchoice`,
   `/{fork}/bodies`). Tracks consensus-protocol changes that ride
   along with fork activations.
3. **Per-endpoint revisions** — selected via a `/vN` URL segment on
   endpoints whose protocol evolves independently of the fork
   schedule (currently just `/blobs/vN`). Tracks engine-API protocol
   changes that don't align with fork activations.

The server advertises which forks and which `/vN` revisions it
understands via `GET /engine/v2/capabilities`.

`engine_exchangeCapabilities` is **removed**. Instead the server lists
its supported fork schemas and endpoint set in a single JSON document
at `/engine/v2/capabilities`.

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

- **Protocol:** HTTP/2 is **required**. Both TCP and IPC transports
  use **h2c** (HTTP/2 cleartext); JWT-on-every-request provides
  authentication, so TLS termination is left to a reverse proxy if
  the operator wants it. HTTP/2 multiplexing means a single CL→EL
  connection can carry the full request mix (forkchoice, payload
  submission, blob fetches, body fetches) without head-of-line
  blocking. HTTP/1.1 is not supported.
- **Default port:** `8551`, shared with the legacy JSON-RPC engine API.
  The two surfaces are distinguished by path: legacy JSON-RPC remains
  at `/` (and accepts JSON-RPC method calls), the new API lives under
  `/engine/v2/...`. The same JWT secret authenticates both.
- **Base path:** `/engine/v2/{fork}/...`. The `/v2` segment is the
  major-protocol version; the `{fork}` segment selects the fork-scoped
  body schema (`paris`, `shanghai`, `cancun`, `prague`, `osaka`,
  `amsterdam`, …). Adding a fork = adding one path prefix and one set
  of SSZ schemas. See [Versioning](#versioning-model).
- **Trailing slashes are forbidden.** `/engine/v2/payloads` is the
  canonical form; `/engine/v2/payloads/` MUST return
  `404 method-not-found`. No automatic redirect.
- **Request body encoding:** `application/octet-stream` carrying SSZ
  bytes for hot-path endpoints. JSON for diagnostic / metadata
  endpoints (capabilities, identity, error bodies).
- **Response body encoding:** SSZ for hot-path data, JSON
  (`application/json`) for diagnostics and error bodies.
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

### Why fork-in-URL instead of method versioning?

Today every change of a single field bumps the method version
(`engine_newPayloadV1..V5`). The new API puts the fork in the URL:

```
POST /engine/v2/amsterdam/payloads
Content-Type: application/octet-stream
Authorization: Bearer <JWT>

<SSZ-encoded ExecutionPayloadEnvelope>
```

The EL routes by fork segment, parses the body according to that fork's
SSZ schema, and returns a fork-shaped response. Adding a fork = adding
one path prefix and one set of SSZ schemas. URLs stay greppable and
discoverable in logs.

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

### Cross-fork response containers

Endpoints that return data spanning multiple block-eras come in two
flavours:

1. **Fork-scoped** (e.g. `/bodies`): the URL `{fork}` selects the
   container schema. Within that schema, fields that didn't exist in
   earlier block-eras are `Optional[T]` (= `[]` for those blocks).
   The outer entry carries an explicit `available` flag so
   "pruned / unavailable" stays distinct from "field-not-applicable":

   ```
   # /amsterdam/bodies/hash response
   BodyEntry {
       available: boolean
       body:      ExecutionPayloadBody
   }

   ExecutionPayloadBody {
       transactions:       List[Transaction, MAX_TXS]
       withdrawals:        Optional[List[Withdrawal, MAX_WITHDRAWALS]]  # [] pre-Shanghai
       block_access_list:  Optional[ByteList[MAX_BAL_BYTES]]            # [] pre-Amsterdam or if pruned
   }
   ```

   A CL on the Cancun schema calls `/cancun/bodies/hash` and
   receives the Cancun container (no `block_access_list` field at
   all). Old CLs never see schemas they don't know.

2. **Independently versioned** (e.g. `/blobs/vN`): each revision is
   its own container, no nullable optionals across revisions. Old
   CLs keep using `/blobs/v1`; new shapes ship as `/blobs/vN+1`
   alongside.

Per-entry fork tags (a `Union` of fork-shaped variants) were
rejected: every fork would bump the union and break old decoders.

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

  This matches today's [`common.md`](./common.md) "Message ordering"
  guarantee in spirit; it makes explicit that h2 multiplexing does
  not relax it. There is **no sequence number on the wire** — the
  protocol stays simple and CL bugs that break ordering are CL bugs.

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
   URL (`/engine/v2/{fork}/...`) so a single endpoint accepts whatever
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
  field. With a REST endpoint, the fork is part of the URL
  (`/engine/v2/amsterdam/payloads`) and the body schema is selected by
  routing, not by method-name suffix.
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
  `eth_syncing`) is **not** mirrored under `/engine/v2/...`. CLs that
  need state / log access continue to call them via the legacy
  JSON-RPC root.

#### Transport

- **HTTP/2 required**, h2c (cleartext) for both TCP and IPC. No
  HTTP/1.1 fallback. JWT-on-every-request authenticates; TLS
  termination is left to a reverse proxy.
- **IPC** is h2c over UNIX socket — same paths and headers as TCP,
  single code path.
- **Default port `8551`**, shared with the legacy JSON-RPC API
  (distinguished by path).
- **Trailing slashes are forbidden** — return `404 method-not-found`.
- **Flow-control:** SHOULD set `INITIAL_WINDOW_SIZE` ≥ 1 MiB.
  `MAX_FRAME_SIZE` and `MAX_HEADER_LIST_SIZE` use HTTP/2 defaults.
- **Connection lifecycle:** CLs MAY open fresh h2 connections per
  request or reuse a long-lived connection.
- **Compression:** `zstd` and `gzip` MAY be implemented. CLs MUST
  tolerate uncompressed responses.

#### Versioning

- **Fork-scoped endpoints:** `/{fork}/payloads`, `/{fork}/forkchoice`,
  `/{fork}/bodies`. Fork in the URL, no `Eth-Consensus-Version`
  header.
- **Independently versioned endpoints:** `/blobs/vN`. Legacy
  `engine_getBlobsVN` numbers carry forward onto the URL. ELs MUST
  serve at least the revision matching their current fork
  (`/blobs/v4` for Amsterdam) and MAY serve older revisions
  alongside. Future blob-shape changes ship as `/blobs/v5`, `/v6`,
  etc.
- **Unscoped endpoints:** `/capabilities`, `/identity`.
- **Major version `/v2`** is bumped only for breaking transport
  changes (e.g. dropping REST or SSZ).

#### Encoding

- **Hot-path bodies use SSZ.** Diagnostic / metadata endpoints
  (`/capabilities`, `/identity`, error bodies) use JSON.
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
  fork-scoped (`/bodies`) uses the URL `{fork}` to pick a schema,
  with `Optional[T]` for fields absent in pre-fork blocks;
  independently versioned (`/blobs/vN`) gives each revision its own
  dedicated container. Both wrap their entries in
  `BodyEntry { available, body }` / `BlobEntry { available, contents }`
  with an outer `Optional[List[...]]` for the syncing /
  all-or-nothing channel. Per-entry fork tags were rejected.

#### Error model

- **RFC 7807 with two fields:** `type` (required, relative URI rooted
  at `/engine-api/errors/...`) and `detail` (optional). Drop `title`,
  `status`, `instance`, `engine_code`.
- **CLs MUST NOT dereference `type`** — opaque strings.
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

#### Forkchoice update (`POST /{fork}/forkchoice`)

- **Single atomic call** carrying forkchoice state, optional
  `payload_attributes`, and optional `custody_columns`.
- **Skip-allowed semantics:** EL MAY skip applying state when the
  new `head` is a `VALID` ancestor of the latest finalized block,
  guarding against malformed CL FCUs.
- **Stale-fork URL** is allowed when `payload_attributes` is absent;
  with `payload_attributes` present, URL `{fork}` MUST match the
  timestamp's fork (otherwise `400 unsupported-fork`).
- **No HTTP-layer body cap** beyond SSZ `MAX_*` constants.
- **Custody-set updates** run independently of the forkchoice flow;
  custody errors do not affect `payload_status`.
- **Custody-set lifetime:** set until the next FCU that includes a
  `custody_columns` field. FCUs that omit it leave the set unchanged.

#### Payload submission (`POST /{fork}/payloads`)

- **`expectedBlobVersionedHashes` removed.** EL recomputes from
  `payload.transactions`; block-hash check covers transactions.
- **`INVALID_BLOCK_HASH` removed** from the status enum.
- **`ACCEPTED` kept** — CLs use it during sync.
- **Transaction min-length** ("at least 1 byte") remains a
  receiver-side validation rule, not an SSZ schema invariant.

#### Payload retrieval (`GET /{fork}/payloads/{payloadId}`)

- **Poll-only**, same semantics as today's `engine_getPayload`. No
  SSE / long-poll.
- **`payload_id` is an opaque server-assigned token** issued by
  `POST /forkchoice`. CLs MUST NOT recompute or validate it.
- **`payload_id` TTL ≥ 10 minutes.** After expiry the EL MAY GC and
  reuse the token namespace; within the TTL no collisions.
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
