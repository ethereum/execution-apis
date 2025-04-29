# eth_simulate
This document contains some extra information that couldn't be fit to the specification document directly.

## Default block values
Unlike `eth_call`, `eth_simulateV1`'s calls are conducted inside blocks. We don't require user to define all the fields of the blocks so here are the defaults that are assumed for blocks parameters:

| parameter name | default value |
-----------------|-----------------------
| prevRandao | `0x0000000000000000000000000000000000000000000000000000000000000000` |
| feeRecipient | `0x0000000000000000000000000000000000000000` |
| mixHash | `0x0000000000000000000000000000000000000000000000000000000000000000` |
| nonce | `0x0` |
| extraData | `0x0000000000000000000000000000000000000000000000000000000000000000` |
| difficulty | The same as the base block defined as the second parameter in the call |
| gasLimit | The same as the base block defined as the second parameter in the call |
| hash | Calculated normally |
| parentHash | Previous blocks hash |
| timestamp | The timestamp of previous block + `network block time` (12s on Ethereum Mainnet) |
| baseFeePerGas | When validation mode is true, baseFeePerGas is calculated on what it should be according to Ethereum's spec. When validation mode is false, the baseFeePerGas is set to zero |
| sha3Uncles | Empty trie root |
| withdrawals | Empty array |
| uncles | Empty array |
| blobBaseFee | When validation mode is true, blobBaseFee is calculated on what it should be according to EIP-4844 spec. When validation mode is false, the blobBaseFee is set to zero |
| number | Previous block number + 1 |
| logsBloom | Calculated normally. ETH logs are not part of the calculation |
| receiptsRoot | Calculated normally |
| transactionsRoot | Calculated normally |
| size | Calculated normally |
| withdrawalsRoot | Calculated normally |
| gasUsed | Calculated normally |
| stateRoot | Calculated normally |

## Default values for transactions
As eth_simulate is an extension to `eth_call` we want to enable the nice user experience that the user does not need to provide all required values for a transaction. We are assuming following defaults if the variable is not provided by the user:
| parameter name | description |
-----------------|-----------------------
| type | `0x2` |
| nonce | Take the correct nonce for the account prior eth_simulate and increment by one for each transaction by the account |
| to | `null` |
| from | `0x0000000000000000000000000000000000000000` |
| gasLimit | blockGasLimit - soFarUsedGasInBlock |
| value | `0x0` |
| input | no data |
| gasPrice | `0x0` |
| maxPriorityFeePerGas | `0x0` |
| maxFeePerGas | `0x0` |
| accessList | empty array |
| blobVersionedHashes | empty array |
| chainId | The chain id of the current chain |
| r | `0x0` |
| s | `0x0` |
| yParity | even |
| v | `0x0` |

