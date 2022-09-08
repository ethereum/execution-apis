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


## JWT specifications

- The execution layer client **MUST** expose the authenticated Engine API at a port independent from existing JSON-RPC API.
  - The default port for the authenticated Engine API is `8551`. The Engine API is exposed under the `engine` namespace.
- The execution layer client **MUST** support at least the following `alg` `HMAC + SHA256` (`HS256`)
- The execution layer client **MUST** reject the `alg` `none`.


The HMAC algorithm implies that several consensus layer clients will be able to use the same key, and from an authentication perspective, be able to impersonate each other. From a deployment perspective, it means that an EL does not need to be provisioned with individual keys for each consensus layer client.

## Key distribution

The execution layer and consensus layer clients **SHOULD** accept a configuration parameter: `jwt-secret`, which designates a file containing the hex-encoded 256 bit secret key to be used for verifying/generating JWT tokens.

If such a parameter is not given, the client **SHOULD** generate such a token, valid for the duration of the execution, and **SHOULD** store the hex-encoded secret as a `jwt.hex` file on the filesystem.  This file can then be used to provision the counterpart client.

If such a parameter _is_ given, but the file cannot be read, or does not contain a hex-encoded key of `256` bits, the client **SHOULD** treat this as an error: either abort the startup, or show error and continue without exposing the authenticated port.

## JWT Claims

This specification utilizes the following list of JWT claims:

- Required: `iat` (issued-at) claim. The execution layer client **SHOULD** only accept `iat` timestamps which are within +-60 seconds from the current time.
- Optional: `id` claim. The consensus layer client **MAY** use this to communicate a unique identifier for the individual consensus layer client.
- Optional: `clv` claim. The consensus layer client **MAY** use this to communicate the consensus layer client type/version.

Other claims **MAY** be included in the JWT payload. If the execution layer client sees claims it does not recognize, these **MUST** be ignored.

## Examples

Todo, add some examples of JWT authentication here.
