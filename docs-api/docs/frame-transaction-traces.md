# Frame transaction traces

This page specifies how `debug_trace*` with `callTracer` and the `trace_*`
methods render an [EIP-8141](https://eips.ethereum.org/EIPS/eip-8141) frame
transaction. It is provisional until at least two clients implement it.

A frame transaction has no single top-level call. Each frame runs as its own
top-level call, in frame order. The trace keeps one tree per transaction by
putting every frame under one synthetic root call. Tooling that walks `calls`
or `traceAddress` recursively therefore picks up all frames with no change.

The per-frame fields of the receipt (`payer`, `frameReceipts`) are the source of
truth for frame status, payer, and per-dimension gas. A trace does not repeat
them. A consumer joins frame `i` of the trace to `frameReceipts[i]` by index.

## Terms

- `ENTRY_POINT` is `0x00000000000000000000000000000000000000aa`, as defined by
  EIP-8141.
- The resolved target of a frame is `frame.target`, or `tx.sender` when
  `frame.target` is empty.
- The gas limit of a frame is `frame.limits.execution + frame.limits.state`.

## `callTracer`

The root call frame of a frame transaction MUST be:

| field | value |
|---|---|
| `type` | `CALL` |
| `from` | `tx.sender` |
| `to` | `ENTRY_POINT` |
| `value` | `0x0` |
| `input` | `0x` |
| `gas` | the transaction gas limit |
| `gasUsed` | the receipt `gasUsed` |

The root `calls` MUST have exactly one entry per entry of `tx.frames`, in frame
order. `calls[i]` is frame `i`:

| field | value |
|---|---|
| `type` | `STATICCALL` for a `VERIFY` frame, `CALL` otherwise |
| `from` | `ENTRY_POINT` for a `DEFAULT` or `VERIFY` frame, `tx.sender` for a `SENDER` frame |
| `to` | the resolved target |
| `value` | `frame.value` |
| `input` | `frame.data` |
| `gas` | the gas limit of the frame |
| `gasUsed` | `frameReceipts[i].gasUsed` |

Calls made by the frame are nested under `calls[i]` as for any call frame.

Status and errors:

- `calls[i]` has an `error` if and only if `frameReceipts[i].status` is not
  `1`.
- A skipped frame (`frameReceipts[i].status` is `2`) has `error` set to
  `frame skipped`, `gasUsed` `0x0`, and no `calls`.
- A frame that fails before it enters the EVM is still rendered from the
  transaction fields, with the error that ended it.
- The root has an `error` if and only if the receipt `status` is `0`. Its
  message is not specified.

Logs, when `withLog` is set:

- The logs of `calls[i]` and its nested calls MUST equal `frameReceipts[i].logs`.
- A frame rolled back by a failed atomic batch keeps status `1` in its receipt
  and has no logs, so its trace shows no logs and no `error`.

With `onlyTopCall`, the result is the root and its frame entries, without the
calls nested under them.

## `trace_*`

The same tree is used in parity encoding:

- The trace at `traceAddress` `[]` is a `call` from `tx.sender` to
  `ENTRY_POINT` with `value` `0x0`, `input` `0x`, `action.gas` equal to the
  transaction gas limit, `result.gasUsed` equal to the receipt `gasUsed`, and
  `subtraces` equal to the number of frames.
- Frame `i` is at `traceAddress` `[i]`, and its nested calls are at
  `[i, ...]`.
- For frame `i`, `callType` is `staticcall` for a `VERIFY` frame and `call`
  otherwise. `from`, `to`, `value`, `input`, and `action.gas` follow the
  `callTracer` rules above, and `result.gasUsed` is
  `frameReceipts[i].gasUsed`.
- A frame that reverts has `error` `Reverted`. A skipped frame, or a frame
  that fails before it enters the EVM, has its `action` and `error`, no
  `result`, and no subtraces. A skipped frame has `error` `frame skipped`.

Because the root and every `DEFAULT` and `VERIFY` frame name `ENTRY_POINT`,
`trace_filter` with `toAddress` `[ENTRY_POINT]` returns the root of every frame
transaction, and `fromAddress` `[ENTRY_POINT]` returns every `DEFAULT` and
`VERIFY` frame.

The frame layout of `vmTrace` is not specified.

## `prestateTracer`

A frame transaction has no `to` and creates no contract from the sender nonce.
The accounts it touches MUST include every resolved target that executed and
the payer, and MUST NOT include an address derived from `tx.sender` and its
nonce unless execution touched it.

## Example

A sponsored transaction with a `VERIFY` frame for the sender, a `VERIFY` frame
for the payer, and a `SENDER` frame that reverts:

```json
{
  "type": "CALL",
  "from": "0x1111111111111111111111111111111111111111",
  "to": "0x00000000000000000000000000000000000000aa",
  "value": "0x0",
  "gas": "0x7d000",
  "gasUsed": "0x1f7a4",
  "input": "0x",
  "error": "execution reverted",
  "calls": [
    {
      "type": "STATICCALL",
      "from": "0x00000000000000000000000000000000000000aa",
      "to": "0x1111111111111111111111111111111111111111",
      "gas": "0x30d40",
      "gasUsed": "0x5208",
      "input": "0x"
    },
    {
      "type": "STATICCALL",
      "from": "0x00000000000000000000000000000000000000aa",
      "to": "0x2222222222222222222222222222222222222222",
      "gas": "0x30d40",
      "gasUsed": "0x61a8",
      "input": "0x"
    },
    {
      "type": "CALL",
      "from": "0x1111111111111111111111111111111111111111",
      "to": "0x3333333333333333333333333333333333333333",
      "value": "0x0",
      "gas": "0x186a0",
      "gasUsed": "0x2710",
      "input": "0xa9059cbb",
      "output": "0x",
      "error": "execution reverted"
    }
  ]
}
```

Its receipt has `payer` `0x2222...2222` and `frameReceipts[2].status` `0`.
