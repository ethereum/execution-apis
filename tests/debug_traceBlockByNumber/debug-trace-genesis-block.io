// trace genesis block with callTracer
>> {"jsonrpc":"2.0","method":"debug_traceBlockByNumber","params":["0x0",{"tracer":"callTracer","tracerConfig":{}}],"id":1}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"genesis is not traceable"}}
