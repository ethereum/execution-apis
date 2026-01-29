# Eth API - Address `Appearance` specification

Specification for defining if a transaction or block field is relevant to a particular address.

A transaction or block field is an `Appearance` for an address if that address is involved in a way that
meets specific criteria. This specification defines those criteria.

## Introduction

An address may arise in the course of an Ethereum state transition. If an address appears in a transaction, that transaction could be meaningful to examine as a historical record.

For example, an `Appearance` may be "an address that is a recipient of a transfer of Ether during
the transaction execution". Such an `Appearance` (the presence of the address in that transaction in that
way) makes that transaction meaningful in an examination of the historical balances of that
address.

A collection of address `Appearance`s (defined in subsequent section) consistutes a set of transactions
and block field that are sufficient to form a complete historical analysis of "activity" for that address.
This "activity" may take many meanings (programs in the Ethereum may do arbitrary things), but can
be identified structurally as will be shown.

## Overview

An address may appear in different parts of a transaction. This might include being the sender or
recipient of a transfer, a block reward recipient, or other categories. One main category
is the address of a piece of code that was run during the transaction. This code address can be
readily identified in the transaction without a need to understand the purpose or nature of
the code.

The identification of an `Appearance` solves a discovery problem. Once an important transaction
for a particular address have been found, an analysis of what the `Appearance` means can be performed,
although this is beyond the scope of this specification.

## Terminology

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119 and RFC 8174.

### Type aliases

|Name|Example|Description|
|-|-|-|
|Hex-string|`0x0abc`|Hex encoded, 0x-prefixed, leading 0's permitted|
|Hex-number|`0xabcd`|Hex encoded, 0x-prefixed, leading 0's omitted|

### Design parameters
|Name|Value|Description|
|-|-|-|
|`MAX_VANITY_ZERO_CHARS`|`8`|Leading or trailing '0's permitted in an address|

### Derived parameters
|Name|Definition|Value|Description|
|-|-|-|-|
|`MIN_NONZERO_BYTES`|`20 - MAX_VANITY_ZERO_CHARS // 2`|`16`| Smallest possible nonzero component of an address|

## Address definition

An address is informally defined as 40 hexadecimal characters that
- May include some leading or trailing zeros (for vanity addresses)
- Appears in a transaction within a 32 byte section with left padding.
or example, in transaction calldata, in this case the calldata is trimmed to a 32 byte multiple,
divided into 32 byte sections with checked separately for "address" or "not address" classification.

An address has the following formal definition:
- MUST be 20 bytes
- MUST begin or end with `MIN_NONZERO_BYTES` non-zero bytes.
This allows for vanity address inclusion of up to `MAX_VANITY_ZERO_CHARS`.
- MUST NOT be a known precompile
- When detected within byte sequence of greater than 20 bytes, the bytes MUST appear with left padding (12 leading zero-bytes for the case of a 32 byte sequence).
The address that meets the criteria is extracted from the source bytes.
- When inspecting more than 32 bytes, the bytes MUST first trimmed to a multiple of 32 bytes and divided
into 32 byte sections (length modulo 32) to be examined separately.

The following examples show the detection and extraction of an address/addresses
from a sequence of bytes.

### Example "address"
Deposit contract address in a 32 byte hex string with left padding:
```command
"0x00000000000000000000000000000000219ab540356cbb839cbe05303d7705fa"
```
Result:
```console
["0x00000000219ab540356cbb839cbe05303d7705fa"]
```

Deposit contract address bytes, but shifted left by 7 hex characters.
A valid address is detected, but is not the deposit contract address.
Shifting 8 characters would be invalid.
```command
"0x0000000000000000000000000219ab540356cbb839cbe05303d7705fa0000000"
```
Result:
```console
["0x0219ab540356cbb839cbe05303d7705fa0000000"]
```

Data from the two examples above, concatenated:
```command
"0x00000000000000000000000000000000219ab540356cbb839cbe05303d7705fa0000000000000000000000000219ab540356cbb839cbe05303d7705fa0000000"
```
Result:
```console
["0x0219ab540356cbb839cbe05303d7705fa0000000", "0x00000000219ab540356cbb839cbe05303d7705fa]
```

