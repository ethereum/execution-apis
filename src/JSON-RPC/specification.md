# Ethereum Execution Layer JSON-RPC API Specification
## Working Draft: Updated December 9th  
---
### **Author:**
Jared Doro(jareddoro@gmail.com)

### **Editors:**


### **Abstract:**
This document provides a detailed description of Ethereum's Execution Layer API. This document also provides the minimum requirements and functionality that is needed for a piece of software to be considered a valid Ethereum Execution Layer API.
### **Keywords:**
The keywords **MUST**, **MUST NOT**, **REQUIRED**, **SHALL**, **SHALL NOT**, **SHOULD**, **SHOULD NOT**, **RECOMMENDED**, **NOT RECOMMENDED**, **MAY**, and **OPTIONAL** in this document are to be interpreted as described in [[RFC2119](http://www.ietf.org/rfc/rfc2119.txt)] when, and only when, they appear in all capitals, as shown here.

# 1 Introduction 

The Ethereum execution layer API is one of the key components of Ethereum. It acts as an intermediary between the users of Ethereum and the execution layer where the transactions are received and executed. It provides a way for users to send transactions, request data, and execute smart contracts. 
## 1.1 Purpose and Intended Audience

The purpose of this document is to act as a centralized source of information regarding the functional and non-functional requirements for Ethereum's execution layer API. This document is intended for development teams that are planning on implementing a version of the execution layer API. This document would also be beneficial to but is not intended for anyone interested in learning how the user interacts with Ethereum clients and at the most basic level. 

## 1.2 Scope
 
The Ethereum execution layer API provides the basis for all external interactions with the Ethereum blockchain. The interactions with the Ethereum network can be divided into four categories. Transferring ETH, deploy contracts, executing contracts, and administrative tasks. The API also provides endpoints that returns historical network and block data. 

## 1.3 Definitions and Terms
* ETH
* Smart contracts
* user
* client
* EVM
* Account
* block
* block-chain

### 1.3.1 Not Specified vs Null

when the term not specified is used, it is describing the case where the parameter is not a part of the call vs when the term null is used it describes the case where the parameter is part of the call and has not been given a value

An example where `to` is not specified and `value` is null
```
{
	"jsonrpc": "2.0",
	"method": "eth_call",
	"params": [{
		"from": "0x3a5509015e0193adf435a761a6ce160f900034b5",
		"value": ""
}, "latest"],
	"id": 1
}
```
### 1.3.2 Default Block Parameter
Some endpoints use the `defaultBlockParameter` to specify the block that the data is being requested from.
The `defaultBlockParameter` allows the following options to specify a block:
* Block Number
* Block Hash
* Block Tag
  * earliest: for the earliest/genesis block.
  * latest: for the latest block received by the client.
  * pending: for the pending state/transactions.

### 1.3.3 Unavailable vs Does not exist
* A block is unavailable when the client does not currently have the requested data.
* A block does not exist when 
## 1.3.4 Input Parameters
 Input parameters for functions will be denoted by using the inline code feature of markdown and will look like `this`
## 1.4 References
* [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
* [Ethereum Yellow Paper](https://ethereum.github.io/yellowpaper/paper.pdf)
* [JSON Standard](https://www.json.org/json-en.html)
* [HTTP/2](https://http2.github.io/)
---

# 2 Overall Description

## 2.1 Product Perspective
The JSON-RPC API must interface with an Ethereum execution client to send and receive data to the Ethereum network
The users of the API will interact with the API through HTTP requests or through a web socket connection

## 2.2 Product Features 
This is a list of standard endpoints for the Ethereum execution layer API
### Web3
* web3_clientVersion
* web3_sha3
### Net
* net_version
* net_peerCount
* net_listening
### Eth
* eth_protocolVersion
* eth_syncing
* eth_coinbase
* eth_etherbase
* eth_mining
* eth_getWork
* eth_gasPrice
* eth_accounts
* eth_blockNumber
* eth_getBalance
* eth_getStorageAt
* eth_getTransactionCount
* eth_getBlockTransactionCountByHash
* eth_getBlockTransactionCountByNumber
* eth_getUncleCountByHash 
* eth_getUncleCountByNumber 
* eth_getCode
* eth_feeHistory
* eth_sign
* eth_signTransaction
* eth_sendTransaction
* eth_sendRawTransaction
* eth_call
* eth_estimateGas
* eth_getBlockByHash
* eth_getBlockByNumber
* eth_getTransactionByHash
* eth_getTransactionByBlockHashAndIndex
* eth_getTransactionByBlockNumberAndIndex
* eth_getTransactionReceipt
* eth_getUncleByBlockHashAndIndex 
* eth_getUncleByBlockNumberAndIndex 
* eth_newFilter
* eth_newBlockFilter
* eth_newPendingTransactionFilter
* eth_uninstallFilter
* eth_getFilterChanges
* eth_getFilterLogs
* eth_getLogs
* eth_hashrate
* eth_submitHashrate
* eth_submitWork
* eth_getRawTransactionByHash
* eth_getRawTransactionByBlockHashAndIndex
* eth_getRawTransactionByBlockNumberAndIndex
* eth_maxPriorityFeePerGas
* eth_getProof
* eth_createAccessList
* eth_getHeaderByNumber
* eth_getHeaderByHash
* eth_pendingTransactions
* eth_resend
* eth_fillTransaction
___
## 2.3 User Classes and Characteristics

## 2.4 Operating Environment how, where, under what conditions the system would be used.
a version of the Execution layer API **MUST** be implemented with each Ethereum Execution Client.
This means that for each Ethereum node running there **SHOULD** be an instance of the Execution layer API running with it to allow users to interact with it.
## 2.5 Design Implementations and constraints
The Execution layer API is currently using the JSON-RPC 2.0 standard. 
The execution layer API also supports interaction using both HTTP2.0 and WebSockets. 
## 2.6 Assumptions and dependencies

# 3 System Features
## web3_clientVersion
* [WC-1] web3_clientVersion **MUST** return a string containing information about the client version.
* [WC-2] web3_clientVersion **MUST** start with the name of the client.
* [WC-3] web3_clientVersion **MUST** return the client version after the name.
* [WC-4] web3_clientVersion **MUST** return the operating system and architecture the client is running on.
* [WC-5] web3_clientVersion **MUST** return the name of the language being used.
## web3_sha3
* [WS-1] web3_sha3 **MUST** return the Keccak256 hash of the given string. 
* [WS-2] web3_sha3 **MUST** return the Keccak256 hash of null when given "0x".
## net_version
* [NV-1] net_version **MUST** return the id number of the network the client is currently connected to.
## net_peerCount
* [NP-1] net_peerCount **MUST** return the number of peer nodes that the client is currently connected to. 
## net_listening
* [NL-1] net_listening **MUST** return a boolean indicating whether the client is currently listening for network connections.
## eth_protocolVersion
* [EP-1] eth_protocolVersion **MUST** return the current Ethereum Wire Protocol (eth protocol) version he network is using.
## eth_syncing
* [ESY-1] eth_syncing **MUST** return the progress of the client's sync of the network data.
* [ESY-2] eth_syncing **MUST** return false when the client is not syncing or already synced to the network.
* [ESY-3] eth_syncing **MUST** return the following for the sync progress:
   * The current block being synced on the client.
   * The current highest block known by the client.
   * The number of known and pulled states.
   * The block number that the client started syncing from.
## eth_coinbase
* [ECB-1] eth_coinbase **MUST** return the public address where the client's mining rewards are sent to.
* [ECB-2] eth_coinbase **MUST** error with code -32000 when the client does not have an address for the block reward to be sent to when not mining.
## eth_etherbase
* [EEB-1] eth_etherbase **MUST** return the public address where the client's mining rewards are sent to.
* [EEB-2] eth_etherbase **MUST** error with code -32000 when the client does not have an address for the block reward to be sent to when not mining.
## eth_mining
* [EM-1] eth_mining **MUST** return true when the client has mining enabled, otherwise it **MUST** return false.
## eth_gasPrice
* [EGP-1] eth_gasPrice **MUST** return the current price per unit of gas in wei that the client is charging.
## eth_accounts
* [EA-1] eth_accounts **MUST** return the public addresses for each Ethereum account that the client you are using manages.
## eth_blockNumber
* [EBN-1] eth_blockNumber **MUST** return the block number for the most recent block mined.
* [EBN-2] eth_blockNumber **MUST** return "0x0" when the client is not synced to the network.
## eth_getBalance
* [EGB-1] eth_getBalance **MUST** return the account balance of the `address` at the given `defaultBlockParameter`. 
* [EGB-2] eth_getBalance **MUST** return "0x0" when the client is not synced to the network.
* [EGB-3] eth_getBalance **MUST** error with code -32000 checking the balance at block that the client does not know about.
## eth_getStorageAt
* [EGS-1] eth_getStorageAt **MUST** return the data stored within the `storage slot` of the given `address` at the given `defaultBlockParameter`. 
* [EGS-2] eth_getStorageAt **MUST**  error with code -32000 when the client does not have the state of the `defaultBlockParameter` requested. 
* [EGS-3] eth_getStorageAt **MUST** return 0x00...0 when using `block tag` while syncing to the network.
* [EGS-4] eth_getStorageAt **MUST** error with code -32000 when using `block number` and `block hash` while syncing to the network.
## eth_getTransactionCount
* [EGTC-1] eth_getTransactionCount **MUST** return the nonce of the account is with the given `address` at the block requested by the `defaultBlockParameter`. 
* [EGTC-2] eth_getTransactionCount **MUST** error with code -32000 when calling a block that does not exist or is unavailable.
* [EGTC-3] eth_getTransactionCount **MUST** error with code -32000 when using a block number or block hash when syncing to the network.
## eth_getBlockTransactionCountByHash
* [EGTCH-1] eth_getBlockTransactionCountByHash **MUST** return the number of transactions within the block with the given `block hash`.
* [EGTCH-2] eth_getBlockTransactionCountByHash **MUST** return null when the `block hash` does not correspond to a block.
* [EGTCH-3] eth_getBlockTransactionCountByHash **MUST** return null when the client is currently syncing to the network.
## eth_getBlockTransactionCountByNumber
* [EGTCN-1] eth_getBlockTransactionCountByNumber **MUST** return the number of transactions within the block with the given `block number` or `block tag`.
* [EGTCN-2] eth_getBlockTransactionCountByNumber **MUST** return null when the `block number` is not known by the client or the block does not exist.
* [EGTCN-3] eth_getBlockTransactionCount **MUST** return 0x0 when block tags are used while syncing to the network.
## eth_getUncleCountByHash
* [EGUCH-1] eth_getUncleCountByHash **MUST** return number of uncle blocks that the block with the given `blockHash` has.
* [EGUCH-2] eth_getUncleCountByHash **MUST** return null the block with the given `blockHash` does not exist or is not available.
## eth_getUncleCountByNumber
* [EGUCN-1] eth_getUncleCountByNumber **MUST** return number of uncle blocks that the block with the given `block number` or `block tag` has.
* [EGUCN-2] eth_getUncleCountByNumber **MUST** return null when the requested block does not exist or is not available. 
* [EGUCN-3] eth_getUncleCountByNumber **MUST** return 0x0 when using `block tag` "pending".
## eth_getUncleByBlockHashAndIndex
* [EGUHI-1] eth_getUncleCountByHashAndIndex **MUST** return the uncle block information at the `uncle index` of the block with the given `block hash`.
* [EGUHI-2] eth_getUncleCountByHashAndIndex **MUST** return null when the block is unavailable or does not exist at the given `blockHash`.
* [EGUHI-3] eth_getUncleCountByHashAndIndex **MUST** return null when the block has no uncles at the `uncle index`.
## eth_getUncleByBlockNumberAndIndex 
* [EGUNI-1] eth_getUncleByBlockNumberAndIndex **MUST** return the uncle block information at the `Uncle index` of the block with the given `block number` or `block tag`.
* [EGUNI-2] eth_getUncleByBlockNumberAndIndex **MUST** return null when the block requested is unavailable or does not exist.
* [EGUNI-3] eth_getUncleByBlockNumberAndIndex **MUST** return null when the block has no uncles at the `uncle index`.
## eth_getCode
* [EGC-1] eth_getCode **MUST** return the deployed smart contract code at the given `address` and `defaultBlockParameter`.
* [EGC-2] eth_getCode **MUST**  error with code -32000 when the state information is not available for the requested block.
* [EGC-3] eth_getCode **MUST** error with code -32000 when using block numbers or block hashes while syncing to the network.
* [EGC-4] eth_getCode **MUST** return 0x0 when using block tags when syncing to the network.
## eth_feeHistory
* [EFH-1] eth_feeHistory **MUST** return the following information for number of blocks specified by the `blockCount` parameter stopping at the `highestBlock` parameter.
  * An array containing the base fee per gas for each block plus then block after `highestBlock`.
  * An array containing the ratio of the gas used by each block.
  * An array containing arrays with the requested `rewardPercentiles` for each block.
  * The oldest block used for the request.
* [EFH-2] eth_feeHistory **MUST** use the available range of blocks when the requested `blockCount` range can't be retrieved.
* [EFH-3] eth_feeHistory **MUST** allow block tags to be used for `highestBlock`.
* [EFH-4] eth_feeHistory **MUST** error with code -32000 when `highestBlock` is ahead of the chain.
## eth_sign
* [ESN-1] eth_sign **MUST** return the Ethereum specific signature detailed in [EIP-191](https://eips.ethereum.org/EIPS/eip-191) for the given unlocked `address` and `message`.
* [ESN-2] eth_sign **MUST** error with code -32000 when the account corresponding to the `address` is not unlocked.
* [ESN-3] eth_sign **MUST** error with code -32000 when the account corresponding to the `address` is not owned by the client.
## eth_signTransaction
* [ESNT-1] eth_signTransaction **MUST** return both the signed encoded and signed unencoded format of the `transaction` given.
* [ESNT-2] eth_signTransaction **MUST NOT** use given transaction `type` for transaction type, **MUST** determine transaction type from given transaction parameters.
* [ESNT-3] eth_signTransaction **MUST** return encoded legacy transactions in the RLP format.
* [ESNT-4] eth_signTransaction **MUST** return encoded [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718) transactions as the transaction `type` concatenated with the transaction type's defined encoding method.
* [ESNT-5] eth_signTransaction **MUST** use 0x0000000000000000000000000000000000000000 as default `from` address when `from` is null or not specified.
* [ESNT-6] eth_signTransaction **MUST** error with code -32000 when the `from` address is locked.
* [ESNT-7] eth_signTransaction **MUST** error with code -32000 when sending transaction using a `from` address that the client does not have the private key for.
* [ESNT-8] eth_signTransaction **MUST** error with code -32000 when the `gas` is not specified.
* [ESNT-9] eth_signTransaction **MUST** use 0x0 for `gas` when parameter is null.
* [ESNT-10] eth_signTransaction **MUST** error with code -32000 when the `gasPrice` or `maxFeePerGas` and `maxPriorityFeePerGas` are not specified.
* [ESNT-11] eth_signTransaction **MUST** error with code -32000 when `gasPrice` is used with `maxFeePerGas` and/or `maxPriorityFeePerGas`.
* [ESNT-12] eth_signTransaction **MUST** use 0x0 for `gasPrice` or `maxFeePerGas` and `maxPriorityFeePerGas` when the parameter is null. 
* [ESNT-13] eth_signTransaction **MUST** use null for `gasPrice` when using a type 2 transaction. 
* [ESNT-14] eth_signTransaction **MUST** error with code -32000 when the `maxPriorityFeePerGas` has a larger value than the `maxFeePerGas`
* [ESNT-15] eth_signTransaction **MUST** error with code -32000 when the `nonce` is not specified.
* [ESNT-16] eth_signTransaction **MUST** use 0x0 for `nonce` when parameter is null.
* [ESNT-17] eth_signTransaction **MUST** allow `to` address to be the same as `from` address.
* [ESNT-18] eth_signTransaction **MUST** allow user to enter extra key value pairs within the `transaction` object that are not used by the selected transaction.
* [ESNT-19] eth_signTransaction **MUST NOT** add any extra key value pairs sent by the user to the signed transaction sent to the network.
* [ESNT-20] eth_signTransaction **MUST** allow user to use duplicate parameters in the `transaction` object and **MUST** use the last of the duplicate parameters.
* [ESNT-21] eth_signTransaction **MUST** allow user to use `data` or `input` for contract deployment or contract interactions.
* [ESNT-22] eth_signTransaction **MUST** error with code -32000 when `data` and `input` are both used and are not equal.
* [ESNT-23] eth_signTransaction **MUST** error with code -32000 when `gas` exceeds block gas limit.
* [ESNT-24] eth_signTransaction **MUST** error with code -32000 when `gasPrice` causes transaction to exceed the transaction fee cap.
* [ESNT-25] eth_signTransaction **MUST** error with code -32000 when `maxFeePerGas` causes transaction to exceed the transaction fee cap.
* [ESNT-26] eth_signTransaction **MUST** error with code -32000 when deploying contract with no `data`/`input`.
* [ESNT-27] eth_signTransaction **MUST** error with code -32000 when specifying a `chainId` that is different from the network's chain Id.
## eth_sendRawTransaction
* [ESRT-1] eth_sendRawTransaction **MUST** return the transaction hash after submitting an encoded signed transaction to the network.
* [ESRT-2] eth_sendRawTransaction **MUST** allow users to send transaction where `gasPrice` or `maxFeePerGas` or `maxPriorityFeePerGas` are below network average and may never be executed.
* [ESRT-3] eth_sendRawTransaction **MUST** error with code -32000 when the `from` address does not have enough Ether to pay for the transaction.
* [ESRT-4] eth_sendRawTransaction **MUST** error with code -32000 when nonce is too low.
* [ESRT-5] eth_sendRawTransaction **MUST** error with code -32000 when the `gas` is too low.
* [ESRT-6] eth_sendRawTransaction **MUST** error with code -32000 when the user did not raise the `maxFeePerGas` enough when trying to replace a pending transaction.
* [ESRT-7] eth_sendRawTransaction **MUST** error with code -32000 when `transaction` is not properly encoded.
* [ESRT-8] eth_sendRawTransaction **MUST** error with code -3200 when sending an encoded transaction while syncing to the network.
* [ESRT-9] eth_sendRawTransaction **MUST** allow sending of contract creation transactions with code that causes the EVM to error. Resulting in a contract with no code. [this would error on eth_call and eth_estimateGas]
## eth_sendTransaction
* [EST-1] eth_sendTransaction **MUST** return the transaction hash of the `transaction` when the transaction is successfully sent to the network.
* [EST-2] eth_sendTransaction **MUST NOT** use given transaction `type` for transaction type, **MUST** determine transaction type from given transaction parameters.
* [EST-3] eth_sendTransaction **MUST** allow `to` address to be the same as `from` address
* [EST-4] eth_sendTransaction **MUST** allow user to enter extra key value pairs within the `transaction` object that are not used by the selected transaction.
* [EST-5] eth_sendTransaction **MUST NOT** add any extra key value pairs sent by the user to the signed transaction sent to the network.
* [EST-6] eth_sendTransaction **MUST** allow user to use duplicate parameters in the `transaction` object and **MUST** use the last of the duplicate parameters.
* [EST-7] eth_sendTransaction **MUST** use 0x0000000000000000000000000000000000000000 as default `from` address when `from` is null or not specified.
* [EST-8] eth_sendTransaction **MUST** error with code -32000 when sending transaction using a `from` address that the client does not have the private key for.
* [EST-9] eth_sendTransaction **MUST** error with code -32000 when sending transaction using a `from` address that is locked.
* [EST-10] eth_sendTransaction **MUST** allow user to use `value` parameter when deploying a contract.
* [EST-11] eth_sendTransaction **MUST NOT** send `value` to created contract when used during contract deployment
* [EST-12] eth_sendTransaction **MUST NOT** cost any extra when `value` is added during contract deployment.
* [EST-13] eth_sendTransaction **MUST** allow user to use `data` or `input` for contract deployment or contract interactions.
* [EST-14] eth_sendTransaction **MUST** error with code -32000 when `data` and `input` are both used and are not equal.
* [EST-15] eth_sendTransaction **MUST** allow users to send transaction where `gasPrice` or `maxFeePerGas` or `maxPriorityFeePerGas` are below network average and may never be executed.
* [EST-16] eth_sendTransaction **MUST** use legacy transaction anytime `gasPrice` is specified without `maxFeePerGas` and `maxPriorityFeePerGas`, otherwise type 2 transaction is used.
* [EST-17] eth_sendTransaction **MUST** use 0x0 for `value` when not specified in the transaction.
* [EST-18] eth_sendTransaction **MUST** use the `from` address's nonce when `nonce` is not specified.
* [EST-19] eth_sendTransaction **MUST** error with code -32000 when `gasPrice` is used with `maxFeePerGas` and/or `maxPriorityFeePerGas`.
* [EST-20] eth_sendTransaction **MUST** error with code -32000 when `gas` is not enough to complete the transaction.
* [EST-21] eth_sendTransaction **MUST** error with code -32000 when `gas` exceeds block gas limit.
* [EST-22] eth_sendTransaction **MUST** error with code -32000 when `gasPrice` causes transaction to exceed the transaction fee cap.
* [EST-23] eth_sendTransaction **MUST** error with code -32000 when the `from` address does not have enough Ether to pay for the transaction.
* [EST-24] eth_sendTransaction **MUST** error with code -32000 when `nonce` is too low.
* [EST-25] eth_sendTransaction **MUST** error with code -32000 when `maxFeePerGas` causes transaction to exceed the transaction fee cap.
* [EST-26] eth_sendTransaction **MUST** error with code -32000 when `maxPriorityFeePerGas` is greater than `maxFeePerGas`.
* [EST-27] eth_sendTransaction **MUST** error with code -32000 when the user did not increase the `maxFeePerGas` enough when trying to replace a pending transaction.
* [EST-28] eth_sendTransaction **MUST** error with code -32000 when deploying contract with no `data`/`input`.
* [EST-29] eth_sendTransaction **MUST** error with code -3200 when `data`/`input` provided caused an EVM error when the `gas` is not specified, otherwise it is ignored.
* [EST-30] eth_sendTransaction **MUST** estimate the amount of gas needed to complete the transaction and use that value for `gas` when not specified.
* [EST-31] eth_sendTransaction **MUST** use 0x0 for `gas` when null.
* [EST-32] eth_sendTransaction **MUST** use 0x0 for `gasPrice` when null.
* [EST-33] eth_sendTransaction **MUST** error with code -32000 if the transaction's `chain Id` is different than the network's Id.
* [EST-34] eth_sendTransaction **MUST** use the estimated price per gas and max priority fee per gas for `maxFeePerGas` and `maxPriorityFeePerGas` when not specified.
* [EST-35] eth_sendTransaction **MUST** error with code -32000 when trying to send a transaction while syncing to the network.
## eth_estimateGas
* [EEG-1] eth_estimateGas **MUST** return the estimated amount of gas the given `transaction` will take to execute.
* [EEG-2] eth_estimateGas **MUST** error with code -32000 when the `from` address has insufficient Ether to execute the given `transaction`.
* [EEG-3] eth_estimateGas **MUST NOT** check if the `from` account has sufficient funds when estimating contract deployment.
* [EEG-4] eth_estimateGas **MUST** use 0x0000000000000000000000000000000000000000 for `from` when it is not specified.
* [EEG-5] eth_estimateGas **MUST** error with code -32000 when estimating a contract creation that causes an error within the EVM.
## eth_getBlockByHash
* [EGBH-1] eth_getBlockByHash **MUST** return the block information for the block with the given `block hash`.
* [EGBH-2] eth_getBlockByHash **MUST** return null when the given `block hash` is unavailable or does not correspond to a block.
* [EGBH-3] eth_getBlockByHash **MUST** return block information with only transaction hashes when `hydrated transactions` is false. Otherwise, it should include full transaction objects.
---
## eth_getBlockByNumber
* [EGBN-1] eth_getBlockByNumber **MUST** return the block information for the block with the given `block number` or `block tag` when available.
* [EGBN-2] eth_getBlockByNumber **MUST** return null when the given `block number` is unavailable or does not correspond to a block. 
* [EGBN-3] eth_getBlockByNumber **MUST** return block information with only transaction hashes when `hydrated transactions` is false. Otherwise, it should include full transaction objects.
---
## eth_getTransactionByHash
* [EGTH-1] eth_getTransactionByHash **MUST** return the transaction object for the transaction with the given `transaction hash`.
* [EGTH-2] eth_getTransactionByHash **MUST** return null when the transaction with the given `transaction hash` does not exist or is not available. 
## eth_getTransactionByBlockHashAndIndex
* [EGTHI-1] eth_getTransactionByBlockHashAndIndex **MUST** return the transaction object with the given `block hash` and `transaction index`.
* [EGTHI-2] eth_getTransactionByBlockHashAndIndex **MUST** return null block with the given `blockHash` does not exist or is not available.
* [EGTHI-3] eth_getTransactionByBlockHashAndIndex **MUST** return null when there is no transaction at the given `transaction index` in the requested block.
## eth_getTransactionByBlockNumberAndIndex
* [EGTNI-1] eth_getTransactionByBlockNumberAndIndex **MUST** return the transaction object with the given `block number` or `block tag` and `transaction index`.
* [EGTNI-2] eth_getTransactionByBlockNumberAndIndex **MUST** return null when the block with given `block number` does not exist or is not available.
* [EGTNI-3] eth_getTransactionByBlockNumberAndIndex **MUST** return null when the given `transaction index` does not exist in the requested block.
## eth_getTransactionReceipt
* [EGTR-1] eth_getTransactionReceipt **MUST** return the transaction receipt for the transaction with the given `transaction hash`.
* [EGTR-2] eth_getTransactionReceipt **MUST** return null when the transaction with the given `transaction hash` does not exist or is not available.
* [EGTR-3] eth_getTransactionReceipt **MUST** return null when the transaction has not been included in a block.
## eth_newFilter
* [ENF-1] eth_newFilter **MUST** create a filter on the client that looks through each transaction log to see if it contains any the of the requested events.
* [ENF-2] eth_newFilter **MUST** allow `fromBlock` and `toBlock` to use both block numbers and block tags.
* [ENF-3] eth_newFilter **MUST** allow `from` and `to` to be used instead of `fromBlock` and `toBlock`.
* [ENF-4] eth_newFilter **MUST** give precedence to `toBlock` and `fromBlock` when used with `to` and `from`.
* [ENF-5] eth_newFilter **MUST** use "latest" for `fromBlock` and or `toBlock` when it is not specified.
* [ENF-6] eth_newFilter **MUST** error with code -32000 when the `fromBlock` is greater than the `toBlock`, except when the `toBlock` is set to latest and `fromBlock` is ahead of the chain.
* [ENF-7] eth_newFilter **MUST** allow `blockhash` to be used instead of `fromBlock` and `toBlock`.
* [ENF-8] eth_newFilter **MUST** error with code -32000 when `blockhash` is used with `fromBlock` and or `toBlock` in the same request.
* [ENF-9] eth_newFilter **MUST** allow `address` to be a single address or an array of addresses.
* [ENF-10] eth_newFilter **MUST** use null for `address` when it is not specified or when it is an empty array.
* [ENF-11] eth_newFilter **MUST** allow `topics` array to contain more than 4 values.
## eth_newBlockFilter
* [ENBF-1] eth_newBlockFilter **MUST** create a filter on the client that tracks when the client receives new blocks.
* [ENBF-2] eth_newBlockFilter **MUST** return the id of the newly created block filter. 
## eth_newPendingTransactionFilter
* [ENPTF-1] eth_newPendingTransactionFilter **MUST** create a filter on the client that tracks the hash of each pending transaction that the client receives.
* [ENPTF-2] eth_newPendingTransactionFilter **MUST** return the id of the newly created pending transaction filter.
## eth_uninstallFilter
* [EUF-1] eth_uninstallFilter **MUST** delete the filter with the given `filter id` from the client.
* [EUF-2] eth_uninstallFilter **MUST** return true when the given filter has been successfully uninstalled, otherwise it **MUST** return false.
## eth_getFilterChanges
* [EGFC-1] eth_getFilterChanges **MUST** return the block hashes of new blocks the client received since the filter was called last or first created, when the `filter id` corresponds to a block filter.
* [EGFC-3] eth_getFilterChanges **MUST** return the transaction hashes or each pending transaction received since the filter was last called or first created, when the `filter id` corresponds to a pending transaction filter.
* [EGFC-4] eth_getFilterChanges **MUST** return an empty array when calling a pending transaction filter while syncing to the network.
* [EGFC-5] eth-getFilterChanges **MUST** return all the logs that match the filters topics since the filter was last called or first created, when the `filter id` corresponds to a regular filter.
* [EGFC-7] eth_getFilterChanges **MUST** error with code -32000 when the given `filter id` does not correspond to an active filter on the client.
## eth_getFilterLogs
* [EGFL-1] eth_getFilterLogs **MUST** return all the logs that match the filters topics for the given `filter id`'s specified range.
* [EGFL-2] eth_getFilterLogs **MUST** only return the logs that match the filters parameters from the latest synced block when syncing to the network.
* [EGFL-3] eth_getFilterLogs **MUST** error with code -32000 when the given `filter id` does not correspond to an active filter on the client.
* [EGFL-4] eth_getFilterLogs **MUST** error with code -32000 when the given `filter id` corresponds to an active block filter or pending transaction filter on the client.
* [EGFL-5] eth_getFilterLogs **MUST** error with code -32005 when trying to return more than 1000 logs.
## eth_getLogs
* [EGL-1] eth_getLogs **MUST** look through all of the transaction logs of the client within the specified range.
* [EGL-2] eth_getLogs **MUST** return all of the logs that meet the filter requirements.
* [EGL-3] eth_getLogs **MUST** allow `fromBlock` and `toBlock` to use both block numbers and block tags.
* [EGL-4] eth_getLogs **MUST** allow `from` and `to` to be used instead of `fromBlock` and `toBlock`.
* [EGL-5] eth_getLogs **MUST** give precedence to `toBlock` and `fromBlock` when used with `to` and `from`.
* [EGL-6] eth_getLogs **MUST** error with code -32000 when the `fromBlock` is greater than the `toBlock`, except when the `toBlock` is set to latest and `fromBlock` is ahead of the current block.
* [EGL-7] eth_getLogs **MUST** use latest for `fromBlock` and or `toBlock` when it is not specified.
* [EGL-8] eth_getLogs **MUST** allow `blockhash` to be used in place of `toBlock` and `fromBlock`.
* [EGL-9] eth_getLogs **MUST** error with code -32602 when using `blockhash` with `fromBlock` and or `toBlock` in the same request.
* [EGL-10] eth_getLogs **MUST** use null for `address` when it is not specified or when it is an empty array.
* [EGL-11] eth_getLogs **MUST** allow `topics` array to contain more than 4 values.
* [EGL-12] eth_getLogs **MUST** return logs that match the parameters from only the latest synced block when syncing to the network.
* [EGL-13] eth_getLogs **MUST** error with code -32005 when trying to return more than 1000 logs.
## 4.2 eth_call
* [EC-1] eth_call **MUST** return the result of the given transaction.
* [EC-2] eth_call **MUST** accept all current transaction types. Legacy transactions and [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718) "typed" transactions.
* [EC-3] eth_call **MUST NOT** mine any transaction on the blockchain.
* [EC-4] eth_call **MUST** use the block requested by the `defaultBlockParameter` when interacting with contracts.
* [EC-5] eth_call **MUST** error with code -32000 when the requested block does not exist or is not available.
* [EC-6] eth_call **SHOULD NOT** be allowed to be called from an address where CODEHASH != EMPTYCODEHASH. [EIP-3607](https://eips.ethereum.org/EIPS/eip-3607)
* [EC-7] eth_call **MUST NOT** error if the transaction exceeds the gas or fee cap.
* [EC-8] eth_call **MUST** use 0x0000000000000000000000000000000000000000 as default `from` address when `from` is null or not specified. [Nethermind uses 0xf...fe] 
* [EC-9] eth_call **MUST** check `from` account balance has sufficient funds to "pay" for the transaction.
* [EC-10] eth_call **MUST** error with code -32000 account has insufficient funds to execute the transaction call.
* [EC-11] eth_call **MUST NOT** calculate cost of transactions when the `gas` and a price per unit of gas is not specified.   
* [EC-12] eth_call **MUST** allow `data` or `input` to be used when testing deployment or interacting with contracts.
* [EC-13] eth_call **MUST** use `input` when both `input` and `data` are specified. 
* [EC-14] eth_call **MUST** error with code -32000 when the `data`/`input` being deployed causes and error in the EVM.
* [EC-15] eth_call **MUST** error with code -32000 when the `gas` is too low to execute the call.
* [EC-16] eth_call **MUST** error with code -32000 when using `gasPrice` with `maxFeePerGas` or `maxPriorityFeePerGas`. 
* [EC-17] eth_call **MUST** use networks chain id for `chainId`
* [EC-18] eth_call **MUST** error with code -32000 when using block number or block hash for `defaultBlockParameter` while syncing to the network.
* [EC-19] eth_call **MUST** return 0x0 when using block tags for `defaultBlockParameter` while syncing to the network.
## eth_hashrate
* [EH-1] eth_hashrate **MUST** return the hashes per second that the client is using to mine blocks.
* [EH-2] eth_hashrate **MUST** return 0x0 when the client does not have mining enabled.
## eth_submitHashrate
* [ESH-1] eth_submitHashrate **MUST** return true when the client successfully submits a `hashrate` and an `id`.
* [ESH-2] eth_submitHashrate **MUST** return true when the client submits their hashrate while not mining.
* [ESH-3] eth_submitHashrate **MUST** return true when the client submits their hashrate while syncing to the network.
* [ESH-4] eth_submitHashrate **MUST** submit 0x0 when `hashrate` is null.
* [ESH-5] eth_submitHashrate **MUST NOT** error when `id` is equal to the `id` or another mining client.
## eth_getWork
* [EGW-1] eth_getWork **MUST** error with code -32000 when mining is not enabled.
* [EGW-2] eth_getWork **MUST** return the block header POW-hash, the seed hash for the DAG, the target condition, and the block number for the block being mined.
## eth_submitWork
* [ESW-1] eth_submitWork **MUST** return true when submitting the correct parameters to claim the block reward, otherwise false.
## eth_getRawTransactionByHash
* [EGRTH-1] eth_getRawTransactionByHash **MUST** return the encoded transaction associated with the given `transaction hash`.
* [EGRTH-2] eth_getRawTransactionByHash **MUST** return 0x0 when the transaction with the given `transaction hash` does not exist or is not available.
## eth_getRawTransactionByBlockHashAndIndex
* [EGRTHI-1] eth_getRawTransactionByBlockNumberAndIndex **MUST** return the encoded transaction associated with the given `block hash` and `transaction index`.
* [EGRTHI-2] eth_getRawTransactionByBlockNumberAndIndex **MUST** return 0x0 the block with the given `blockHash` does not exist or is unavailable.
* [EGRTHI-3] eth_getRawTransactionByBlockNumberAndIndex **MUST** return 0x0 when no transaction exists at the given `transaction index`.
## eth_getRawTransactionByBlockNumberAndIndex
* [EGRTNI-1] eth_getRawTransactionByBlockNumberAndIndex **MUST** return the encoded transaction associated with the given `block number` or `block tag` and `transaction index`.
* [EGRTNI-2] eth_getRawTransactionByBlockNumberAndIndex **MUST** return 0x0 when the requested block does not exist or is unavailable.
* [EGRTNI-3] eth_getRawTransactionByBlockNUmberAndIndex **MUST** return 0x0 when no transaction exists at the given `transaction index`.
## eth_maxPriorityFeePerGas
* [EMPFPG-1] eth_maxPriorityFeePerGas **MUST** return the clients price per unit of gas - 7.
## eth_getProof
* [EGPR-1] eth_getProof **MUST** return information about the given `account` that allows the `account`'s information to be verified.
* [EGPR-2] eth_getProof **MUST** return the following information about the given `account`.
  * Address
  * Account proof
  * Balance
  * Nonce
  * Code hash
  * Storage hash
  * Storage proof
* [EGPR-3] eth_getProof **MUST** return an array of RLP encoded MerkleTree-Nodes starting with the stateRoot following the given `account` to its source for the account proof.
* [EGPR-4] eth_getProof **MUST** error with code -32000 when the requested block is unavailable.
## eth_createAccessList
* [ECAL-1] eth_createAccessList **MUST** return an [EIP-2930](https://eips.ethereum.org/EIPS/eip-2930) access list biased off the given `transaction` and estimated gas cost when using the access list in the `transaction`.
* [ECAL-2] eth_createAccessList **MUST** use "latest" when the `defaultBlockParameter` is not specified.
* [ECAL-3] eth_createAccessList **MUST** use the estimated gas when the `gas` is not specified.
* [ECAL-4] eth_createAccessList **MUST** error with code -32000 when `gas` is too low.
* [ECAL-5] eth_createAccessList **MUST** error with code -32000 when the `from` address does not have enough Ether to pay for the transaction.
## eth_getHeaderByNumber
* [EGHN-1] eth_getHeaderByNumber **MUST** return the block header with the given `blockNumber` or `block tag`.
* [EGHN-2] eth_getHeaderByNumber **MUST** return null when the requested block does not exist or is unavailable.
## eth_getHeaderByHash
* [EGHH-1] eth_getHeaderByNumber **MUST** return the block header with the given `blockHash`.
* [EGHH-2] eth_getHeaderByNumber **MUST** return null when the requested block does not exist or is unavailable.
## eth_pendingTransactions
* [EPT-1] eth_pendingTransactions **MUST** return the transactions sent by accounts that are owned by the client that are currently in the transaction pool.
* [EPT-2] eth_pendingTransactions **MUST** return an empty array when syncing to the network.
## eth_resend
* [ERS-1] eth_resend **MUST** error with code -32000 when given any transaction. [geth Issue](https://github.com/ethereum/go-ethereum/issues/23964)
## eth_fillTransaction
* [EFT-1] eth_fillTransaction **MUST** fill in the missing transaction parameters of the given transaction.
* [EFT-2] eth_fillTransaction **MUST** return the return the raw transaction and JSON transaction object of the filled transaction.
* [EFT-3] eth_filterTransaction **MUST NOT** sign the transaction.
* [EFT-4] eth_filterTransaction **MUST** error with code -32000 when the `from` address does not have enough Ether to pay for the transaction.
* [EFT-5] eth_filterTransaction **MUST** allow both `data` and `input` to be used for contract creation,
* [EFT-6] eth_filterTransaction **MUST** error with code -32000 when the `data` and `input` both specified and not equal
* [EFT-7] eth_filterTransaction **MUST** error with code -32000 when `data`/`input` or `to` is not specified.
* [EFT-8] eth_filterTransaction **MUST** error with code -3200 when `data`/`input` provided caused an EVM error when the `gas` is not specified, otherwise it is ignored. 
# Errors
Error codes between-32768 and -32000 are reserved for JSON-RPC errors, where -32000 to -32099 are for Execution layer API errors
This table has been taken from the initial version of the JSON-RPC API spec that was never finalized. 
|Code|Message|Meaning|Category|
|-|-|-|-|
|-32700|Parse error|Invalid JSON|standard|
|-32600|Invalid request|JSON is not a valid request object|standard|
|-32601|Method not found|Method does not exist|standard|
|-32602|Invalid params|Invalid method parameters|standard|
|-32603|Internal error|Internal JSON-RPC error|standard|
|-32000|Invalid input|Missing or invalid parameters|non-standard|
|-32001|Resource not found|Requested resource not found|non-standard|
|-32002|Resource unavailable|Requested resource not available|non-standard|
|-32003|Transaction rejected|Transaction creation failed|non-standard|
|-32004|Method not supported|Method is not implemented|non-standard|
|-32005|Limit exceeded|Request exceeds defined limit|non-standard|
|-32006|JSON-RPC version not supported|Version of JSON-RPC protocol is not supported|non-standard|

[table source](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1474.md)
# Appendix
