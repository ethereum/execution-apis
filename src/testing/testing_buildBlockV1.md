# `testing_buildBlockV1`

This document specifies the `testing_buildBlockV1` RPC method. This method is a debugging and testing tool that simplifies the block production process into a single call. It is intended to replace the multi-step workflow of sending transactions, calling `engine_forkchoiceUpdated` with `payloadAttributes`, and then calling `engine_getPayload`.

This method is considered sensitive and is intended for testing environments only. See [**Security Considerations**](#securityconsiderations) for more details.

## Method: `testing_buildBlockV1`

### Parameters

1. `buildBlockParams`: `object` - An object containing the parameters for building the block:

   * `parentBlockHash`: `DATA, 32 Bytes` - (Required) The hash of the parent block upon which to build the new block.

   * `payloadAttributes`: `OBJECT` - (Required) The [`PayloadAttributesV3`](../engine/cancun.md#payloadattributesv3) object, as defined in [`engine_forkchoiceUpdatedV3`](../engine/cancun.md#engine_forkchoiceupdatedv3).

   * `transactions`: `Array of DATA` - (Required) An array of raw, signed transactions (hex-encoded `0x...` strings) to forcibly include in the generated block.

   * `extraData`: `DATA, 0 to 32 Bytes` - (Optional) A hex-encoded `0x...` string to be set as the `extraData` field of the built block. If this field is provided, it MUST override any default `extraData` the client would normally insert.

   * `skipValidation`: `boolean` - (Optional, default: `false`) If `true`, the client MUST attempt to build the block even if the provided transactions are invalid (e.g., incorrect nonce, insufficient balance, or invalid signature). If `false`, the method MUST return an error if any transaction fails state validation.

### Returns

`result`: `OBJECT` - The constructed object matching the response to [`engine_getPayloadV4`](../engine/prague.md#response-1).

### Behavior

* The client MUST build a new execution payload using the block specified by `parentBlockHash` as its parent.

* The client MUST use the provided `payloadAttributes` to define the context of the new block (e.g., `timestamp`, `prevRandao`, `suggestedFeeRecipient`).

* The client MUST include all transactions from the `transactions` array at the beginning of the block's transaction list, in the order they were provided.

* The client MUST NOT include any transactions from its local transaction pool. The resulting block MUST only contain the transactions specified in the `transactions` array.

* If `extraData` is provided, the client MUST set the `extraData` field of the resulting payload to this value.

* If `skipValidation` is `false` (or omitted), the client MUST validate all transactions in the `transactions` array against the state at the `parentBlockHash`. If any transaction is invalid, the method MUST return a standard JSON-RPC error.

* If `skipValidation` is `true`, the client MUST skip state-dependent validation (e.g., nonce, balance) for the transactions in the `transactions` array. The client SHOULD still perform basic format validation. The client MUST attempt to build and return the payload, even if this payload would be considered invalid by consensus rules.

* This method MUST NOT modify the client's canonical chain or head block. It is a read-only method for payload generation. It does not run the equivalent of `engine_newPayload` or `engine_forkchoiceUpdated`.

### Security Considerations

* **HIGHLY SENSITIVE**: This method is a powerful debugging tool intended for testing environments ONLY.

* It allows for the creation of arbitrary, and potentially invalid, blocks. It can be in used to bypass transaction pool validation and forcibly include any transaction.

* This method MUST NOT be exposed on public-facing RPC APIs.

* It is strongly recommended that this method, and its `testing_` namespace, be disabled by default. Enabling it should require an explicit command-line flag or configuration setting by the node operator.

### Example

#### Request

```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "method": "testing_buildBlockV1",
  "params": [
    {
      "parentBlockHash": "0x3b8fb240d288781d4f1e1d32f4c1550BEEFDEADBEEFDEADBEEFDEADBEEFDEAD",
      "payloadAttributes": {
        "timestamp": "0x6705D918",
        "prevRandao": "0x0000000000000000000000000000000000000000000000000000000000000000",
        "suggestedFeeRecipient": "0x0000000000000000000000000000000000000000",
        "withdrawals": [],
        "parentBeaconBlockRoot": "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884365149a42212e8822"
      },
      "transactions": [
        "0xf86c0a8504a817c80082520894...cb61163c917540a0c64c12b8f50015",
        "0xf86b0f8504a817c80082520894...5443c01e69d3b0c5112101564914a2"
      ],
      "extraData": "0x746573745F6E616D65",
      "skipValidation": false
    }
  ]
}
```

#### Response (Success)

```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "result": {
    "executionPayload": {
      "blockHash": "0x8980a3a7f8b16053c4dec86e1050e6378e176272517852f8c5b56f34e9a0f9b6",
      "parentHash": "0x3b8fb240d288781d4f1e1d32f4c15509a2b7538b8e0e719541a50a31a2631a01",
      "feeRecipient": "0x0000000000000000000000000000000000000000",
      "stateRoot": "0xca3699d05d3369680315af1d03c6f8f73188945f3c83756bde2575ddc25c6040",
      "receiptsRoot": "0x2e0617b0c3545d1668e8d8cb530a6c71b0a8f82875b1d0b38c352718016feff6",
      "logsBloom": "0x00...00",
      "prevRandao": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "blockNumber": "0x1b4",
      "gasLimit": "0x1c9c380",
      "gasUsed": "0xa410",
      "timestamp": "0x6705D918",
      "extraData": "0x746573745F6E616D65",
      "baseFeePerGas": "0x7",
      "excessBlobGas": "0x0",
      "blobGasUsed": "0x0",
      "transactions": [
        "0xf86c0a8504a817c80082520894...cb61163c917540a0c64c12b8f50015",
        "0xf86b0f8504a817c80082520894...5443c01e69d3b0c5112101564914a2"
      ],
      "parentBeaconBlockRoot": "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884365149a42212e8822",
      "depositRequests": [],
      "consolidationRequests": []
    },
    "blockValue": "0x1b681a965684000",
    "blobsBundle": {
      "commitments": [],
      "blobs": [],
      "kzgProofs": []
    },
    "shouldOverrideBuilder": false,
    "executionRequests": []
  }
}
```
