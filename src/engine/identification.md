# Engine API -- Client Identification

Engine API structures and methods specified for client identification.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Client Identification](#client-identification)
- [Structures](#structures)
  - [ClientCode](#clientcode)
  - [ClientIdentificationV1](#clientidentificationv1)
- [Methods](#methods)
  - [engine_clientIdentificationV1](#engine_clientidentificationv1)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Client Identification

To facilitate a more accurate measurement of execution layer client diversity statistics, execution clients **SHOULD** support the methods described in this document.

## Structures

### ClientCode

This enum defines a standard for specifying a client with just two letters. Clients teams which have a code reserved in this list **MUST** use this code when identifying themselves. The code is specified here only to facilitate standardization and NOT to imply that these are the only supported Ethereum clients. Any clients not listed here are free to use any two letters which don't collide with an existing client code. They are encouraged to make a PR to this repo to reserve their own code. Existing codes are as follows:

 - `BU`: besu
 - `EJ`: ethereumJS
 - `EG`: erigon
 - `GE`: go-ethereum
 - `LH`: lighthouse
 - `LS`: lodestar
 - `NM`: nethermind
 - `NB`: nimbus
 - `TK`: teku
 - `PM`: prysm
 - `RH`: reth
 
### ClientIdentificationV1

This structure contains information which identifies a client implementatiopn. The fields are encoded as follows:

- `code`: `ClientCode`, e.g. `NB` or `BU`
- `clientName`: `string`, Human-readable name of the client, e.g. `Lighthouse` or `go-ethereum`
- `version`: `string`, the standard semantic version string of the current implementation e.g. `v4.6.0` or `1.0.0-alpha.1` or `1.0.0+20130313144700`
- `commit`: `string`, the hex of the first 4 bytes of the latest commit hash of this build e.g. `fa4ff922`

Rationale: Human-readable fields like `clientName` and `version` are useful for log messages while fields like `code` and `commit` are useful for uniquely specifying clients within a limited space (e.g. in block `graffiti`).

## Methods

### engine_clientIdentificationV1

#### Request

* method: `engine_clientIdentificationV1`
* params:
  1. [`ClientIdentificationV1`](#ClientIdentificationV1) - identifies the consensus client
* timeout: 1s

#### Response

* result: [`ClientIdentificationV1`](#ClientIdentificationV1) - identifies the execution client

#### Specification

1. Consensus and execution layer clients **MAY** exchange `ClientIdentificationV1` objects. Execution clients **MUST NOT** log any error messages if this method has either never been called or hasn't been called for a significant amount of time.
2. Clients **MUST** accomodate receiving any two-letter `ClientCode`, even if they are not reserved in the list above. Clients **MAY** log messages upon receiving an unlisted client code. 