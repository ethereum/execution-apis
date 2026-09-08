// replaces a pending transaction without the required price bump
>> {"jsonrpc":"2.0","id":1,"method":"eth_sendRawTransaction","params":["0xf86c058405763d658261a894aa000000000000000000000000000000000000000a8222228718e5bb3abd10a0a00143985bad80a4c76d32fd0deca585931c5998b00b00b1f76c0377bb2592b7fda05f1be04cd9e422fb80f4acee750812bec39db8d9782256dbac21cc217793ebbb"]}
<< {"jsonrpc":"2.0","id":1,"result":"0xdf2b049904f977cff86ae87ab3fbd82efafc5bf98c8c285994f415ae09347ba5"}
>> {"jsonrpc":"2.0","id":2,"method":"eth_sendRawTransaction","params":["0xf86c058405763d658261a894aa00000000000000000000000000000000000000148233338718e5bb3abd109fa0dda8fd585c96a17f49cf8c6efa2a5ebdd9ea21f1e5e78f23cb45ddc4ecdc002fa01d48d97ee713c096ed1f5478775e5140fb93b688128ee56d669f0b75c9b5f0ae"]}
<< {"jsonrpc":"2.0","id":2,"error":{"code":1002,"message":"replacement transaction underpriced"}}
