# Engine API -- Transition Configuration

## Table of contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Structures](#structures)
  - [TransitionConfigurationV1](#transitionconfigurationv1)
- [Methods](#methods)
  - [engine_exchangeTransitionConfigurationV1](#engine_exchangetransitionconfigurationv1)
    - [Request](#request)
    - [Response](#response)
    - [Specification](#specification)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Structures

### TransitionConfigurationV1

This structure contains configurable settings of the transition process. The fields are encoded as follows:
- `terminalTotalDifficulty`: `QUANTITY`, 256 Bits - maps on the `TERMINAL_TOTAL_DIFFICULTY` parameter of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#client-software-configuration)
- `terminalBlockHash`: `DATA`, 32 Bytes - maps on `TERMINAL_BLOCK_HASH` parameter of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#client-software-configuration)
- `terminalBlockNumber`: `QUANTITY`, 64 Bits - maps on `TERMINAL_BLOCK_NUMBER` parameter of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#client-software-configuration)

## Methods

### engine_exchangeTransitionConfigurationV1

* status: **`Final`**

#### Request

* method: `engine_exchangeTransitionConfigurationV1`
* params:
  1. `transitionConfiguration`: `Object` - instance of [`TransitionConfigurationV1`](#TransitionConfigurationV1)
* timeout: 1s

#### Response

* result: [`TransitionConfigurationV1`](#TransitionConfigurationV1)
* error: code and message set in case an exception happens while getting a transition configuration.

#### Specification

1. Execution Layer client software **MUST** respond with configurable setting values that are set according to the Client software configuration section of [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675#client-software-configuration).

2. Execution Layer client software **SHOULD** surface an error to the user if local configuration settings mismatch corresponding values received in the call of this method, with exception for `terminalBlockNumber` value.

3. Consensus Layer client software **SHOULD** surface an error to the user if local configuration settings mismatch corresponding values obtained from the response to the call of this method.

4. Consensus Layer client software **SHOULD** poll this endpoint every 60 seconds.

5. Execution Layer client software **SHOULD** surface an error to the user if it does not recieve a request on this endpoint at least once every 120 seconds.

6. Considering the absence of the `TERMINAL_BLOCK_NUMBER` setting, Consensus Layer client software **MAY** use `0` value for the `terminalBlockNumber` field in the input parameters of this call.

7. Considering the absence of the `TERMINAL_TOTAL_DIFFICULTY` value (i.e. when a value has not been decided), Consensus Layer and Execution Layer client software **MUST** use `115792089237316195423570985008687907853269984665640564039457584007913129638912` value (equal to`2**256-2**10`) for the `terminalTotalDifficulty` input parameter of this call.
