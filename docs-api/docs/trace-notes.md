# trace_* notes

This document contains extra information that couldn't be fit to the specification document
directly: rules shared by several trace methods, and the reasons behind them.

The specification covers the nine traditional trace methods and their three output families: call
traces (`trace`), state changes (`stateDiff`) and VM traces (`vmTrace`). The target is a useful,
precise contract; historical implementations explain compatibility costs but do not decide it.
Intentional departures from Parity are called out below.

## Methods and parameters

- The nine methods are one contract, with no optional subset or discovery mechanism: a client
  that serves the trace namespace implements every method. A method a client does not serve
  returns -32601, never an empty result.
- An explicit `null` for an optional parameter, or for an optional member of a call object or a
  `trace_filter` object, is the same as omitting it: the default applies. Objects reused from
  `eth_call` and `eth_simulateV1`, such as block selectors and block overrides, keep their own
  schemas. Clients already treat `null` this way in `eth_call`, `eth_estimateGas` and `eth_getLogs`
  fields. `null` keeps its own meaning only where the schema gives it one: a null `to` creates a
  contract.
- Empty trace selection is valid. The envelope always preserves output.
- Client execution caps are server policy. This specification does not define a block selector for
  `trace_rawTransaction`, or simulation parameters beyond the override positions. Clients must
  document the additional parameters and fields they support rather than rely on a successful
  response as feature detection. Full semantic conformance cannot be inferred from schema validity.

## Error responses

Error codes named in this specification are recommended, except -32001 for an unknown block, which
is required; conformance requires the error response itself. Where the specification requires an
error, a client returns one complete JSON-RPC error response, never a result, `null`, `[]`, a
clamped or partial result, or a malformed or aborted response. What is required is the behavior:
whether a request executes or is rejected, whether an answer is an error, `null` or `[]`, and which
requests are rejected at all. A rejection is for the request's own violation and reports it: a
transaction signed for another chain is rejected for its chain identity, not for insufficient funds
of the different sender that its signature then recovers. Where several rules are violated, the
precedence stated below decides which violation a rejection reports. The recommended codes reuse
the standard JSON-RPC and `eth_simulateV1` codes, the `eth_sendRawTransaction` error groups and the
pruned-history code, so callers can branch without parsing messages, but a different code is not a
conformance failure. An unknown block, by number, tag or hash, in any method is -32001 (Resource not
found). Only the JSON-RPC 2.0 base protocol is exempt: a request that is not valid JSON returns
-32700 (Parse error), a method a client does not serve returns -32601 (Method not found), and an
internal error (-32603) reports a server failure, not a rejection, so it never serves as the
required error. Malformed parameters recommend the standard -32602 (Invalid params) without
requiring it, because clients reject some malformed input through their own validation paths.

EVM failure is trace output; invalid parameters or unavailable state are RPC errors.

### Response integrity

Before result bytes are committed, validation or execution setup failures must produce one
complete JSON-RPC error, not a partially serialized result. Buffering may be necessary to preserve
that boundary: parsing all `trace_callMany` inputs does not validate later calls against earlier
calls' state. If a failure occurs after output is committed, explicitly abort the transport rather
than append a second error envelope or finish a partial result as success. An aborted request is
incomplete, not a successful trace or a complete JSON-RPC error. Tests must distinguish deliberate
transport termination from malformed completed responses.

## Storage and resource policy

Implementations may generate traces on demand or use stored traces or indexes. This specification
does not require additional indexes, trace persistence, archive retention, or particular latency,
throughput, or minimum query-range guarantees. Historical replay still requires the block data and
suitable execution pre-state, or the ability to reconstruct that state.

Clients may impose implementation-defined limits on query ranges, output size, execution work, and
duration. Unavailable required history or a resource limit that prevents completing the requested
result produces an explicit error, never a silently narrowed selection or an incomplete success.
Explicit pagination through `count` and `after` retains its defined meaning.

## Block selection

