# Ethereum Execution Layer JSON-RPC API
## Technical Specification
## Working Draft: Updated December 19th  
---
### **Editor:**
Jared Doro(jareddoro@gmail.com)
### **Abstract:**
This document describes the requirements and expected behavior for the standard endpoint of the Ethereum Execution layer JSON-RPC API.
### **Keywords:**
The keywords **MUST**, **MUST NOT**, **REQUIRED**, **SHALL**, **SHALL NOT**, **SHOULD**, **SHOULD NOT**, **RECOMMENDED**, **NOT RECOMMENDED**, **MAY**, and **OPTIONAL** in this document are to be interpreted as described in [[RFC2119](http://www.ietf.org/rfc/rfc2119.txt)] when, and only when, they appear in all capitals, as shown here.

### **Citation format:**
When referencing this specification the following citation format should be used:

[ethereum-execution-layer-JSON-RPC-API]

Ethereum execution layer JSON-RPC API Edited by Jared Doro. 19th December 2021.
# Table of Contents
[1 Introduction](#1-introduction)\
&nbsp;&nbsp;[1.1 Overview](#11-overview)\
&nbsp;&nbsp;[1.2 Terminology](#12-terminology)\
&nbsp;&nbsp;[1.3 References](#13-references)\
&nbsp;&nbsp;&nbsp;&nbsp;[1.3.1 Normative References](#131-normative-references)\
&nbsp;&nbsp;&nbsp;&nbsp;[1.3.2 Non-normative References](#132-non-normative-references)\
&nbsp;&nbsp;[1.4 Document Conventions](#14-document-conventions)\
&nbsp;&nbsp;&nbsp;&nbsp;[1.4.1 Input Parameters](#141-input-parameters)\
&nbsp;&nbsp;&nbsp;&nbsp;[1.4.2 Not Specified vs Null vs Empty](#142-not-specified-vs-null-vs-empty)\
&nbsp;&nbsp;&nbsp;&nbsp;[1.4.3 Default Block Parameter](#143-default-block-parameter)\
&nbsp;&nbsp;&nbsp;&nbsp;[1.4.4 Unavailable vs Does Not Exist](#144-unavailable-vs-does-not-exist)\
[2 Standard endpoint specification](#2-standard-endpoint-specifications)\
[3 Errors](#3-errors)\
[Appendix](#appendix)
# 1 Introduction 
## 1.1 Overview
The Ethereum execution layer API is one of the key components of Ethereum. It acts as an intermediary between the users of Ethereum and the execution layer where the blocks are created and executed. It provides a way for users to send transactions to the blockchain, query data from the current and historical states of the blockchain, and monitor events and changes happening within the blockchain

The purpose of this document is to act as a centralized source of information regarding the functional and non-functional requirements for Ethereum's execution layer API. This document is intended for development teams that are planning on implementing a version of the execution layer API. This document would also be beneficial to but is not intended for anyone interested in learning how the user interacts with Ethereum clients and at the most basic level. 
## 1.2 Terminology
|Term| Definition|
|---|---|
|Account||
|Address||
|Block||
|Blockchain||
|call||
|client| Software that allows users to run an instance of the Ethereum blockchain|
|cryptography||
|Ether|Cryptographic token created as a reward for participating in securing and running the Ethereum blockchain. Ether is also used to pay for all transactions on the Ethereum blockchain.|
|EVM||
|Execution layer||
|gas||
|genisis block||
|hash||
|header||
|merkle tree||
|mining||
|network||
|node| A running client connected to the Ethereum blockchain.|
|nonce| An integer corresponding to the number of transactions an account has made on the blockchain.|
|peer nodes||
|private key||
|public key||
|rewards||
|signing||
|Smart Contract| Code|
|state||
|storage slot||
|token||
|transaction||
|Uncle block| A block that was |
|user||
|Wei|The smallest unit of Ether where **1Eth = 1e18Wei**|
## 1.3 References
### 1.3.1 Normative References
### 1.3.2 Non-normative References
## 1.4 Document Conventions
### 1.4.1 Input Parameters
Input parameters will be displayed like `this` when referred to in the document.
### 1.4.2 Not Specified vs Null vs Empty
* Not specified describes when a parameter is not part of the call.
* Null describes when a parameter is part of the call, but has the value of null.
* Empty describes when a parameter is part of the call, but has only an empty string.

An example where `gas` is not specified `value` is empty, and `data` is null.
```
{
	"jsonrpc": "2.0",
	"method": "eth_sendTransaction",
	"params": [
		{
			"from": "0x3a5509015e0193adf435a761a6ce160f900034b5",
			"to": "0xe64fac7f3df5ab44333ad3d3eb3fb68be43f2e8c",
			"value": "",
			"data": null
		}
	],
	"id": 1
}
```
### 1.4.3 Default Block Parameter
Some endpoints use the `defaultBlockParameter` to specify the block that the data is being requested from.
The `defaultBlockParameter` allows the following options to specify a block:
* Block Number
* Block Hash
* Block Tag
  * earliest: for the earliest/genesis block.
  * latest: for the latest block received by the client.
  * pending: for the pending state/transactions.

### 1.4.4 Unavailable vs Does Not Exist
* Requested data is unavailable when the client does not contain the necessary information to respond to the request, but the information does exist on the network.
* Requested data does not exist when neither the client or the network contain the necessary information to respond to the request.
# 2 Standard Endpoint Specifications
The following list does not contain every module that is available on all clients. This only contains the standard modules that each client needs to function properly.
### Web3
* <a href="#clientVersion">web3_clientVersion</a>
* <a href="#sha3">web3_sha3</a>
### Net
* <a href="#version">net_version</a>
* <a href="#peerCount">net_peerCount</a>
* <a href="#listening">net_listening</a>
### Eth
* <a href="#protocolVersion">eth_protocolVersion</a>
* <a href="#syncing">eth_syncing</a>
* <a href="#coinbase">eth_coinbase</a>
* <a href="#etherbase">eth_etherbase</a>
* <a href="#accounts">eth_accounts</a>
* <a href="#mining">eth_mining</a>
* <a href="#getWork">eth_getWork</a>
* <a href="#submitWork">eth_submitWork</a>
* <a href="#hashrate">eth_hashrate</a>
* <a href="#submitHashrate">eth_submitHashrate</a>
* <a href="#gasPrice">eth_gasPrice</a>
* <a href="#maxPriorityFeePerGas">eth_maxPriorityFeePerGas</a>
* <a href="#feeHistory">eth_feeHistory</a>
* <a href="#blockNumber">eth_blockNumber</a>
* <a href="#sign">eth_sign</a>
* <a href="#call">eth_call</a>
* <a href="#fillTransaction">eth_fillTransaction</a>
* <a href="#createAccessList">eth_createAccessList</a>
* <a href="#estimateGas">eth_estimateGas</a>
* <a href="#signTransaction">eth_signTransaction</a>
* <a href="#sendRawTransaction">eth_sendRawTransaction</a>
* <a href="#sendTransaction">eth_sendTransaction</a>
* <a href="#pendingTransactions">eth_pendingTransactions</a>
* <a href="#resend">eth_resend</a>
* <a href="#getProof">eth_getProof</a>
* <a href="#getBalance">eth_getBalance</a>
* <a href="#getTransactionCount">eth_getTransactionCount</a>
* <a href="#getStorageAt">eth_getStorageAt</a>
* <a href="#getCode">eth_getCode</a>
* <a href="#getBlockByHash">eth_getBlockByHash</a>
* <a href="#getBlockByNumber">eth_getBlockByNumber</a>
* <a href="#getHeaderByNumber">eth_getHeaderByNumber</a>
* <a href="#getHeaderByHash">eth_getHeaderByHash</a>
* <a href="#getBlockTransactionCountByHash">eth_getBlockTransactionCountByHash</a>
* <a href="#getBlockTransactionCountByNumber">eth_getBlockTransactionCountByNumber</a>
* <a href="#getTransactionReceipt">eth_getTransactionReceipt</a>
* <a href="#getTransactionByHash">eth_getTransactionByHash</a>
* <a href="#getTransactionByBlockHashAndIndex">eth_getTransactionByBlockHashAndIndex</a>
* <a href="#getTransactionByBlockNumberAndIndex">eth_getTransactionByBlockNumberAndIndex</a>
* <a href="#getRawTransactionByHash">eth_getRawTransactionByHash</a>
* <a href="#getRawTransactionByBlockHashAndIndex">eth_getRawTransactionByBlockHashAndIndex</a>
* <a href="#getRawTransactionByBlockNumberAndIndex">eth_getRawTransactionByBlockNumberAndIndex</a>
* <a href="#getUncleCountByHash">eth_getUncleCountByHash</a> 
* <a href="#getUncleCountByNumber">eth_getUncleCountByNumber</a>
* <a href="#getUncleByBlockHashAndIndex">eth_getUncleByBlockHashAndIndex</a> 
* <a href="#getUncleByBlockNumberAndIndex">eth_getUncleByBlockNumberAndIndex</a>
* <a href="#newFilter">eth_newFilter</a>
* <a href="#newBlockFilter">eth_newBlockFilter</a>
* <a href="#newPendingTransactionFilter">eth_newPendingTransactionFilter</a>
* <a href="#uninstallFilter">eth_uninstallFilter</a>
* <a href="#getFilterChanges">eth_getFilterChanges</a>
* <a href="#getFilterLogs">eth_getFilterLogs</a>
* <a href="#getLogs">eth_getLogs</a>
___
## <p id="clientVersion">web3_clientVersion</p>
* [WC-1] web3_clientVersion **MUST** return a string containing information about the client version.
* [WC-2] web3_clientVersion **MUST** start with the name of the client.
* [WC-3] web3_clientVersion **MUST** return the client version after the name.
* [WC-4] web3_clientVersion **MUST** return the operating system and architecture the client is running on.
* [WC-5] web3_clientVersion **MUST** return the name of the language being used.
## <p id="sha3">web3_sha3</p>
* [WS-1] web3_sha3 **MUST** return the Keccak256 hash of the given string. 
* [WS-2] web3_sha3 **MUST** return the Keccak256 hash of null when given "0x".
## <p id="version">net_version</p>
* [NV-1] net_version **MUST** return the id number of the network the client is currently connected to.
## <p id="peerCount">net_peerCount</p>
* [NP-1] net_peerCount **MUST** return the number of peer nodes that the client is currently connected to. 
## <p id="listening">net_listening</p>
* [NL-1] net_listening **MUST** return a boolean indicating whether the client is currently listening for network connections.
## <p id="protocolVersion">eth_protocolVersion</p>
* [EP-1] eth_protocolVersion **MUST** return the current Ethereum Wire Protocol (eth protocol) version he network is using.
## <p id="syncing">eth_syncing</p>
* [ESY-1] eth_syncing **MUST** return the progress of the client's sync of the network data.
* [ESY-2] eth_syncing **MUST** return false when the client is not syncing or already synced to the network.
* [ESY-3] eth_syncing **MUST** return the following for the sync progress:
   * The current block being synced on the client.
   * The current highest block known by the client.
   * The number of known and pulled states.
   * The block number that the client started syncing from.
## <p id="coinbase">eth_coinbase</p>
* [ECB-1] eth_coinbase **MUST** return the public address where the client's mining rewards are sent to.
* [ECB-2] eth_coinbase **MUST** error with code -32000 when the client does not have an address for the block reward to be sent to when not mining.
## <p id="etherbase">eth_etherbase</p>
* [EEB-1] eth_etherbase **MUST** return the public address where the client's mining rewards are sent to.
* [EEB-2] eth_etherbase **MUST** error with code -32000 when the client does not have an address for the block reward to be sent to when not mining.
## <p id="accounts">eth_accounts</p>
* [EA-1] eth_accounts **MUST** return the public addresses for each Ethereum account that the client you are using manages.
## <p id="mining">eth_mining</p>
* [EM-1] eth_mining **MUST** return true when the client has mining enabled, otherwise it **MUST** return false.
## <p id="getWork">eth_getWork</p>
* [EGW-1] eth_getWork **MUST** error with code -32000 when mining is not enabled.
* [EGW-2] eth_getWork **MUST** return the block header POW-hash, the seed hash for the DAG, the target condition, and the block number for the block being mined.
## <p id="submitWork">eth_submitWork</p>
* [ESW-1] eth_submitWork **MUST** return true when submitting the correct parameters to claim the block reward, otherwise false.
## <p id="hashrate">eth_hashrate</p>
* [EH-1] eth_hashrate **MUST** return the hashes per second that the client is using to mine blocks.
* [EH-2] eth_hashrate **MUST** return 0x0 when the client does not have mining enabled.
## <p id="submitHashrate">eth_submitHashrate</p>
* [ESH-1] eth_submitHashrate **MUST** return true when the client successfully submits a `hashrate` and an `id`.
* [ESH-2] eth_submitHashrate **MUST** return true when the client submits their hashrate while not mining.
* [ESH-3] eth_submitHashrate **MUST** return true when the client submits their hashrate while syncing to the network.
* [ESH-4] eth_submitHashrate **MUST** submit 0x0 when `hashrate` is null.
* [ESH-5] eth_submitHashrate **MUST NOT** error when `id` is equal to the `id` or another mining client.
## <p id="gasPrice">eth_gasPrice</p>
* [EGP-1] eth_gasPrice **MUST** return the current price per unit of gas in wei that the client is charging.
## <p id="maxPriorityFeePerGas">eth_maxPriorityFeePerGas</p>
* [EMPFPG-1] eth_maxPriorityFeePerGas **MUST** return the clients price per unit of gas - 7.
## <p id="feeHistory">eth_feeHistory</p>
* [EFH-1] eth_feeHistory **MUST** return the following information for number of blocks specified by the `blockCount` parameter stopping at the `highestBlock` parameter.
  * An array containing the base fee per gas for each block plus then block after `highestBlock`.
  * An array containing the ratio of the gas used by each block.
  * An array containing arrays with the requested `rewardPercentiles` for each block.
  * The oldest block used for the request.
* [EFH-2] eth_feeHistory **MUST** use the available range of blocks when the requested `blockCount` range can't be retrieved.
* [EFH-3] eth_feeHistory **MUST** allow block tags to be used for `highestBlock`.
* [EFH-4] eth_feeHistory **MUST** error with code -32000 when `highestBlock` is ahead of the chain.
## <p id="blockNumber">eth_blockNumber</p>
* [EBN-1] eth_blockNumber **MUST** return the block number for the most recent block mined.
* [EBN-2] eth_blockNumber **MUST** return "0x0" when the client is not synced to the network.
## <p id="sign">eth_sign</p>
* [ESN-1] eth_sign **MUST** return the Ethereum specific signature detailed in [EIP-191](https://eips.ethereum.org/EIPS/eip-191) for the given unlocked `address` and `message`.
* [ESN-2] eth_sign **MUST** error with code -32000 when the account corresponding to the `address` is not unlocked.
* [ESN-3] eth_sign **MUST** error with code -32000 when the account corresponding to the `address` is not owned by the client.
## <p id="call">eth_call</p>
* [EC-1] eth_call **MUST** return the result of the given transaction.
* [EC-2] eth_call **MUST** accept all current transaction types. Legacy transactions and [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718) "typed" transactions.
* [EC-3] eth_call **MUST NOT** sign nor propagate the transaction (if it happens to be signed) to the network.
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
* [EC-14] eth_call **MUST** error with code -32000 when the `data` or `input` being deployed causes and error in the EVM.
* [EC-15] eth_call **MUST** error with code -32000 when the `gas` is too low to execute the call.
* [EC-16] eth_call **MUST** error with code -32000 when using `gasPrice` with `maxFeePerGas` or `maxPriorityFeePerGas`. 
* [EC-17] eth_call **MUST** use networks chain id for `chainId`
* [EC-18] eth_call **MUST** error with code -32000 when using block number or block hash for `defaultBlockParameter` while syncing to the network.
* [EC-19] eth_call **MUST** return 0x0 when using block tags for `defaultBlockParameter` while syncing to the network.
## <p id="fillTransaction">eth_fillTransaction</p>
* [EFT-1] eth_fillTransaction **MUST** fill in the missing transaction parameters of the given transaction.
* [EFT-2] eth_fillTransaction **MUST** return the return the raw transaction and JSON transaction object of the filled transaction.
* [EFT-3] eth_filterTransaction **MUST NOT** sign the transaction.
* [EFT-4] eth_filterTransaction **MUST** error with code -32000 when the `from` address does not have enough Ether to pay for the transaction.
* [EFT-5] eth_filterTransaction **MUST** allow both `data` and `input` to be used for contract creation,
* [EFT-6] eth_filterTransaction **MUST** error with code -32000 when the `data` and `input` both specified and not equal
* [EFT-7] eth_filterTransaction **MUST** error with code -32000 when `data` or `input` or `to` is not specified.
* [EFT-8] eth_filterTransaction **MUST** error with code -3200 when `data` or `input` provided caused an EVM error when the `gas` is not specified, otherwise it is ignored.
## <p id="createAccessList">eth_createAccessList</p>
* [ECAL-1] eth_createAccessList **MUST** return an [EIP-2930](https://eips.ethereum.org/EIPS/eip-2930) access list biased off the given `transaction` and estimated gas cost when using the access list in the `transaction`.
* [ECAL-2] eth_createAccessList **MUST** use "latest" when the `defaultBlockParameter` is not specified.
* [ECAL-3] eth_createAccessList **MUST** use the estimated gas when the `gas` is not specified.
* [ECAL-4] eth_createAccessList **MUST** error with code -32000 when `gas` is too low.
* [ECAL-5] eth_createAccessList **MUST** error with code -32000 when the `from` address does not have enough Ether to pay for the transaction.
## <p id="estimateGas">eth_estimateGas</p>
* [EEG-1] eth_estimateGas **MUST** return the estimated amount of gas the given `transaction` will take to execute.
* [EEG-2] eth_estimateGas **MUST** error with code -32000 when the `from` address has insufficient Ether to execute the given `transaction`.
* [EEG-3] eth_estimateGas **MUST NOT** check if the `from` account has sufficient funds when estimating contract deployment.
* [EEG-4] eth_estimateGas **MUST** use 0x0000000000000000000000000000000000000000 for `from` when it is not specified.
* [EEG-5] eth_estimateGas **MUST** error with code -32000 when estimating a contract creation that causes an error within the EVM.
## <p id="signTransaction">eth_signTransaction</p>
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
* [ESNT-26] eth_signTransaction **MUST** error with code -32000 when deploying contract with no `data` or `input`.
* [ESNT-27] eth_signTransaction **MUST** error with code -32000 when specifying a `chainId` that is different from the network's chain Id.
## <p id="sendRawTransaction">eth_sendRawTransaction</p>
* [ESRT-1] eth_sendRawTransaction **MUST** return the transaction hash after submitting an encoded signed transaction to the network.
* [ESRT-2] eth_sendRawTransaction **MUST** allow users to send transaction where `gasPrice` or `maxFeePerGas` or `maxPriorityFeePerGas` are below network average and may never be executed.
* [ESRT-3] eth_sendRawTransaction **MUST** error with code -32000 when the `from` address does not have enough Ether to pay for the transaction.
* [ESRT-4] eth_sendRawTransaction **MUST** error with code -32000 when nonce is too low.
* [ESRT-5] eth_sendRawTransaction **MUST** error with code -32000 when the `gas` is too low.
* [ESRT-6] eth_sendRawTransaction **MUST** error with code -32000 when the user did not raise the `maxFeePerGas` enough when trying to replace a pending transaction.
* [ESRT-7] eth_sendRawTransaction **MUST** error with code -32000 when `transaction` is not properly encoded.
* [ESRT-8] eth_sendRawTransaction **MUST** error with code -3200 when sending an encoded transaction while syncing to the network.
* [ESRT-9] eth_sendRawTransaction **MUST** allow sending of contract creation transactions with code that causes the EVM to error. Resulting in a contract with no code.
## <p id="sendTransaction">eth_sendTransaction</p>
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
* [EST-28] eth_sendTransaction **MUST** error with code -32000 when deploying contract with no `data` or `input`.
* [EST-29] eth_sendTransaction **MUST** error with code -3200 when `data` or `input` provided caused an EVM error when the `gas` is not specified, otherwise it is ignored.
* [EST-30] eth_sendTransaction **MUST** estimate the amount of gas needed to complete the transaction and use that value for `gas` when not specified.
* [EST-31] eth_sendTransaction **MUST** use 0x0 for `gas` when null.
* [EST-32] eth_sendTransaction **MUST** use 0x0 for `gasPrice` when null.
* [EST-33] eth_sendTransaction **MUST** error with code -32000 if the transaction's `chainId` is different than the network's Id.
* [EST-34] eth_sendTransaction **MUST** use the estimated price per gas and max priority fee per gas for `maxFeePerGas` and `maxPriorityFeePerGas` when not specified.
* [EST-35] eth_sendTransaction **MUST** error with code -32000 when trying to send a transaction while syncing to the network.
## <p id="pendingTransactions">eth_pendingTransactions</p>
* [EPT-1] eth_pendingTransactions **MUST** return the transactions sent by accounts that are owned by the client that are currently in the transaction pool.
* [EPT-2] eth_pendingTransactions **MUST** return an empty array when syncing to the network.
## <p id="resend">eth_resend</p>
* [ERS-1] eth_resend **MUST** error with code -32000 when given any transaction. [geth Issue](https://github.com/ethereum/go-ethereum/issues/23964)
## <p id="getProof">eth_getProof</p>
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
## <p id="getBalance">eth_getBalance</p>
* [EGB-1] eth_getBalance **MUST** return the account balance of the `address` at the given `defaultBlockParameter`. 
* [EGB-2] eth_getBalance **MUST** return "0x0" when the client is not synced to the network.
* [EGB-3] eth_getBalance **MUST** error with code -32000 checking the balance at block that the client does not know about.
## <p id="getTransactionCount">eth_getTransactionCount</p>
* [EGTC-1] eth_getTransactionCount **MUST** return the nonce of the account is with the given `address` at the block requested by the `defaultBlockParameter`. 
* [EGTC-2] eth_getTransactionCount **MUST** error with code -32000 when calling a block that does not exist or is unavailable.
* [EGTC-3] eth_getTransactionCount **MUST** error with code -32000 when using a block number or block hash when syncing to the network.
## <p id="getStorageAt">eth_getStorageAt</p>
* [EGS-1] eth_getStorageAt **MUST** return the data stored within the `storageSlot` of the given `address` at the given `defaultBlockParameter`. 
* [EGS-2] eth_getStorageAt **MUST**  error with code -32000 when the client does not have the state of the `defaultBlockParameter` requested. 
* [EGS-3] eth_getStorageAt **MUST** return 0x00...0 when using `blockTag` while syncing to the network.
* [EGS-4] eth_getStorageAt **MUST** error with code -32000 when using `blockNumber` and `blockHash` while syncing to the network.
## <p id="getCode">eth_getCode</p>
* [EGC-1] eth_getCode **MUST** return the deployed smart contract code at the given `address` and `defaultBlockParameter`.
* [EGC-2] eth_getCode **MUST**  error with code -32000 when the state information is not available for the requested block.
* [EGC-3] eth_getCode **MUST** error with code -32000 when using block numbers or block hashes while syncing to the network.
* [EGC-4] eth_getCode **MUST** return 0x0 when using block tags when syncing to the network.
## <p id="getBlockByHash">eth_getBlockByHash</p>
* [EGBH-1] eth_getBlockByHash **MUST** return the block information for the block with the given `blockHash`.
* [EGBH-2] eth_getBlockByHash **MUST** return null when the given `blockHash` is unavailable or does not correspond to a block.
* [EGBH-3] eth_getBlockByHash **MUST** return block information with only transaction hashes when `hydratedTransactions` is false. Otherwise, it should include full transaction objects.
## <p id="getBlockByNumber">eth_getBlockByNumber</p>
* [EGBN-1] eth_getBlockByNumber **MUST** return the block information for the block with the given `blockNumber` or `blockTag` when available.
* [EGBN-2] eth_getBlockByNumber **MUST** return null when the given `blockNumber` is unavailable or does not correspond to a block. 
* [EGBN-3] eth_getBlockByNumber **MUST** return block information with only transaction hashes when `hydratedTransactions` is false. Otherwise, it should include full transaction objects.
## <p id="getHeaderByNumber">eth_getHeaderByNumber</p>
* [EGHN-1] eth_getHeaderByNumber **MUST** return the block header with the given `blockNumber` or `blockTag`.
* [EGHN-2] eth_getHeaderByNumber **MUST** return null when the requested block does not exist or is unavailable.
## <p id="getHeaderByHash">eth_getHeaderByHash</p>
* [EGHH-1] eth_getHeaderByNumber **MUST** return the block header with the given `blockHash`.
* [EGHH-2] eth_getHeaderByNumber **MUST** return null when the requested block does not exist or is unavailable.
## <p id="getBlockTransactionCountByHash">eth_getBlockTransactionCountByHash</p>
* [EGTCH-1] eth_getBlockTransactionCountByHash **MUST** return the number of transactions within the block with the given `blockHash`.
* [EGTCH-2] eth_getBlockTransactionCountByHash **MUST** return null when the `blockHash` does not correspond to a block.
* [EGTCH-3] eth_getBlockTransactionCountByHash **MUST** return null when the client is currently syncing to the network.
## <p id="getBlockTransactionCountByNumber">eth_getBlockTransactionCountByNumber</p>
* [EGTCN-1] eth_getBlockTransactionCountByNumber **MUST** return the number of transactions within the block with the given `blockNumber` or `blockTag`.
* [EGTCN-2] eth_getBlockTransactionCountByNumber **MUST** return null when the `blockNumber` is not known by the client or the block does not exist.
* [EGTCN-3] eth_getBlockTransactionCount **MUST** return 0x0 when block tags are used while syncing to the network.
## <p id="getTransactionReceipt">eth_getTransactionReceipt</p>
* [EGTR-1] eth_getTransactionReceipt **MUST** return the transaction receipt for the transaction with the given `transactionHash`.
* [EGTR-2] eth_getTransactionReceipt **MUST** return null when the transaction with the given `transactionHash` does not exist or is not available.
* [EGTR-3] eth_getTransactionReceipt **MUST** return null when the transaction has not been included in a block.
## <p id="getTransactionByHash">eth_getTransactionByHash</p>
* [EGTH-1] eth_getTransactionByHash **MUST** return the transaction object for the transaction with the given `transactionHash`.
* [EGTH-2] eth_getTransactionByHash **MUST** return null when the transaction with the given `transactionHash` does not exist or is not available.
## <p id="getTransactionByBlockHashAndIndex">eth_getTransactionByBlockHashAndIndex</p>
* [EGTHI-1] eth_getTransactionByBlockHashAndIndex **MUST** return the transaction object with the given `blockHash` and `transactionIndex`.
* [EGTHI-2] eth_getTransactionByBlockHashAndIndex **MUST** return null block with the given `blockHash` does not exist or is not available.
* [EGTHI-3] eth_getTransactionByBlockHashAndIndex **MUST** return null when there is no transaction at the given `transactionIndex` in the requested block.
## <p id="getTransactionByBlockNumberAndIndex">eth_getTransactionByBlockNumberAndIndex</p>
* [EGTNI-1] eth_getTransactionByBlockNumberAndIndex **MUST** return the transaction object with the given `blockNumber` or `blockTag` and `transactionIndex`.
* [EGTNI-2] eth_getTransactionByBlockNumberAndIndex **MUST** return null when the block with given `blockNumber` does not exist or is not available.
* [EGTNI-3] eth_getTransactionByBlockNumberAndIndex **MUST** return null when the given `transactionIndex` does not exist in the requested block.
## <p id="getRawTransactionByHash">eth_getRawTransactionByHash</p>
* [EGRTH-1] eth_getRawTransactionByHash **MUST** return the encoded transaction associated with the given `transactionHash`.
* [EGRTH-2] eth_getRawTransactionByHash **MUST** return 0x0 when the transaction with the given `transactionHash` does not exist or is not available.
## <p id="getRawTransactionByBlockHashAndIndex">eth_getRawTransactionByBlockHashAndIndex</p>
* [EGRTHI-1] eth_getRawTransactionByBlockNumberAndIndex **MUST** return the encoded transaction associated with the given `blockHash` and `transactionIndex`.
* [EGRTHI-2] eth_getRawTransactionByBlockNumberAndIndex **MUST** return 0x0 the block with the given `blockHash` does not exist or is unavailable.
* [EGRTHI-3] eth_getRawTransactionByBlockNumberAndIndex **MUST** return 0x0 when no transaction exists at the given `transactionIndex`.
## <p id="getRawTransactionByBlockNumberAndIndex">eth_getRawTransactionByBlockNumberAndIndex</p>
* [EGRTNI-1] eth_getRawTransactionByBlockNumberAndIndex **MUST** return the encoded transaction associated with the given `blockNumber` or `blockTag` and `transactionIndex`.
* [EGRTNI-2] eth_getRawTransactionByBlockNumberAndIndex **MUST** return 0x0 when the requested block does not exist or is unavailable.
* [EGRTNI-3] eth_getRawTransactionByBlockNUmberAndIndex **MUST** return 0x0 when no transaction exists at the given `transactionIndex`.
## <p id="getUncleCountByHash">eth_getUncleCountByHash</p>
* [EGUCH-1] eth_getUncleCountByHash **MUST** return number of uncle blocks that the block with the given `blockHash` has.
* [EGUCH-2] eth_getUncleCountByHash **MUST** return null the block with the given `blockHash` does not exist or is not available.
## <p id="getUncleCountByNumber">eth_getUncleCountByNumber</p>
* [EGUCN-1] eth_getUncleCountByNumber **MUST** return number of uncle blocks that the block with the given `blockNumber` or `blockTag` has.
* [EGUCN-2] eth_getUncleCountByNumber **MUST** return null when the requested block does not exist or is not available. 
* [EGUCN-3] eth_getUncleCountByNumber **MUST** return 0x0 when using `blockTag` "pending".
## <p id="getUncleByBlockHashAndIndex">eth_getUncleByBlockHashAndIndex</p>
* [EGUHI-1] eth_getUncleCountByHashAndIndex **MUST** return the uncle block information at the `uncleIndex` of the block with the given `blockHash`.
* [EGUHI-2] eth_getUncleCountByHashAndIndex **MUST** return null when the block is unavailable or does not exist at the given `blockHash`.
* [EGUHI-3] eth_getUncleCountByHashAndIndex **MUST** return null when the block has no uncles at the `uncleIndex`.
## <p id="getUncleByBlockNumberAndIndex">eth_getUncleByBlockNumberAndIndex</p> 
* [EGUNI-1] eth_getUncleByBlockNumberAndIndex **MUST** return the uncle block information at the `UncleIndex` of the block with the given `blockNumber` or `blockTag`.
* [EGUNI-2] eth_getUncleByBlockNumberAndIndex **MUST** return null when the block requested is unavailable or does not exist.
* [EGUNI-3] eth_getUncleByBlockNumberAndIndex **MUST** return null when the block has no uncles at the `uncleIndex`.
## <p id="newFilter">eth_newFilter</p>
* [ENF-1] eth_newFilter **MUST** create a filter on the client that looks through each transaction log to see if it contains any the of the requested events.
* [ENF-2] eth_newFilter **MUST** allow `fromBlock` and `toBlock` to use both block numbers and block tags.
* [ENF-3] eth_newFilter **MUST** allow `from` and `to` to be used instead of `fromBlock` and `toBlock`.
* [ENF-4] eth_newFilter **MUST** give precedence to `toBlock` and `fromBlock` when used with `to` and `from`.
* [ENF-5] eth_newFilter **MUST** use "latest" for `fromBlock` and or `toBlock` when it is not specified.
* [ENF-6] eth_newFilter **MUST** error with code -32000 when the `fromBlock` is greater than the `toBlock`, except when the `toBlock` is set to latest and `fromBlock` is ahead of the chain.
* [ENF-7] eth_newFilter **MUST** allow `blockHash` to be used instead of `fromBlock` and `toBlock`.
* [ENF-8] eth_newFilter **MUST** error with code -32000 when `blockHash` is used with `fromBlock` and or `toBlock` in the same request.
* [ENF-9] eth_newFilter **MUST** allow `address` to be a single address or an array of addresses.
* [ENF-10] eth_newFilter **MUST** use null for `address` when it is not specified or when it is an empty array.
* [ENF-11] eth_newFilter **MUST** allow `topics` array to contain more than 4 values.
## <p id="newBlockFilter">eth_newBlockFilter</p>
* [ENBF-1] eth_newBlockFilter **MUST** create a filter on the client that tracks when the client receives new blocks.
* [ENBF-2] eth_newBlockFilter **MUST** return the id of the newly created block filter. 
## <p id="newPendingTransactionFilter">eth_newPendingTransactionFilter</p>
* [ENPTF-1] eth_newPendingTransactionFilter **MUST** create a filter on the client that tracks the hash of each pending transaction that the client receives.
* [ENPTF-2] eth_newPendingTransactionFilter **MUST** return the id of the newly created pending transaction filter.
## <p id="uninstallFilter">eth_uninstallFilter</p>
* [EUF-1] eth_uninstallFilter **MUST** delete the filter with the given `filterId` from the client.
* [EUF-2] eth_uninstallFilter **MUST** return true when the given filter has been successfully uninstalled, otherwise it **MUST** return false.
## <p id="getFilterChanges">eth_getFilterChanges</p>
* [EGFC-1] eth_getFilterChanges **MUST** return the block hashes of new blocks the client received since the filter was called last or first created, when the `filterId` corresponds to a block filter.
* [EGFC-3] eth_getFilterChanges **MUST** return the transaction hashes or each pending transaction received since the filter was last called or first created, when the `filterId` corresponds to a pending transaction filter.
* [EGFC-4] eth_getFilterChanges **MUST** return an empty array when calling a pending transaction filter while syncing to the network.
* [EGFC-5] eth-getFilterChanges **MUST** return all the logs that match the filters topics since the filter was last called or first created, when the `filterId` corresponds to a regular filter.
* [EGFC-7] eth_getFilterChanges **MUST** error with code -32000 when the given `filterId` does not correspond to an active filter on the client.
## <p id="getFilterLogs">eth_getFilterLogs</p>
* [EGFL-1] eth_getFilterLogs **MUST** return all the logs that match the filters topics for the given `filterId`'s specified range.
* [EGFL-2] eth_getFilterLogs **MUST** only return the logs that match the filters parameters from the latest synced block when syncing to the network.
* [EGFL-3] eth_getFilterLogs **MUST** error with code -32000 when the given `filterId` does not correspond to an active filter on the client.
* [EGFL-4] eth_getFilterLogs **MUST** error with code -32000 when the given `filterId` corresponds to an active block filter or pending transaction filter on the client.
* [EGFL-5] eth_getFilterLogs **MUST** error with code -32005 when trying to return more than 1000 logs.
## <p id="getLogs">eth_getLogs</p>
* [EGL-1] eth_getLogs **MUST** look through all of the transaction logs of the client within the specified range.
* [EGL-2] eth_getLogs **MUST** return all of the logs that meet the filter requirements.
* [EGL-3] eth_getLogs **MUST** allow `fromBlock` and `toBlock` to use both block numbers and block tags.
* [EGL-4] eth_getLogs **MUST** allow `from` and `to` to be used instead of `fromBlock` and `toBlock`.
* [EGL-5] eth_getLogs **MUST** give precedence to `toBlock` and `fromBlock` when used with `to` and `from`.
* [EGL-6] eth_getLogs **MUST** error with code -32000 when the `fromBlock` is greater than the `toBlock`, except when the `toBlock` is set to latest and `fromBlock` is ahead of the current block.
* [EGL-7] eth_getLogs **MUST** use latest for `fromBlock` and or `toBlock` when it is not specified.
* [EGL-8] eth_getLogs **MUST** allow `blockHash` to be used in place of `toBlock` and `fromBlock`.
* [EGL-9] eth_getLogs **MUST** error with code -32602 when using `blockHash` with `fromBlock` and or `toBlock` in the same request.
* [EGL-10] eth_getLogs **MUST** use null for `address` when it is not specified or when it is an empty array.
* [EGL-11] eth_getLogs **MUST** allow `topics` array to contain more than 4 values.
* [EGL-12] eth_getLogs **MUST** return logs that match the parameters from only the latest synced block when syncing to the network.
* [EGL-13] eth_getLogs **MUST** error with code -32005 when trying to return more than 1000 logs.
# 3 Errors
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
# References
# Appendix
* [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
* [Ethereum Yellow Paper](https://ethereum.github.io/yellowpaper/paper.pdf)
* [JSON Standard](https://www.json.org/json-en.html)
* [HTTP/2](https://httpwg.org/specs/rfc7540.html)
