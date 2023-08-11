# Address JSON-RPC API

## Table of Contents

- [Address JSON-RPC API](#address-json-rpc-api)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Separate `address_` namespace](#separate-address_-namespace)
  - [Database requirement](#database-requirement)
  - [Rationale for the index existing inside the node](#rationale-for-the-index-existing-inside-the-node)
  - ["Address appearances" concept](#address-appearances-concept)
  - [Security considerations](#security-considerations)
    - [Increased resource use](#increased-resource-use)

## Overview

The Address JSON-RPC API is a collection of methods that Ethereum archival execution clients MAY implement.

This interface allows the use of an archive node from the perspective of an end user.
It provides information about which transactions are important for a given address.
This includes any address, including an externally owned account, a smart wallet, an account abstraction wallet or an application contract.

The API is desinged around an index mapping addresses to transaction identifiers. This index is created in advance so that queries are responded to quickly and do not require block re-execution.

The API allows introspection into otherwise opaque history. An example is an externally
owned account that received ether via a `CALL` opcode.
The API can return this transaction id, which can then be used with other methods (e.g., `debug_traceTransaction`) to identify the meaning of the transaction,
in this case an ether transfer.

## Separate `address_` namespace

A separace `address_` namespace exists because the method(s) within the namespace
require additional considerations that the `eth_` namespace does not require.
Clients that support the `address_` therefore have a clear way to toggle on this feature
to explicitly opt in to these requirements.

The requirements are:
- Archival node (store history for all blocks)
- Additional database requirement (see below)

## Database requirement

A node that supports the `address_` must store an index consisting of a mapping of address
to transaction identifiers. Each block must be parsed and addresses detected. Then for each
address the block number and transaction index must be stored.

The space required for a basic implementation is estimated at 80GB. Client implementations
may be able to reduce this amount significantly.

## Rationale for the index existing inside the node

The rationale for storing the index alongside other chain data is to standardise a new pattern
of data access. Archive nodes that support the `address_` namespace will be able to provide
users with the ability to introspect on particular addresses out of the box. This
means that a frontend can connect to a node and have historical data access for an
arbitrary combination of addresses.

This provides application developers with a new tool for decentralised front ends that
include historical information.

As the `address_` namespace methods apply to all addresses, this is a generalised solution
that makes inclusion inside the node an immediate utility for every user and protocol.

## "Address appearances" concept

`address_getAppearances` is the basic method to discover which transactions are relevant to a given address.

For a given address, this method returns an array of transaction identifiers in which the
address "appears". An "address appearance" is defined in [../../specs/appearance](../../specs/appearance.md)

## Security considerations

See [../../specs/appearance](../../specs/appearance.md) for additional security
considerations.

### Increased resource use

The provision of the `address_` namespace will result in the index being created by
the client. This can cause a temporary high resource use (CPU/disk), and then an
additional amount of work as the chain grows.

The amount of new work for a new block may impact a resource constrained node. The
task involves the equivalent of calling `debug_traceTransaction` with a `callTracer`,
parsing for addresses, and then adding these to a local database.
