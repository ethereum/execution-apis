>> {"jsonrpc": "2.0", "id": 1, "method": "eth_multicall", "params": [ [ { "calls": [ { "to": "0x4B62D7C9C4e5c7150Eda45F7552a25C7Cd726bF6", "data": "0x42cbb15c" } ] }, { "blockOverrides": { "number": "0x4999999" }, "calls": [ { "to": "0x4B62D7C9C4e5c7150Eda45F7552a25C7Cd726bF6", "data": "0x42cbb15c" } ] } ], "latest" ] }
<< {"jsonrpc":"2.0","id":1,"result":[{"status":"0x1","returnData":"0x0000000000000000000000000000000000000000000000000000000001031b64","gasUsed":"0x5338","logs":[]},{"status":"0x1","returnData":"0x0000000000000000000000000000000000000000000000000000000004999999","gasUsed":"0x5338","logs":[]}]}
