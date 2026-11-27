// a 1-wei transfer with 32 zero bytes of calldata to an existing account runs no code, so the EIP-7623 calldata floor binds and the estimate is at least the floor; an authorization costs more
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_estimateGas","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","input":"0x0000000000000000000000000000000000000000000000000000000000000000","nonce":"0x0","to":"0x14e46043e63d0e3cdcf2530519f4cfaf35058cb2","value":"0x1"}]}
<< {"jsonrpc":"2.0","id":1,"result":"0x5b58"}
>> {"jsonrpc":"2.0","id":2,"method":"eth_estimateGas","params":[{"authorizationList":[{"address":"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","chainId":"0x1","nonce":"0x0","r":"0x1111111111111111111111111111111111111111111111111111111111111111","s":"0x2222222222222222222222222222222222222222222222222222222222222222","yParity":"0x0"}],"from":"0x0c2c51a0990aee1d73c1228de158688341557508","input":"0x0000000000000000000000000000000000000000000000000000000000000000","nonce":"0x0","to":"0x14e46043e63d0e3cdcf2530519f4cfaf35058cb2","type":"0x4","value":"0x1"}]}
<< {"jsonrpc":"2.0","id":2,"result":"0x726c"}
--

const floor = BigInt(23048);
const plain = BigInt(messages[1].response.result);
const auth = BigInt(messages[3].response.result);
if (plain < floor) throw new Error("estimate " + plain + " is below the calldata floor " + floor);
if (plain > floor * 11n / 10n) throw new Error("estimate " + plain + " is more than 10% above the calldata floor " + floor);
if (auth <= plain) throw new Error("estimate with an authorization " + auth + " is not above " + plain);