Data from the example above, with additional bytes that are truncated as modulo 32 bytes:
```
"0x00000000000000000000000000000000219ab540356cbb839cbe05303d7705fa0000000000000000000000000219ab540356cbb839cbe05303d7705fa000000000000000"
```
Result:
```console
["0x219ab540356cbb839cbe05303d7705fa00000000", "0x00000000219ab540356cbb839cbe05303d7705fa]
```

### Example "not address"

Deposit contract address in a 32 byte hex string with right padding. No address is detected because the leftmost 24 characters (12 bytes) must all be zeros:
```command
"0x00000000219ab540356cbb839cbe05303d7705fa000000000000000000000000"
```
Result:
```console
[]
```
Deposit address shifted right to have 8 zeros (not allowed):
```command
"0x000000000000000000000000219ab540356cbb839cbe05303d7705fa00000000"
```
Result:
```console
[]
```

No address is detected because the leftmost 24 characters (12 bytes) are ignored and the
remaining string has too many trailing 0's (`0x9cbe05303d7705fa000000000000000000000000`):
```command
"0x0000000000000000000000009cbe05303d7705fa000000000000000000000000"
```
Result:
```console
[]
```
Nonzero characters long enough for an address, but rejected for spanning a 32 byte boundary
(read as `0x...000219ab540356cbb83` and `0x9cbe05303d7705fa000...`):
```command
"0x000000000000000000000000000000000000000000000000219ab540356cbb839cbe05303d7705fa000000000000000000000000000000000000000000000000"
```
Result:
```console
[]
```
## `Appearance` definition

An address `Appearance` is informally defined as a transaction identifier or block field
that contains that particular address. That is, transaction "A" is an `Appearance` of address "B"
if address "B" is part of transaction "A" in an important way, E.g., One of sender, recipient,
code address, etc..

For a given address, a transaction MUST be classified as an `Appearance` if any any of the following
conditions are met. Conditions are divided into different sections for clarity.

### Intra-transaction `Appearance`s
An address MAY appear in any of the following:
|Short description|Description|Access Comment|
|-|-|-|
|Access list|Transaction "accessList" field object "address" key| Transaction body|
|Sender|Transaction "from" field|Transaction body|
|Target|Transaction "to" field|Transaction body|
|Calldata|Transaction "input" field|Transaction body. Transaction. 32 byte aligned, modulo 32 bytes|
|Log origin|Log family (LOG0, LOG1, LOG2, LOG3, LOG4) address|Transaction receipt|
|Log topics|Log topic family (LOG1, LOG2, LOG3, LOG4) topic index 1, 2, 3 or 4|Transaction receipt|
|Log data|Log family (LOG0, LOG1, LOG2, LOG3, LOG4) data|Transaction receipt. 32 byte aligned, modulo 32 bytes|
|Opcode address argument|Opcode "address" parameter (including but not limited to CALL, CALLCODE, STATICCALL, DELEGATECALL, SELFDESTRUCT)|Accessible via call tracer "to" field|
|Internal return data|RETURN data defined by "offset" and "size" fields |Accessible via call tracer "output" field. 32 byte aligned, modulo 32 bytes|
|Create address|Create family opcode (CREATE or CREATE2) return "address" field|Accessible via call tracer "to" field|
|Internal calldata|Call family opcode (CALL, CALLCODE, STATICCALL or DELGATECALL) argument data defined by opcode "argsOffset" and "argsSize" fields|Accessible via call tracer "input" field. 32 byte aligned, modulo 32 bytes|
|Internal return data|Call family opcode (CALL, CALLCODE, STATICCALL or DELGATECALL) return data defined by opcode "retOffset" and "retSize" fields|Accessible via call tracer "output" field. 32 byte aligned, modulo 32 bytes|
|Internal create data|Create-family (CREATE or CREATE2) data defined by opcode "offset" and "size" fields|Accessible via call tracer "input" field. 32 byte aligned, modulo 32 bytes|

Note that the call tracer "to", "from", "input" and "output" fields are sufficient to capture all
the required data not present in the transaction body and receipts. See below for an algorithm
for finding `Appearance`s.

### Extra-transaction `Appearance`s
An address MAY appear in any of the following:
|Short description|Description|Comment|
|-|-|-|
|Block reward|An address in the "miner" field of a block header|-|
|Uncle reward|An address in the "miner" field of a block header within the block "uncles" field array|-|
|Withdrawal|An address in the "address" field of a block "withdrawals" field array object|-|
|Alloc|An address in the "alloc" field of the genesis block (a key in that object)|-|

