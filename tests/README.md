# Tests

The Execution API has a comprehensive test suite to verify conformance of
clients. The tests in this repository are loaded into the [`hive`][hive] test
simulator [`rpc-compat`][rpc-compat] and validated against every major client.

The test suite is run daily and results are always available [here][hivetests2]
under the tag `rpc-compat`. 

To learn more about the `rpc-compat` simulator, please see its
[documentation][rpc-compat].

## Format

Tests are written to describe the round-trip of a single request-response
cycle. A test starts with a `>> ` denoting the request portion. It is delimited
by `\n` and then a `<< ` is used to denote the response. All together, it looks
something like this:

```javascript
>> {"jsonrpc":"2.0","id":1,"method":"eth_blockNumber"}
<< {"jsonrpc":"2.0","id":1,"result":"0x3"}
```

For organizational purposes, tests are stored at a path following the template
`tests/{method-name}/{test-name}.io`. The path does not affect the validity of
the test and is only used to describe what the test is aiming to test.

## Generation

Test generation can be broken down into two parts. First is the generation of a
chain against which tests will be executed. Second is executing the actual
tests and recording their round-trip.

Although the `io` format is agnostic to the generation tool, it is preferred
test contributors use the generation tool [`rpctestgen`][rpctestgen].
`rpctestgen` takes care of both pieces of test generation.

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
simulator can be found with there.

[hive]: https://github.com/ethereum/hive
[hivetests2]: https://hivetests2.ethdevops.io
[rpc-compat]: https://github.com/ethereum/hive/tree/master/simulators/ethereum/rpc-compat
[rpctestgen]: https://github.com/lightclient/rpctestgen
