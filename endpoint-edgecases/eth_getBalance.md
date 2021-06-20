# `eth_getBalance`
The key words “MUST”, “MUST NOT”, “REQUIRED”, “SHALL”, “SHALL NOT”, “SHOULD”, “SHOULD NOT”, “RECOMMENDED”, “MAY”, and “OPTIONAL” in this document are to be interpreted as described in [RFC-2119](https://www.ietf.org/rfc/rfc2119.txt).
Specification | Description 
---|---
1 |eth_getBalance MUST return the balance of the account or contract specified in the address parameter  in wei as type quantity|
1.1 |eth_getBalance MUST return 0x0 if the account or contract does not exist|
2 |eth_getBalance's returned balance MUST be the account balance at the block specified in the block parameter|
3 |eth_getBalance's returned balance MUST not be affected by its size|

Error Type | Error Code
---|---
Invalid parameters |eth_getBalance MUST return error code -32602|
Invalid Input |eth_getBalance MUST return  error code -32000|

# Copyright
Copyright and related rights waived via [CC0](https://creativecommons.org/publicdomain/zero/1.0/).
