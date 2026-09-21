# Proposed Parity trace profile

This branch is a draft for client review, not adopted RPC policy. It proposes the nine
traditional trace methods and three output families. Support/conformance policy (H01)
remains a standards decision; adding these files does not assert that every client must
implement every method. Geth support is not assumed.

The companion [trace-interop project](https://github.com/banteg/trace-interop) retains
the dated client observations, reproducible cases and per-client impact reports.
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

## Decision records

### H01 — Method coverage

**Proposed rule:** Specify each method independently; make the implemented subset discoverable or explicitly declared. An unsupported method should return -32601, never a fabricated empty result. Add individual replay to Besu when practical rather than removing it from the common API.

**Rationale:** A caller should be able to distinguish an unsupported operation from a valid query with no trace.

**Review:** Agree the optional-method/conformance model at the meeting; no assumption of adopted optional status or Geth support.

### H02 — trace_get selector and return shape

**Proposed rule:** Treat the argument as one traceAddress path: [] is root, [0] first child, [6,0] a nested child. Return one trace object, or null when the transaction/path does not exist. Specify the wire encoding of indices explicitly (prefer hex quantities for compatibility with these requests).

**Rationale:** The lookup argument should identify the same path that traceAddress describes, with a singular result matching the method name.

**Review:** Choose the tree-path contract and align Reth/Nethermind; the positive nested discriminator is now available.

### H03 — Filter composition and mode

**Proposed rule:** Use OR within each address list and AND between fromAddress and toAddress. One unconstrained side should not suppress matches. If union is retained, expose it explicitly as an optional extension and reject unsupported mode values rather than silently ignoring their meaning.

**Rationale:** Supplying a second filter normally narrows a search; combining two sets of alternatives should be explicit.

**Review:** Settle boolean composition and extension validation together; one-sided and unknown-mode fixtures are available.

### H04 — Empty address lists

**Proposed rule:** Treat an omitted list or [] as no restriction. If null is accepted, give it the same meaning; document whether null is accepted at all. A supplied nonempty list restricts that side.

**Rationale:** Programmatically adding an empty optional filter should not erase all results.

**Review:** Agree whether null is legal; distinguish that schema choice from empty-array matching and special-action membership.

### H05 — Post-merge reward records

**Proposed rule:** Do not emit a synthetic PoW block reward on a PoS block. Represent real protocol balance changes only under explicitly defined semantics; keep transaction tips in transaction accounting. Define nullable localization fields separately for non-transaction records.

**Rationale:** A reward trace should correspond to a real protocol operation, not a zero-value placeholder that changes pagination.

**Review:** Specify which protocol operations belong in this namespace. Keep withdrawals and system transitions separate from transaction call traces unless explicitly included.

### H06 — Missing transactions and paths

**Proposed rule:** Return null for a well-formed lookup of an unknown transaction or trace path, consistently across these methods. Reserve [] for a successfully evaluated collection with no elements. Preserve -32601 for unsupported methods and use errors for unavailable/pruned state instead of pretending data is absent.

**Rationale:** Not found, empty, unsupported and unavailable are different states that callers need to distinguish.

**Review:** Choose an unavailable-history error contract. Extend the retention fixture to Erigon, Nethermind and Besu; the current pruning result is Reth-only.

### H07 — Replay transactionHash field

**Proposed rule:** Include transactionHash in both individual and per-transaction block-replay envelopes. Do not require it for unsigned trace_call or trace_callMany simulations.

**Rationale:** A replay object should keep its identity when stored or moved out of its original request context.

**Review:** Confirm the envelope schema and add one serialization test; treat missing support in Besu separately under H01.

### H08 — Empty output and unrequested components

**Proposed rule:** Keep a stable envelope: output is the execution return bytes ("0x" when empty); trace is [] when not requested; stateDiff and vmTrace are null when not requested. Requesting a different tracer must not change or erase the returned execution output.

**Rationale:** Empty bytes and an unavailable value are different, and selecting an auxiliary view should not change the primary call result.

**Review:** Preserve the execution return bytes independently of which optional trace components are requested.

### H09 — Failed frame results and error labels

**Proposed rule:** Always mark failure explicitly, preserve REVERT return bytes and measured gasUsed when available, and define a small stable set of failure kinds. Human-readable explanations or decoded revertReason may be optional. Use null for an inapplicable result; never manufacture a successful result for an exceptional halt.

**Rationale:** Consumers need the raw revert data and failure identity without parsing implementation-specific English or mistaking a failure for success.

**Review:** Agree exact field placement and stable failure kinds. Separate REVERT, exceptional halt and SELFDESTRUCT/no-result cases in fixtures; do not require identical free-text descriptions.

### H10 — Creation result field names

**Proposed rule:** Use result.address, result.code and result.gasUsed for successful creation. result.code is deployed runtime bytecode, including "0x" for an empty deployment. Use result.output for CALL results; vmTrace.code for executing initcode.

**Rationale:** Creation has two different bytecodes—initcode and deployed code—and their field meanings should remain distinct.

**Review:** Specify creationMethod if it is required; retain the successful mixed-creation probe and investigate historical-method metadata separately.

### H11 — Empty trace-type selection

**Proposed rule:** Accept an empty selection, execute the call and return its output with empty/unrequested components. Never leak a collection-reduction exception.

**Rationale:** An empty list of auxiliary views is a natural request for just the execution output, and already works identically in three clients.

**Review:** A focused Nethermind fix and regression fixture can proceed independently of the broader API design.

### H12 — Raw-transaction block argument

**Proposed rule:** Standardize the existing two-argument baseline with an explicit default of latest state. Treat an optional third block selector as a separately negotiated extension until clients adopt it; reject unsupported extra arguments consistently.

**Rationale:** The baseline should be usable on all four today without silently changing which state is traced.

**Review:** Decide whether to add the block selector later. Keep its acceptance test separate from raw transaction execution fidelity.

### H13 — Signed transaction nonce validation

**Proposed rule:** For trace_rawTransaction, honor the signed transaction’s nonce and validity rules at the selected state. Reject nonce mismatch with a stable transaction-validation error. Use unsigned trace_call for hypothetical execution that deliberately relaxes admission rules.

**Rationale:** A signed transaction is a concrete transaction, so tracing should not silently rewrite its nonce or return a partial-looking success.

**Review:** Agree admission semantics for signed raw transactions; fix malformed responses independently of that policy choice.

### H14 — Invalid-parameter error codes

**Proposed rule:** Use -32602 for malformed encodings, wrong types and unsupported argument shapes. Use a documented execution/validation error for a well-formed but invalid transaction. Keep messages implementation-specific; omit internal stack traces.

**Rationale:** Callers should distinguish a malformed request from a valid request that cannot execute, without text matching.

**Review:** Specify accepted types, bounds and unknown-field handling, then align parameter validation; do not conflate validation with EVM errors.

### H15 — Unsigned simulation fees and block environment

**Proposed rule:** Allow explicit zero-fee unsigned simulations without changing the selected block’s BASEFEE or other block fields. With nonzero supplied fees, use those fees consistently in opcode context and simulated accounting. Apply the same policy to trace_call and trace_callMany; keep strict signed validation separate.

**Rationale:** Relaxing transaction admission for a hypothetical call should not silently rewrite the block the caller selected.

**Review:** Agree the fee-free simulation rule, then fix the block-environment discrepancy. The source contract confirms BASEFEE is returned as word 3; this is more than an error-message difference.

### H16 — Fee accounting and sequential state diffs

**Proposed rule:** For signed/raw and mined replay, report the complete actual transition: sender pays value plus gas, fee recipient gets the tip, base fee is burned. For unsigned calls use H15’s documented policy consistently. Each callMany result should describe its own transition from the preceding call’s post-state, not a cumulative diff from the original state.

**Rationale:** A state diff should be internally coherent and reconstructible; an ordered multi-call simulation should have one predictable state timeline.

**Review:** Resolve fee and environment rules while preserving the now-tested storage carry-over, revert rollback and simulation isolation. Broader warm-access/refund behavior remains to test.

### H17 — New-account stateDiff encoding

**Proposed rule:** Make account existence explicit: + for fields of a newly created account (including empty code and nonce zero), - for deletion, * for a changed existing value, = for unchanged existing fields. If an address already existed with empty code, do not mark it as newly created.

**Rationale:** Absent and existing-but-empty are different states; consumers should be able to reconstruct the change without guessing.

**Review:** Confirm the historical delta model and add pre-funded empty account and deletion fixtures. Current evidence isolates a previously absent recipient.

### H18 — EIP-7702 code changes in stateDiff

**Proposed rule:** Report every actual delegation-code transition, including set, replace and clear, independently of account-creation flags. Preserve authorization changes when later execution reverts, according to execution semantics.

**Rationale:** stateDiff must include code that changed, even if the account already existed.

**Review:** Fix account code-change reporting separately; these cases persist after the VM-delta fix.

### H19 — vmTrace executing bytecode

**Proposed rule:** Record the bytecode actually executed in each frame: initcode for creation, resolved implementation code for delegation, and the relevant code source for delegatecall/callcode. Do not populate it from an address-only pre-state lookup after tracing.

**Rationale:** The operations must be decodable against the bytes in their own frame.

**Review:** Fix bytecode population separately from the already-retested instruction deltas.

### H20 — vmTrace step timing and deltas

**Proposed rule:** Define ex.used as post-operation gas remaining, and mem/store/push as effects of that same operation. Use null for no memory write, with actual offset and bytes for a write. Specify call/return gas boundaries using dedicated nested-call fixtures before choosing detailed gas formulas.

**Rationale:** A consumer applying deltas in order should reconstruct the state after every opcode without shifting effects to neighboring steps.

**Review:** Retain the fixed regression cases. Nested warm-access, stipend, refund and output-memory edge cases still require broader coverage.

### H21 — vmTrace numeric and optional metadata encoding

**Proposed rule:** Encode stack words as minimal hex quantities (zero is 0x0), memory/code as even-length byte strings. Make op/idx optional metadata; require consumers to accept their absence and validate consistency when present. Keep required execution fields independent of these annotations.

**Rationale:** Numeric values and byte arrays have different formatting needs; optional convenience labels should not determine whether a trace can be consumed.

**Review:** Agree schema types and extension policy. A diagnostic comparison may show a separate typed-semantic projection, but must retain the exact wire differences.

### H22 — Precompile return bytes

**Proposed rule:** Report the actual return bytes in both the execution envelope and the corresponding successful call-frame result, including precompiles with no EVM bytecode.

**Rationale:** Two views of the same successful call should not disagree about its return value.

**Review:** A focused Besu reporting fix and regression fixture can proceed without waiting on broad API policy.

### H23 — Special-action address matching

**Proposed rule:** Define from/to equivalents per action: creator/created address for CREATE, destroyed account/refund beneficiary for SELFDESTRUCT. Include these records when the corresponding address matches.

**Rationale:** Address-based history should not silently omit a transfer or creation because its action uses different field names.

**Review:** Agree the action mapping and add the discriminating filter fixtures.

### H24 — Sibling failure isolation

**Proposed rule:** Attach error and result data to the actual call frame. A sibling failure must not change another frame's status.

**Rationale:** A successful caller can handle a failed subcall and continue; frame results must preserve that distinction.

**Review:** Prepare a focused Besu regression and fix without waiting for broad schema policy.

### H25 — Well-formed errors for rejected raw transactions

**Proposed rule:** Return one complete valid JSON-RPC error response on validation failure. Do not begin serializing a result before validation has succeeded.

**Rationale:** Clients must be able to parse failures and associate them with a request, regardless of the chosen signed-transaction admission policy.

**Review:** Fix the serializer/error path; retain raw malformed response text in regression evidence.

### H26 — Account deletion across Cancun

**Proposed rule:** Represent genuine account deletion with the deletion marker, including code, nonce and storage. After EIP-6780, an existing account survives SELFDESTRUCT; do not report code deletion.

**Rationale:** Deletion and a zero balance with retained code are observably different states. Apply the rules of the selected block.

**Review:** Align deletion encoding and fix missing Reth deletion data. The separate Besu prestateTracer probe had no complete response and is not used as an execution oracle.

### H27 — Filter execution across fork boundaries

**Proposed rule:** Execute each block with its own fork rules and state, then filter and paginate the resulting ordered records. A range query should agree with the corresponding per-block queries.

**Rationale:** The same block must not trace differently depending on the other blocks included in a query.

**Review:** The repeated fixtures are ready for a focused processor-context investigation and regression.

### H28 — Historical state at system-operation boundaries

**Proposed rule:** Historical state at block N should include that block's changes, including system operations, but none from N+1. Replay should start from the appropriate parent state and apply each system transition once.

**Rationale:** A future system write must not appear in an earlier block's state. This is an execution-context issue, not a cosmetic trace encoding choice.

**Review:** Minimize the historical-state boundary discrepancy and investigate Erigon history indexing separately from the Parity schema.


## Recursive schema rendering

`openrpc.json` retains the recursive `vmTrace` schema and is the validation artifact.
The documentation renderer currently cannot expand recursive schemas. Its generated
`docs-openrpc.json` display input replaces resource-local self references with labeled
recursive object descriptions; it must not be used for conformance validation.
