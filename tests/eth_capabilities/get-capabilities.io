// retrieves the node's effective routing capabilities
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_capabilities"}
<< {"jsonrpc":"2.0","id":1,"result":{"head":{"number":"0x2d","hash":"0xe27a3e81bd7cfe2aec2cc9e832c73a17c93e7efcf659cf4b39883b96c48708c2"},"state":{"disabled":false,"oldestBlock":"0x0"},"tx":{"disabled":false,"oldestBlock":"0x0"},"logs":{"disabled":false,"oldestBlock":"0x0","deleteStrategy":{"type":"window","retentionBlocks":"0x23dbb0"}},"receipts":{"disabled":false,"oldestBlock":"0x0"},"blocks":{"disabled":false,"oldestBlock":"0x0"},"stateproofs":{"disabled":false,"oldestBlock":"0x0"}}}
