* **MUST NOT** affect on-chain state 

* **MUST** accept a `from` account that does not exist on-chain and if `from` is not defined it **MUST** be set to `0x0000000000000000000000000000000000000000`
 
* If the `to` is null or not defined on-chain it **MUST** return an empty hex string

* If `max_fee_per_gas` or `max_priority_fee_per_gas` is set the other must be set manually, otherwise they both **MUST** be set to the `gasPrice` or 0 when `gasPrice` is null

* **MUST** consider gas to equal 0 if the `gas` parameter is equal to `null` 

* If any non-zero fee or `value` is provided the `from` account balance **MUST** be checked to ensure account has enough funds
