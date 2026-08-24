// Calls net_peerCount to retrieve the number of connected peers. The test client runs without peers, so the expected value is zero.
>> {"jsonrpc":"2.0","id":1,"method":"net_peerCount"}
<< {"jsonrpc":"2.0","id":1,"result":"0x0"}
