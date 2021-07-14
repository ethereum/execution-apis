import RefParser from "@apidevtools/json-schema-ref-parser"
import fs from "fs"
import * as openrpc from "../openrpc.json"

const _parse = async (_schema) => {
  try {
    let schema = await RefParser.dereference(_schema);
    if (!fs.existsSync("./build")) {
      fs.mkdirSync("./build");
    }

    fs.writeFileSync("./build/openrpc.json", JSON.stringify(schema.default));
  } catch (err) {
    console.error(err);
  }
};

_parse(openrpc);
