# rpctestgen

`rpctestgen` is a test fixture generator for execution client's JSON-RPC API. 

Conceptually, it is similar to [`retesteth`][retesteth], which generates
consensus tests, in that it takes test definitions (in rpctestgen's case, go
functions), executes them against a client, and outputs the resulting
interaction.

The full API specification can be found in [`ethereum/exeuction-apis`][exeuciton-apis]

## Usage

rpctestgen runs with sane defaults. The tests will be filled with whatever
binary `geth` matches in the `PATH`. See `rpctestgen --help` for a full list of
options.

```console
$ rpctestgen --help
Usage: rpctestgen [--client CLIENT] [--bin BIN] [--out OUT] [--verbose] [--log-level LOG-LEVEL]

Options:
  --client CLIENT        client type [default: geth]
  --bin BIN              path to client binary [default: geth]
  --out OUT              directory where test fixtures will be written [default: tests]
  --verbose, -v          verbosity level of rpctestgen
  --log-level LOG-LEVEL
                         log level of client [default: info]
  --help, -h             display this help and exit
  ```

## Fixture format

The fixtures are very simple. Each statement is delimited by a newline. The `>>` prefix denotes
a request sent to the client. The `<<` prefix denotes the client's response.

```js
>> {"jsonrpc":"2.0","id":1,"method":"eth_blockNumber"}
<< {"jsonrpc":"2.0","id":1,"result":"0x3"}
```

[retesteth]: https://github.com/ethereum/retesteth
[execution-apis]: https:github.com/ethereum/execution-apis
