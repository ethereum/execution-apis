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
- Trace calls default to latest post-state and accept a block number, tag or hash. Raw signed
  simulation uses latest and two parameters. Accepting an explicitly supplied block-selector
  extension is outside this baseline, not a conformance failure. The third selector existed in
  Parity’s implementation, although its guide omitted it.
- Empty trace selection is valid. The envelope always preserves output.
- Unknown single selected blocks (`trace_block`, `trace_replayBlockTransactions`, and the
  simulation block) return -32001 (Resource not found). Unknown transactions and valid but absent
  tree paths return null. Known blocks with pruned required state return 4444. If pruned indexing
  prevents establishing whether a hash is absent, return 4444 rather than claiming a definitive
  not-found result. This extends the pruned-history code adopted for eth/debug methods in
  [#636](https://github.com/ethereum/execution-apis/pull/636) to trace methods and execution state;
  that extension remains a proposal.
- `trace_filter` ranges follow `eth_getLogs`
  ([#875](https://github.com/ethereum/execution-apis/pull/875)): if either bound resolves beyond
  the current head, or `fromBlock` resolves above `toBlock`, return -32602 (Invalid params). Never
  clamp the range or return a partial result. Bounds use `BlockNumberOrTagForRange`, which excludes
  `pending`. `earliest` is the lowest block the client has available, as the shared tag defines it;
  an explicit number below retained history returns 4444.
- Omitted filter bounds both mean `latest`, resolved against the same canonical head for the
  request. A `toBlock` earlier than the omitted `fromBlock` is a reversed range (-32602);
  callers searching history must state `fromBlock`. This follows Parity’s original
  `trace_filter` default and the `eth_getLogs` convention. Query limits produce an explicit error,
  never an incomplete success. `count` limits returned records, not scan or replay work.
  [H30](https://github.com/banteg/trace-interop/blob/main/reports/decisions/H30.md)
- `trace_block` and `trace_replayBlockTransactions` reject `pending` with -32602 until a pending
  contract exists. Localized records never carry a block hash that is not canonical.

## Blocks, rewards and pagination

The genesis block has no transaction or reward records: `trace_block(0)` and
`trace_replayBlockTransactions(0)` return `[]`, and filters over genesis contribute nothing. In
every other block, rewards follow all transaction records of that block: the block reward, then
uncle rewards in ommer order. A reward matches `toAddress` by its `author` and has no sender side,
so it is excluded when `fromAddress` is populated in intersection mode; in union mode a `toAddress`
match suffices. `trace_filter` filters first, then applies `after` and `count` in block,
transaction, preorder, then reward order. `after` and `count` are JSON integers bounded by uint64.
Offsets are stable only for a fixed, numerically resolved range.

System operations (EIP-4788 and EIP-2935 pre-block calls, EIP-7002 and EIP-7251 post-block
calls), withdrawals and rewards are applied to replay state in protocol order, but they are not call
trace records and belong to no transaction’s `stateDiff`. Concatenated transaction diffs therefore do
not reconstruct the block post-state. Each block is traced from its parent’s post-block state with
its own pre-transaction system operations applied, under its own fork rules, whether a request
covers one block or a range. Replay arrays hold exactly one envelope per transaction.

## Simulation requests

Call objects use the `eth_simulateV1` `GenericCallTransaction` fields, together with `chainId` and
`authorizationList` from `GenericTransaction` and `data` as an alias for `input`. When `data` and
`input` are both present they must be equal (-32602). Every field the schema defines either takes
effect with its `eth_call` meaning or causes a rejection (-32602, or the fork’s validity error);
only fields outside the schema are ignored. A supplied nonce is accepted but neither validated nor
used, so CREATE addresses derive from the state nonce. `gas` is a uint64.

Fees follow `eth_call` and `eth_simulateV1` (H15). Omitted fee fields default to zero. The zero-fee
rule applies to the effective gas price after defaulting: a zero price means GASPRICE 0 and BASEFEE
0. BLOBBASEFEE is 0 exactly when `maxFeePerBlobGas` is supplied as 0 or defaulted to 0; calls without blob
fields keep the block’s BLOBBASEFEE. Fee validation is skipped only when both fee caps are zero.
Positive prices are validated against the base fee, funded and charged, with refunds, base-fee burn
and tips simulated. This removes Parity’s virtual balance top-up for unsigned calls.

Parameter positions for overrides are reserved: `trace_call` takes `StateOverrides` fourth and
`BlockOverrides` fifth; `trace_callMany` takes them third and fourth. Both use the `eth_simulateV1`
schemas. A client that does not implement them must reject a non-null value with -32602. With a
block override, the zero-fee rule uses the overridden base fee.

In `trace_callMany` every item is a separate transaction: EIP-2200 and EIP-3529 original storage
values, EIP-2929 access sets, EIP-1153 transient storage, the refund counter and EIP-6780’s
same-transaction scope start afresh. All items share one block environment, so NUMBER and TIMESTAMP
do not advance. The first item runs against the same state and environment as `trace_call` at the
selected block, and each diff is relative to the preceding item’s post-state. If any item fails
validation, the request returns one error whose `error.data.index` is the zero-based item index, and
no partial results. Servers may cap items or total gas with an explicit -38026 error, never by
truncating.

Validation rejections of `trace_call` and `trace_callMany` use the `eth_simulateV1` codes: -38010
nonce too low, -38011 nonce too high, -38012 base fee too low, -38013 intrinsic gas, -38014
insufficient funds, -38015 block gas limit, -38024 sender not an EOA, -38025 init-code size and
-38026 client limit. A priority fee above the fee cap is -32602, as Geth reports it for `eth_simulateV1`.
`trace_rawTransaction` uses the `eth_sendRawTransaction` error groups
([#650](https://github.com/ethereum/execution-apis/pull/650)): 1 nonce too low, 2 nonce too high,
800 intrinsic gas, 804 priority fee above fee cap, 806 fee cap below base fee and 809 insufficient
funds, with -32003 (Transaction rejected) as the generic fallback. Malformed input is -32602.

## Execution results and state changes

Frame numbers follow Parity. The root `action.gas` is the transaction gas minus intrinsic gas,
including access-list and authorization costs. The root `result.gasUsed` is execution gas before
refunds, excluding intrinsic gas and the EIP-7623 floor. A nested CALL-family `action.gas` is the gas
forwarded after the 63/64 cap plus the 2300 stipend for a value transfer. CREATE `gasUsed` includes
the code deposit, and an exceptional halt consumes `action.gas`. CALL and STATICCALL report the
caller as `from` and the target as `to`; DELEGATECALL and CALLCODE report the executing address and
the code address. DELEGATECALL `value` is the inherited value and STATICCALL `value` is 0. Create
actions require `creationMethod` (`create` or `create2`). A `suicide` frame records the opcode and
its transfer, not account deletion. Calls and creates that fail their precheck (depth limit,
insufficient balance) emit no frame; a CREATE address collision emits a create frame with error
`Contract address collision`.

`error` alone determines failure. A REVERT frame has error `Reverted` and requires
`result: {gasUsed, output}`, for CREATE too, without an address or code. An exceptional halt may
omit `result` or set it to null. Erigon and Reth already emit REVERT results; the new part is the
failed-CREATE shape, which must not report the would-be address. Failure labels are `Reverted`,
`Out of gas`, `Bad instruction`, `Bad jump destination`, `Stack underflow`, `Out of stack`,
`Mutable Call In Static Context`, `Built-in failed` and `Out of bounds`, plus the post-Parity
`Contract address collision`, `Code size limit exceeded`, `Invalid code prefix 0xEF` and
`Nonce overflow`. Other strings are extensions that consumers treat as generic failure.
`revertReason`, if present, is decoded `Error(string)` text; raw revert bytes live in
`result.output`. The execution envelope carries root output, which cannot substitute for a nested
frame’s revert bytes. A locally successful child remains successful even if an ancestor later
reverts; its state changes then do not survive in `stateDiff`.

`stateDiff` compares execution endpoints, not intermediate writes. An account appears only if its
balance, nonce, code, storage or existence changed. Existence means presence in the state trie, so
an existing EIP-161-empty account removed by touch-clearing is a deletion. `+` and `-` appear only
when the account is born or dies. Slots of an account that exists at both endpoints use `*` with
32-byte words, including zero words; slots never use `=`. A born account lists its nonzero
post-state slots as `+`. A deleted account has `storage: {}`, and its `-` implies that all storage
is wiped, as Parity, Besu and Erigon already report. Multiple EIP-7702 authorizations may restore the
original code while changing nonce; only net code changes appear. Accepted authorization effects
survive a later EVM revert. Accounts absent at both endpoints have no account diff even if created
and destroyed. The sender pays value, `gasUsed` times the effective price and `blobGasUsed` times
the blob base fee; base and blob fees are burned or credited to a chain-defined collector, and the
tip goes to the fee recipient.

VM `mem` is the post-operation contents of the memory range the opcode’s operands designate:
MSTORE and MLOAD `[off, off+32)`, MSTORE8 `[off, off+1)`, the destination of CALLDATACOPY,
CODECOPY, EXTCODECOPY, RETURNDATACOPY and MCOPY, and the full CALL-family output window. It is null
when that range is empty or the opcode has none (RETURN, REVERT, LOG, KECCAK256, CREATE). This
matches Parity, Erigon, Nethermind and Besu for MLOAD, and reverses an earlier draft that narrowed
`mem` to written bytes. `cost` is the total gas deducted from the caller when the operation starts,
including gas made available to a child frame (the 63/64-capped forwarded gas for the CALL family,
excluding the value stipend; all but 1/64 for the CREATE family). `ex.used` is the caller’s
remaining gas after the operation, including unused child gas returned. An operation that began
executing and halted exceptionally keeps its pre-execution `cost` and has `ex: null` and `sub: null`.
Operations rejected before execution are omitted, and no synthetic STOP is added, so every `pc`
lies inside `code`. `push` is the top k stack words after execution, deepest first, where k is the
number of words the opcode leaves in place of its inputs (DUPn and SWAPn n+1, CALL and CREATE 1).
`store` is set only by SSTORE, from its operands, whenever the SSTORE completes, even if the value is
unchanged. A selected `vmTrace` is always an object. An operation that entered a child frame,
including a precompile or empty-code target, has that frame’s trace as `sub`
(`{code: "0x", ops: []}` if no instruction ran); others have `sub: null`. These deltas do not
independently encode allocated memory size. `pc`, `cost`, `ex.used` and output trace paths remain
JSON integers; stack words use hex quantities. Optional `idx` needs a declared numbering convention
to be checked.

Filter membership describes actions, not necessarily committed transfers. Failed CREATE
can match its creator but has no created-address match; an explicit union may
still retain the creator match. SELFDESTRUCT matches the executing (self-destructing) account and
the beneficiary; after EIP-6780 the account usually survives. Range results must match
fork-correct per-block records, whether those records come from replay or an index.

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
A precompile is an address in the precompile set active at the executing block’s fork.
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

## Signed transaction validation

H13 proposes execution validity at the selected state for signed raw transactions:
signature, chain identity, nonce equality, balance for value and upfront gas, intrinsic
gas, fees and the selected fork's sender-code restrictions, including EIP-7702's
delegation exception. Reject both low and high nonces without modifying signed fields
or implicitly changing the sender's nonce or balance. A valid transaction that REVERTs
or runs out of execution gas still returns a trace. Local pool policies such as
replacement pricing, already-known rejection and minimum tips do not apply. Full block
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

## Open details requiring focused review

Optional method discovery; client execution caps for omitted gas; `pending` for simulations
and its localization; the shared semantics of a raw-transaction block selector (H12); and
simulation extensions beyond the reserved override positions need further agreement. Clients
must declare supported extensions rather than relying on a successful response as feature
detection. Full semantic conformance cannot be inferred from schema validity.

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