## Overriding default values
The default values of blocks and transactions can be overriden. For Transactions we allow overriding of variables `type`, `nonce`, `to`, `from`, `gas limit`, `value`, `input`, `gasPrice`, `maxPriorityFeePerGas`, `maxFeePerGas`, `accessList`, and for blocks we allow modifications of `number`, `time`, `gasLimit`, `feeRecipient`, `prevRandao`, `baseFeePerGas` and `blobBaseFee`:
```json
"blockOverrides": {
	"number": "0x14",
	"time": "0xc8",
	"gasLimit": "0x2e631",
	"feeRecipient": "0xc100000000000000000000000000000000000000",
	"prevRandao": "0x0000000000000000000000000000000000000000000000000000000000001234",
	"baseFeePerGas": "0x14",
	"blobBaseFee": "0x15"
},
```
All the other fields are computed automatically (eg, `stateRoot` and `gasUsed`) or kept as their default values (eg. `uncles` or `withdrawals`). When overriding `number` and `time` variables for blocks, we automatically check that the block numbers and time fields are strictly increasing (we don't allow decreasing, or duplicated block numbers or times). If the block number is increased more than `1` compared to the previous block, new empty blocks are generated in between.

An interesting note here is that an user can specify block numbers and times of some blocks, but not for others. When block numbers of times are left unspecified, the default values will be used. After the blocks have been constructed, and default values are calculated, the blocks are checked that their block numbers and times are still valid.

## ETH transfer logs
When `traceTransfers` setting is enabled on `eth_simulateV1`, eth_simulate will return logs for ethereum transfers along with the normal logs sent by contracts. The ETH transfers are identical to ERC20 transfers, except the "sending contract" is address `0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee`.

For example, here's a query that will simply send ether from one address to another (with a state override that gives us the ETH initially):
```json
{
	"jsonrpc": "2.0",
	"id": 1,
	"method": "eth_simulateV1",
	"params": [
		{
			"blockStateCalls": [
				{
					"stateOverrides": {
						"0xc000000000000000000000000000000000000000": {
							"balance": "0x7d0"
						}
					},
					"calls": [
						{
							"from": "0xc000000000000000000000000000000000000000",
							"to": "0xc100000000000000000000000000000000000000",
							"value": "0x3e8"
						}
					]
				}
			],
			"traceTransfers": true
		},
		"latest"
	]
}
```

The output of this query is:
```json
{
	"jsonrpc": "2.0",
	"id": 1,
	"result": [
		{
			"number": "0x4",
			"hash": "0x859c932c5cf0dabf8d12eb2518e063966ac1a25e2fc49f1f02574a37f358d0b5",
			"timestamp": "0x1f",
			"gasLimit": "0x4c4b40",
			"gasUsed": "0x5208",
			"feeRecipient": "0x0000000000000000000000000000000000000000",
			"baseFeePerGas": "0x2310a91d",
			"prevRandao": "0x0000000000000000000000000000000000000000000000000000000000000000",
			"calls": [
				{
					"returnData": "0x",
					"logs": [
						{
							"address": "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
							"topics": [
								"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
								"0x000000000000000000000000c000000000000000000000000000000000000000",
								"0x000000000000000000000000c100000000000000000000000000000000000000"
							],
							"data": "0x00000000000000000000000000000000000000000000000000000000000003e8",
							"blockNumber": "0x4",
							"transactionHash": "0xa4d41019e71335f8567e17746b708ddda8b975a9a61f909bd3df55f4866cc913",
							"transactionIndex": "0x0",
							"blockHash": "0x859c932c5cf0dabf8d12eb2518e063966ac1a25e2fc49f1f02574a37f358d0b5",
							"logIndex": "0x0",
							"removed": false
						}
					],
					"gasUsed": "0x5208",
					"status": "0x1"
				}
			]
		}
	]
}
```

Here the interesting part is:
```json
"address": "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
"topics": [
	"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
	"0x000000000000000000000000c000000000000000000000000000000000000000",
	"0x000000000000000000000000c100000000000000000000000000000000000000"
],
"data": "0x00000000000000000000000000000000000000000000000000000000000003e8",
```
In the observed event, the sender address is denoted as the `0xee...` address. The first topic (`0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef`) aligns with the event signature `Transfer(address,address,uint256)`, while the second topic (`0x000000000000000000000000c000000000000000000000000000000000000000`) corresponds to the sending address, and the third topic (`0x000000000000000000000000c100000000000000000000000000000000000000`) represents the receiving address. The quantity of ETH transacted is stored in the data field.

The ETH logs will contain following types of ETH transfers:
 - Transfering ETH from EOA
 - Transfering ETH via contract
 - Selfdestructing contract sending ETH

But not following ones:
 - Gas fees
 - eth_simulates eth balance override  

ETH logs are not part of the calculation for logs bloom filter. Also, similar to normal logs, if the transaction sends ETH but the execution reverts, no log gets issued.

## Validation Mode
The `eth_simulate` includes a feature to enable or disable validation using the `Validation` setting. By default, validation is off, and `eth_simulate` behaves similarly to `eth_call`. When validation is enabled, the simulation closely approximates real EVM block creation, except it skips transaction signature checks and allows direct transactions from contracts.

### Zero Base Fee
When validation mode is disabled, the block's `baseFee` is set to zero. This affects the BASEFEE opcode and the `baseFee` returned by `eth_simulate`. Additionally, the base fee does not auto-adjust based on the parent block, even if the user overrides the previous block's `baseFee`.

This behavior is intentional to enable free transactions in `eth_simulate`, allowing users to avoid adding boilerplate code to give themselves ETH using a balance override. While similar results could be achieved by skipping balance validation, it would be inconsistent with EVM operations.

## Failures
It is possible that user defines a transaction that cannot be included in the Ethereum block as it breaks the rules of EVM. For example, if transactions nonce is too high or low, baseFeePerGas is too low etc. In these situations the execution of eth_simulate ends and an error is returned.

## Version number
The method name for eth_simulate `eth_simulateV1` the intention is that after release of eth_simulate, if new features are wanted the `eth_simulateV1` is kept as it is, and instead `eth_simulateV2` is published with the new wanted features.

## Clients can set their own limits
Clients may introduce their own limits to prevent DOS attacks using the method. We have thought of three such standard limits
- How many blocks can be defined in `BlockStateCalls`. The suggested default for this is 256 blocks
- A global gas limit (similar to the same limit for `eth_call`). The eth_simulate cannot exceed the global gas limit over its lifespan
- The clients can set their own limit on how big the input JSON payload can be. A suggested default for this is 30mb

## Rationale

### Pre-computed calls

When it comes to contract override behavior, specifically precompile override, there were two approaches:

1. As specified above to allow replacement of precompiles by EVM code and to allow those same precompiles to be relocated to another access for fallback behavior.
2. Allow users to pass in a set of pre-computed calls for an address, i.e. direct mapping of input to output.

The second approach has better UX for simple use-cases such as faking a signature via ecrecover. It also allows for getting the same gas usage as a real precompile execution. The simpler UX comes at cost of flexibility. Overriding with EVM code is general. It also imposes changes to the EVM interpreter code, which has been otherwise avoided, without enabling new features. Hence the spec proposes the first alternative.
