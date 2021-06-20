# `eth_coinbase`
The key words “MUST”, “MUST NOT”, “REQUIRED”, “SHALL”, “SHALL NOT”, “SHOULD”, “SHOULD NOT”, “RECOMMENDED”, “MAY”, and “OPTIONAL” in this document are to be interpreted as described in [RFC-2119](https://www.ietf.org/rfc/rfc2119.txt).
Specification | Description 
---|---
1 |eth_coinbase MUST return the benificary address of the client as type DATA|
1.1| eth_coinbase's benificary address MUST be set before  the client can start mining any blocks|
2 |eth_coinbase MUST throw an exception if the client has no benificary address |
# Copyright
Copyright and related rights waived via [CC0](https://creativecommons.org/publicdomain/zero/1.0/).