# Frame transaction components

These proposed JSON components give execution-client developers a shared representation of an EIP-8141 envelope. They follow [EIP-8141 at revision b75cbe6](https://github.com/ethereum/EIPs/blob/b75cbe61150f09a44c38843be916417283d5b7bf/EIPS/eip-8141.md). RPC naming and representation require client consensus. This change does not add type 6 to any method's accepted or returned transaction union.

## Complete envelopes

`Transaction8141` contains the complete envelope fields. `sender` represents the encoded sender. `fees` groups `maxPriorityFeePerGas`, `maxFeePerGas`, and `maxFeePerBlobGas`. Each frame groups its `execution` and `state` budgets under `limits`. `type`, modes, flags, and schemes use canonical hex quantities. The JSON type is `"0x6"`; the serialized transaction prefix is the byte `0x06`.

JSON objects follow the envelope's grouping while retaining camelCase RPC names such as `chainId` and `blobVersionedHashes`. Frames and signatures remain ordered lists of objects.

All properties are required, including zero values and empty lists. `target: null` selects the sender; it does not create a contract. Empty signature metadata uses `"0x"`, not null or an omitted property. Empty `signer` selects the sender for SECP256K1/P256; ARBITRARY requires empty `signer`. Empty `msg` selects the canonical signing hash.

`FrameSignature` preserves raw wire bytes. SECP256K1 uses 65 bytes with recovery ID 0 or 1 first. P256 uses 128 bytes in `r || s || qx || qy` order. ARBITRARY permits arbitrary whole bytes, including empty bytes. Empty cryptographic signatures are rejected; simulation placeholders need a separate request contract.

For example, this complete envelope has no protocol signatures. Schema acceptance does not establish approval or execution validity:

```json
{
  "type": "0x6",
  "chainId": "0x1",
  "nonce": "0x0",
  "sender": "0x1111111111111111111111111111111111111111",
  "frames": [{
    "mode": "0x1",
    "flags": "0x3",
    "target": null,
    "limits": { "execution": "0x10000", "state": "0x0" },
    "value": "0x0",
    "data": "0x"
  }],
  "signatures": [],
  "fees": {
    "maxPriorityFeePerGas": "0x1",
    "maxFeePerGas": "0x2",
    "maxFeePerBlobGas": "0x0"
  },
  "blobVersionedHashes": []
}
```

Chain ID, values, and fees fit unsigned 256-bit integers; nonce and individual frame budgets fit unsigned 64-bit integers. Blob hashes use version `0x01`. Without blobs, `fees.maxFeePerBlobGas` is explicitly zero. Blob sidecars are transport data and are not part of this component.

## Gas and validation

There is no independently signed outer `gas` budget. Using the intrinsic and calldata-floor calculations from the pinned EIP, the total reservation is:

```text
executionReservation = max(intrinsicGas + sum(frame.limits.execution), calldataFloorGas)
stateReservation = sum(frame.limits.state)
maxGas = executionReservation + stateReservation
```

The execution reservation must not exceed the EIP-7825 cap of 16,777,216. The sum of declared frame execution and state limits must be less than 2^64. Block capacity constrains both dimensions. This reservation is neither an estimate nor receipt gas usage. A later response contract will specify its RPC representation.

Schema checks cover field presence, encoding, widths, frame count, mode-local value/flag constraints, signature lengths, and the no-blob fee condition. Clients must additionally validate signature scalars, low-s canonicality, signer recovery, fee relationships, gas sums/caps, blob availability, and all execution-dependent approval and payment requirements.

Clients must also check relationships between frames: approval targets, batch successors, approval scope on every batch member, and expiry-verifier constraints. A locally valid frame is not necessarily valid in a particular envelope. These schemas do not prove consensus validity or mempool admission.

## Scope and review decisions

The envelope rejects unknown properties, including outer `to`, `value`, `input`, `data`, `gas`, `gasPrice`, access lists, authorization lists, `from`, flat fee fields, and ECDSA `r`/`s`/`v`/`yParity`. Lookup metadata, compatibility response fields, partial preparation inputs, and method integration belong to subsequent changes.

Before merging, client maintainers must agree on `sender`, `target`, nested `limits` and `fees`, hex quantities for modes/flags/schemes, explicit defaults, and signature metadata encoding. Null/zero compatibility fields and derived gas in responses remain separate decisions. This component does not settle simulation, estimation, receipts, or managed signing behavior.
