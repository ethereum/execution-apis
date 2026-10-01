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

## Format

Tests are written to describe the round-trip of a single request-response
cycle. A test starts with `>>` followed by a space, denoting the request portion.
It is delimited by `\n` and then `<<` followed by a space denotes the response.
All together, it looks something like this:

```javascript
>> {"jsonrpc":"2.0","id":1,"method":"eth_blockNumber"}
<< {"jsonrpc":"2.0","id":1,"result":"0x3"}
```

For organizational purposes, tests are stored at a path following the template
`tests/{method-name}/{test-name}.io`. The path does not affect the validity of
the test and is only used to describe what the test is aiming to test.

## Generation

### Engine API fixtures

`rpctestgen` sends `engine_*` methods to the authenticated Engine API endpoint.
It records only JSON-RPC request and response bodies; authentication headers
are not part of the fixture. Hive rpc-compat needs Engine API routing support
to replay these fixtures on the authenticated endpoint.

For `engine_forkchoiceUpdatedV*`, payload IDs are opaque and client-specific.
A recorded non-null `payloadId` requires a non-null 8-byte DATA value on replay,
while the other response fields are compared exactly. These tests do not use
`speconly`: a `SYNCING` response or a null payload ID must not satisfy a fixture
that expects `VALID` and a started payload build.

The `engine_forkchoiceUpdatedV4/target-gas-limit-*` fixtures require the Amsterdam
chain from [#867](https://github.com/ethereum/execution-apis/pull/867). They cover
zero, one, the signed 64-bit maximum, and the upper half of the unsigned 64-bit
range. They keep the known head unchanged and check successful payload-building
initiation, not equality between the target and the next block's gas limit.
The generator reports an explicit error if run against a pre-Amsterdam head.

### Generating fixtures

Test generation can be broken down into two parts. First is the generation of a
chain against which tests will be executed. Second is executing the actual
tests and recording their round-trip.

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
