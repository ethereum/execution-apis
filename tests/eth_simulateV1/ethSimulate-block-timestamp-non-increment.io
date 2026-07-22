// Error: simulates calls with timestamp staying the same
>> {"jsonrpc":"2.0","id":1,"method":"eth_simulateV1","params":[{"blockStateCalls":[{"blockOverrides":{"time":"0x228"}},{"blockOverrides":{"time":"0x228"}}]},"latest"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-38021,"message":"block timestamps must be in order: 552 \u003c= 552"}}
