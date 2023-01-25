# Engine API -- Capabilities

Specification of `engine_exchangeCapabilities` method exchanging with a list of Engine API methods supported by the server (execution layer client) and the client (consensus layer client) down to a version of each method.

The proposed method should become a part of [`common.md`](../common.md) document if accepted.

## Methods

### engine_exchangeCapabilities

*Note:* The method itself doesn't have a version suffix.

#### Request

* method: `engine_exchangeCapabilities`
* params:
    1. `Array of string` -- Array of strings, each string is a name of a method supported by consensus layer client software.
* timeout: 1s

#### Response

`Array of string` -- Array of strings, each string is a name of a method supported by execution layer client software.

#### Specification

1. Consensus and execution layer client software **MAY** exchange with a list of currently supported Engine API methods. Execution layer client software **MUST NOT** log any error messages if this method has either never been called or haven't been called for a significant amount of time.

2. Request and response lists **MUST** contain Engine API methods that are currently supported by consensus and execution client software respectively. Name of each method in both lists **MUST** include suffixed version. Consider the following examples:
    * Client software of both layers currently supports `V1` and `V2` versions of `engine_newPayload` method:
        * params: `["engine_newPayloadV1", "engine_newPayloadV2", ...]`,
        * response: `["engine_newPayloadV1", "engine_newPayloadV2", ...]`.
    * `V1` method has been deprecated and `V3` method has been introduced on execution layer side since the last call:
        * params: `["engine_newPayloadV1", "engine_newPayloadV2", ...]`,
        * response: `["engine_newPayloadV2", "engine_newPayloadV3", ...]`.
    * The same capabilities modification has happened in consensus layer client, so, both clients have the same capability set again:
        * params: `["engine_newPayloadV2", "engine_newPayloadV3", ...]`,
        * response: `["engine_newPayloadV2", "engine_newPayloadV3", ...]`.

3. The `engine_exchangeCapabilities` method **MUST NOT** be returned in the response list.