- Where clients already agree, this specification keeps their answer. Unknown transactions and
  valid but absent tree paths return null, as Parity did and as `eth_getTransactionByHash` does:
  `trace_replayTransaction`, `trace_transaction` and `trace_get` return null for a transaction that
  is not in a canonical block. Unknown blocks return an error in every method, as an unknown
  simulation block does in `eth_call`; a successful collection or null never represents an unknown
  block. The code is -32001 (Resource not found): in `trace_call` and `trace_callMany` clients use
  -32000 for transaction-validation failures, so a dedicated code lets callers tell an unknown block
  from an invalid call without parsing messages. Known blocks with pruned required state return an
  error, with 4444 recommended. For a known transaction, unavailable history or state is an error,
  not null. If pruned indexing prevents establishing whether a hash is absent, return an error (4444
  recommended) rather than claiming a definitive not-found result. This extends the pruned-history
  code adopted for eth/debug methods in
  [#636](https://github.com/ethereum/execution-apis/pull/636) to trace methods and execution state.
- `trace_call` and `trace_callMany` default to latest post-state and accept a block number, tag or
  hash. A block hash, as a string or an object, selects that block even when it is not canonical,
  as in `eth_call`; with `requireCanonical: true` a noncanonical hash returns an error (-32001).
  `trace_replayBlockTransactions` and `trace_block` require a block number or tag; there is no
  implicit default block.
- `trace_rawTransaction` uses latest state and takes two parameters. A client may also accept an
  explicitly supplied block selector; this specification does not define its meaning, and
  accepting it is not a conformance failure. The third selector existed in Parity's
  implementation, although its guide omitted it.
- Correct hash-string and block selector object (`blockHash` or `blockNumber`) forms on other trace
  methods remain optional, beyond the selector schemas this specification defines. An accepted selector must preserve its
  requested identity. Accepting `requireCanonical: true` must enforce canonicality or reject the
  request; ignoring it is not grandfathered. Method-specific legacy acceptance does not require
  every client to implement every selector form.
- `pending` selects the client's pending block: the next block number, built on `latest` with the
  client's pending transactions. `trace_block`, `trace_replayBlockTransactions`, `trace_call` and
  `trace_callMany` accept it only when the client has such a pending environment; a simulation then
  runs on the pending block's post-state and environment. A client without one rejects it (-32602
  recommended) rather than evaluate `latest` or another block. Records traced from the pending
  block carry its number and the hash of the block the client built, which is not canonical; all
  other localized records carry canonical block hashes. `trace_filter` bounds reject `pending`, with
  -32602 recommended.

## Simulation: trace_call and trace_callMany

`trace_call` executes against the state at the end of the selected block, default latest, under its
fork rules and block environment. Simulated changes are discarded, and neither method changes
canonical state.

### Call objects

Call objects use the `eth_simulateV1` `GenericCallTransaction` fields, together with `chainId` and
`authorizationList` from `GenericTransaction` and `data` as an alias for `input`. When `data` and
`input` are both present they must be equal, or the request is rejected (-32602 recommended). Every
field the schema defines either takes effect with its `eth_call` meaning or causes a rejection
(recommended: -32602, or -32003 when the selected fork does not support the field); only fields
outside the schema are ignored. A supplied nonce is accepted but neither validated nor used, so
CREATE addresses derive from the state nonce. `gas` is a uint64.

### Fees

Ordinary execution fees follow `eth_call`. Omitted execution-fee fields default to zero. The
effective gas price is `gasPrice`, or min(`maxPriorityFeePerGas` + base fee, `maxFeePerGas`),
computed after defaulting. The zero-fee rule applies to that effective price: when it is zero,
GASPRICE and BASEFEE are both 0 and no gas fee is paid. Other block fields are unchanged. The blob
fee is decided independently of the execution fee. A blob call carries a non-empty
`blobVersionedHashes` list; when its `maxFeePerBlobGas` is omitted, null or zero, it runs with
BLOBBASEFEE 0 and pays no blob fee, so an explicit zero is accepted, matching Geth’s `eth_call`
defaults and the `eth_simulateV1` schema default of 0. A positive `maxFeePerBlobGas` is validated
against the selected block's blob base fee and charged. Calls without blob fields keep the selected
block's BLOBBASEFEE. Defined blob fields must take effect or cause rejection, never be ignored: a
`maxFeePerBlobGas` without `blobVersionedHashes` is rejected.

Execution-fee validation is skipped only when both execution fee caps are zero, as with
`gasPrice` 0. Otherwise the price must cover the base fee (-38012 recommended),
`maxPriorityFeePerGas` must not exceed `maxFeePerGas` (-32602 recommended), and the sender must fund
the upfront cost (-38014 recommended); execution then charges the upfront gas payment, applies
refunds, burns the base fee and pays the tip. A zero price does not exempt value transfers from
funding. This removes Parity's virtual balance top-up for unsigned calls.

