# Error groups overview

## Goal
A standard for JSON-RPC error codes & ship a shared catalog of JSON-RPC error codes and messages for EVM clients to unlock consistent tooling and developer ergonomics.

#### Motivation
Client implementations and EVM-compatible chains currently reuse codes or return generic error messages, making cross-client debugging brittle.

## Current implementation
This repository uses method-level `error-groups` references (for example in `src/eth/submit.yaml`) that point to reusable group definitions under `#/components/error-groups/*`.

During spec generation, `tools/specgen` resolves those references into each method's final `errors` array.

## Layout
- `src/error-groups/*.yaml` – reusable error-group definitions.
- `src/eth/*.yaml`, `src/debug/*.yaml`, `src/engine/openrpc/methods/*.yaml` – methods that can reference groups via `error-groups`.
- `tools/specgen` – builds `refs-openrpc.json` and `openrpc.json` and resolves `error-groups` references.
- `Makefile` – passes `-error-groups 'src/error-groups'` to `tools/specgen`.

## Implemented methods
Currently, only below methods import all the error groups via `$ref` and may include inline method-specific codes while still inheriting the standard set.
- `eth_sendTransaction` in `src/eth/submit.yaml`
- `eth_sendRawTransaction` in `src/eth/submit.yaml`
## Reserved ranges at a glance
| Extension group | Reserved range | Source |
| --- | --- | --- |
| JSON-RPC standard | $-32768$ to $-32000$ | JSON-RPC 2.0 spec |
| JSON-RPC non-standard | $-32099$ to $-32000$ | JSON-RPC 2.0 addendum |
| Execution errors | $1$ to $199$ | `execution-errors.yaml` |
| Gas errors | $800$ to $999$ | `gas-errors.yaml` |
| TxPool errors | $1000$ to $1199$ | `txpool-errors.yaml` |
| ZK execution errors | $2000$ to $2199$ | `zk-execution-errors.yaml` |

## Error data
The `data` field of an error is machine-readable, so the spec defines its content.
- A server MUST NOT include `data` unless the spec declares a `data` schema for that error. This also applies to the standard codes such as `-32602` and `-32000`. A `data` field with the value `null` counts as present.
- When the spec declares a `data` schema for the error, the server MUST include `data`, and the value MUST match the schema.
- Put human-readable detail in `message`.

For example, the `3` (execution reverted) error of `eth_call` and `eth_estimateGas` carries the raw revert bytes in `data`. The same applies to a code `3` error in `eth_simulateV1` call results. Empty revert output is `"0x"`.

The Engine API defines its own error data in [common.md](../engine/common.md#errors).

## Extending the catalog
1. Add or update a group file in `src/error-groups/` with `category`, `range`, and `errors`.
2. Keep codes within that group's declared range.
3. Reference the group from a method with `$ref: '#/components/error-groups/<GroupName>'` in `error-groups`.
4. Rebuild specs using `make build` (or run `./tools/specgen ...` with the same flags from `Makefile`).

This keeps method definitions concise while preserving consistent error semantics across clients.