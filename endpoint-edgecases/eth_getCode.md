# `eth_getCode`
The key words “MUST”, “MUST NOT”, “REQUIRED”, “SHALL”, “SHALL NOT”, “SHOULD”, “SHOULD NOT”, “RECOMMENDED”, “MAY”, and “OPTIONAL” in this document are to be interpreted as described in [RFC-2119](https://www.ietf.org/rfc/rfc2119.txt).
Spefication | Description
---|---
1 |eth_getCode MUST return code stored in given  address|
1.1 |eth_getCode MUST return 0x if no contract exists at given address|
1.2 |eth_getCode MUST return 0x when contract exists but has no bytecode deployed to it|
2 |eth_getCode Must return code if it exists at or before the block given|
2.1 |eth_getCode MUST throw an exception if given block does not exist|
# Copyright
Copyright and related rights waived via [CC0](https://creativecommons.org/publicdomain/zero/1.0/).
