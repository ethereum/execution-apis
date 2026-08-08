// retrieves the node's effective routing capabilities
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_capabilities"}
<< {"jsonrpc":"2.0","id":1,"result":{"head":{"number":"0x36","hash":"0xfbfbdd694e7ab88387234ef9ba3df04aef99b089870f5fd37fe4cb5998e5d4ce"},"state":{"disabled":false,"oldestBlock":"0x0"},"tx":{"disabled":false,"oldestBlock":"0x0"},"logs":{"disabled":false,"oldestBlock":"0x0","deleteStrategy":{"type":"window","retentionBlocks":"0x23dbb0"}},"receipts":{"disabled":false,"oldestBlock":"0x0"},"blocks":{"disabled":false,"oldestBlock":"0x0"},"stateproofs":{"disabled":false,"oldestBlock":"0x0"}}}
