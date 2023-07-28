# Address JSON-RPC API

The Address JSON-RPC API is a collection of methods that Ethereum archival execution clients MAY implement.

This interface allows the use of an archive node from the perspective of an end user.
It provides information about which transactions are important for a given address. The address
may be an externally owned account, a smart wallet, an account abstraction wallet or an application contract.

The API is desinged around an index mapping addresses to transaction ids. This index is created
in advance so that queries are responded to quickly and do not require EVM re-execution.

The API allows introspection into otherwise opaque history. An example is an EOA that received
ether via a CALL opcode. The API can return this transaction id, which can then be used with
other methods (debug_traceTransaction) to identify the transfer.

# Address API -- Common Definitions

This document specifies common definitions and requirements affecting Address API specification in general.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Address JSON-RPC API](#address-json-rpc-api)
- [Address API -- Common Definitions](#address-api----common-definitions)
  - [Table of contents](#table-of-contents)
  - [Errors](#errors)


<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Errors

The list of error codes introduced by this specification can be found below.

| Code | Message | Meaning |
| - | - | - |
| -32700 | Parse error | Invalid JSON was received by the server. |
| -32600 | Invalid Request | The JSON sent is not a valid Request object. |
| -32601 | Method not found | The method does not exist / is not available. |
| -32602 | Invalid params | Invalid method parameter(s). |
| -32603 | Internal error | Internal JSON-RPC error. |
| -32000 | Server error | Generic client error while processing request. |
| -38004 | Too large request | Number of requested entities is too large. |


Each error returns a `null` `data` value, except `-32000` which returns the `data` object with a `err` member that explains the error encountered.

