# Proposed Parity trace profile

This branch is a draft for client review, not adopted RPC policy. It proposes the nine
traditional trace methods and three output families. Support/conformance policy (H01)
remains a standards decision; adding these files does not assert that every client must
implement every method. Geth support is not assumed.

The companion [trace-interop project](https://github.com/banteg/trace-interop) retains
client observations, reproducible cases and per-client impact reports.
Observed agreement is not a correctness oracle. Recommendations below are proposals.

## Explicit choices in this draft

- Query path inputs use hex quantities; traceAddress outputs retain JSON integers.
- Address filters compose with AND between lists. Empty lists are unrestricted;
  null and the mode extension are rejected by this baseline.
- Historical records include localization fields; protocol rewards have null transaction
  hash/index. Synthetic PoS rewards and withdrawals are not call traces.
- Trace calls default to latest post-state. Raw signed simulation uses latest and two
  parameters. Unsigned zero-fee calls preserve the block environment.
- Empty trace selection is valid. The envelope always preserves output.
- The existing upstream pruned-history error 4444 is reused as the draft proposal;
  other transaction/execution error codes need agreement.

## Open details requiring focused review

Stable frame error kinds; failed-operation ex conventions; nested CALL stipend/refund
gas accounting; client execution caps; protocol reward ordering; optional method
discovery; pending-state behavior and simulation extensions need further agreement.
The call schema intentionally excludes blob/authorization override extensions in this
first profile. These are scope gaps, not statements that clients must remove extensions.
Full semantic conformance cannot be inferred from schema validity.

The VM schema has its own JSON Schema resource identity and a local self-reference,
so nested execution receives the same validation as the root. Build tooling preserves
that bounded self-reference while still rejecting cyclic component expansion.

## Decisions and client impact

The [decision ledger](https://github.com/banteg/trace-interop/blob/main/decisions/README.md)
contains the rationale, evidence and open questions for H01–H28. The
[client impact tables](https://github.com/banteg/trace-interop/blob/main/reports/README.md)
show the changes required by each tested build, with individual requests and responses.

## Recursive schema rendering

`openrpc.json` retains the recursive `vmTrace` schema and is the validation artifact.
The documentation renderer currently cannot expand recursive schemas. Its generated
`docs-openrpc.json` display input replaces resource-local self references with labeled
recursive object descriptions; it must not be used for conformance validation.
