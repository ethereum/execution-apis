const RefParser = require("@apidevtools/json-schema-ref-parser");
const fs = require("fs");

const _parse = async (_schema) => {
  try {
    let schema = await RefParser.dereference(_schema);
    if (!fs.existsSync("./build")) {
      fs.mkdirSync("./build");
    }
    fs.writeFileSync("./build/openrpc.json", JSON.stringify(schema));
  } catch (err) {
    console.error(err);
  }
};

_parse(require("./openrpc.json"));
