# rpctestgen

`rpctestgen` is a test fixture generator for the execution layer JSON-RPC API. 

Conceptually, it is similar to [`retesteth`][retesteth], which generates
consensus tests, in that it takes test definitions (in rpctestgen's case, go
functions), executes them against a client, and outputs the exchange.

The full API specification can be found in
[`ethereum/exeuction-apis`][execution-apis].

## Usage

rpctestgen runs with sane defaults. The tests will be filled with whatever
binary `geth` matches in the `$PATH`. By default, tests are not sealed. To seal
with ethash, run with the `--ethash` flag. For a full list of options, see
`rpctestgen --help`.

### Quick Start

To fill all tests with ethash seal, simply run `make fill`.

```console
$ make fill
go build .
./rpctestgen  --ethash
sealing block 1
sealing block 2
sealing block 3
starting client
filling tests...
generating tests/eth_blockNumber/simple-test.io  done.
generating tests/eth_getBlockByNumber/get-genesis.io  done.
generating tests/eth_getBlockByNumber/get-block-n.io  done.
```

This will write the generated test fixtures to `tests/` directory. In addition
to JSON-RPC exchange, a `chain.rlp` and `genesis.json` will be included so that
the exchange can be verified on all clients.

## Fixture format

The fixtures are very simple. Each statement is delimited by a newline. The
`>>` prefix denotes a request sent to the client. The `<<` prefix denotes the
client's response.

```js
>> {"jsonrpc":"2.0","id":1,"method":"eth_blockNumber"}
<< {"jsonrpc":"2.0","id":1,"result":"0x3"}
```

[retesteth]: https://github.com/ethereum/retesteth
[execution-apis]: https://github.com/ethereum/execution-apis
