// requests the witness of the pending tag; an error is expected because pending does not identify an executed block
>> {"jsonrpc":"2.0","id":1,"method":"debug_executionWitness","params":["pending"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"pending block has no witness"}}