### Gas

Omitted gas follows the client's `eth_call` default at the selected state, bounded by the server's
execution cap. Only omitted or null gas takes that default: an explicit `"gas": "0x0"` is a supplied
limit of 0, which fails the intrinsic-gas check, so the call is rejected (-38013 recommended) with
no trace, and in `trace_callMany` the whole request errors as for any invalid item. This matches
`eth_simulateV1` and every client's `eth_call` except Erigon's, which treats 0 as omitted. A block
limit or sender allowance may lower the budget but never raise it above the cap. Supplied gas above
the cap runs with the cap. Callers requiring a portable budget supply gas explicitly; the cap is
server policy, not a truncated result.

### Validation and error codes

Validation rejections of `trace_call` and `trace_callMany` are errors that report their violation.
The recommended codes are those of `eth_simulateV1`: -38012 base fee too low, -38013 intrinsic gas,
-38014 insufficient funds, -38025 init-code size and -38026 client limit. Malformed requests are
rejected (-32602 recommended). A priority fee above the fee cap, and any other defect that makes the
call object invalid regardless of state (a field combination no transaction can carry, an empty
authorization list), takes precedence when several rules are violated: the rejection reports that
defect, with -32602 recommended. `gasPrice` with `maxFeePerGas` or `maxPriorityFeePerGas` is such a
combination. A legacy `gasPrice` otherwise serves as both `maxFeePerGas` and
`maxPriorityFeePerGas`, also with an `authorizationList`, whose authorizations apply, or with blob
fields, whose blob fee is priced on its own. For transaction types 0x0 to 0x4, `type` never changes
a simulated call: the fields present decide, no supplied field is dropped because of it, and the
type adds no requirement of its own. An access list before Berlin, `maxFeePerGas` or
`maxPriorityFeePerGas` before London, blob fields before Cancun and an `authorizationList` before
Prague are rejected (-32003 recommended), even with zero or empty values, since no transaction at
the selected fork can carry them: before London only `gasPrice` priced a transaction, so dynamic
fees are neither ignored nor reinterpreted there. A `maxFeePerBlobGas` without
`blobVersionedHashes`, or an empty `blobVersionedHashes` list, is rejected at every fork, since a
blob transaction carries at least one versioned hash. Methods that build or sign a transaction, and
other type values, which chains define as extensions, are outside this rule. Any other validation
failure without a listed code, such as a positive blob fee cap below the blob base fee, recommends
-32003 (Transaction rejected), the fallback
`trace_rawTransaction` also uses. A supplied nonce is not validated and unsigned calls skip the
EIP-3607 sender-code check, as `eth_call` does, so the nonce and sender-not-EOA codes do not apply.

### Overrides

Overrides are optional: a client need not accept either form below. The preferred form is
positional: `trace_call` takes `StateOverrides` fourth and `BlockOverrides` fifth; `trace_callMany`
takes them third and fourth, using the `eth_simulateV1` field schemas, with storage optional for
balance/code/nonce-only overrides and `state`/`stateDiff` mutually exclusive. Existing recognized
configuration wrappers may also be accepted where their contents actually take effect. Reject
unsupported non-null override shapes (-32602 recommended), never silently treat a state map as an
empty configuration. Accepted overrides apply before execution, once before the first batch item.
Block overrides replace fields of the selected block environment; with a block override, the
zero-fee rule uses the overridden base fee.

### trace_callMany

Each item is a separate transaction executed in order on the preceding item's post-state, so all
transaction-scoped execution state starts afresh for each item. All items share one block
environment, as selected for `trace_call`, so NUMBER and TIMESTAMP do not advance. The first item
runs against the same state and environment as `trace_call` at the selected block, and each diff is
relative to the preceding item's post-state. The `trace_call` fee rules, including the zero-fee
BASEFEE and blob-fee BLOBBASEFEE adjustments, apply to each item independently, as do its gas rules.
Reverted writes roll back and subsequent items continue; transaction-level nonce and fee effects
survive. The result holds one envelope per input, in order; empty input returns `[]`.

