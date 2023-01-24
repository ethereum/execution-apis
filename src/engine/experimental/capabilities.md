# Engine API -- Capabilities

Specification of `engine_getCapabilities` method returning a list of Engine API methods supported by the server (execution layer client) down to a version of each method.

The proposed method should become a part of [`common.md`](../common.md) document if accepted.

## Methods

### engine_getCapabilities

*Note:* The method itself doesn't have a version suffix.

#### Request

* method: `engine_getCapabilities`
* timeout: 1s

#### Response

`Array of string` -- Array of strings, each string is a name of a method supported by execution layer client software.

#### Specification

1. Client software **MUST** return a list of currently supported Engine API methods down to a version of each method. Consider the following response examples: 
    * `["engine_newPayloadV1", "engine_newPayloadV2", ...]` -- the software currently supports `V1` and `V2` versions of `engine_newPayload` method,
    * `["engine_newPayloadV2", "engine_newPayloadV3", ...]` -- `V1` version has been deprecated, and `V3` have been introduced with respect to the above response.

2. The `engine_getCapabilities` method **MUST NOT** be returned in the response list.
