# Proposed Parity trace profile

This branch is a draft for client review, not adopted RPC policy. It proposes the nine
traditional trace methods and three output families. A client that serves the trace
namespace implements all nine methods (H01). Geth support is not assumed.

The companion [trace-interop project](https://github.com/banteg/trace-interop) retains
client observations, reproducible cases and per-client impact reports.
Observed agreement is not a correctness oracle. Recommendations below are proposals.
The target is a useful, precise contract; historical implementations explain compatibility
costs but do not decide it. Intentional departures are called out below.

## Error responses

Error codes named in this profile are recommended; conformance requires the error response itself.
Where the profile requires an error, a client returns one complete JSON-RPC error response, never a
result, `null`, `[]`, a clamped or partial result, or a malformed or aborted response. What is
required is the behavior: whether a request executes or is rejected, whether an answer is an error,
`null` or `[]`, and which requests are rejected at all. A rejection is for the request's own
violation and reports it: a transaction signed for another chain is rejected for its chain identity,
not for insufficient funds of the different sender that its signature then recovers. Where several
rules are violated, the precedence stated below decides which violation a rejection reports. The
recommended codes reuse the standard JSON-RPC and `eth_simulateV1` codes, the `eth_sendRawTransaction`
error groups and the pruned-history code, so callers can branch without parsing messages, but a different code is not a
conformance failure. Only the JSON-RPC 2.0 base protocol is exempt: a request that is not valid JSON
returns -32700 (Parse error), a method a client does not serve returns -32601 (Method not found), and
an internal error (-32603) reports a server failure, not a rejection, so it never serves as the
required error. Malformed parameters recommend the standard -32602 (Invalid params) without requiring
it, because clients reject some malformed input through their own validation paths.

## Explicit choices in this draft

- The nine methods are one contract, with no optional subset or discovery mechanism: a client
  that serves the trace namespace implements every method. A method a client does not serve
  returns -32601, never an empty result (H01).
- An explicit `null` for an optional parameter or an optional member of a parameter object is the
  same as omitting it: the default applies. Clients already treat `null` this way in `eth_call`,
  `eth_estimateGas` and `eth_getLogs` fields. `null` keeps its own meaning only where the schema
  gives it one: a null `to` creates a contract (H14).
- Query path inputs use hex quantities for caller compatibility; traceAddress outputs retain JSON
  integers. Convert each output integer to a minimal hex quantity before passing it to trace_get.
  Integer path entries are rejected with an error (-32602 recommended), rather than null.
- Address filters use OR within each list and AND between lists, as `eth_getLogs` composes topic
  positions. Missing, null and empty lists are unrestricted. Addresses compare by bytes, irrespective
  of hex letter case. Optional `mode` accepts `intersection` (the default, also when `mode` is
  null) and `union` (match either populated list); other values are rejected (-32602 recommended).
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
- Where clients already agree, the draft keeps their answer (H06). Unknown transactions and valid but
  absent tree paths return null, as Parity did and as `eth_getTransactionByHash` does. Unknown blocks
  in `trace_block` and `trace_replayBlockTransactions` return null or an error, preserving both
  established client conventions. An unknown simulation block always returns an error, as `eth_call`
  does. A successful collection never represents an unknown block. -32001 (Resource not found) is recommended:
  in `trace_call` and `trace_callMany` clients use -32000 for transaction-validation failures, so a
  dedicated code lets callers tell an unknown block from an invalid call without parsing messages. Known blocks with pruned required state return an error, with 4444 recommended. If pruned
  indexing prevents establishing whether a hash is absent, return an error (4444 recommended) rather than claiming a
  definitive not-found result. This extends the pruned-history code adopted for eth/debug methods in
  [#636](https://github.com/ethereum/execution-apis/pull/636) to trace methods and execution state;
  that extension remains a proposal.
- `trace_filter` ranges follow `eth_getLogs`
  ([#875](https://github.com/ethereum/execution-apis/pull/875)): if either bound resolves beyond
  the current head, or `fromBlock` resolves above `toBlock`, return an error (-32602, Invalid params,
  recommended). Never clamp the range or return a partial result. Bounds use `BlockNumberOrTagForRange`, which excludes
  `pending` and block hashes. Rejecting hash bounds and EIP-1898 block objects is a choice: Parity,
  Erigon and Nethermind accept them, but `eth_getLogs` range bounds do not; use a number or the optional
  exact-block member below. `earliest` is the lowest block the client has available, as the shared tag defines it;
  an explicit number below retained history returns an error (4444 recommended).
- Omitted filter bounds both mean `latest`, resolved against the same canonical head for the
  request. A `toBlock` earlier than the omitted `fromBlock` is a reversed range and is rejected;
  callers searching history must state `fromBlock`. This follows Parity’s original
  `trace_filter` default and the `eth_getLogs` convention. Query limits produce an explicit error,
  never an incomplete success. `count` limits returned records, not scan or replay work.
  [H30](https://github.com/banteg/trace-interop/blob/main/reports/decisions/H30.md)
- `trace_filter` may implement a `blockHash` member, using the `eth_getLogs` selector shape.
  An unsupported non-null member is explicitly rejected (-32602 recommended), never silently ignored.
  Null means omission even on clients without support (H14). A non-null hash is mutually exclusive
  with non-null bounds (-32602 recommended). Accepted results, including `[]`, describe exactly that
  block, with its own hash on each record; matching, ordering and pagination retain their usual rules.
  Accurate noncanonical results are allowed where the required body, branch state and execution
  environment are available. Orphan support and retention are optional. Unknown, unexecuted or
  otherwise unavailable hashes return an error (-32001 recommended; 4444 for pruned required history),
  never null, `[]`, partial results or another block's records. Validate the selector before any
  `count: 0` shortcut and preserve identity in one coherent view, rather than resolving a hash to a
  height and scanning a replacement. This exact-block-or-error proposal is narrower than EIP-234's
  orphan-serving contract. Universal support remains a separate portability question (H33).
- Existing correct hash-string and EIP-1898 selector forms on other trace methods remain optional
  extensions beyond the baseline selector schemas. An accepted selector must preserve its requested
  identity. Accepting `requireCanonical: true` must enforce canonicality or reject the request;
  ignoring it is not grandfathered. Method-specific legacy acceptance does not require every client
  to implement every selector form (H33).
- `pending` selects the client's pending block: the next block number, built on `latest` with the
  client's pending transactions. `trace_block`, `trace_replayBlockTransactions`, `trace_call` and
  `trace_callMany` accept it only when the client has such a pending environment; a simulation then
  runs on the pending block's post-state and environment. A client without one rejects it
  (-32602 recommended) rather than evaluate `latest` or another block. Records traced from the pending block carry its number and
  the hash of the block the client built, which is not canonical; numeric and tag selections otherwise
  carry canonical block hashes. Optional exact hash selections may return accurate noncanonical
  records carrying the requested block's own hash (H33). `trace_filter` bounds reject `pending`,
  with -32602 recommended (H32).

## Blocks, rewards and pagination

The genesis block has no transaction or reward records: `trace_block(0)` and
`trace_replayBlockTransactions(0)` return `[]`, and filters over genesis contribute nothing. In
every other block, rewards follow all transaction records of that block: the block reward, then
uncle rewards in ommer order. A reward matches `toAddress` by its `author` and has no sender side,
so it is excluded when `fromAddress` is populated in intersection mode, even when other records of
that block match `fromAddress`; in union mode a `toAddress` match suffices. `trace_filter` filters
first, then applies `after` and `count` in block, transaction, preorder, then reward order.
`after` and `count` are JSON integers bounded by uint64. Offsets are stable only for a fixed,
numerically resolved range.

System operations (EIP-4788 and EIP-2935 pre-block calls, EIP-7002 and EIP-7251 post-block
calls), withdrawals and rewards are applied to replay state in protocol order, but they are not call
trace records and belong to no transaction’s `stateDiff`. Concatenated transaction diffs therefore do
not reconstruct the block post-state. Each block is traced from its parent’s post-block state with
its own pre-transaction system operations applied, under its own fork rules, whether a request
covers one block or a range. Replay arrays hold exactly one envelope per transaction.

## Simulation requests

Call objects use the `eth_simulateV1` `GenericCallTransaction` fields, together with `chainId` and
`authorizationList` from `GenericTransaction` and `data` as an alias for `input`. When `data` and
`input` are both present they must be equal, or the request is rejected (-32602 recommended). Every
field the schema defines either takes effect with its `eth_call` meaning or causes a rejection
(recommended: -32602, or -32003 when the selected fork does not support the field); only fields
outside the schema are ignored. A supplied nonce is accepted but neither validated nor
used, so CREATE addresses derive from the state nonce. `gas` is a uint64.

Fees follow `eth_call` and `eth_simulateV1` (H15). Omitted fee fields default to zero. The zero-fee
rule applies to the effective gas price after defaulting: a zero price means GASPRICE 0 and BASEFEE
0. BLOBBASEFEE is 0 exactly when `maxFeePerBlobGas` is supplied as 0 or defaulted to 0; calls without blob
fields keep the block’s BLOBBASEFEE. Fee validation is skipped only when both fee caps are zero.
Positive prices are validated against the base fee, funded and charged, with refunds, base-fee burn
and tips simulated. This removes Parity’s virtual balance top-up for unsigned calls.

Parameter positions for overrides are reserved: `trace_call` takes `StateOverrides` fourth and
`BlockOverrides` fifth; `trace_callMany` takes them third and fourth. Both use the `eth_simulateV1`
schemas. A client that does not implement them must reject a non-null value (-32602 recommended). With a
block override, the zero-fee rule uses the overridden base fee.

In `trace_callMany` every item is a separate transaction: EIP-2200 and EIP-3529 original storage
values, EIP-2929 access sets, EIP-1153 transient storage, the refund counter and EIP-6780’s
same-transaction scope start afresh. All items share one block environment, so NUMBER and TIMESTAMP
do not advance. The first item runs against the same state and environment as `trace_call` at the
selected block, and each diff is relative to the preceding item’s post-state. If any decoded item
fails validation, the request returns one error whose `error.data.index` is the zero-based item index,
and no partial results. The single error without partial results follows Erigon, Reth and Parity.
`error.data.index` is new in this profile: no client reports the index in structured form today
(Erigon names it only in message text), and it lets callers locate the failing item without matching
text. A malformed item is rejected without an index (-32602 recommended). Servers may cap items or
total gas with an explicit error (-38026 recommended), never by truncating.

Validation rejections of `trace_call` and `trace_callMany` are errors that report their violation.
The recommended codes are those of `eth_simulateV1`: -38012 base fee too low, -38013 intrinsic gas,
-38014 insufficient funds, -38025 init-code size and -38026 client limit. A priority fee above the
fee cap, and any other defect that makes the call object invalid regardless of state (a transaction
type or field combination no transaction can carry, an empty authorization list), takes precedence
when several rules are violated: the rejection reports that defect, with -32602 recommended. Any
other validation failure without a listed code, such as a blob fee cap below the blob base fee or a
transaction type not active at the selected fork, recommends -32003 (Transaction rejected), the
fallback `trace_rawTransaction` also uses. A supplied nonce is not validated and unsigned calls skip
the EIP-3607 sender-code check, as `eth_call` does, so the nonce and sender-not-EOA codes do not
apply. Omitted gas follows the client's `eth_call` default at the selected state, bounded by the
server's execution cap. A block limit or sender allowance may lower the budget but never raise it
above the cap. Supplied gas above the cap runs with the cap. Supply gas explicitly for a portable
budget; the cap is server policy, not a truncated result.
`trace_rawTransaction` rejections recommend the `eth_sendRawTransaction` error groups
([#650](https://github.com/ethereum/execution-apis/pull/650)): 1 nonce too low, 2 nonce too high,
800 intrinsic gas, 804 priority fee above fee cap, 806 fee cap below base fee and 809 insufficient
funds, with -32003 (Transaction rejected) as the generic fallback. A signed gas limit above the
server's execution cap is not reduced, because that would change the transaction: it is rejected
(-38026 recommended). Malformed input is rejected (-32602 recommended).

## Execution results and state changes

Frame numbers follow Parity. The root `action.gas` is the transaction gas minus intrinsic gas,
including access-list and authorization costs. The root `result.gasUsed` is execution gas before
refunds, excluding intrinsic gas and the EIP-7623 floor. A nested CALL-family `action.gas` is the gas
forwarded after the 63/64 cap plus the 2300 stipend for a value transfer. CREATE `gasUsed` includes
the code deposit, and an exceptional halt consumes `action.gas`. CALL and STATICCALL report the
caller as `from` and the target as `to`; DELEGATECALL and CALLCODE report the executing address and
the code address. DELEGATECALL `value` is the inherited value and STATICCALL `value` is 0. Create
actions require `creationMethod` (`create` or `create2`). A `suicide` frame records the opcode and
its transfer, not account deletion. A CALL-family call that fails its precheck (depth limit,
insufficient balance) emits a frame with its action and error, no `result` and no subtraces; the
forwarded gas returns to the caller, which continues after pushing 0. A CREATE or CREATE2 that fails
its precheck (depth limit, insufficient balance, nonce overflow) emits the same kind of frame before
Amsterdam; from Amsterdam the check runs in the creating opcode and no frame is emitted. This follows
the call tracers of Geth, Erigon, Reth and Besu, which record the rejected attempt. A CREATE address
collision emits a create frame with error `Contract address collision`.

`error` alone determines failure. A REVERT frame has error `Reverted` and requires
`result: {gasUsed, output}`, for CREATE too, without an address or code. An exceptional halt may
omit `result` or set it to null. Erigon and Reth already emit REVERT results; the new part is the
failed-CREATE shape, which must not report the would-be address. No contract exists at that
address, which is also why a failed CREATE has no created-address match in `trace_filter` (H23).
Failure labels are `Reverted`, `Out of gas`, `Bad instruction`, `Bad jump destination`,
`Stack underflow`, `Out of stack`, `Mutable Call In Static Context`, `Built-in failed` and
`Out of bounds`. As in Parity and OpenEthereum, a code-deposit failure, including code above the
EIP-170 size limit, is `Out of gas`, and returned code starting with 0xEF (EIP-3541) is
`Invalid code`. Beyond Parity, which reported an address collision as `Out of gas` and emitted no
frame for a failed precheck, the profile distinguishes `Contract address collision`,
`Nonce overflow`, `Insufficient balance for transfer` and `Max call depth exceeded`. EIP-3860
oversized initcode aborts the creating frame with `Out of gas`, as the EIP specifies. The
designated invalid instruction 0xFE is an undefined opcode: it is omitted from `vmTrace` ops and its
frame fails with `Bad instruction`. Other strings are extensions that consumers treat as generic
failure. `revertReason` is not part of this profile; raw revert bytes live in `result.output`.
The execution envelope carries root output, which cannot substitute for a nested
frame’s revert bytes. A locally successful child remains successful even if an ancestor later
reverts; its state changes then do not survive in `stateDiff`.

`stateDiff` compares execution endpoints, not intermediate writes. An account appears only if its
balance, nonce, code, storage or existence changed. Existence means presence in the state trie, so
an existing EIP-161-empty account removed by touch-clearing is a deletion. `+` and `-` appear only
when the account is born or dies. Slots of an account that exists at both endpoints use `*` with
32-byte words, including zero words; slots never use `=`. A born account lists its nonzero
post-state slots as `+`. A deleted account may report `storage: {}` or accurate optional old-slot
values with `-`; account deletion implies that all storage is wiped in either case. OpenEthereum
3.3.5 reports such entries when earlier uncommitted calls wrote a slot. Enumeration is optional
and callers must not treat the listed slots as exhaustive. Multiple EIP-7702 authorizations may restore the
original code while changing nonce; only net code changes appear. Accepted authorization effects
survive a later EVM revert. Accounts absent at both endpoints have no account diff even if created
and destroyed. The sender pays value, `gasUsed` times the effective price and `blobGasUsed` times
the blob base fee; base and blob fees are burned or credited to a chain-defined collector, and the
tip goes to the fee recipient.

VM `mem` is the post-operation contents of the memory range the opcode’s operands designate:
MSTORE and MLOAD `[off, off+32)`, MSTORE8 `[off, off+1)`, the destination of CALLDATACOPY,
CODECOPY, EXTCODECOPY, RETURNDATACOPY and MCOPY; for the CALL family, starting at the output offset,
either the full output window (preferred) or exactly the bytes copied from return data. It is null
when that range is empty or the opcode has none (RETURN, REVERT, LOG, KECCAK256, CREATE). This
matches Parity, Erigon, Nethermind and Besu for MLOAD. The CALL-family range may be the full window
(Parity, Erigon) or the copied prefix (Besu, revm-inspectors): both reconstruct memory exactly when
applied as a write, and no known consumer distinguishes them. Consumers must not infer returned-data
length from `mem`. `cost` is the total gas deducted from the caller when the operation starts,
including gas made available to a child frame (the 63/64-capped forwarded gas for the CALL family,
excluding the value stipend; all but 1/64 for the CREATE family), also when the call or creation
fails its balance or depth precheck and that gas is returned immediately. `ex.used` is the caller’s
remaining gas after the operation, including unused child gas returned. An operation that began
executing and halted exceptionally has `ex: null` and `sub: null`; its `cost` is an
implementation-defined non-negative integer, since where an interpreter detects the fault varies
and a halted frame returns no gas.
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

Filter membership describes actions, not necessarily committed transfers. A record matches only
addresses it reports. A failed CREATE reports no created address, because its result has no address
(H09), so it can match its creator but never a populated recipient list, even when a client knows
the would-be address; an explicit union may still retain the creator match. SELFDESTRUCT matches the
executing (self-destructing) account and the beneficiary; after EIP-6780 the account usually
survives. Range results must match
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
delegation exception. Each rejection reports the rule the transaction violates: a
transaction signed for another chain fails chain identity, whatever its recovered sender
holds. Reject both low and high nonces without modifying signed fields
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

The value of client execution caps; the shared semantics of a raw-transaction block selector
(H12); and simulation extensions beyond the reserved override positions need further agreement.
Clients must declare supported extensions rather than relying on a successful response as feature
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
