# Proposed Parity trace profile

This branch is a draft for client review, not adopted RPC policy. It proposes the nine
traditional trace methods and three output families. Support/conformance policy (H01)
remains a standards decision; adding these files does not assert that every client must
implement every method. Geth support is not assumed.

The companion [trace-interop project](https://github.com/banteg/trace-interop) retains
client observations, reproducible cases and per-client impact reports.
Observed agreement is not a correctness oracle. Recommendations below are proposals.
The target is a useful, precise contract; historical implementations explain compatibility
costs but do not decide it. Intentional departures are called out below.

## Explicit choices in this draft

- Query path inputs use hex quantities for caller compatibility; traceAddress outputs retain JSON
  integers. Convert each output integer to a minimal hex quantity before passing it to trace_get.
  Integer path entries return -32602, rather than null.
- Address filters use OR within each list and AND between lists, as `eth_getLogs` composes topic
  positions. Missing, null and empty lists are unrestricted. Addresses compare by bytes, irrespective
  of hex letter case. Optional `mode` accepts `intersection` (the default) and `union` (match either
  populated list); other values return -32602.
- Standalone historical records (`trace_block`, `trace_filter`, `trace_transaction`,
  `trace_get`) require non-null block localization. Mined transaction frames also require
  non-null transaction hash/index; protocol rewards require null transaction hash/index
  and occur only in block/filter results. Frames inside simulation and replay envelopes
  omit localization and rewards; both individual and block replay envelopes carry the transaction
  hash so results retain their identity outside the request. Requiring it for individual replay
  deliberately extends the original Parity response shape.
  Synthetic PoS rewards and withdrawals are not call traces.
- Trace calls default to latest post-state. Raw signed simulation uses latest and two
  parameters. Accepting an explicitly supplied block-selector extension is outside this
  baseline, not a conformance failure. The third selector existed in Parity’s implementation,
  although its guide omitted it. Unsigned zero-fee calls preserve the block environment.
- Empty trace selection is valid. The envelope always preserves output.
- Unknown selected blocks and range endpoints return -32001 (Resource not found). Unknown
  transactions and valid but absent tree paths return null. Known blocks with pruned required state
  return 4444. If pruned indexing prevents establishing whether a hash is absent, return 4444
  rather than claiming a definitive not-found result. This extends the pruned-history code
  adopted for eth/debug methods in [#636](https://github.com/ethereum/execution-apis/pull/636)
  to trace methods and execution state; that extension remains a proposal.
- Omitted filter bounds mean genesis (`earliest`) through `latest`, resolved once against the
  request’s chain view. Retention must not silently raise the lower bound. This intentionally
  changes Parity’s latest/latest default to make omission a historical search. Required history
  that is unavailable produces 4444; query limits produce an explicit error, never an incomplete
  success. `count` limits returned records, not scan or replay work.

## Execution results and state changes

Failed CALL/CREATE frames require `error` and an explicit `result`: preserve available
revert bytes and measured gas, or use null where result data does not apply. This is a
richer contract than Parity’s error-only failures. A populated result does not imply
success; `error` determines failure. Failed CREATE must not report a successful address
or deployed code. The execution envelope carries root output, which cannot substitute
for a nested frame’s revert bytes. A locally successful child remains successful even
if an ancestor later reverts; its state changes then do not survive in `stateDiff`.

`stateDiff` compares execution endpoints, not intermediate writes. In particular,
multiple EIP-7702 authorizations may restore the original code while changing nonce;
only net code changes appear. Accepted authorization effects survive a later EVM revert.
Accounts absent at both endpoints have no account diff even if created and destroyed;
a prefunded address existed before creation and can instead require deletion markers.
Deletion implies clearing all storage. Whether every untouched prior slot must be
enumerated still needs agreement; empty-storage fixtures do not answer that question.
Each callMany diff is relative to the preceding call’s post-state, including surviving
nonce and fee effects but excluding reverted EVM writes.

VM `mem` describes bytes actually written by that opcode, not its whole accessed or
expanded memory window. MLOAD alone has no byte write; CALL copies only the bytes
actually returned into its output range. An empty copy has no write. This deliberately
refines Parity’s post-step memory snapshots, which could include MLOAD and the entire
requested CALL output window. These deltas do not independently encode allocated memory
size. `pc`, `cost`, `ex.used` and output trace paths remain JSON integers; stack words
use hex quantities. Optional `idx` needs a declared numbering convention to be checked.

Filter membership describes actions, not necessarily committed transfers. Failed CREATE
can match its creator but has no successful recipient address; an explicit union may
still retain the creator match. Range results must match fork-correct per-block records,
whether those records come from replay or an index.

## Response integrity

Before result bytes are committed, validation or execution setup failures must produce
one complete JSON-RPC error. Buffering may be necessary to preserve that boundary:
parsing all callMany inputs does not validate later calls against earlier calls’ state.
If a failure occurs after output is committed, explicitly abort the transport rather
than append a second error envelope or finish a partial result as success. An aborted
request is incomplete, not a successful trace or a complete JSON-RPC error. Tests must
distinguish deliberate transport termination from malformed completed responses.

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
H13 proposes execution validity at the selected state for signed raw transactions:
signature, chain identity, nonce equality, balance for value and upfront gas, intrinsic
gas, fees and the selected fork's sender-code restrictions, including EIP-7702's
delegation exception. Reject both low and high nonces without modifying signed fields
or implicitly changing the sender's nonce or balance. A valid transaction that REVERTs
or runs out of execution gas still returns a trace. Local pool policies such as
replacement pricing, already-known rejection and minimum tips do not apply.

The proposed validation-failure code is `-32003` (Transaction rejected); malformed
encodings and request parameters use `-32602`. Clients currently also use `-32000`, so
error-code alignment requires review separately from the validation policy. Full block
admissibility is not established by a successful simulation. Authorization tuples skipped
under EIP-7702 do not themselves make the outer transaction invalid; validation must
follow the selected fork’s distinction between rejected transactions and skipped tuples.

This deliberately tightens legacy permissive simulation: queued high-nonce transactions
and already-mined low-nonce transactions may stop tracing at latest state. Historical
OpenEthereum skipped nonce checks and could supplement balance. Using a state nonce
different from the signed nonce can change CREATE's execution address; overriding
state instead invents a pre-state. The baseline chooses neither behavior. Hypothetical
execution belongs in explicitly documented simulation facilities. Erigon's
[feedback](https://github.com/ethereum/execution-apis/issues/890#issuecomment-5784143408)
supports strict validation despite the compatibility cost; this is not unanimous client
approval. Malformed JSON on rejection is independently a reporting defect.

The two-argument baseline does
not require clients to remove an explicitly selected third-argument extension (H12).
Call objects accept standard eth_call transaction fields, including blob and authorization
fields with their usual semantics at the selected fork. Unknown object fields are ignored
for shared call-object compatibility; known fields still require validation. Ignoring extras
can hide typos, and acceptance does not establish extension support. Clients must declare
supported extensions rather than relying on a successful response as feature detection.
Additional positional state and block overrides remain outside this profile.
Full semantic conformance cannot be inferred from schema validity.

The VM schema has its own JSON Schema resource identity and a local self-reference,
so nested execution receives the same validation as the root. Build tooling preserves
that bounded self-reference while still rejecting cyclic component expansion.

## Decisions and client impact

The [decision ledger](https://github.com/banteg/trace-interop/blob/main/decisions/README.md)
contains the rationale, evidence and open questions for each decision. The
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