## `Appearance` components

An address `Appearance` is defined as having the following components:
- Block number
    - MUST be included for any `Appearance`
    - Type: Hex-number
- Location, one of
    - Transaction index
        - MUST be used for an intra-transaction `Appearance`.
        - Type: Hex-number
    - Block field
        - MUST be used for an extra-transaction `Appearance`.
        - Type: String, one of: "alloc", "withdrawals", "uncles" or "miner"

An address MAY have multiple `Appearance`s in one block. For example, the set of "0x12", "0x3c" and
"withdrawals" is valid.

`Appearance`s MUST NOT be duplicated. This applies to an address that appears multiple times in a
single transaction, or multiple times in a block field. For example, the set of "0x12", "0x3c", "withdrawals" and "withdrawals" is not valid.

## Algorithm

### Address detection


A 32 byte string may be inspected to determine if it meets criteria for an address as follows:

```go
// Source: UnchainedIndex Specification, trueblocks-core@v0.51.0, Go implementation
func potentialAddress(addr string) bool {
 // Any 32 byte value smaller than this number (including precompiles)
 // are assumed to be baddresses. While there are technically a very
 // large number of addresses in this range, we choose to eliminate them
 // in an effort to keep the index small.
 //
 // While this may seem drastic—that a lot of addresses are being excluded,
 // the number is actually a quite small number--less than two out of
 // every 10000000000000000000000000000000000000000000000 20-bytes strings
 // are excluded, and almost every one of these are actually numbers such
 // account balance or number of tokens transferred. It’s worth it.
 small := "00000000000000000000000000000000000000ffffffffffffffffffffffffff"
 // -------+-------+-------+-------+-------+-------+-------+-------+
 if addr <= small {
 return false
 }
 // Any 32 byte value with less than this many leading zeros assumed to be
 // a baddress. (Most addresses are 20-bytes long and left-padded with zeros
 // Note: we’re processing these as strings, so 24 characters is 12 bytes.
 largePrefix := "000000000000000000000000"
 // -------+-------+-------+
 if !strings.HasPrefix(address, largePrefix) {
 return false
 }
 // A large number of what would normally be considered valid addresses
 // happen to end with eight zeros. We’re not sure why, but we identify
 // these as badresses as well in a final effort to lower the size of
 // the index. We’ve seen no obvious ill-effects from this choice.
 if strings.HasSuffix(address, "00000000") {
 return false
 }
 return true
}
```

### `Appearance` detection

`Appearance`s are be detected by inspecting a block. For demonstration purposes,
the following procedure can be performed by use of existing JSON-RPC endpoints (availability depends
on the client used).

1. Call `eth_getBlockByNumber` with params `[block_number, true]` to include transactions.
Extract addresses from block header (e.g., miner, uncles, withdrawal, alloc)
2. Get the call tracer via the command below. Extract addresses from fields as described in the
`Appearance`s table in the prior section.
3. For each address found, record the transaction(s) it appeared in (if appropriate).


Command to obtain the call tracer with logs:
```command
$ curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc": "2.0", "method": "debug_traceBlockByNumber", "params": [17796114, {"tracer": "callTracer", "tracerConfig":{"withLog": true}}], "id":1}' http://127.0.0.1:8545 | jq
```

