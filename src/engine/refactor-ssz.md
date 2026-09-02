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
- [Per-fork container catalogue](#per-fork-container-catalogue)
  - [`ExecutionPayload` per fork](#executionpayload-per-fork)
  - [`PayloadAttributes` per fork](#payloadattributes-per-fork)
  - [`ExecutionPayloadBody` per fork](#executionpayloadbody-per-fork)
  - [`BlobsBundle` per revision](#blobsbundle-per-revision)
  - [`BuiltPayload` per fork](#builtpayload-per-fork)
  - [`ExecutionPayloadEnvelope` per fork](#executionpayloadenvelope-per-fork)
  - [`ForkchoiceUpdate` per fork](#forkchoiceupdate-per-fork)
  - [`BlobAndProof` per revision](#blobandproof-per-revision)
  - [Identification & capabilities](#identification--capabilities)
- [Endpoint containers](#endpoint-containers)
  - [`POST /payloads`](#post-payloads)
  - [`POST /forkchoice`](#post-forkchoice)
  - [`GET /payloads/{payloadId}`](#get-payloadspayloadid)
  - [`POST /bodies/hash` and `GET /bodies?...`](#post-bodieshash-and-get-bodies)
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

| Constant | Value | Source |
| - | - | - |
| `MAX_BYTES_PER_TX` | `2**30` (1,073,741,824) | [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844) |
| `MAX_TXS_PER_PAYLOAD` | `2**20` (1,048,576) | [Bellatrix](https://github.com/ethereum/consensus-specs/blob/dev/specs/bellatrix/beacon-chain.md) |
| `MAX_WITHDRAWALS_PER_PAYLOAD` | `2**4` (16) | [Capella](https://github.com/ethereum/consensus-specs/blob/dev/specs/capella/beacon-chain.md) |
| `BYTES_PER_LOGS_BLOOM` | `256` | [Bellatrix](https://github.com/ethereum/consensus-specs/blob/dev/specs/bellatrix/beacon-chain.md) |
| `MAX_EXTRA_DATA_BYTES` | `2**5` (32) | [Bellatrix](https://github.com/ethereum/consensus-specs/blob/dev/specs/bellatrix/beacon-chain.md) |
| `MAX_BLOB_COMMITMENTS_PER_BLOCK` | `2**12` (4,096) | [Deneb](https://github.com/ethereum/consensus-specs/blob/dev/specs/deneb/beacon-chain.md) |
| `FIELD_ELEMENTS_PER_BLOB` | `4096` | [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844) |
| `BYTES_PER_FIELD_ELEMENT` | `32` | [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844) |
| `BYTES_PER_BLOB` | `FIELD_ELEMENTS_PER_BLOB * BYTES_PER_FIELD_ELEMENT` (131,072) | derived |
| `CELLS_PER_EXT_BLOB` | `128` | [EIP-7594](https://eips.ethereum.org/EIPS/eip-7594) |
| `FIELD_ELEMENTS_PER_CELL` | `64` | [EIP-7594](https://eips.ethereum.org/EIPS/eip-7594) |
| `BYTES_PER_CELL` | `FIELD_ELEMENTS_PER_CELL * BYTES_PER_FIELD_ELEMENT` (2,048) | [EIP-7594](https://eips.ethereum.org/EIPS/eip-7594) |
| `MAX_BAL_BYTES` | `MAX_BYTES_PER_TX` | [EIP-7928](https://eips.ethereum.org/EIPS/eip-7928) (placeholder until EIP pins a tighter bound) |
| `MAX_EXECUTION_REQUESTS_PER_PAYLOAD` | `2**8` (256) | [EIP-7685](https://eips.ethereum.org/EIPS/eip-7685) |
| `MAX_BYTES_PER_EXECUTION_REQUEST` | `MAX_BYTES_PER_TX` | this spec (placeholder; reuse the tx bound) |
| `MAX_VERSIONED_HASHES_PER_REQUEST` | `128` | [Osaka](./osaka.md#engine_getblobsv2) |
| `MAX_BLOBS_REQUEST` | `MAX_VERSIONED_HASHES_PER_REQUEST` (128) | derived |
| `MAX_BODIES_REQUEST` | `2**5` (32) | [Shanghai](./shanghai.md#engine_getpayloadbodiesbyhashv1) |
| `MAX_REQUEST_BODY_SIZE` | `2**26` (67,108,864) | this spec (64 MiB; advertised as `limits.payload.max_bytes`) |
| `MAX_ERROR_BYTES` | `1024` | this spec |
| `MAX_CLIENT_CODE_LENGTH` | `2` | this spec |
| `MAX_CLIENT_NAME_LENGTH` | `64` | this spec |
| `MAX_CLIENT_VERSION_LENGTH` | `64` | this spec |
| `MAX_CLIENT_VERSIONS` | `4` | this spec |
| `MAX_CAPABILITY_NAME_LENGTH` | `64` | this spec |
| `MAX_CAPABILITIES` | `64` | this spec |

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
    target_gas_limit:         Uint64
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
| `0` | `VALID` | both |
| `1` | `INVALID` | both |
| `2` | `SYNCING` | both |
| `3` | `ACCEPTED` | `POST /payloads` only |

Numbering starts at `0` so a default-initialised SSZ `PayloadStatus`
deserialises as `VALID` rather than as a reserved sentinel.

`INVALID_BLOCK_HASH` is removed (already supplanted by `INVALID`).
`POST /forkchoice` MUST return `0`/`1`/`2` only; CLs MUST treat a
`3` from `/forkchoice` as a protocol error.

`Optional[String]` resolves to `List[List[byte, MAX_ERROR_BYTES], 1]`.
This nesting is subtle, so two worked byte examples follow. (Some
implementations have mistakenly shipped a plain `List[byte, 1024]`,
which cannot distinguish an absent error from an empty-string error;
the wire shape below is the conformant one.)

`PayloadStatus` is a variable-size SSZ container with one fixed field
(`status: uint8`, 1 byte) and two variable-size fields
(`latest_valid_hash`, `validation_error`), each contributing a 4-byte
offset in the fixed part. The fixed part is therefore `1 + 4 + 4` = `9`
bytes, and the variable parts follow in field order.

**Example A — `VALID`, no error** (the 41-byte response from
[refactor.md § Example: submit a payload](./refactor.md#example-submit-a-payload)):

```
status            : 0x00                              # 1 byte, VALID
offset[lvh]       : 0x09000000                        # 4 bytes -> 9
offset[verr]      : 0x29000000                        # 4 bytes -> 41
latest_valid_hash : [<32-byte hash>]                  # Optional present:
                                                       #   inner offset omitted because
                                                       #   Hash32 is fixed-size; the List[T,1]
                                                       #   body is just the 32 bytes
validation_error  :                                   # Optional absent: List length 0, 0 bytes
```

Total = `1 + 4 + 4 + 32 + 0` = `41` bytes. The `latest_valid_hash`
optional is *present* (length-1 list of a fixed-size element, so no
inner offset), and `validation_error` is *absent* (length-0 list).

**Example B — `INVALID`, error present** (`"bad state root"`, 14
bytes of UTF-8):

```
status            : 0x01                              # 1 byte, INVALID
offset[lvh]       : 0x09000000                        # 4 bytes -> 9
offset[verr]      : 0x09000000                        # 4 bytes -> 9 (lvh is empty)
latest_valid_hash :                                   # Optional absent: List length 0, 0 bytes
validation_error  : 0x04000000                        # outer List[..,1] present: one element,
                                                       #   so one 4-byte offset -> 4 (relative to
                                                       #   the start of this variable region)
                  : 0x6261642073746174... (14 bytes)  # inner List[byte,1024] body: the UTF-8 text
```

The `validation_error` variable region is `4 + 14` = `18` bytes: a
single 4-byte offset (because the outer `List[String, 1]` has one
variable-size element, `String` = `List[byte, 1024]`, so its offset is
emitted) pointing at the start of the 14-byte text. Total =
`1 + 4 + 4 + 0 + 18` = `27` bytes. Note the inner 4-byte offset that
precedes the text whenever the error is present — this is exactly the
byte that a plain `List[byte, 1024]` implementation omits, and the
source of the divergence.

---

## Per-fork container catalogue

Each fork-scoped endpoint (`/payloads`, `/forkchoice`, `/bodies`)
uses its own SSZ container shape, selected by the
`Eth-Execution-Version` request header. ELs handling
`Eth-Execution-Version: cancun` MUST use the Cancun containers; ELs
handling `Eth-Execution-Version: amsterdam` MUST use the Amsterdam
containers; etc. This section catalogues every fork-scoped variant.

### `ExecutionPayload` per fork

Used by `POST /payloads` (the inner `payload` field of
`ExecutionPayloadEnvelope`) and `GET /payloads/{payloadId}`
(the inner `payload` field of `BuiltPayload`).

```
# Paris
ExecutionPayloadParis {
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
}

# Shanghai = Paris + withdrawals
ExecutionPayloadShanghai {
    ...Paris fields...
    withdrawals:        List[Withdrawal, MAX_WITHDRAWALS_PER_PAYLOAD]
}

# Cancun = Shanghai + blob_gas_used + excess_blob_gas
ExecutionPayloadCancun {
    ...Shanghai fields...
    blob_gas_used:      Uint64
    excess_blob_gas:    Uint64
}

# Prague = Cancun (no payload-shape change; execution_requests is at the envelope level)
ExecutionPayloadPrague = ExecutionPayloadCancun

# Osaka = Prague (no payload-shape change; blobs bundle moved to BlobsBundleV2)
ExecutionPayloadOsaka = ExecutionPayloadPrague

# Amsterdam = Cancun + block_access_list + slot_number
ExecutionPayloadAmsterdam {
    ...Cancun fields...
    block_access_list:  ByteList[MAX_BAL_BYTES]
    slot_number:        Uint64
}
```

The Amsterdam variant is identical to the
[`ExecutionPayload` (Amsterdam)](#executionpayload-amsterdam) shape
above; this section just makes the progression explicit.

### `PayloadAttributes` per fork

Used by the `payload_attributes` field of `ForkchoiceUpdate` (the
request body of `POST /forkchoice`).

```
# Paris
PayloadAttributesParis {
    timestamp:               Uint64
    prev_randao:             Bytes32
    suggested_fee_recipient: Address
}

# Shanghai = Paris + withdrawals
PayloadAttributesShanghai {
    ...Paris fields...
    withdrawals:             List[Withdrawal, MAX_WITHDRAWALS_PER_PAYLOAD]
}

# Cancun = Shanghai + parent_beacon_block_root
PayloadAttributesCancun {
    ...Shanghai fields...
    parent_beacon_block_root: Root
}

# Prague = Cancun (no shape change)
PayloadAttributesPrague = PayloadAttributesCancun

# Osaka = Cancun (no shape change)
PayloadAttributesOsaka = PayloadAttributesCancun

# Amsterdam = Cancun + slot_number + target_gas_limit
PayloadAttributesAmsterdam {
    ...Cancun fields...
    slot_number:      Uint64
    target_gas_limit: Uint64
}
```

### `ExecutionPayloadBody` per fork

Used by the inner `body` field of `BodyEntry`. Each fork URL serves
only blocks from its own time range, so every field is
unconditionally present (no `Optional[T]`).

```
# Paris
ExecutionPayloadBodyParis {
    transactions: List[ByteList[MAX_BYTES_PER_TX], MAX_TXS_PER_PAYLOAD]
}

# Shanghai = Paris + withdrawals
ExecutionPayloadBodyShanghai {
    ...Paris fields...
    withdrawals: List[Withdrawal, MAX_WITHDRAWALS_PER_PAYLOAD]
}

# Cancun, Prague, Osaka = Shanghai (no shape change for the body)
ExecutionPayloadBodyCancun  = ExecutionPayloadBodyShanghai
ExecutionPayloadBodyPrague  = ExecutionPayloadBodyShanghai
ExecutionPayloadBodyOsaka   = ExecutionPayloadBodyShanghai

# Amsterdam = Shanghai + block_access_list
ExecutionPayloadBodyAmsterdam {
    ...Shanghai fields...
    block_access_list: ByteList[MAX_BAL_BYTES]
}
```

### `BlobsBundle` per revision

Used by the `blobs_bundle` field of `BuiltPayload`. The bundle shape
follows the consensus-specs progression (V1 single proof, V2 cell
proofs).

```
# Cancun (V1) — one proof per blob
BlobsBundleV1 {
    commitments: List[Bytes48, MAX_BLOB_COMMITMENTS_PER_BLOCK]
    proofs:      List[Bytes48, MAX_BLOB_COMMITMENTS_PER_BLOCK]
    blobs:       List[ByteVector[BYTES_PER_BLOB], MAX_BLOB_COMMITMENTS_PER_BLOCK]
}

# Osaka+ (V2) — CELLS_PER_EXT_BLOB cell proofs per blob
BlobsBundleV2 {
    commitments: List[Bytes48, MAX_BLOB_COMMITMENTS_PER_BLOCK]
    proofs:      List[Bytes48, MAX_BLOB_COMMITMENTS_PER_BLOCK * CELLS_PER_EXT_BLOB]
    blobs:       List[ByteVector[BYTES_PER_BLOB], MAX_BLOB_COMMITMENTS_PER_BLOCK]
}
```

`BuiltPayload` for Cancun / Prague carries `BlobsBundleV1`;
Osaka / Amsterdam carries `BlobsBundleV2`.

### `BuiltPayload` per fork

Returned by `GET /payloads/{payloadId}`. Like the
`ExecutionPayload` catalogue above, the shape evolves per fork; this
section pins every variant so there is no ambiguity for pre-Amsterdam
forks. **All variants use the same field order; SSZ fields are
positional, so the order below is normative and MUST be followed
exactly.**

Two points where implementations have diverged, settled here:

- **Field order.** `execution_requests` (when present) comes
  **before** `should_override_builder`. This intentionally differs
  from the legacy JSON-RPC `getPayload` envelope, which appended
  `executionRequests` last. Implementations that kept the legacy order
  are non-conformant.
- **Paris shape.** Paris `BuiltPayload` is **not** a bare
  `ExecutionPayload`; it is the `{payload, block_value}` container
  below. (`engine_getPayloadV1` returned a bare payload; `V2`
  introduced the `{executionPayload, blockValue}` wrapper. The v2 API
  uses the wrapper uniformly from Paris on, so every fork has the same
  outer container and a CL never has to special-case Paris.)

Field-introduction history (from the legacy `engine_getPayloadV{1..5}`
response evolution; see [shanghai.md](./shanghai.md),
[cancun.md](./cancun.md), [prague.md](./prague.md),
[osaka.md](./osaka.md)):

| Field | Introduced | Notes |
| - | - | - |
| `payload`, `block_value` | Shanghai (`getPayloadV2`) | the wrapper itself; v2 API back-applies it to Paris |
| `blobs_bundle` | Cancun (`getPayloadV3`) | `BlobsBundleV1` (single proof) |
| `should_override_builder` | Cancun (`getPayloadV3`) | introduced alongside `blobs_bundle` — **not** Shanghai |
| `execution_requests` | Prague (`getPayloadV4`) | placed before `should_override_builder` |
| `blobs_bundle` → `BlobsBundleV2` | Osaka (`getPayloadV5`) | cell proofs replace the single proof |

```
# Paris — payload + block_value only
BuiltPayloadParis {
    payload:                 ExecutionPayloadParis
    block_value:             Uint256
}

# Shanghai — Paris + nothing new on the wrapper (getPayloadV2 added the
# {executionPayload, blockValue} wrapper; that's already the Paris shape here)
BuiltPayloadShanghai {
    payload:                 ExecutionPayloadShanghai
    block_value:             Uint256
}

# Cancun — Shanghai + blobs_bundle (V1) + should_override_builder
# (both introduced together in engine_getPayloadV3/Cancun)
BuiltPayloadCancun {
    payload:                 ExecutionPayloadCancun
    block_value:             Uint256
    blobs_bundle:            BlobsBundleV1
    should_override_builder: Boolean
}

# Prague — Cancun + execution_requests (note: before should_override_builder)
BuiltPayloadPrague {
    payload:                 ExecutionPayloadPrague
    block_value:             Uint256
    blobs_bundle:            BlobsBundleV1
    execution_requests:      List[ByteList[MAX_BYTES_PER_EXECUTION_REQUEST], MAX_EXECUTION_REQUESTS_PER_PAYLOAD]
    should_override_builder: Boolean
}

# Osaka — Prague but blobs_bundle is V2 (cell proofs)
BuiltPayloadOsaka {
    payload:                 ExecutionPayloadOsaka
    block_value:             Uint256
    blobs_bundle:            BlobsBundleV2
    execution_requests:      List[ByteList[MAX_BYTES_PER_EXECUTION_REQUEST], MAX_EXECUTION_REQUESTS_PER_PAYLOAD]
    should_override_builder: Boolean
}

# Amsterdam — Osaka with the Amsterdam ExecutionPayload (BAL + slot_number)
BuiltPayloadAmsterdam {
    payload:                 ExecutionPayloadAmsterdam
    block_value:             Uint256
    blobs_bundle:            BlobsBundleV2
    execution_requests:      List[ByteList[MAX_BYTES_PER_EXECUTION_REQUEST], MAX_EXECUTION_REQUESTS_PER_PAYLOAD]
    should_override_builder: Boolean
}
```

The Amsterdam variant is the one shown in the
[endpoint section below](#get-forkpayloadspayloadid).

### `ExecutionPayloadEnvelope` per fork

The request body of `POST /payloads`. `parent_beacon_block_root`
exists from Cancun on (it was a separate `engine_newPayload` parameter
since Cancun); `execution_requests` from Prague on. Field order is
normative.

```
# Paris / Shanghai — bare payload, no envelope fields
ExecutionPayloadEnvelopeParis {
    payload: ExecutionPayloadParis
}
ExecutionPayloadEnvelopeShanghai {
    payload: ExecutionPayloadShanghai
}

# Cancun — + parent_beacon_block_root
ExecutionPayloadEnvelopeCancun {
    payload:                  ExecutionPayloadCancun
    parent_beacon_block_root: Root
}

# Prague — Cancun + execution_requests
ExecutionPayloadEnvelopePrague {
    payload:                  ExecutionPayloadPrague
    parent_beacon_block_root: Root
    execution_requests:       List[ByteList[MAX_BYTES_PER_EXECUTION_REQUEST], MAX_EXECUTION_REQUESTS_PER_PAYLOAD]
}

# Osaka = Prague shape, with ExecutionPayloadOsaka inner
# Amsterdam = Prague shape, with ExecutionPayloadAmsterdam inner
```

### `ForkchoiceUpdate` per fork

The request body of `POST /forkchoice`. `payload_attributes`
selects the matching per-fork `PayloadAttributes` shape;
`custody_columns` exists only from Amsterdam on.

```
# Paris .. Osaka
ForkchoiceUpdate {
    forkchoice_state:   ForkchoiceState
    payload_attributes: Optional[PayloadAttributes]   # the fork's PayloadAttributes shape
}

# Amsterdam — + custody_columns
ForkchoiceUpdateAmsterdam {
    forkchoice_state:   ForkchoiceState
    payload_attributes: Optional[PayloadAttributesAmsterdam]
    custody_columns:    Optional[Bitvector[CELLS_PER_EXT_BLOB]]
}
```

### `BlobAndProof` per revision

Used by `BlobEntry.contents` on the blob-pool endpoints (`/blobs/vN`).

```
# /blobs/v1 — Cancun whole-blob, single proof
BlobAndProofV1 {
    blob:  ByteVector[BYTES_PER_BLOB]
    proof: Bytes48
}

# /blobs/v2 and /blobs/v3 — Osaka cell proofs
BlobAndProofV2 {
    blob:   ByteVector[BYTES_PER_BLOB]
    proofs: List[Bytes48, CELLS_PER_EXT_BLOB]
}

# /blobs/v4 — Amsterdam cell-range selection (per-cell nullable)
BlobCellsAndProofs {
    blob_cells: List[Optional[ByteVector[BYTES_PER_CELL]], CELLS_PER_EXT_BLOB]
    proofs:     List[Optional[Bytes48], CELLS_PER_EXT_BLOB]
}
```

### Identification & capabilities

Used by `GET /identity` and `GET /capabilities`. These are JSON on
the wire (see [refactor.md § Resource model](./refactor.md#resource-model-overview)),
but we list the SSZ shapes for completeness so future versions could
switch to SSZ if desired.

```
ClientVersion {
    code:    ByteList[MAX_CLIENT_CODE_LENGTH]
    name:    ByteList[MAX_CLIENT_NAME_LENGTH]
    version: ByteList[MAX_CLIENT_VERSION_LENGTH]
    commit:  Bytes4
}

IdentityResponse {
    versions: List[ClientVersion, MAX_CLIENT_VERSIONS]
}

CapabilitiesResponse {
    capabilities: List[ByteList[MAX_CAPABILITY_NAME_LENGTH], MAX_CAPABILITIES]
    # ... plus the structured fields documented in refactor.md
}
```

---

## Endpoint containers

The endpoint sketches below use the **Amsterdam** shapes as the worked
example. Every fork-scoped endpoint (`/payloads`, `/forkchoice`,
`/bodies`) is defined for **every fork from Paris
onward**; substitute the matching entry from the
[per-fork container catalogue](#per-fork-container-catalogue) for the
value of the `Eth-Execution-Version` request header. For instance
`POST /payloads` with `Eth-Execution-Version: cancun` takes an
`ExecutionPayloadEnvelopeCancun` wrapping an `ExecutionPayloadCancun`,
and `GET /payloads/{id}` with `Eth-Execution-Version: shanghai`
returns a `BuiltPayloadShanghai`.

> **Fork-invariant containers.** `PayloadStatus`, `ForkchoiceState`,
> `ForkchoiceUpdateResponse`, and `Withdrawal` have the **same shape
> across all forks** — only the fork-scoped payload/attributes/body
> containers and the `BuiltPayload` / `ExecutionPayloadEnvelope` /
> `ForkchoiceUpdate` wrappers that embed them vary by fork.

> **Implementation note (monolithic vs. per-fork types).** This
> catalogue names a distinct container per fork
> (`ExecutionPayloadParis`, `…Shanghai`, …) so the wire shape of each
> fork is unambiguous. Implementations are free to model these as a
> single *monolithic* superset container per type whose fields are
> gated on the active fork (e.g. one `ExecutionPayload` struct where
> `withdrawals` participates only from Shanghai, `block_access_list`
> only from Amsterdam, etc.), driving the gate from the
> `Eth-Execution-Version` header.
> This is a valid strategy **as long as the bytes on the wire are
> identical** to the per-fork shape for that fork — i.e. a gated-off
> field contributes neither an offset nor content. go-ethereum's
> implementation takes this monolithic approach.

### `POST /payloads`

Replaces `engine_newPayloadV{1..5}` (Amsterdam shown; `engine_newPayloadV5`).
Each fork uses its `ExecutionPayloadEnvelope{Fork}` from the catalogue
above — Paris/Shanghai carry the bare payload, Cancun+ add
`parent_beacon_block_root`, Prague+ add `execution_requests`.

#### Request (Amsterdam)

```
ExecutionPayloadEnvelopeAmsterdam {
    payload:                  ExecutionPayloadAmsterdam
    parent_beacon_block_root: Root
    execution_requests:       List[ByteList[MAX_BYTES_PER_EXECUTION_REQUEST], MAX_EXECUTION_REQUESTS_PER_PAYLOAD]
}
```

`expected_blob_versioned_hashes` is removed (the EL recomputes it
from `payload.transactions`).

#### Response

`PayloadStatus` (full enum, `0`/`1`/`2`/`3`).

### `POST /forkchoice`

Replaces `engine_forkchoiceUpdatedV{1..4}` (Amsterdam shown;
`engine_forkchoiceUpdatedV4`). Each fork uses its `ForkchoiceUpdate{Fork}`
and `PayloadAttributes{Fork}` from the catalogue; `custody_columns`
exists only from Amsterdam on. `ForkchoiceState` and the response are
fork-invariant.

#### Request (Amsterdam)

```
ForkchoiceUpdateAmsterdam {
    forkchoice_state:    ForkchoiceState
    payload_attributes:  Optional[PayloadAttributesAmsterdam]
    custody_columns:     Optional[Bitvector[CELLS_PER_EXT_BLOB]]
}
```

#### Response (all forks)

```
ForkchoiceUpdateResponse {
    payload_status: PayloadStatus      # restricted: VALID | INVALID | SYNCING
    payload_id:     Optional[Bytes8]
}
```

### `GET /payloads/{payloadId}`

Replaces `engine_getPayloadV{1..6}` (Amsterdam shown;
`engine_getPayloadV6`). Each fork returns its `BuiltPayload{Fork}` from
the catalogue.

#### Response (Amsterdam)

```
BuiltPayloadAmsterdam {
    payload:                 ExecutionPayloadAmsterdam
    block_value:             Uint256
    blobs_bundle:            BlobsBundleV2          # see consensus-specs Osaka
    execution_requests:      List[ByteList[MAX_BYTES_PER_EXECUTION_REQUEST], MAX_EXECUTION_REQUESTS_PER_PAYLOAD]
    should_override_builder: Boolean
}

BlobsBundleV2 {
    commitments: List[Bytes48, MAX_BLOB_COMMITMENTS_PER_BLOCK]
    proofs:      List[Bytes48, MAX_BLOB_COMMITMENTS_PER_BLOCK * CELLS_PER_EXT_BLOB]
    blobs:       List[ByteVector[BYTES_PER_BLOB], MAX_BLOB_COMMITMENTS_PER_BLOCK]
}
```

`commitments` and `blobs` MUST have equal length; `proofs` MUST
have length `len(blobs) * CELLS_PER_EXT_BLOB` (mirrors the
`engine_getPayloadV5` rule from osaka.md).

### `POST /bodies/hash` and `GET /bodies?...`

Replace `engine_getPayloadBodiesByHashV{1,2}` and
`engine_getPayloadBodiesByRangeV{1,2}` (Amsterdam shown). Both return
the same response container; the inner `ExecutionPayloadBody` follows
the `Eth-Execution-Version` request header per the catalogue.

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
```

`available` is `false` when the requested block is unavailable /
pruned, **or** when the block's timestamp falls outside the header
fork's active range. When `available=false`, `body` is zero-valued
and CLs MUST ignore its contents.

For **range queries**, blocks past the latest known block are **not**
represented by an `available=false` entry — they are **omitted**, and
the response is truncated at the latest known block (the legacy
`engine_getPayloadBodiesByRange` "no trailing nulls" rule). The
`entries` list therefore has length `min(count, head - from + 1)` for
`from <= head`, and is empty when `from > head`. Only `available=false`
appears for in-range-but-out-of-era or pruned blocks; never as
trailing padding. See
[refactor.md § Historical bodies](./refactor.md#historical-bodies).

Each `Eth-Execution-Version` value pairs with its own
`ExecutionPayloadBody` schema. The Amsterdam variant carries every
field unconditionally:

```
# Amsterdam ExecutionPayloadBody (Eth-Execution-Version: amsterdam)
ExecutionPayloadBody {
    transactions:      List[ByteList[MAX_BYTES_PER_TX], MAX_TXS_PER_PAYLOAD]
    withdrawals:       List[Withdrawal, MAX_WITHDRAWALS_PER_PAYLOAD]
    block_access_list: ByteList[MAX_BAL_BYTES]
}
```

Earlier-fork variants drop the fields their fork didn't have. For
reference:

```
# Cancun ExecutionPayloadBody (Eth-Execution-Version: cancun)
ExecutionPayloadBody {
    transactions: List[ByteList[MAX_BYTES_PER_TX], MAX_TXS_PER_PAYLOAD]
    withdrawals:  List[Withdrawal, MAX_WITHDRAWALS_PER_PAYLOAD]
}

# Paris ExecutionPayloadBody (Eth-Execution-Version: paris)
ExecutionPayloadBody {
    transactions: List[ByteList[MAX_BYTES_PER_TX], MAX_TXS_PER_PAYLOAD]
}
```

No `Optional[T]` cross-fork nullability anywhere — each fork
returns only blocks from its own era, so every field is always
present.

### `POST /blobs/v1`

Replaces `engine_getBlobsV1` (Cancun whole-blob).

#### Request

```
BlobsV1Request {
    versioned_hashes: List[VersionedHash, MAX_BLOBS_REQUEST]
}
```

#### Response

`200 OK` carries the SSZ body below; `204 No Content` (with empty
body) signals "EL cannot serve this request" (e.g. syncing).

```
BlobsV1Response {
    entries: List[BlobV1Entry, MAX_BLOBS_REQUEST]
}

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
`BYTES_PER_BLOB`-byte zero blob and a 48-byte zero proof) and CLs
MUST ignore them.

### `POST /blobs/v2`

Replaces `engine_getBlobsV2` (Osaka all-or-nothing cell proofs).

#### Request — same as `/v1`

```
BlobsV2Request {
    versioned_hashes: List[VersionedHash, MAX_BLOBS_REQUEST]
}
```

#### Response

`200 OK` carries the SSZ body below; `204 No Content` (with empty
body) signals either "EL cannot serve this request at all" or
"at least one requested blob is missing" (V2 is all-or-nothing).

```
BlobsV2Response {
    entries: List[BlobV2Entry, MAX_BLOBS_REQUEST]
}

BlobV2Entry {
    available: Boolean      # always true for /v2 (all-or-nothing); included for shape symmetry
    contents:  BlobAndProofV2
}

BlobAndProofV2 {
    blob:   ByteVector[BYTES_PER_BLOB]
    proofs: List[Bytes48, CELLS_PER_EXT_BLOB]
}
```

CLs that need partial responses use `/v3`.

### `POST /blobs/v3`

Replaces `engine_getBlobsV3` (Osaka partial responses with cell
proofs).

#### Request — same as `/v2`

#### Response

`200 OK` carries the SSZ body; missing blobs surface as
`available=false` per entry. `204 No Content` only when the EL
cannot serve the request at all (e.g. syncing).

```
BlobsV3Response {
    entries: List[BlobV2Entry, MAX_BLOBS_REQUEST]
}
```

`/v3` reuses `BlobV2Entry` (and therefore `BlobAndProofV2`)
**verbatim** — the wire encoding of a `/v3` entry is byte-identical to
a `/v2` entry; only the response-level semantics differ (`/v3` allows
per-entry `available=false`, `/v2` is all-or-nothing). There is **no**
separate `BlobV3Entry` type: implementations **MUST NOT** define one,
to avoid the two drifting apart. The only difference between the
revisions lives in the endpoint's response semantics, not the entry
container.

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

`200 OK` carries the SSZ body; `204 No Content` signals "EL cannot
serve this request at all."

```
BlobsV4Response {
    entries: List[BlobV4Entry, MAX_BLOBS_REQUEST]
}

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

`BYTES_PER_CELL` = `FIELD_ELEMENTS_PER_CELL * BYTES_PER_FIELD_ELEMENT`
= `2048` ([EIP-7594](https://eips.ethereum.org/EIPS/eip-7594)). A cell
spans `FIELD_ELEMENTS_PER_CELL` (64) field elements of the *extended*
blob (`FIELD_ELEMENTS_PER_EXT_BLOB` = `2 * FIELD_ELEMENTS_PER_BLOB` =
`8192`), so there are `8192 / 64` = `128` = `CELLS_PER_EXT_BLOB` cells,
each `64 * 32` = `2048` bytes — `c-kzg-4844`'s `compute_cells` writes
exactly this. The earlier `BYTES_PER_BLOB / CELLS_PER_EXT_BLOB`
derivation was wrong: it divided the *original*-blob byte count over
the *extended*-blob cell count, halving the true cell size.

---

## Open sketch questions

1. **`MAX_BAL_BYTES`.** EIP-7928 defines the BAL but hasn't pinned
   a numeric upper bound yet. The catalogue currently uses
   `MAX_BYTES_PER_TX` as a placeholder; this should be tightened
   once the EIP lands.
2. **`MAX_BYTES_PER_EXECUTION_REQUEST`.** EIP-7685 hasn't pinned a
   numeric per-element bound either. Same placeholder pattern as
   `MAX_BAL_BYTES`; needs a concrete value.
3. **`Bitvector` SSZ encoding for `indices_bitarray` and
   `custody_columns`.** Both are `Bitvector[CELLS_PER_EXT_BLOB]`
   = `Bitvector[128]` = 16 bytes packed. Double-check that's the
   reading the Amsterdam spec wants (it currently describes it as
   "16 bytes interpreted as a bitarray").
4. **`PayloadStatus` enum encoding.** A `uint8` with sentinel
   values matches the JSON-RPC enum; SSZ has no native enum type
   so this is the cleanest mapping. Alternative: `Container { ... }`
   wrapping a `uint8`.
5. **Naming convention.** The legacy spec used `camelCase`; this
   sketch uses `snake_case` to match consensus-specs. Worth
   confirming once before publication.
