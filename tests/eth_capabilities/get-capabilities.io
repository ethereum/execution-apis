// retrieves the node's effective routing capabilities
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_capabilities"}
<< {"jsonrpc":"2.0","id":1,"result":{"head":{"number":"0x3c","hash":"0x52bccfc9c4d4874b1898c0fbd0ae3b598a0edc7b94d9b796857dbbe1093c32e0"},"state":{"disabled":false,"oldestBlock":"0x0"},"tx":{"disabled":false,"oldestBlock":"0x0"},"logs":{"disabled":false,"oldestBlock":"0x0","deleteStrategy":{"type":"window","retentionBlocks":"0x23dbb0"}},"receipts":{"disabled":false,"oldestBlock":"0x0"},"blocks":{"disabled":false,"oldestBlock":"0x0"},"stateproofs":{"disabled":false,"oldestBlock":"0x0"}}}
