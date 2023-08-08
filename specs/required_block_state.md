## `RequiredBlockState` specification

Specification of a data format that contains state required to
trace a single Ethereum block.

This is the format of the data returned by the `eth_getRequiredBlockState` JSON-RPC method.

## Table of Contents

  - [`RequiredBlockState` specification](#requiredblockstate-specification)
  - [Table of Contents](#table-of-contents)
  - [Abstract](#abstract)
  - [Motivation](#motivation)
  - [Overview](#overview)
    - [General Structure](#general-structure)
    - [Notation](#notation)
    - [Endianness](#endianness)
  - [Constants](#constants)
    - [Variable-size type parameters](#variable-size-type-parameters)
  - [Definitions](#definitions)
    - [`RequiredBlockState`](#requiredblockstate)
    - [`CompactEip1186Proof`](#compacteip1186proof)
    - [`Contract`](#contract)
    - [`TrieNode`](#trienode)
    - [`RecentBlockHash`](#recentblockhash)
    - [`CompactStorageProof`](#compactstorageproof)
  - [Algorithms](#algorithms)
    - [`construct_required_block_state`](#construct_required_block_state)
    - [`get_state_accesses`](#get_state_accesses)
    - [`get_proofs`](#get_proofs)
    - [`get_block_hashes`](#get_block_hashes)
    - [`use_required_block_state`](#use_required_block_state)
    - [`verify_required_block_state`](#verify_required_block_state)
    - [`trace_block_locally`](#trace_block_locally)
    - [`compression_procedure`](#compression_procedure)
  - [Security](#security)
    - [Future protocol changes](#future-protocol-changes)
    - [Canonicality](#canonicality)
    - [Post-block state root](#post-block-state-root)


## Abstract

An Ethereum block returned by `eth_getBlockByNumber` can be considered a program that executes
a state transition. The input to that program is the state immediately prior to that block.
Only a small part of that state is required to run the program (re-execute the block).
The state values can be accompanied by merkle proofs to prevent tampering.

The specification of that state (values and proofs as `RequiredBlockState`) facilitates
data transfer between two parties. The transfer represents the minimum amount of data
required for the holder of an Ethereum block to re-execute that block.

Re-execution is required for basic accounting (examination of the history of the global
shared ledger). Trustless accounting of single Ethereum blocks allows for lightweight
distributed block exploration.


## Motivation

State is rooted in the header. A merkle multiproof for all state required for all
transactions in one block enables is sufficient to trace any historical block.

In addition to the proof, BLOCKHASH opcode reads are also included.

Together, anyone with an ability to verify that a historical block header is canonical
can trustlessly trace a block without posession of an archive node.

The format of the data is deterministic, so that two peers creating the same
data will produce identical structures.

The primary motivation is that data may be distributed in a peer-to-peer content delivery network.
This would represent the state for a sharded archive node, where users may host subsets of the
data useful to them.

A secondary benefit is that traditional node providers could serve users the ability to
re-execute a block, rather than provide the result of re-execution. Transfer
of `RequiredBlockState` is approximately 167kb/Mgas (~2.5MB per block). Transfer of
a `debug_TraceBlock` result is on the order of hundreds of megabytes per block with memory
disabled, and with memory enabled can be tens of gigabytes. Local re-execution with an EVM
implementation of choice can produce the identical re-execution (including memory or custom
tracers), and can be processed and discarded on the fly.

## Overview

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT",
"RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted
as described in RFC 2119 and RFC 8174.

### General Structure

The `RequiredBlockState` consists of account state values as Merkle proofs, contract bytecode
and recent block hashes.

### Notation
Code snippets appearing in `this style` are to be interpreted as Python 3 pseudocode. The
style of the document is intended to be readable by those familiar with the
Ethereum consensus [https://github.com/ethereum/consensus-specs](https://github.com/ethereum/consensus-specs)
and Simple Serialize (SSZ) ([https://github.com/ethereum/consensus-specs/blob/dev/ssz/simple-serialize.md](https://github.com/ethereum/consensus-specs/blob/dev/ssz/simple-serialize.md))
specifications.

Where a list/vector is said to be sorted, it indicates that the elements are ordered
lexicographically when in hexadecimal representation (e.g., `[0x12, 0x3e, 0xe3]`) prior
to conversion to ssz format. For elements that are containers, the ordering is determined by
the first element in the container.

### Endianness

Big endian form is used as most data relates to the Ethereum execution context.

## Constants

### Variable-size type parameters

Helper values for SSZ operations. SSZ variable-size elements require a maximum length field.

Most values are chosen to be the approximately the smallest possible value.

| Name | Value | Description |
| - | - | - |
| MAX_ACCOUNT_NODES_PER_BLOCK | uint16(32768) | - |
| MAX_BLOCKHASH_READS_PER_BLOCK | uint16(256) | A BLOCKHASH opcode may read up to 256 recent blocks |
| MAX_BYTES_PER_NODE | uint16(32768) | - |
| MAX_BYTES_PER_CONTRACT | uint16(32768) | - |
| MAX_CONTRACTS_PER_BLOCK | uint16(2048) | - |
| MAX_NODES_PER_PROOF | uint16(64) | - |
| MAX_STORAGE_NODES_PER_BLOCK | uint16(32768) | - |
| MAX_ACCOUNT_PROOFS_PER_BLOCK | uint16(8192) | - |
| MAX_STORAGE_PROOFS_PER_ACCOUNT | uint16(8192) | - |

## Definitions

### `RequiredBlockState`

The entire `RequiredBlockState` data format is represented by the following (SSZ-encoded and
snappy-compressed) container.

As proofs sometimes have common internal nodes, all internal nodes for proofs are aggregated
for deduplication. They are located in the `account_nodes` and `storage_nodes` members.
Proofs refer to those nodes by index. A "compact" proof consists of a list of indices, indicating
which node is used.

The proof data represents values in the historical chain immediately prior to the execution of
the block (sometimes referred to as "prestate"). That is, `RequiredBlockState` for block `n`
contains proofs rooted in the state root of block `n - 1`.

```python
class RequiredBlockState(Container):
    #sorted (by address)
    compact_eip1186_proofs: List[CompactEip1186Proof, MAX_ACCOUNT_PROOFS_PER_BLOCK]
    #sorted
    contracts: List[Contract, MAX_CONTRACTS_PER_BLOCK]
    #sorted
    account_nodes: List[TrieNode, MAX_ACCOUNT_NODES_PER_BLOCK]
    #sorted
    storage_nodes: List[TrieNode, MAX_STORAGE_NODES_PER_BLOCK]
    #sorted (by block number)
    block_hashes: List[RecentBlockHash, MAX_BLOCKHASH_READS_PER_BLOCK]
```
The `RequiredBlockState` is compressed using snappy encoding (see algorithms section). The
`eth_getRequiredBlockState` JSON-RPC method returns the SSZ-encoded container with snappy encoding.

### `CompactEip1186Proof`

Represents the proof data whose root is the state root in the block header of the preceeding block.

The `account_proof` member consists of indices that refer to items in the `account_nodes` member
of the `RequiredBlockState` container.

```python
class CompactEip1186Proof(Container):
    address: Vector[uint8, 20]
    balance: List[uint8, 32]
    code_hash: Vector[uint8, 32]
    nonce: List[uint8, 8]
    storage_hash: Vector[uint8, 32]
    #sorted: node nearest to root first
    account_proof: List[uint16, MAX_NODES_PER_PROOF]
    #sorted
    storage_proofs: List[CompactStorageProof, MAX_STORAGE_PROOFS_PER_ACCOUNT]
```

### `Contract`

An alias for contract bytecode.
```python
Contract = List[uint8,  MAX_BYTES_PER_CONTRACT]
```

### `TrieNode`

An alias for a node in a merkle patricia proof.

Merkle Patricia Trie proofs consist of a list of witness nodes that correspond to each trie node that consists of various data elements depending on the type of node (e.g. blank, branch, extension, leaf).  When serialized, each witness node is represented as an RLP serialized list of the component elements.

```python
TrieNode = List[uint8,  MAX_BYTES_PER_NODE]
```

### `RecentBlockHash`

A block hash accessed by the "BLOCKHASH" opcode.
```python
class RecentBlockHash(Container):
    block_number: List[uint8, 8]
    block_hash: Vector[uint8, 32]
```

### `CompactStorageProof`

The `proof` member consists of indices that refer to items in the `storage_nodes` member
of the `RequiredBlockState` container.

The proof consists of a list of indices, one per node. The indices refer to the nodes in `TrieNode`.
```python
class CompactStorageProof(Container):
    key: Vector[uint8, 32]
    value: List[uint8, 8]
    #sorted: node nearest to root first
    proof: List[uint16, MAX_NODES_PER_PROOF]
```

## Algorithms

This section contains descriptions of procedures relevant to `RequiredBlockState`, including their
production (`construct_required_block_state`) and use (`use_required_block_state`).

### `construct_required_block_state`

For a given block, `RequiredBlockState` can be constructed using existing JSON-RPC methods by
using the following algorithms/steps:
1. `get_state_accesses` algorithm
2. `get_proofs`
3. `get_block_hashes`
4. Create the `RequiredBlockState` SSZ container
5. Use `compression_procedure` to compress the `RequiredBlockState`

### `get_state_accesses`

Call `debug_TraceBlock` with the prestate tracer, record key/value pairs where
they are first encountered in the block.

```
curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc": "2.0", "method": "debug_traceBlock", "params": ["finalized", {"tracer": "prestateTracer"}], "id":1}' http://127.0.0.1:8545 | jq
```
This will return state objects consisting of a key (account address), and value (state, which
may include contract bytecode and storage key/value pairs). See two objects for reference:
```json
[
    "0x58803db3cc22e8b1562c332494da49cacd94c6ab": {
        "balance": "0x13befe42b38a40",
        "nonce": 54
    },
    "0xae7ab96520de3a18e5e111b5eaab095312d7fe84": {
        "balance": "0x4558214a60e751c3a",
        "code": "0x608060/* Snip (entire contract bytecode) */410029",
        "nonce": 1,
        "storage": {
        "0x1b6078aebb015f6e4f96e70b5cfaec7393b4f2cdf5b66fb81b586e48bf1f4a26": "0x0000000000000000000000000000000000000000000000000000000000000000",
        "0x4172f0f7d2289153072b0a6ca36959e0cbe2efc3afe50fc81636caa96338137b": "0x000000000000000000000000b8ffc3cd6e7cf5a098a1c92f48009765b24088dc",
        "0x644132c4ddd5bb6f0655d5fe2870dcec7870e6be4758890f366b83441f9fdece": "0x0000000000000000000000000000000000000000000000000000000000000001",
        "0xd625496217aa6a3453eecb9c3489dc5a53e6c67b444329ea2b2cbc9ff547639b": "0x3ca7c3e38968823ccb4c78ea688df41356f182ae1d159e4ee608d30d68cef320"
        }
    },
    ...
]
```

### `get_proofs`

Call the `eth_getProof` JSON-RPC method for each state key (address) returned by the
`get_state_accesses` algorithm, including
storage keys if appropriate.

The block number used is the block prior to the block of interest (state is stored as post-block
state).

For all account proofs, aggregate and sort the proof nodes and represent each proof as a list of
indices to those nodes. Repeat for all storage proofs.

### `get_block_hashes`

Call `debug_TraceBlock` with the default tracer, record any use of the "BLOCKHASH" opcode.
Record the block number (top of stack in the "BLOCKHASH" step), and the block hash (top
of stack in the subsequent step).

### `use_required_block_state`

1. Obtain `RequiredBlockState`, for example by calling `eth_getRequiredBlockState`
2. Use `compression_procedure` to decompress the `RequiredBlockState`
3. `verify_required_block_state`
4. `trace_block_locally`

### `verify_required_block_state`

Check block hashes are canonical such as a node or against an accumulator of canonical
block hashes. Check merkle proofs in the required block state.

### `trace_block_locally`

Obtain a block (`eth_getBlockByNumber` JSON-RPC method) with transaction bodies. Use an EVM
and load it with the `RequiredBlockState` and the block. Execute
the transactions in the block and observe the trace.

### `compression_procedure`

The `RequiredBlockState` returned by the `eth_getRequiredBlockState` JSON-RPC method is
compressed. Snappy compression is used ([https://github.com/google/snappy](https://github.com/google/snappy)).

The encoding and decoding procedures are the same as that used in the Ethereum consensus specifications
([https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/p2p-interface.md#ssz-snappy-encoding-strategy](https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/p2p-interface.md#ssz-snappy-encoding-strategy)).

For encoding (compression), data is first SSZ-encoded and then snappy-encoded.
For decoding (decompression), data is first snappy-decoded and then SSZ-decoded.

## Security

### Future protocol changes

Merkle patricia proofs may be replaced by verkle proofs after some hard fork.
This would not invalidate `RequiredBlockState` data prior to that fork.
The new proof format could be added to this specification for data after that fork.

### Canonicality

A recipient of `RequiredBlockState` must check that the blockhashes are part of the real
Ethereum chain history. Failure to verify (`verify_required_block_state`) can result in invalid
re-execution (`trace_block_locally`).

### Post-block state root

A user that has access to canonical block hashes and a sound EVM implementation has strong
guarantees about the integrity of the block re-execution (`trace_block_locally`).

However, there is no guarantee to be able to compute a new block state root for this post-execution
state. For example, with the aim to check against the state root in the block header of that block
and thereby audit the state changes that were applied.

This is because the state changes may involve an arbitrary number of state deletions. State
deletions may change the structure of the merkle trie in a way that requires knowledge of
internal nodes that are not present in the proofs obtained by `eth_getProof` JSON-RPC method.
Hence, while the complete post-block trie can sometimes be created, it is not guaranteed.


