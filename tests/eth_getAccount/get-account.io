>> {"jsonrpc":"2.0","id":9,"method":"eth_getAccount","params":["0x000000000000000000000000000000000000dead","latest"]}
<< {"jsonrpc":"2.0","id":9,"result": null}
>> {"jsonrpc":"2.0","id":10,"method":"eth_getAccount","params":["0xaa00000000000000000000000000000000000000","latest"]}
<< {"jsonrpc":"2.0","id":10,"result": { "codeHash": "0xce92c756baff35fa740c3557c1a971fd24d2d35b7c8e067880d50cd86bb0bc99", "root": "0x8afc95b7d18a226944b9c2070b6bda1c3a36afcc3730429d47579c94b9fe5850", "balance": "0x1", "nonce": "0x1"}}
>> {"jsonrpc":"2.0","id":11,"method":"eth_getAccount","params":["0xaa00000000000000000000000000000000000000","0xffff"]}
<< {"jsonrpc":"2.0","id":11,"error":{"code":-32000,"message":"header not found"}}