If any decoded item fails validation, the request returns one error, with the `trace_call` codes
recommended, and no partial results. A client SHOULD include the failing item's zero-based index in
`error.data.index`; if present, the index MUST identify that item. The single error without partial
results follows Erigon, Reth and Parity. `error.data.index` is new in this specification: Parity and
Erigon do not report the index in structured form (Erigon names it only in message text). It is
recommended diagnostics, so its absence is not a conformance failure; when supplied it lets callers
locate the failing item without matching text. A malformed item is rejected without an index
(-32602 recommended). Servers may cap the item count or total gas only with an explicit error
(-38026 recommended), never by truncating.

## Signed transactions: trace_rawTransaction

`trace_rawTransaction` decodes the signed transaction and validates it for execution against latest
state before simulating it, under the transaction validity rules of the selected fork: signature,
chain identity, nonce equality, balance for value and upfront gas, intrinsic gas, fees and the
selected fork's sender-code restrictions, including EIP-7702's delegation exception. Each rejection
reports the rule the transaction violates: a transaction signed for another chain fails chain
identity, whatever its recovered sender holds. Reject both low and high nonces without modifying
signed fields or implicitly changing the sender's nonce or balance. A valid transaction that REVERTs
or exceptionally halts, including running out of execution gas, returns an execution trace, not a
validation error. This is execution validation, not transaction-pool admission: local pool policies
such as replacement pricing, already-known rejection and minimum tips do not apply. Full block
admissibility is not established by a successful simulation. Authorization tuples skipped under
EIP-7702 do not themselves make the outer transaction invalid; validation must follow the selected
fork's distinction between rejected transactions and skipped tuples.

