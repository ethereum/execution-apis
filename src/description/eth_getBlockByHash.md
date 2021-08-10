* If the client has a block associated with the hash in the `block_hash` parameter it **MUST** return it, otherwise it **MUST** return null.

* If the `full_transaction` parameter is true, **MUST** include the full transaction details for every transaction in the block, otherwise it **MUST** return only the transaction hash for every transaction in the block.