It can be seen in the example below (a subset of the response from the above call)
```json
{
    "result": {
    "from": "0x9853fda0b5e99eac2968dc59ad37cded61cb1bf5",
    "gas": "0xd3c7",
    "gasUsed": "0xcc7c",
    "to": "0x1a0ad011913a150f69f6a19df447a0cfd9551054",
    "input": "0xe9e05c420000000000000000000000009853fda0b5e99eac2968dc59ad37cded61cb1bf500000000000000000000000000000000000000000000000000038d7ea4c6800000000000000000000000000000000000000000000000000000000000000186a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000",
    "calls": [
        {
        "from": "0x1a0ad011913a150f69f6a19df447a0cfd9551054",
        "gas": "0xbd57",
        "gasUsed": "0x623a",
        "to": "0x43260ee547c3965bb2a0174763bb8fecc650ba4a",
        "input": "0xe9e05c420000000000000000000000009853fda0b5e99eac2968dc59ad37cded61cb1bf500000000000000000000000000000000000000000000000000038d7ea4c6800000000000000000000000000000000000000000000000000000000000000186a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000",
        "calls": [
            {
            "from": "0x1a0ad011913a150f69f6a19df447a0cfd9551054",
            "gas": "0x9257",
            "gasUsed": "0x1f0b",
            "to": "0xa3cab0126d5f504b071b81a3e8a2bbbf17930d86",
            "input": "0xcc731b02",
            "output": "0x0000000000000000000000000000000000000000000000000000000001312d00000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000003b9aca0000000000000000000000000000000000000000000000000000000000000f424000000000000000000000000000000000ffffffffffffffffffffffffffffffff",
            "calls": [
                {
                "from": "0xa3cab0126d5f504b071b81a3e8a2bbbf17930d86",
                "gas": "0x7d0a",
                "gasUsed": "0xb7c",
                "to": "0x17fb7c8ce213f1a7691ee41ea880abf6ebc6fa95",
                "input": "0xcc731b02",
                "output": "0x0000000000000000000000000000000000000000000000000000000001312d00000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000003b9aca0000000000000000000000000000000000000000000000000000000000000f424000000000000000000000000000000000ffffffffffffffffffffffffffffffff",
                "type": "DELEGATECALL"
                }
            ],
            "type": "STATICCALL"
            }
        ],
        "logs": [
            {
            "address": "0x1a0ad011913a150f69f6a19df447a0cfd9551054",
            "topics": [
                "0xb3813568d9991fc951961fcb4c784893574240a28925604d09fc577c55bb7c32",
                "0x0000000000000000000000009853fda0b5e99eac2968dc59ad37cded61cb1bf5",
                "0x0000000000000000000000009853fda0b5e99eac2968dc59ad37cded61cb1bf5",
                "0x0000000000000000000000000000000000000000000000000000000000000000"
            ],
            "data": "0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000004900000000000000000000000000000000000000000000000000038d7ea4c6800000000000000000000000000000000000000000000000000000038d7ea4c6800000000000000186a0000000000000000000000000000000000000000000000000"
            }
        ],
        "type": "DELEGATECALL"
        }
    ],
    "value": "0x38d7ea4c68000",
    "type": "CALL"
    }
}
```
Some of the `Appearance`s in the above result are:
- `0x17fb7c8ce213f1a7691ee41ea880abf6ebc6fa95`
- `0x1a0ad011913a150f69f6a19df447a0cfd9551054`
- `0x43260ee547c3965bb2a0174763bb8fecc650ba4a`
- `0x9853fda0b5e99eac2968dc59ad37cded61cb1bf5`
- `0xa3cab0126d5f504b071b81a3e8a2bbbf17930d86`

## Security Considerations

Ethereum allows for use of data that has the same structure as a 20 byte address.
The algorithm used to find address `Appearance`s ideally minimises missing `Appearance`s
(false negatives) and so may include false positives.

### False positives: `Appearance`s returned for addresses that do not exist

Some non-address data with a particular encoding may be used and be misclassified
as an address. For example, calldata for a specific application that is not
encoding for an address.

False positives can lead to a larger response size, and wasted effort if performing
semantic analysis of transactions for addresses found.

### False negatives: Vanity address

An address may be found with more zeros than the specification allows. This
process involves generating private keys and checking the associated address.

The presence of such a "vanity address" will not be detected by algorithms
conforming to this specification. This design tradeoff allows fewer false positives.

### False negatives: Address composition

A contract may manipulate separate bytes to construct an address and then use this
for an opcode that does not get detected by the algorithm, such as BALANCE.

This will be missed by the algorithm, which constitutes an `Appearance` of the address
(its balance was checked) that will be absent in semantic analysis. This impact is
limited a the subset of opcodes (BALANCE, EXTCODESIZE, ...) and any analysis should take
this into account.

### Precompiles

Existing precompiles are located at addresses `0x01` to `0x09`. They do
not meet definitional criteria and SHOULD NOT be included.

Future precompiles may be deployed at addresses that meet criteria however, SHOULD NOT be included
as addresses.

## Copyright

Copyright and related rights waived via [CC0](../LICENSE.md).
