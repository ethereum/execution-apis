# Engine API v2 -- SSZ Container Sketches (Amsterdam)

> **Status:** Sketch. This is a working draft of the concrete SSZ
> container definitions referenced by the Engine API v2 spec
> ([refactor.md](./refactor.md)). Field types, names, and `MAX_*`
> constants are placeholders and need a final review before
> publication.
>
> All conventions in this document follow
> [refactor.md § SSZ encoding conventions](./refactor.md#ssz-encoding-conventions):
>
> - `Optional[T]` ≡ `List[T, 1]` (length 0 = absent, length 1 = present)
> - `String` ≡ `List[byte, MAX_ERROR_BYTES]`, `MAX_ERROR_BYTES = 1024`
> - `ByteList[N]` ≡ `List[byte, N]`
> - `ByteVector[N]` is fixed-size, `Bytes32` etc. are aliases
> - All uints are little-endian

---

## Table of contents

- [Primitive aliases](#primitive-aliases)
- [`MAX_*` constants](#max-constants)
- [Shared structures](#shared-structures)
  - [`Withdrawal`](#withdrawal)
  - [`ExecutionPayload` (Amsterdam)](#executionpayload-amsterdam)
  - [`PayloadAttributes` (Amsterdam)](#payloadattributes-amsterdam)
  - [`ForkchoiceState`](#forkchoicestate)
  - [`PayloadStatus`](#payloadstatus)
- [Endpoint containers](#endpoint-containers)
  - [`POST /amsterdam/payloads`](#post-amsterdampayloads)
  - [`POST /amsterdam/forkchoice`](#post-amsterdamforkchoice)
  - [`GET /amsterdam/payloads/{payloadId}`](#get-amsterdampayloadspayloadid)
  - [`POST /amsterdam/bodies/hash` and `GET /amsterdam/bodies?...`](#post-amsterdambodieshash-and-get-amsterdambodies)
  - [`POST /blobs/v1`](#post-blobsv1)
  - [`POST /blobs/v2`](#post-blobsv2)
  - [`POST /blobs/v3`](#post-blobsv3)
  - [`POST /blobs/v4`](#post-blobsv4)
- [Open sketch questions](#open-sketch-questions)

---

## Primitive aliases

| Alias | SSZ type | Notes |
| - | - | - |
| `Hash32` | `ByteVector[32]` | block / payload hashes |
| `Root` | `ByteVector[32]` | beacon-block roots, merkle roots |
| `Address` | `ByteVector[20]` | execution-layer 160-bit address |
| `Bloom` | `ByteVector[256]` | logs bloom filter |
| `VersionedHash` | `ByteVector[32]` | EIP-4844 versioned blob hash |
| `Bytes8` | `ByteVector[8]` | `payload_id` |
| `Bytes32` | `ByteVector[32]` | `prevRandao`, generic 32-byte values |
| `Bytes48` | `ByteVector[48]` | KZG commitments and proofs |
| `Uint64` | `uint64` | LE on the wire |
| `Uint256` | `uint256` | LE on the wire (`block_value`, `base_fee_per_gas`) |
| `Boolean` | `bool` | one byte, `0x00` / `0x01` |

## `MAX_*` constants

These are sketch values — final values come from a follow-up that
matches the consensus-specs `Amsterdam` preset. They are listed here
for completeness so readers can size the on-wire bounds.

| Constant | Sketch value | Where it's used |
| - | - | - |
| `MAX_TXS_PER_PAYLOAD` | `1048576` | `ExecutionPayload.transactions` |
| `MAX_BYTES_PER_TX` | `1073741824` | element bound inside `transactions` |
| `MAX_WITHDRAWALS_PER_PAYLOAD` | `16` | `ExecutionPayload.withdrawals`, `PayloadAttributes.withdrawals` |
| `MAX_EXTRA_DATA_BYTES` | `32` | `ExecutionPayload.extra_data` |
| `MAX_BAL_BYTES` | TBD (EIP-7928) | `ExecutionPayload.block_access_list` |
| `MAX_EXECUTION_REQUESTS_PER_PAYLOAD` | TBD (EIP-7685) | `ExecutionPayloadEnvelope.execution_requests` |
| `MAX_BYTES_PER_EXECUTION_REQUEST` | TBD | element bound inside `execution_requests` |
| `MAX_VERSIONED_HASHES_PER_REQUEST` | `128` | `BlobsRequest.versioned_hashes` |
| `MAX_BODIES_REQUEST` | `128` | bodies request and response lists |
| `MAX_BLOBS_REQUEST` | `128` | blobs request and response lists |
| `MAX_BLOBS_PER_PAYLOAD` | `MAX_VERSIONED_HASHES_PER_REQUEST` | `BlobsBundle.commitments`, `.blobs` |
| `CELLS_PER_EXT_BLOB` | `128` (EIP-7594) | cell-proof and custody bitvectors |
| `BYTES_PER_BLOB` | `131072` | one blob (`4096 * 32`) |
| `MAX_ERROR_BYTES` | `1024` | `validation_error`, JSON error `detail` |

---

## Shared structures

These containers are used by multiple endpoints. They map directly
onto today's JSON-RPC structures with field renaming
(`camelCase` → `snake_case`) and the type changes that follow from
the SSZ encoding conventions.

### `Withdrawal`

Same as the consensus-specs `Withdrawal` container. The `amount`
field is now natively LE in SSZ; the `withdrawals.amount` LE-vs-BE
note in shanghai.md goes away.

```
Withdrawal {
    index:           Uint64
    validator_index: Uint64
    address:         Address
    amount:          Uint64    # gwei
}
```

### `ExecutionPayload` (Amsterdam)

Reflects today's [`ExecutionPayloadV4`](./amsterdam.md#executionpayloadv4).
`block_access_list` is a fixed field for Amsterdam (no `Optional[T]`
here — that's only used for cross-fork `BodyEntry` responses; the
Amsterdam `ExecutionPayload` always carries the BAL).

```
ExecutionPayload {
    parent_hash:        Hash32
    fee_recipient:      Address
    state_root:         Hash32
    receipts_root:      Hash32
    logs_bloom:         Bloom
    prev_randao:        Bytes32
    block_number:       Uint64
    gas_limit:          Uint64
    gas_used:           Uint64
    timestamp:          Uint64
    extra_data:         ByteList[MAX_EXTRA_DATA_BYTES]
    base_fee_per_gas:   Uint256
    block_hash:         Hash32
    transactions:       List[ByteList[MAX_BYTES_PER_TX], MAX_TXS_PER_PAYLOAD]
    withdrawals:        List[Withdrawal, MAX_WITHDRAWALS_PER_PAYLOAD]
    blob_gas_used:      Uint64
    excess_blob_gas:    Uint64
    block_access_list:  ByteList[MAX_BAL_BYTES]   # RLP-encoded EIP-7928 BAL
    slot_number:        Uint64
}
```

Notes:

- `block_access_list` is RLP-encoded inside an SSZ `ByteList`. EIP-7928's
  encoding is RLP and we don't try to re-encode it as SSZ — the EL
  treats it as opaque bytes for transport, decodes it as RLP for
  validation. Same pattern as `transactions`.
- `transactions` elements remain RLP-encoded `TransactionType ||
  TransactionPayload` per EIP-2718. Receiver-side rule: each element
  MUST be ≥ 1 byte (see refactor.md § Payload submission).

### `PayloadAttributes` (Amsterdam)

Reflects today's [`PayloadAttributesV4`](./amsterdam.md#payloadattributesv4).

```
PayloadAttributes {
    timestamp:                Uint64
    prev_randao:              Bytes32
    suggested_fee_recipient:  Address
    withdrawals:              List[Withdrawal, MAX_WITHDRAWALS_PER_PAYLOAD]
    parent_beacon_block_root: Root
    slot_number:              Uint64
}
```

### `ForkchoiceState`

Same fields as today's [`ForkchoiceStateV1`](./paris.md#forkchoicestatev1).

```
ForkchoiceState {
    head_block_hash:      Hash32
    safe_block_hash:      Hash32
    finalized_block_hash: Hash32
}
```

### `PayloadStatus`

Used by `POST /payloads` (full enum) and `POST /forkchoice`
(restricted enum — `ACCEPTED` not allowed).

```
PayloadStatus {
    status:            uint8                    # see enum below
    latest_valid_hash: Optional[Hash32]
    validation_error:  Optional[String]
}
```

Status enum:

| Value | Name | Used by |
| - | - | - |
| `1` | `VALID` | both |
| `2` | `INVALID` | both |
| `3` | `SYNCING` | both |
| `4` | `ACCEPTED` | `POST /payloads` only |

`INVALID_BLOCK_HASH` is removed (already supplanted by `INVALID`).
`POST /forkchoice` MUST return `1`/`2`/`3` only; CLs MUST treat a
`4` from `/forkchoice` as a protocol error.

`Optional[String]` resolves to `List[List[byte, MAX_ERROR_BYTES], 1]`.

---

## Endpoint containers

### `POST /amsterdam/payloads`

Replaces `engine_newPayloadV5`.

#### Request

```
ExecutionPayloadEnvelope {
    payload:                  ExecutionPayload
    parent_beacon_block_root: Root
    execution_requests:       List[ByteList[MAX_BYTES_PER_EXECUTION_REQUEST], MAX_EXECUTION_REQUESTS_PER_PAYLOAD]
}
```

`expected_blob_versioned_hashes` is removed (the EL recomputes it
from `payload.transactions`).

#### Response

`PayloadStatus` (full enum, `1`/`2`/`3`/`4`).

### `POST /amsterdam/forkchoice`

Replaces `engine_forkchoiceUpdatedV4`.

#### Request

```
ForkchoiceUpdate {
    forkchoice_state:    ForkchoiceState
    payload_attributes:  Optional[PayloadAttributes]
    custody_columns:     Optional[Bitvector[CELLS_PER_EXT_BLOB]]
}
```

#### Response

```
ForkchoiceUpdateResponse {
    payload_status: PayloadStatus      # restricted: VALID | INVALID | SYNCING
    payload_id:     Optional[Bytes8]
}
```

### `GET /amsterdam/payloads/{payloadId}`

Replaces `engine_getPayloadV6`.

#### Response

```
BuiltPayload {
    payload:                 ExecutionPayload
    block_value:             Uint256
    blobs_bundle:            BlobsBundleV2          # see consensus-specs Osaka
    execution_requests:      List[ByteList[MAX_BYTES_PER_EXECUTION_REQUEST], MAX_EXECUTION_REQUESTS_PER_PAYLOAD]
    should_override_builder: Boolean
}

BlobsBundleV2 {
    commitments: List[Bytes48, MAX_BLOBS_PER_PAYLOAD]
    proofs:      List[Bytes48, MAX_BLOBS_PER_PAYLOAD * CELLS_PER_EXT_BLOB]
    blobs:       List[ByteVector[BYTES_PER_BLOB], MAX_BLOBS_PER_PAYLOAD]
}
```

`commitments` and `blobs` MUST have equal length; `proofs` MUST
have length `len(blobs) * CELLS_PER_EXT_BLOB` (mirrors the
`engine_getPayloadV5` rule from osaka.md).

### `POST /amsterdam/bodies/hash` and `GET /amsterdam/bodies?...`

Replace `engine_getPayloadBodiesByHashV2` and
`engine_getPayloadBodiesByRangeV2`. Both return the same response
container.

#### Request — `/bodies/hash`

```
BodiesByHashRequest {
    block_hashes: List[Hash32, MAX_BODIES_REQUEST]
}
```

#### Request — `/bodies?from=N&count=M`

URL query parameters; no SSZ request body.

#### Response

```
BodiesResponse {
    entries: List[BodyEntry, MAX_BODIES_REQUEST]
}

BodyEntry {
    available: Boolean
    body:      ExecutionPayloadBody
}

# /amsterdam/bodies/... uses this Amsterdam-fork ExecutionPayloadBody
ExecutionPayloadBody {
    transactions:      List[ByteList[MAX_BYTES_PER_TX], MAX_TXS_PER_PAYLOAD]
    withdrawals:       Optional[List[Withdrawal, MAX_WITHDRAWALS_PER_PAYLOAD]]  # [] pre-Shanghai
    block_access_list: Optional[ByteList[MAX_BAL_BYTES]]                        # [] pre-Amsterdam or pruned
}
```

A CL on the Cancun schema would call `/cancun/bodies/...` and receive
a Cancun-shaped `ExecutionPayloadBody` (no `block_access_list` field
at all). The Cancun-fork variant is sketched here for clarity:

```
# /cancun/bodies/... ExecutionPayloadBody (for reference)
ExecutionPayloadBody {
    transactions: List[ByteList[MAX_BYTES_PER_TX], MAX_TXS_PER_PAYLOAD]
    withdrawals:  Optional[List[Withdrawal, MAX_WITHDRAWALS_PER_PAYLOAD]]   # [] pre-Shanghai
}
```

### `POST /blobs/v1`

Replaces `engine_getBlobsV1` (Cancun whole-blob).

#### Request

```
BlobsV1Request {
    versioned_hashes: List[VersionedHash, MAX_BLOBS_REQUEST]
}
```

#### Response

```
BlobsV1Response = Optional[List[BlobV1Entry, MAX_BLOBS_REQUEST]]

BlobV1Entry {
    available: Boolean
    contents:  BlobAndProofV1
}

BlobAndProofV1 {
    blob:  ByteVector[BYTES_PER_BLOB]
    proof: Bytes48
}
```

When `available == false`, `contents` carries zero-valued bytes (a
`BYTES_PER_BLOB`-byte zero blob and a 48-byte zero proof). The outer
`Optional` returns `[]` when the EL cannot serve the request at all.

### `POST /blobs/v2`

Replaces `engine_getBlobsV2` (Osaka all-or-nothing cell proofs).

#### Request — same as `/v1`

```
BlobsV2Request {
    versioned_hashes: List[VersionedHash, MAX_BLOBS_REQUEST]
}
```

#### Response

```
BlobsV2Response = Optional[List[BlobV2Entry, MAX_BLOBS_REQUEST]]

BlobV2Entry {
    available: Boolean      # always true for /v2 (all-or-nothing); included for shape symmetry
    contents:  BlobAndProofV2
}

BlobAndProofV2 {
    blob:   ByteVector[BYTES_PER_BLOB]
    proofs: List[Bytes48, CELLS_PER_EXT_BLOB]
}
```

All-or-nothing: if any requested blob is missing, the outer
`Optional` returns `[]` and no per-entry data is sent. CLs that need
partial responses use `/v3`.

### `POST /blobs/v3`

Replaces `engine_getBlobsV3` (Osaka partial responses with cell
proofs).

#### Request — same as `/v2`

#### Response

Same shape as `/v2` (`BlobV2Entry` reused), but missing blobs
surface as `available=false` per entry rather than collapsing the
whole response to `[]`. Outer `Optional` returns `[]` only when the
EL cannot serve the request at all (e.g. syncing).

```
BlobsV3Response = Optional[List[BlobV2Entry, MAX_BLOBS_REQUEST]]
```

### `POST /blobs/v4`

Replaces `engine_getBlobsV4` (Amsterdam cell-range selection).

#### Request

```
BlobsV4Request {
    versioned_hashes: List[VersionedHash, MAX_BLOBS_REQUEST]
    indices_bitarray: Bitvector[CELLS_PER_EXT_BLOB]
}
```

#### Response

```
BlobsV4Response = Optional[List[BlobV4Entry, MAX_BLOBS_REQUEST]]

BlobV4Entry {
    available: Boolean
    contents:  BlobCellsAndProofs
}

BlobCellsAndProofs {
    blob_cells: List[Optional[ByteVector[BYTES_PER_CELL]], CELLS_PER_EXT_BLOB]
    proofs:    List[Optional[Bytes48], CELLS_PER_EXT_BLOB]
}
```

Per the Amsterdam spec: only the indices set in the request's
`indices_bitarray` carry a value; all other indices are `[]`. Within
the requested indices, individual missing cells are also `[]`, and
the corresponding `proofs` entry MUST also be `[]` (`null` in the
old spec).

`BYTES_PER_CELL` = `BYTES_PER_BLOB / CELLS_PER_EXT_BLOB` = `1024`
(EIP-7594).

---

## Open sketch questions

These are the items left to decide before promoting this sketch to
the canonical Amsterdam SSZ schema:

1. **`MAX_*` placeholder values.** Several constants above are
   `TBD` or sketch-only. They need to be pinned to the
   consensus-specs `Amsterdam` preset values once those land.
2. **`MAX_BAL_BYTES`.** EIP-7928 defines the BAL but doesn't yet
   pin a numeric upper bound that's friendly for SSZ. We need a
   concrete number; otherwise the SSZ schema can't round-trip.
3. **`Bitvector` SSZ encoding for `indices_bitarray` and
   `custody_columns`.** Both are `Bitvector[CELLS_PER_EXT_BLOB]`
   = `Bitvector[128]` = 16 bytes packed. Double-check that's the
   reading the Amsterdam spec wants (it currently describes it as
   "16 bytes interpreted as a bitarray").
4. **`should_override_builder` typing.** SSZ has `bool` but it's
   a 1-byte field. Keeping it inside `BuiltPayload` (rather than
   moving to a header) was the [refactor.md](./refactor.md)
   decision; this sketch follows that.
5. **`PayloadStatus` enum encoding.** A `uint8` with sentinel
   values matches the JSON-RPC enum; SSZ has no native enum type
   so this is the cleanest mapping. Alternative: `Container { ... }`
   wrapping a `uint8`. Open for discussion.
6. **`ExecutionPayloadBody` shared definition.** Today every fork
   redefines `ExecutionPayloadBody` from scratch. The new spec
   would benefit from a small set of fork-named containers
   (`ExecutionPayloadBodyParis`, `ExecutionPayloadBodyShanghai`,
   `ExecutionPayloadBodyAmsterdam`, …) with the URL `{fork}`
   selecting which one. Not worked out here.
7. **Naming convention.** The legacy spec used `camelCase`; this
   sketch uses `snake_case` to match consensus-specs. Worth
   confirming.
8. **`ByteVector[BYTES_PER_BLOB]` vs `ByteList[BYTES_PER_BLOB]`.**
   A blob is fixed-size (131072 bytes), so `ByteVector` is the
   correct typing. Verify against consensus-specs to keep
   alignment.
