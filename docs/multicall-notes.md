# Multicall
This documenent contains some extra information that couldn't be fit to the specification document directly.

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
						"number": "0xb"
					},
					"calls": [
						{
							"from": "0xc000000000000000000000000000000000000000",
							"input": "0x4360005260206000f3"
						}
					]
				},
				{
					"blockOverrides": {
						"number": "0xc"
					},
					"calls": [
						{
							"from": "0xc100000000000000000000000000000000000000",
							"input": "0x4360005260206000f3"
						}
					]
				}
			]
		},
		"0xa"
	]
}
```

Here we want our calls to be executed in blocks 11 (0xb) and in 12 (0xc). This block numbers can be anything as long as they are increasing and higher than the block we are building from 10 (0xa). So, we could set the blocks to be 100 and 200 for example. Now we end up in a situation that there exists block ranges 13-99 and 101-199 that are not defined anywhere. These blocks are called "phantom blocks". What happens if you try to to request block hash of any of such block in the EVM? How can we calculate block hash of future blocks when we don't know the block hash of the previous block?

Our solution to this problem is to define block hash of a phantom block to be:

```
phantom block at block #i = keccac(rlp(hash_of_previous_non_phantom_block, #i))
```

So for example in our example, you could get block hash of block #142 as 
```
phantom block at block #i = keccac(rlp(hash of block #12, #142))
```

The phantom blocks other properties are set to their default properties as defined by the multicall specification. We came to this definition by wanting phantom blocks hashes to be unique if things prior to the phantom block changes, so if a tooling is storing block hashes somewhere, they should remain unique if things change in the simulation.

One other approah to this problem would be to really calculate the real block hashes for all the phantom blocks, but this would make generating blocks far in future really expensive, as to generate 100 phantom blocks, you would need to calculate 100 block hashes that all depend on each other. And in most cases, no one really cares about these blocks.
