// retrieves the node's effective routing capabilities
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_capabilities"}
<< {"jsonrpc":"2.0","id":1,"result":{"head":{"number":"0x36","hash":"0xa2223685e7dbcf60dbf07cbfc4e8836f584e3994011cbe97d3d0afc574f2b257"},"state":{"disabled":false,"oldestBlock":"0x0"},"tx":{"disabled":false,"oldestBlock":"0x0"},"logs":{"disabled":false,"oldestBlock":"0x0","deleteStrategy":{"type":"window","retentionBlocks":"0x23dbb0"}},"receipts":{"disabled":false,"oldestBlock":"0x0"},"blocks":{"disabled":false,"oldestBlock":"0x0"},"stateproofs":{"disabled":false,"oldestBlock":"0x0"}}}
