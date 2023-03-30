# Engine API -- Builder Override

Engine API spec for builder override feature.

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Methods](#methods)
  - [engine_getPayloadV3](#engine_getpayloadv3)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Methods

### engine_getPayloadV3

#### Request

* method: `engine_getPayloadV3`
* params:
  1. `payloadId`: `DATA`, 8 Bytes - Identifier of the payload build process
* timeout: 1s

#### Response

* result: `object`
  - `executionPayload`: [`ExecutionPayloadV1`](./paris.md#ExecutionPayloadV1) | [`ExecutionPayloadV2`](#./shanghai.md#ExecutionPayloadV2) where:
      - `ExecutionPayloadV1` **MUST** be returned if the payload `timestamp` is lower than the Shanghai timestamp
      - `ExecutionPayloadV2` **MUST** be returned if the payload `timestamp` is greater or equal to the Shanghai timestamp
  - `blockValue` : `QUANTITY`, 256 Bits - The expected value to be received by the `feeRecipient` in wei
  - `shouldOverrideBuilder` : `BOOLEAN` - Suggestion from the EL to use this `executionPayload` instead of an externally provided one
* error: code and message set in case an exception happens while getting the payload.

#### Specification

This method follows the same specification as [`engine_getPayloadV2`](./shanghai.md#engine_getpayloadv2) with the addition of the following:

  1. Client software **MAY** use any heuristics to decide `shouldOverrideBuilder`.
