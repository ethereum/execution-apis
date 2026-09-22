# Proposed Parity trace profile

This branch is a draft for client review, not adopted RPC policy. It proposes the nine
traditional trace methods and three output families. Support/conformance policy (H01)
remains a standards decision; adding these files does not assert that every client must
implement every method. Geth support is not assumed.

The companion [trace-interop project](https://github.com/banteg/trace-interop) retains
client observations, reproducible cases and per-client impact reports.
Observed agreement is not a correctness oracle. Recommendations below are proposals.

## Explicit choices in this draft

- Query path inputs use hex quantities for caller compatibility; traceAddress outputs retain JSON
  integers. Convert each output integer to a minimal hex quantity before passing it to trace_get.
  Integer path entries return -32602, rather than null.
- Address filters compose with AND between lists. Missing, null and empty lists are unrestricted.
  Addresses compare by bytes, irrespective of hex letter case. The mode extension is rejected by this baseline.
- Standalone historical records (`trace_block`, `trace_filter`, `trace_transaction`,
  `trace_get`) require non-null block localization. Mined transaction frames also require
  non-null transaction hash/index; protocol rewards require null transaction hash/index
  and occur only in block/filter results. Frames inside simulation and replay envelopes
  omit localization and rewards; replay envelopes carry the transaction hash.
  Synthetic PoS rewards and withdrawals are not call traces.
- Trace calls default to latest post-state. Raw signed simulation uses latest and two
  parameters. Accepting an explicitly supplied block-selector extension is outside this
  baseline, not a conformance failure. Unsigned zero-fee calls preserve the block environment.
- Empty trace selection is valid. The envelope always preserves output.
- Unknown selected blocks and range endpoints return -32001 (Resource not found). Unknown
  transactions and valid but absent tree paths return null. Known blocks with pruned required state
  return the existing upstream error 4444. Other transaction/execution error codes need agreement.

## Precompile frames (H29)

All methods that return call frames use the same inclusion rule: retain root precompile
calls regardless of value. Omit nested precompile frames with zero value, and retain
those with nonzero transferred or inherited value, whether successful or failed.
For CALL and CALLCODE, use the explicit value operand, not the parent transaction value.
For DELEGATECALL, use the inherited call value; STATICCALL has zero value. The inherited
value exception preserves action context even though DELEGATECALL transfers no funds.
`traceAddress` and `subtraces` describe the emitted tree, so omitted frames do not
consume child ordinals. This also determines the paths used by `trace_get` and the
records available to `trace_filter`. An omitted frame must not suppress its caller's
VM return-memory effects. Errors belong to the frame that failed; a handled precompile
failure does not make its successful caller fail.

This is a compatibility choice, subject to client review. Parity's trace database
[omitted nested built-ins to avoid denial-of-service from trace volume](https://github.com/openethereum/parity-ethereum/blob/66477a9476200aaf7e933a52dc64676154adfb2b/ethcore/src/executive.rs#L266),
then [retained actual value transfers](https://github.com/openethereum/parity-ethereum/blob/4255d4c46436d2740c4afb521b773cb95072c7cc/ethcore/src/executive.rs#L431) while citing heavy `IDENTITY` use. The inherited-value
case above follows [current client behavior](https://github.com/banteg/trace-interop/blob/main/reports/decisions/H29.md), not Parity's original
transfer-only rationale. JSON Schema alone cannot enforce this rule; the companion
precompile fixtures check inclusion and path numbering.

## Open details requiring focused review

Stable frame error kinds; failed-operation ex conventions; nested CALL stipend/refund
gas accounting; client execution caps; protocol reward ordering; optional method
discovery; pending-state behavior and simulation extensions need further agreement.
The localized schemas describe mined records; they do not define pending localization.
Signed raw-transaction nonce admission (H13) is also a proposed contract choice; malformed
JSON on rejection is independently a reporting defect. The two-argument baseline does
not require clients to remove an explicitly selected third-argument extension (H12).
Call objects accept standard eth_call transaction fields, including blob and authorization
fields with their usual semantics at the selected fork. Unknown object fields are ignored
for forward compatibility; accepting an unknown field does not establish extension support.
Additional positional state and block overrides remain outside this profile.
Full semantic conformance cannot be inferred from schema validity.

The VM schema has its own JSON Schema resource identity and a local self-reference,
so nested execution receives the same validation as the root. Build tooling preserves
that bounded self-reference while still rejecting cyclic component expansion.

## Decisions and client impact

The [decision ledger](https://github.com/banteg/trace-interop/blob/main/decisions/README.md)
contains the rationale, evidence and open questions for H01–H29. The
[client impact tables](https://github.com/banteg/trace-interop/blob/main/reports/README.md)
show the changes required by each tested build, with individual requests and responses.

## Recursive schema rendering

`openrpc.json` retains the recursive `vmTrace` schema and is the validation artifact.
The documentation renderer currently cannot expand recursive schemas. Its generated
`docs-openrpc.json` display input replaces resource-local self references with labeled
recursive object descriptions. It also materializes shared object fields in union
branches and labels array-item variants for the renderer. Different item variants may
coexist in one array; the display wrappers are not an equivalent validation schema.
This artifact must not be used for conformance validation.

`npm run docs:refresh` rebuilds the spec, refreshes this display projection, and copies
the introductory page. The watcher uses the same chain. Startup and production-build
hooks only refresh the projection so an already prepared release spec is preserved.
`npm run test:docs` tests rendered method output and the refresh chain after `make build`.

The Go trace schema tests build the real YAML and validate positive and negative cases
against reference-preserving and expanded schemas (Draft 7 and Draft 2019-09), and the
actual speccheck parsing/validation path. They include recursive VM resources, fee
conflicts, contextual frame restrictions and nonempty callMany tuples. These are schema
tests, not execution fixtures or evidence of client semantic conformance.
