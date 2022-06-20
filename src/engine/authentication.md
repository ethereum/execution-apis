# Authentication

The `engine` JSON-RPC interface, exposed by EL and consumed by CL, needs to be authenticated. The authentication scheme chosen for this purpose is [JWT](https://jwt.io/).

The type of attacks that this authentication scheme attempts to protect against are the following:

- RPC port exposed towards the internet, allowing attackers to exchange messages with EL engine API.
- RPC port exposed towards the browser, allowing malicious webpages to submit messages to the EL engine API.

The authentication scheme is _not_ designed to

- Prevent attackers with capability to read ('sniff') network traffic from reading the traffic,
- Prevent attackers with capability to read ('sniff') network traffic from performing replay-attacks of earlier messages.

Authentication is performed as follows:

- For `HTTP` dialogue, each `jsonrpc` request is individually authenticated by supplying `JWT` token in the HTTP header.
- For a WebSocket dialogue, only the initial handshake is authenticated, after which the message dialogue proceeds without further use of JWT.
  - Clarification: The websocket handshake starts with the client performing a websocket upgrade request. This is a regular http GET request, and the actual
parameters for the WS-handshake are carried in the http headers.
- For `inproc`, a.k.a raw ipc communication, no authentication is required, under the assumption that a process able to access `ipc` channels for the process, which usually means local file access, is already sufficiently permissioned that further authentication requirements do not add security.


## JWT specifications

- Client software MUST expose the authenticated Engine API at a port independent from existing JSON-RPC API.
  - The default port for the authenticated Engine API is `8551`. The Engine API is exposed under the `engine` namespace.
- The EL **MUST** support at least the following `alg` `HMAC + SHA256` (`HS256`)
- The EL **MUST** reject the `alg` `none`.


The HMAC algorithm implies that several CLs will be able to use the same key, and from an authentication perspective, be able to impersonate each other. From a deployment perspective, it means that an EL does not need to be provisioned with individual keys for each CL.

## Key distribution

The `EL` and `CL` clients **MUST** accept a cli/config parameter: `jwt-secret`, which designates a file containing the hex-encoded 256 bit secret key to be used for verifying/generating JWT tokens.

If such a parameter is not given, the client **SHOULD** generate such a token, valid for the duration of the execution, and store the hex-encoded secret as a `jwt.hex` file on the filesystem.  This file can then be used to provision the counterpart client.

If such a parameter _is_ given, but the file cannot be read, or does not contain a hex-encoded key of `256` bits, the client should treat this as an error: either abort the startup, or show error and continue without exposing the authenticated port.

## JWT Claims

This specification utilizes the following list of JWT claims:

- Required: `iat` (issued-at) claim. The EL **SHOULD** only accept `iat` timestamps which are within +-5 seconds from the current time.
- Optional: `id` claim. The CL **MAY** use this to communicate a unique identifier for the individual CL node.
- Optional: `clv` claim. The CL **MAY** use this to communicate the CL node type/version.

Other claims **MAY** be included in the JWT payload. If the EL sees claims it does not recognize, these **MUST** be ignored.

## Examples

Todo, add some examples of JWT authentication here.
