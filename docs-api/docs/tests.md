# Tests

The Execution API has a comprehensive test suite to verify conformance of
clients. The tests in this repository are loaded into the [`hive`][hive] test
simulator [`rpc-compat`][rpc-compat] and validated against every major client.

The test suite is run daily and results are always available at
[hive.ethpandaops.io][hivetests] under the tag `rpc-compat`.

To learn more about the `rpc-compat` simulator, please see its
[documentation][rpc-compat].

## How Hive Consumes These Tests

The rpc-compat simulator clones execution-apis during its Docker build and copies
the `tests/` directory into the simulator container. By default it fetches the
`main` branch; the `branch` build arg can target a specific ref (e.g., a
version tag once versioned releases are available). Test results are published
at [hive.ethpandaops.io][hivetests] under the `rpc-compat` tag.

## Fixture Format (.io files)

For organizational purposes, tests are stored at a path following the template
`tests/{method-name}/{test-name}.io`. The path does not affect the validity of the test
and is only used to describe what the test is aiming to test.

The test fixtures use a line-delimited format. 

- Lines starting with `>>` denote a message sent to the server
- `<<` starts a response receive, and declares the expected response.
- The test format also supports comments using a `//` line prefix.

```javascript
// this is a comment
>> {"jsonrpc":"2.0","id":1,"method":"eth_blockNumber"}
<< {"jsonrpc":"2.0","id":1,"result":"0x3"}
```

### speconly Tests

To declare that a test's responses are not to be checked against the server's responses
literally, a comment starting with `speconly:` is used. Example:

```javascript
// This test checks gas estimation.
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_estimateGas","params":{"data":"0x00"}}
<< {"jsonrpc":"2.0","id":1,"result":"0x200044"}
```

Tests are marked as `speconly` when the result is non-deterministic and depends on client
configuration. For `speconly` tests, the test harness will verify the response against the
JSON-Schema of the corresponding request method.

### Test Script

Test files can optionally contain a 'validation script' section at the end. Test scripts
are intended for `speconly` tests files, since literal matching of responses gives the
strongest possible testing guarantees.

The script section is introduced by a line containing `--` and nothing else. All remaining
text in the file is JavaScript code. Example:

```javascript
// estimates a successful contract call
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_estimateGas","params":[{"from":"0x0102030000000000000000000000000000000000","input":"0xff01","to":"0x17e7eedce4ac02ef114a7ed9fe6e2f33feba1667"}]}
<< {"jsonrpc":"2.0","id":1,"result":"0x5316"}
--
if (BigInt(messages[1].response.result) <= 5000) {
    throw new Exception("invalid gas estimation result");
}
```

The test harness executes the validation script after message exchanges with the server.
If the script throws an exception, the test is considered to have failed. Within the
script, the `messages` variable contains an array of message objects. Each element of
`messages` is an object where

  - `messages[i].send` is set to the RPC request for send (`>>`) lines.
  - `messages[i].expected` is the expected message for receive (`<<`) lines.
  - `messages[i].response` is the message received from the server

Note the `console` module is available for use in validation scripts. Output from the test
script is printed to the test log.

### Advanced JSON-Schema Validation in Test Scripts

One specific use case of test scripts is forcing schema validation of a specific branch of
an `anyOf` construction in `speconly` tests. For example, consider an RPC method that can
return JSON objects corresponding to different transaction types depending on the input.
Since the result of this method is declared using `anyOf` over all possible transaction
types, a `speconly` test would not be able to verify that a specific type was returned.

To make it possible to verify such responses, the test script can access the OpenRPC
specification and perform schema validation using a specific sub-schema element.

Within the execution environment of the test script, the pre-loaded `openrpc` module
allows such validation. The `validate` method indexes the OpenRPC spec document with the
JSON-Pointer given as the first argument and performs JSON-Schema validation of the value.

```javascript
openrpc.validate("#/components/schemas/AccessList", messages[1].response.value);
```

There is also a boolean test:

```javascript
if (openrpc.isValid("#/components/schemas/AccessList", messages[1].response.value)) {
    // value is valid
}
```

## Generation

Test generation can be broken down into two parts. First is the generation of a chain
against which tests will be executed. Second is executing the actual tests and recording
their round-trip.

Although the `io` format is agnostic to the generation tool, it is preferred
test contributors use [`rpctestgen`][rpctestgen], which takes care of both
chain generation and test execution. See the [Tools README][tools-readme] for
usage and CI requirements.

### Chain making

Inside the `tests` directory are three chain-related files that test authors
must be aware of.

`genesis.json` - a standard genesis config file in the go-ethereum format.
`chain.rlp`    - a newline-delimited list of blocks making up the test chain.
`bad.rlp`      - a newline-delimited list of blocks that are sealed and
                 conduct an invalid transition.

Generally, test authors should ingest `genesis.json` and `chain.rlp` and
generate tests against the head of that chain. If a test requires a certain
condition exist in the chain that does not currently exist, then the author may
append a block to head of the chain and regenerate all tests against the new
`chain.rlp`.

### Test Generation

Once a test chain has been created, test authors may move on to generating the
actual test fixtures. To do so, authors must follow the format defined above.
Tests should be limited to a single round-trip interaction. At this time, this
precludes subscription methods from being tested.

It is also recommended that test authors test their tests. Each interaction
should be validated against the expected values. Due to the number of fixtures
generated, it is easy accept incorrect responses.

A good final verification of tests is to run them in the hive simulator
[`rpc-compat`][rpc-compat]. More information on how to run custom tests in the
simulator can be found in the simulator documentation.

[hive]: https://github.com/ethereum/hive
[hivetests]: https://hive.ethpandaops.io
[rpc-compat]: https://github.com/ethereum/hive/tree/master/simulators/ethereum/rpc-compat
[rpctestgen]: https://github.com/ethereum/execution-apis/tree/main/tools
[tools-readme]: https://github.com/ethereum/execution-apis/blob/main/tools/README.md
