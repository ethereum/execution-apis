# Engine API -- Client Version Specification

Engine API structures and methods specified for client version specification.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Client Version Specification](#client-version-specification)
- [Structures](#structures)
  - [ClientCode](#clientcode)
  - [ClientVersionV1](#clientversionv1)
- [Methods](#methods)
  - [engine_getClientVersionV1](#engine_getclientversionv1)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Client Version Specification

To facilitate a more accurate measurement of execution layer client diversity statistics, execution clients **SHOULD** support the methods described in this document.

## Structures

### ClientCode

This enum defines a standard for specifying a client with just two letters. Clients teams which have a code reserved in this list **MUST** use this code when identifying themselves. The code is specified here only to facilitate standardization and NOT to imply that these are the only supported Ethereum clients. Any clients not listed here are free to use any two letters which don't collide with an existing client code. They are encouraged to make a PR to this repo to reserve their own code. Existing codes are as follows:

 - `BU`: besu
 - `EJ`: ethereumJS
 - `EG`: erigon
 - `GE`: go-ethereum
 - `GR`: grandine
 - `LH`: lighthouse
 - `LS`: lodestar
 - `NM`: nethermind
 - `NB`: nimbus
 - `TE`: trin-execution
 - `TK`: teku
 - `PM`: prysm
 - `RH`: reth

### ClientVersionV1

This structure contains information which identifies a client implementation. The fields are encoded as follows:

- `code`: `ClientCode`, e.g. `NB` or `BU`
- `name`: `string`, Human-readable name of the client, e.g. `Lighthouse` or `go-ethereum`
- `version`: `string`, the version string of the current implementation e.g. `v4.6.0` or `1.0.0-alpha.1` or `1.0.0+20130313144700`
- `commit`: `DATA`, 4 bytes - first four bytes of the latest commit hash of this build e.g. `fa4ff922`

Rationale: Human-readable fields like `clientName` and `version` are useful for log messages while fields like `code` and `commit` are useful for uniquely specifying clients within a limited space (e.g. in block `graffiti`).

## Methods

### engine_getClientVersionV1

#### Request

* method: `engine_getClientVersionV1`
* params:
  1. [`ClientVersionV1`](#ClientVersionV1) - identifies the consensus client
* timeout: 1s

#### Response
* result: `Array of ClientVersionV1` - Array of [`ClientVersionV1`](#ClientVersionV1)

#### Specification

1. Consensus and execution layer clients **MAY** exchange `ClientVersionV1` objects. Execution clients **MUST NOT** log any error messages if this method has either never been called or hasn't been called for a significant amount of time.
2. Clients **MUST** accommodate receiving any two-letter `ClientCode`, even if they are not reserved in the list above. Clients **MAY** log messages upon receiving an unlisted client code.
3. When connected to a single execution client, the consensus client **MUST** receive an array with a single
`ClientVersionV1` object. When connected to multiple execution clients via a multiplexer, the multiplexer **MUST** concatenate the responses from each execution client into a single, flat array before returning the
response to the consensus client.