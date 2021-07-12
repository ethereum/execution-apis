# `eth_call`
The key words “MUST”, “MUST NOT”, “REQUIRED”, “SHALL”, “SHALL NOT”, “SHOULD”, “SHOULD NOT”, “RECOMMENDED”, “MAY”, and “OPTIONAL” in this document are to be interpreted as described in [RFC-2119](https://www.ietf.org/rfc/rfc2119.txt).

Specification | Description 
--|--
1         | `eth_call` MUST NOT affect on-chain state|
2         | `eth_call` MUST accept a `params` parameter [JSON Array](https://tools.ietf.org/html/rfc7159#section-5) of length 2, and its first item containing a [JSON Object](https://tools.ietf.org/html/rfc7159#section-4) with keys known here as the `input parameters` |
2.1     | `eth_call` MUST accept an optional input parameter `from` of type `ADDRESS` |
2.1.1  | `eth_call` MUST consider CALLER account to be `0x0000000000000000000000000000000000000000` if `from` is not defined |
2.1.2    | `eth_call` MUST consider CALLER account equal to the `from` parameter |
2.1.2.1 | `eth_call` MUST accept a `from` account that does not exist on-chain |
2.2       | `eth_call` MUST accept an input parameter `to` of type `ADDRESS` |
2.2.1     | `eth_call` MUST accept requests with `to` equal to null |
2.2.1.1   | `eth_call` MUST return empty hex string if `to` input parameter is equal to null  |
2.2.1.2   | `eth_call` MUST return empty hex string if `to` input parameter does not exist on-chain |
2.3       | `eth_call` MUST accept an optional input parameter `gas` of type `QUANTITY` |
2.3.1 | `eth_call` MUST consider gas to equal 25 million if the `gas` parameter is equal to `null` | 
2.3.2    | `eth_call` MUST accept requests with `gas` equal to a value greater than block `GAS LIMIT` |
2.3.3    | `eth_call` MUST NOT accept requests if the `gas` input parameter is less than `implicit gas` or `actual gas` |
2.3.4  | `eth_call` MUST NOT accept requests if the `gas` input parameter is greater than `2^64 - 1` or `0xffffffffffffffff` |
2.4       | `eth_call` MUST accept an optional input parameter `gasPrice` of type `Quantity` |
2.4.1  | `eth_call` MUST consider `GASPRICE` as equal to 0 if `gasPrice` input parameter is equal to `null` |
2.4.2   | `eth_call` MUST consider `GASPRICE` as equal to `gasPrice` input parameter if `gasPrice` input parameter is not equal to null |
2.5       | `eth_call` MUST accept an optional input parameter `value` of type `QUANTITY` | 
2.6 | `eth_call` MUST accept an optional input parameter `data` of type `DATA` |
3         | `eth_call` MUST accept a `params` parameter [JSON Array](https://tools.ietf.org/html/rfc7159#section-5) of length 2, and its second item (index = 1) containing a `Block Identifier` parameter that specifies the block height of the best block ([yellow paper](https://ethereum.github.io/yellowpaper/paper.pdf)) to assume the state |
3.1       | `eth_call` MUST assume the on-chain state of the latest best block ([yellow paper](https://ethereum.github.io/yellowpaper/paper.pdf)) if the `Block Identifier` parameter is equal to the [JSON String](https://tools.ietf.org/html/rfc7159#section-7) of "latest"; the latest best block height ([yellow paper](https://ethereum.github.io/yellowpaper/paper.pdf)) is equal to the current best block height. |
3.2       | `eth_call` MUST assume the on-chain state of the current best block with the greatest height ([yellow paper](https://ethereum.github.io/yellowpaper/paper.pdf))  if the `Block Identifier` parameter is equal to the [JSON String](https://tools.ietf.org/html/rfc7159#section-7) of "pending"; the pending best block height ([yellow paper](https://ethereum.github.io/yellowpaper/paper.pdf)) is equal to the current best block height plus a block with the best pending transactions. |
3.3       | `eth_call` MUST assume the on-chain state of the genesis best block ([yellow paper](https://ethereum.github.io/yellowpaper/paper.pdf)) if the `Block Identifier` parameter is equal to the [JSON String](https://tools.ietf.org/html/rfc7159#section-7) of "earliest"; the earliest best block height ([yellow paper](https://ethereum.github.io/yellowpaper/paper.pdf)) is equal to 0. |
3.4       | `eth_call` MUST assume the on-chain state of the best blockchain block with height equal to the `Quantity` value of the `Block Identifier` parameter if the `Block Identifier` parameter is equal to a hex encoded `Quantity` value. Height is equal to the number of blocks after the genesis block on the best blockchain (block number). |


# Copyright
Copyright and related rights waived via [CC0](https://creativecommons.org/publicdomain/zero/1.0/).
