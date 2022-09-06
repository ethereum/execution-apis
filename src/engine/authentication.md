# Authentication

The `engine` JSON-RPC interface, exposed by execution layer clients and consumed by consensus layer clients, needs to be authenticated. The authentication scheme chosen for this purpose is [JWT](https://jwt.io/).

The type of attacks that this authentication scheme attempts to protect against are the following:

- RPC port exposed towards the internet, allowing attackers to exchange messages with execution layer engine API.
- RPC port exposed towards the browser, allowing malicious webpages to submit messages to the execution layer engine API.

The authentication scheme is _not_ designed to

- Prevent attackers with capability to read ('sniff') network traffic from reading the traffic,
- Prevent attackers with capability to read ('sniff') network traffic from performing replay-attacks of earlier messages.

Authentication is performed as follows:

- For `HTTP` dialogue, each `jsonrpc` request is individually authenticated by supplying `JWT` token in the HTTP header.
- For a WebSocket dialogue, only the initial handshake is authenticated, after which the message dialogue proceeds without further use of JWT.
  - Clarification: The websocket handshake starts with the consensus layer client performing a websocket upgrade request. This is a regular http GET request, and the actual
parameters for the WS-handshake are carried in the http headers.
- For `inproc`, a.k.a raw ipc communication, no authentication is required, under the assumption that a process able to access `ipc` channels for the process, which usually means local file access, is already sufficiently permissioned that further authentication requirements do not add security.
- If the listening address is `127.0.0.1`, a fixed `JWT` secret must be used, under the assumption that sockets bound to the localhost cannot accept external connections and therefore are not susceptible to manipulation from remote adversaries. The fixed secret serves to protect against malicious webpages that may attempt to make requests against the localhost.

## JWT specifications

- The execution layer client **MUST** expose the authenticated Engine API at a port independent from existing JSON-RPC API.
  - The default port for the authenticated Engine API is `8551`. The Engine API is exposed under the `engine` namespace.
- The execution layer client **MUST** support at least the following `alg` `HMAC + SHA256` (`HS256`)
- The execution layer client **MUST** reject the `alg` `none`.


The HMAC algorithm implies that several consensus layer clients will be able to use the same key, and from an authentication perspective, be able to impersonate each other. From a deployment perspective, it means that an EL does not need to be provisioned with individual keys for each consensus layer client.

## Key distribution

The execution layer and consensus layer clients **SHOULD** accept a configuration parameter: `jwt-secret`, which designates a file containing the hex-encoded 256 bit secret key to be used for verifying/generating JWT tokens.

If such a parameter is not given and the host is `127.0.0.1` the client **MUST** continue with the default `jwt-secret`:

```
0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3584bdb5d4e74fe97f5a5222b533fe1322fd0b6ad3eb03f02c3221984e2c0b4304985f5ca3d2afbec36529aa96f74de3cc10a2a4a6c44f2157a57d2c6059a11bb2086799aeebeae135c246c65021c82b4e15a2c451340993aacfd2751886514f058eff9265aedf8a54da8121de1324e1e0d9aac99f694d16c6a41afffe3817d73b1fcff633029ee18ab6482b58ff8b6e95dd7c82a954c852157152a7a6d32785eeddb0590e1095fbe51205a51a297daef7259e229af0432214ae6cb2c1f750750eddb0590e1095fbe51205a51a297daef7259e229af0432214ae6cb2c1f750750
```

If such a parameter is not given and the host _is not_ `127.0.0.1`, the client **SHOULD** generate such a token, valid for the duration of the execution, and store the hex-encoded secret as a `jwt.hex` file on the filesystem.  This file can then be used to provision the counterpart client.

If such a parameter _is_ given, but the file cannot be read, or does not contain a hex-encoded key of `256` bits, the client **SHOULD** treat this as an error: either abort the startup, or show error and continue without exposing the authenticated port.

## JWT Claims

This specification utilizes the following list of JWT claims:

- Required: `iat` (issued-at) claim. The execution layer client **SHOULD** only accept `iat` timestamps which are within +-60 seconds from the current time.
- Optional: `id` claim. The consensus layer client **MAY** use this to communicate a unique identifier for the individual consensus layer client.
- Optional: `clv` claim. The consensus layer client **MAY** use this to communicate the consensus layer client type/version.

Other claims **MAY** be included in the JWT payload. If the execution layer client sees claims it does not recognize, these **MUST** be ignored.

## Examples

Todo, add some examples of JWT authentication here.
