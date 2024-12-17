// gets block finalized
>> {"jsonrpc":"2.0","id":1,"method":"eth_getBlockByNumber","params":["finalized",true]}
<< {"jsonrpc":"2.0","id":1,"result":{"baseFeePerGas":"0x4426f5d","blobGasUsed":"0x20000","difficulty":"0x0","excessBlobGas":"0x0","extraData":"0x","gasLimit":"0x23f3e20","gasUsed":"0x23114","hash":"0x3890d259f63e7350fa4f747e5627abc7d95ef106b1f80d96c9586383f432cc2b","logsBloom":"0x00000000000000000000000000000000000000000000000000000100800000000000000000000000000000000000000000040000000000000000000000000000000000000000000000004000000000000200000000000000000000000000002000000000000000000040000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000000000000000000009000000000000000800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000","miner":"0x0000000000000000000000000000000000000000","mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000","nonce":"0x0000000000000000","number":"0x14","parentBeaconBlockRoot":"0x3cc9f65fc1f46927eb46fbf6d14bc94af078fe8ff982a984bdd117152cd1549f","parentHash":"0x5a0a138de5ea6f00d6d8bb652a3d66a4adef6f44ceb1394f5c3fd2350e6b50da","receiptsRoot":"0x352a1ceb0040f363d3be40a62c347548c4737a2264a8dc42295d014d29c7c7a2","requestsHash":"0xe3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","size":"0x4ca","stateRoot":"0x42ab203ee8a47d8baa7bac7178486747393952f07887bf436b9814a4e725ec4f","timestamp":"0xc8","transactions":[{"blockHash":"0x3890d259f63e7350fa4f747e5627abc7d95ef106b1f80d96c9586383f432cc2b","blockNumber":"0x14","from":"0x7435ed30a8b4aeb0877cef0c6e8cffe834eb865f","gas":"0x186a0","gasPrice":"0x4426f5e","maxFeePerGas":"0x4426f5e","maxPriorityFeePerGas":"0x1","maxFeePerBlobGas":"0x20000","hash":"0xe1acd16bccdfc280a1022ab1e2baccf663a5e51770d3a2b9acb91176f2446445","input":"0x7346f7df2f4852bf656d6974","nonce":"0x48","to":"0x7dcd17433742f4c0ca53122ab541d0ba67fc27df","transactionIndex":"0x0","value":"0x3","type":"0x3","accessList":[{"address":"0x7dcd17433742f4c0ca53122ab541d0ba67fc27df","storageKeys":["0x0000000000000000000000000000000000000000000000000000000000000000","0xcb55d89f2ee070d017b426876d6072d91c2a7311ade9a1bed2f8200127ec380e"]}],"chainId":"0xc72dd9d5e883e","blobVersionedHashes":["0x015a4cab4911426699ed34483de6640cf55a568afc5c5edffdcbd8bcd4452f68"],"v":"0x0","r":"0xb3ecfcecf2bfbf57c2129bd7dd90ae8d0188f51078937de75141e7cf232de5d1","s":"0x4dd67185c36813d4e03f06d68f557a8f07646ff14282c0bf3936af6dde2e93db","yParity":"0x0"},{"blockHash":"0x3890d259f63e7350fa4f747e5627abc7d95ef106b1f80d96c9586383f432cc2b","blockNumber":"0x14","from":"0x7435ed30a8b4aeb0877cef0c6e8cffe834eb865f","gas":"0x186a0","gasPrice":"0x4426f5e","hash":"0x7d4006cdcb52ab5cd11e68aa3c3f0b9356b399d9f27503b3bf8db01da345f0e6","input":"0x4520339faf61a8c6656d6974","nonce":"0x49","to":"0x7dcd17433742f4c0ca53122ab541d0ba67fc27df","transactionIndex":"0x1","value":"0x2","type":"0x0","chainId":"0xc72dd9d5e883e","v":"0x18e5bb3abd109f","r":"0x9f9359f74a7add08ab65ed7082bd6d7032b2ff4507cf5b861a6e0f49a164c7c9","s":"0x6004559d997c5bdf73237fdfc72c02c0117633ada8c3719a747a9589be61a3f"},{"blockHash":"0x3890d259f63e7350fa4f747e5627abc7d95ef106b1f80d96c9586383f432cc2b","blockNumber":"0x14","from":"0x7435ed30a8b4aeb0877cef0c6e8cffe834eb865f","gas":"0x5208","gasPrice":"0x4426f5e","maxFeePerGas":"0x4426f5e","maxPriorityFeePerGas":"0x1","hash":"0x07f663926bd370f5c558d5c550ea07b66933038e8f034d1106aff535e5c392e0","input":"0x","nonce":"0x4a","to":"0x4a0f1452281bcec5bd90c3dce6162a5995bfe9df","transactionIndex":"0x2","value":"0x1","type":"0x2","accessList":[],"chainId":"0xc72dd9d5e883e","v":"0x0","r":"0x5351822237ac6874ed552603cc0472ec418b597ac17c0ab907787f79e56c67d5","s":"0x7d17ecbfe738c84a9e7a465818b02b7c7cf1f12c84e20db46459e8fb77021abe","yParity":"0x0"},{"blockHash":"0x3890d259f63e7350fa4f747e5627abc7d95ef106b1f80d96c9586383f432cc2b","blockNumber":"0x14","from":"0x7435ed30a8b4aeb0877cef0c6e8cffe834eb865f","gas":"0x5208","gasPrice":"0x4426f5e","hash":"0xe3764461aad4083151dbd686e8d10a5b6ba8589a19a7d1e7bfc67d64ae36ff57","input":"0x","nonce":"0x4b","to":"0x1f5bde34b4afc686f136c7a3cb6ec376f7357759","transactionIndex":"0x3","value":"0x1","type":"0x1","accessList":[],"chainId":"0xc72dd9d5e883e","v":"0x1","r":"0x4a7a18bf2abcf5d7bb392b48676ec55893662deb253fb921c0dce1e77b0b24b6","s":"0x56c10bd3cf6d4141feff6d2e1a934fda29de0b418427a8e7725424ba49f1a1bb","yParity":"0x1"}],"transactionsRoot":"0x4bd2c9f4122998ca263eb98b4a1e42ae242ff8e4024562a5345546ddbdb84f7c","uncles":[],"withdrawals":[],"withdrawalsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"}}