Rejections recommend the `eth_sendRawTransaction` error groups
([#650](https://github.com/ethereum/execution-apis/pull/650)): 1 nonce too low, 2 nonce too high,
800 intrinsic gas, 804 priority fee above fee cap, 806 fee cap below base fee and 809 insufficient
funds, with -32003 (Transaction rejected) as the generic fallback for any other validation failure.
A signed gas limit above the server's execution cap is not reduced, because that would change the
transaction: it is rejected (-38026 recommended). Malformed encodings or request parameters are
rejected (-32602 recommended).

This deliberately tightens legacy permissive simulation: queued high-nonce transactions and
already-mined low-nonce transactions may stop tracing at latest state. Historical OpenEthereum
skipped nonce checks and could supplement balance. Using a state nonce different from the signed
nonce can change CREATE's execution address; overriding state instead invents a pre-state. This
specification chooses neither behavior. Hypothetical execution belongs in explicitly documented
simulation facilities. Erigon's
[feedback](https://github.com/ethereum/execution-apis/issues/890#issuecomment-5784143408) supports
strict validation despite the compatibility cost. Malformed
JSON on rejection is independently a reporting defect.

## Replay and block traces

`trace_replayTransaction` reconstructs the state immediately before the transaction from the parent
block's post-block state, applying the block's pre-transaction system operations and then its
preceding transactions, under the block's fork rules. It reports the transaction's actual committed
state transition.

Each block is traced from its parent's post-block state with its own pre-transaction system
operations applied, under its own fork rules, whether a request covers one block or a range. System
operations (EIP-4788 and EIP-2935 pre-block calls, EIP-7002 and EIP-7251 post-block calls),
withdrawals and rewards are applied to replay state in protocol order, but they are not call trace
records and belong to no transaction's `stateDiff` or replay envelope. Concatenated transaction
diffs therefore do not reconstruct the block post-state. Replay arrays hold exactly one envelope per
transaction.

The genesis block has no transaction or reward records: `trace_block(0)` and
`trace_replayBlockTransactions(0)` return `[]`, as does `trace_replayBlockTransactions` for an empty
block, and filters over genesis contribute nothing. In every other block, rewards follow all
transaction records of that block: the block reward, then uncle rewards in ommer order. Reward
records are the block's historical PoW rewards; do not emit synthetic zero PoW rewards after the
Merge.

## trace_filter

Indexed retrieval and replay must yield the same fork-correct per-block records, each block
traced from its parent's post-block state.

### Block ranges

- `trace_filter` ranges follow `eth_getLogs`
  ([#875](https://github.com/ethereum/execution-apis/pull/875)): if either bound resolves beyond
  the current head, or `fromBlock` resolves above `toBlock`, return an error (-32602, Invalid
  params, recommended). Never clamp the range or return a partial result. Bounds use
  `BlockNumberOrTagForRange`: numbers and the `earliest`, `latest`, `safe` and `finalized` tags,
  excluding `pending`. A block-hash string or block-hash object as `fromBlock` or `toBlock` is
  rejected (-32602 recommended), as `eth_getLogs` range bounds reject them. Rejecting hash bounds is
  a choice: Erigon and Nethermind accept both forms today. Parity accepted only the block-hash
  object, never a bare hash string, and resolved it to a number before scanning its canonical trace
  index, so its hash bounds never pinned a block. To pin one block by hash, use the `blockHash`
  member below. An unresolvable `safe` or `finalized` tag returns an error, never head substitution
  or an empty result. `earliest` is the lowest block the client has available, as the shared tag
  defines it; an explicit number below retained history returns an error (4444 recommended).
- Omitted filter bounds both mean `latest`, resolved against the same canonical head for the
  request. A `toBlock` earlier than the omitted `fromBlock` is a reversed range and is rejected;
  callers searching history must state `fromBlock`. This follows Parity's original `trace_filter`
  default and the `eth_getLogs` convention. Query limits produce an explicit error, never an
  incomplete success. `count` limits returned records, not scan or replay work.
- A request captures one coherent snapshot; numeric and tag ranges select its canonical blocks.

### Selecting a block by hash

`trace_filter` takes a `blockHash` member (a 32-byte hash), as `eth_getLogs` does under EIP-234. It
is part of the method contract, like the other filter members: callers may omit it, and clients
implement it rather than reject or ignore it. A non-null `blockHash` is mutually exclusive with
non-null `fromBlock` and `toBlock` (-32602 recommended); an explicit null for any of the three is
the same as omitting it. It selects exactly the block with that hash: the result, including `[]`,
is for that block and each record carries its hash. Address matching, `mode` (including reward
matching), ordering and `after`/`count` then apply as for a single-block range. The block must be
canonical in the request's chain view and executed: an unknown, noncanonical, not-yet-executed or
otherwise unavailable hash returns an error (-32001; 4444 recommended for pruned history), never
`[]` or another block's records. Resolve the hash and trace the block in one chain view, never
resolving it to a number and scanning that number in a newer view, and validate the selector before
any `count: 0` shortcut. This narrows EIP-234, which also serves noncanonical blocks: serving a side
chain needs its body, its parent's state and its branch environment, which canonical history
retention does not provide. Side-chain selection is left to possible later work.

### Address matching

Address filters use OR within each list and AND between lists, as `eth_getLogs` composes topic
positions. Missing, null and empty lists are unrestricted. Addresses compare by bytes, irrespective
of hex letter case. Optional `mode` accepts `intersection` (the default, also when `mode` is null)
and `union` (match either populated list); other values are rejected (-32602 recommended).

Filter membership describes actions, not necessarily committed transfers. CALL matches its sender
and recipient, CREATE its creator and created address, and SELFDESTRUCT the executing
(self-destructing) account and the beneficiary; after EIP-6780 the account usually survives. A
record matches only addresses it reports. A failed CREATE reports no created address, because its
result has no address, so it can match its creator but never a populated recipient list, even when
a client knows the would-be address; an explicit union may still retain the creator match. A reward
matches `toAddress` by its `author` and has no sender side, so it is excluded when `fromAddress` is
populated in intersection mode, even when other records of that block match `fromAddress`; in union
mode a `toAddress` match suffices.

### Ordering and pagination

Records are ordered by block, then transaction, then preorder, with each block's rewards after its
transaction records (block reward, then uncle rewards in ommer order). `trace_filter` filters
first, then skips `after` records and takes `count`. `after` and `count` are JSON integers bounded
by uint64. Offsets are stable only for a fixed, numerically resolved range.

## Call frames

Frame numbers follow Parity. The root `action.gas` is the transaction gas minus intrinsic gas,
including access-list and authorization costs. The root `result.gasUsed` is execution gas before
refunds, excluding intrinsic gas and the EIP-7623 floor. A nested CALL-family `action.gas` is the gas
forwarded after the 63/64 cap plus the 2300 stipend for a value transfer. CREATE `gasUsed` includes
the code deposit, and an exceptional halt consumes `action.gas`. CALL and STATICCALL report the
caller as `from` and the target as `to`; DELEGATECALL and CALLCODE report the executing address and
the code address. DELEGATECALL `value` is the inherited value and STATICCALL `value` is 0. Create
actions require `creationMethod` (`create` or `create2`). A `suicide` frame records the opcode and
its transfer, not account deletion. A CALL-family call that fails its precheck (depth limit,
insufficient balance) emits a frame with its action and error, `result: null` and no subtraces,
unless it is a nested zero-value precompile call, which is omitted like any other; the forwarded
gas returns to the caller, which continues after pushing 0. A CREATE or CREATE2 that fails its
precheck (depth limit, insufficient balance, nonce overflow) emits the same kind of frame before
Amsterdam; from Amsterdam the check runs in the creating opcode and no frame is emitted. This
follows the call tracers of Geth, Erigon, Reth and Besu, which record the rejected attempt. A
CREATE address collision emits a create frame with error `Contract address collision`.

`error` alone determines failure. A REVERT frame has error `Reverted` and requires
`result: {gasUsed, output}`, for CREATE too, without an address or code. An exceptional halt has
`result: null`. Erigon and Reth already emit REVERT results; the new part is the failed-CREATE
shape, which must not report the would-be address. No contract exists at that address, which is
also why a failed CREATE has no created-address match in `trace_filter`. Failure labels are
`Reverted`, `Out of gas`, `Bad instruction`, `Bad jump destination`, `Stack underflow`,
`Out of stack`, `Mutable Call In Static Context`, `Built-in failed` and `Out of bounds`. As in
Parity and OpenEthereum, a code-deposit failure, including code above the EIP-170 size limit, is
`Out of gas`, and returned code starting with 0xEF (EIP-3541) is `Invalid code`. Beyond Parity,
which reported an address collision as `Out of gas` and emitted no frame for a failed precheck, this
specification distinguishes `Contract address collision`, `Nonce overflow`,
`Insufficient balance for transfer` and `Max call depth exceeded`. EIP-3860 oversized initcode
aborts the creating frame with `Out of gas`, as the EIP specifies. The designated invalid
instruction 0xFE is an undefined opcode: it is omitted from `vmTrace` ops and its frame fails with
`Bad instruction`. Clients may report other strings, which consumers treat as generic failure.
`revertReason` is not defined by this specification; raw revert bytes live in `result.output`. The
execution envelope carries root output, which cannot substitute for a nested frame's revert bytes.
A locally successful child remains successful even if an ancestor later reverts; its state changes
then do not survive in `stateDiff`.

Standalone historical records (`trace_block`, `trace_filter`, `trace_transaction`, `trace_get`)
require non-null block localization. Mined transaction frames also require non-null transaction
hash/index; protocol rewards require null transaction hash/index and occur only in block/filter
results. Frames inside simulation and replay envelopes omit localization and rewards; both
individual and block replay envelopes carry the transaction hash so results retain their identity
outside the request. Requiring it for individual replay deliberately extends the original Parity
response shape.

### Precompile frames

All methods that return call frames use the same inclusion rule: retain root precompile calls
regardless of value. Omit nested precompile frames with zero value, including one whose call fails
its depth or balance precheck, and retain those with nonzero transferred or inherited value,
whether successful or failed. A precompile is an address in the precompile set active at the
executing block's fork. For CALL and CALLCODE, use the explicit value operand, not the parent
transaction value. For DELEGATECALL, use the inherited call value; STATICCALL has zero value. The
inherited value exception preserves action context even though DELEGATECALL transfers no funds.
`traceAddress` and `subtraces` describe the emitted tree, so omitted frames do not consume child
ordinals. This also determines the paths used by `trace_get` and the records available to
`trace_filter`. An omitted frame must not suppress its caller's VM return-memory effects. Errors
belong to the frame that failed; a handled precompile failure does not make its successful caller
fail.

This is a compatibility choice. Parity's trace database
[omitted nested built-ins to avoid denial-of-service from trace volume](https://github.com/openethereum/parity-ethereum/blob/66477a9476200aaf7e933a52dc64676154adfb2b/ethcore/src/executive.rs#L266),
then [retained actual value transfers](https://github.com/openethereum/parity-ethereum/blob/4255d4c46436d2740c4afb521b773cb95072c7cc/ethcore/src/executive.rs#L431)
while citing heavy `IDENTITY` use. The inherited-value case above follows current client behavior,
not Parity's original transfer-only rationale. JSON Schema alone cannot enforce this rule; the
companion precompile fixtures check inclusion and path numbering.

## State changes

`stateDiff` compares execution endpoints, not intermediate writes. An account appears only if its
balance, nonce, code, storage or existence changed. Existence means presence in the state trie, so
an existing EIP-161-empty account removed by touch-clearing is a deletion. `+` and `-` appear only
when the account is born or dies. Slots of an account that exists at both endpoints use `*` with
32-byte words, including zero words; slots never use `=`. A born account lists its nonzero
post-state slots as `+`. A deleted account may report `storage: {}` or accurate optional old-slot
values with `-`; account deletion implies that all storage is wiped in either case. OpenEthereum
3.3.5 reports such entries when earlier uncommitted calls wrote a slot. Enumeration is optional and
callers must not treat the listed slots as exhaustive. Multiple EIP-7702 authorizations may restore
the original code while changing nonce; only net code changes appear. Accepted authorization
effects survive a later EVM revert. Accounts absent at both endpoints have no account diff even if
created and destroyed. The sender pays value, `gasUsed` times the effective price and
`blobGasUsed` times the blob base fee; base and blob fees are burned or credited to a chain-defined
collector, and the tip goes to the fee recipient.

## VM traces

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
fails its balance or depth precheck and that gas is returned immediately. `ex.used` is the caller's
remaining gas after the operation, including unused child gas returned. An operation that began
executing and halted exceptionally has `ex: null` and `sub: null`; its `cost` is an
implementation-defined non-negative integer, since where an interpreter detects the fault varies
and a halted frame returns no gas.

Operations rejected before execution are omitted, and no synthetic STOP is added, so every `pc`
lies inside `code`. `push` is the top k stack words after execution, deepest first, where k is the
number of words the opcode leaves in place of its inputs (DUPn and SWAPn n+1, CALL and CREATE 1).
`store` is set only by SSTORE, from its operands, whenever the SSTORE completes, even if the value
is unchanged. A selected `vmTrace` is always an object. An operation that entered a child frame,
including a precompile or empty-code target, has that frame's trace as `sub`
(`{code: "0x", ops: []}` if no instruction ran); others have `sub: null`. These deltas do not
independently encode allocated memory size. Each operation carries its mnemonic as `op`, and
operations have no other members.

## Integer encoding

These methods keep Parity's wire format, which the clients that serve them and their existing
consumers use. Some integers are therefore JSON numbers rather than the hex quantities used
elsewhere in this specification: `subtraces`, `traceAddress` entries, `blockNumber` and
`transactionPosition` in localized records, the `vmTrace` `pc`, `cost`, `ex.used` and `ex.mem.off`,
and the `trace_filter` `after` and `count` members. Gas, values and stack words remain hex
quantities. Query path inputs use hex quantities for caller compatibility, while `traceAddress`
outputs retain JSON integers: convert each output integer to a minimal hex quantity before passing
it to `trace_get`. Integer path entries are rejected with an error (-32602 recommended), rather than
null.

## Schema rendering and tests

The VM schema has its own JSON Schema resource identity and a local self-reference, so nested
execution receives the same validation as the root. Build tooling preserves that bounded
self-reference while still rejecting cyclic component expansion.

`openrpc.json` retains the recursive `vmTrace` schema and is the validation artifact. The
documentation renderer currently cannot expand recursive schemas. Its generated `docs-openrpc.json`
display input replaces resource-local self references with labeled recursive object descriptions.
It also materializes shared object fields in union branches and labels array-item variants for the
renderer. Different item variants may coexist in one array; the display wrappers are not an
equivalent validation schema. This artifact must not be used for conformance validation.

`npm run docs:refresh` rebuilds the spec, refreshes this display projection, and copies the
introductory page. The watcher uses the same chain. Startup and production-build hooks only refresh
the projection so an already prepared release spec is preserved. `npm run test:docs` tests rendered
method output and the refresh chain after `make build`.

The Go trace schema tests build the real YAML and validate positive and negative cases against
reference-preserving and expanded schemas (Draft 7 and Draft 2019-09), and the actual speccheck
parsing/validation path. They include recursive VM resources, fee conflicts, contextual frame
restrictions and nonempty `trace_callMany` tuples. These are schema tests, not execution fixtures or
evidence of client semantic conformance.
