import { readFileSync, writeFileSync } from "node:fs";
import { argv, exit } from "node:process";

const flags = {};
const positional = [];
for (let i = 2; i < argv.length; i++) {
  const a = argv[i];
  if (a.startsWith("--")) {
    flags[a.slice(2)] = argv[++i];
  } else {
    positional.push(a);
  }
}

if (!flags.version) {
  console.error("missing --version");
  exit(1);
}

const version = flags.version.replace(/^v/, "");
if (!/^\d+\.\d+\.\d+(-[A-Za-z0-9.-]+)?(\+[A-Za-z0-9.-]+)?$/.test(version)) {
  console.error(`invalid semver: ${version}`);
  exit(1);
}

const files = positional.length
  ? positional
  : ["openrpc.json", "refs-openrpc.json"];

for (const f of files) {
  const doc = JSON.parse(readFileSync(f, "utf8"));
  if (!doc.info || typeof doc.info !== "object") {
    console.error(`${f}: missing info object`);
    exit(1);
  }
  const prev = doc.info.version;
  doc.info.version = version;
  writeFileSync(f, JSON.stringify(doc, null, 2) + "\n");
  console.log(`set ${f}: info.version ${prev} -> ${version}`);
}
