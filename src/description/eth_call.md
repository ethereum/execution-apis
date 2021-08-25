* **MUST NOT** affect on-chain state 

* **MUST** allow a `from` parameter that does not exist on-chain and if `from` is not defined it **MUST** be interpreted as `0x0000000000000000000000000000000000000000`

 
* If the `to` is null or not defined on-chain and there is no `data` parameter it **MUST** return an empty hex string

* If the `gasPrice` parameter is used it **MUST** interpret it as the `maxFeePerGas` and `maxPriorityFeePerGas` both equal to the value of the `gasPrice` parameter or 0 when `gasPrice` is null

* **MUST** consider gas to equal 0 if the `gas` parameter is equal to `null` 

* If any non-zero fee or `value` is provided the `from` account balance **MUST** be checked to ensure account has enough funds, in the case that the account has insufficient funds it **MUST** error.
