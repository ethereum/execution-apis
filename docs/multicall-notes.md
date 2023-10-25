# Multicall
This document contains some extra information that couldn't be fit to the specification document directly.

## Default block values
Unlike `eth_call`, `eth_multicallV1`'s calls are conducted inside blocks. We don't require user to define all the fields of the blocks so here are the defaults that are assumed for blocks parameters:

| parameter name | default value |
-----------------|-----------------------
| prevRandao | 0x0 |
| feeRecipient | 0x0 |
| mixHash | 0x0 |
| nonce | 0x0 |
| extraData | 0x0 |
| difficulty | The same as the base block defined as the second parameter in the call |
| gasLimit | The same as the base block defined as the second parameter in the call |
| hash | calculated normally, except for phantom blocks, see below Phantom block section |
| parentHash | previous blocks hash (the real hash, or phantom blocks hash) |
| timestamp | the timestamp of previous block + 1 |
| baseFeePerGas | calculated on what it should be according to ethereum's spec. Note: baseFeePerGas is not adjusted in the phantom blocks. |
| sha3Uncles | empty trie root |
| withdrawals | empty array |
| uncles | empty array |
| logsBloom | calculated normally. ETH logs are not part of the calculation |
| receiptsRoot | calculated normally |
| transactionsRoot | calculated normally |
| number | calculated normally |
| size | calculated normally |
| withdrawalsRoot | calculated normally |
| gasUsed | calculated normally |
| stateRoot | calculated normally |

An interesting note here is that we decide timestamp as `previous block timestamp + 1`, while `previous block timestamp + 12` could also be an assumed default. The reasoning to use `+1` is that it's the minimum amount we have to increase the timestamp to keep them valid. While `+12` is what Mainnet uses, there are other chains that use some other values, and we didn't want to complicate the specification to consider all networks.

## Phantom blocks
The multicall allows you to define on what block number your calls or transactions are being executed on. Eg, consider following call:
```json
{
	"jsonrpc": "2.0",
	"id": 1,
	"method": "eth_multicallV1",
	"params": [
		{
			"blockStateCalls": [
				{
					"blockOverrides": {
						"number": "0x64"
					},
				},
				{
					"blockOverrides": {
						"number": "0xc8"
					},
				}
			]
		},
		"0xa"
	]
}
```

Here we want our calls to be executed in blocks 100 (0x64) and in 200 (0xc8). The block numbers can be anything as long as they are increasing and higher than the block we are building from 10 (0xa). Now we end up in a situation where there exists block ranges 13-99 and 101-199 that are not defined anywhere. These blocks are called "phantom blocks". What happens if you try to request block hash of any of such blocks in the EVM? How can we calculate the block hash of future blocks when we don't know the block hash of the previous block?

Our solution to this problem is to define block hash of a phantom block to be:

```
keccak(rlp([hash_of_previous_non_phantom_block, phantom_block_number]))
```

So for example in our example, you could get block hash of block 142 as follows: 
```
keccac(rlp([hash of block 12, 142]))
```

The phantom blocks other properties are set to their default properties as defined by the multicall specification. We came to this definition by wanting phantom block hashes to be unique if things prior to the phantom block changes, so if tooling is storing block hashes somewhere, they should remain unique if things change in the simulation.

One other approach to this problem would be to really calculate the real block hashes for all the phantom blocks, but this would make generating blocks far in future really expensive, as to generate 100 phantom blocks, you would need to calculate 100 block hashes that all depend on each other. And in most cases, no one really cares about these blocks.

Base fee per gas is not adjusted in the phantom blocks, their base fee remains constant.

## Default values for transactions
As multicall is an extension to `eth_call` we want to enable the nice user experience that the user does not need to provide all required values for a transaction. We are assuming following defaults if the variable is not provided by the user:
| parameter name | description |
-----------------|-----------------------
| type | 0x2 |
| nonce | Defaults to correct nonce |
| to | null |
| from | 0x0 |
| gas limit | Remaining gas in the curent block |
| value | 0x0 |
| input | no data |
| gasPrice | 0x0 |
| maxPriorityFeePerGas | 0x0 |
| maxFeePerGas | 0x0 |
| accessList | empty array |

## ETH transfer logs
When `traceTransfers` setting is enabled on `eth_multicallV1` The multical will return logs for ethereum transfers along with the normal logs sent by contracts. The ETH transfers are identical to ERC20 transfers, except the "sending contract" is address 0x0.

For example, here's a query that will simply send ether from one address to another (with a state override that gives us the ETH initially):
```json
{
	"jsonrpc": "2.0",
	"id": 1,
	"method": "eth_multicallV1",
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
							"address": "0x0000000000000000000000000000000000000000",
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
"address": "0x0000000000000000000000000000000000000000",
"topics": [
	"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
	"0x000000000000000000000000c000000000000000000000000000000000000000",
	"0x000000000000000000000000c100000000000000000000000000000000000000"
],
"data": "0x00000000000000000000000000000000000000000000000000000000000003e8",
```
As can be seen, the sending address is the zero address, `0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef` corresponds to signature `Transfer(address,address,uint256)`, `"0x000000000000000000000000c000000000000000000000000000000000000000"` corresponds the sending address and `0x000000000000000000000000c100000000000000000000000000000000000000` is the receiving address. The amount of ETH moved is stored in the `data` field.

The ETH logs will contain following types of ETH transfers:
 - Transfering ETH from EOA
 - Transfering ETH via contract
 - Selfdestructing contract sending ETH

But not following ones:
 - Gas fees
 - Multicalls eth balance override  

ETH logs are not part of the calculation for logs bloom filter. Also, similar to normal logs, if the transaction sends ETH but the execution reverts, no log gets issued.

## Validation
The multicall has a feature to enable or disable validation with setting `Validation`, by default, the validation is off, and the multicall mimics `eth_call` with reduced number of checks. Validation enabled mode is intended to give as close as possible simulation of real EVM block creation, except there's no checks for transaction signatures and we also allow one to send a direct transaction from a contract.

## Failures
It is possible that user defines a transaction that cannot be included in the Ethereum block as it breaks the rules of EVM. For example, if transactions nonce is too high or low, baseFeePerGas is too low etc. In these situations the execution of multicall ends and an error is returned.

## Version number
The method name for multicall `eth_multicallV1` the intention is that after release of multicall, if new features are wanted the `eth_multicallV1` is kept as it is, and instead `eth_multicallV2` is published with the new wanted features.

## Clients can set their own limits
Clients may introduce their own limits to prevent DOS attacks using the method. We have thought of two such standard limits
- How many blocks can be defined in `BlockStateCalls`. The suggested default for this is 256 blocks
- A global gas limit (similar to the same limit for `eth_call`). The multicall cannot exceed the global gas limit over its lifespan
