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

## Groups
| Group | File |
| --- | --- |
| `TransactionValidationErrors` | `transaction-validation-errors.yaml` |
| `ExecutionResultErrors` | `execution-result-errors.yaml` |
| `SimulationErrors` | `simulation-errors.yaml` |
| `TxPoolErrors` | `txpool-errors.yaml` |

## Extending the catalog
1. Add or update a group file in `src/error-groups/` with `errors` (and an optional `range`).
2. Reference the group from a method with `$ref: '#/components/error-groups/<GroupName>'` in `error-groups`.
3. Rebuild specs using `make build` (or run `./tools/specgen ...` with the same flags from `Makefile`).

This keeps method definitions concise while preserving consistent error semantics across clients.