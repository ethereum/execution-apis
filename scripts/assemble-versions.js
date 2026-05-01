import { execFileSync } from "node:child_process";
import { mkdtempSync, writeFileSync, readdirSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { exit } from "node:process";

const MAX_VERSIONS = 10;

function sh(cmd, args, opts = {}) {
  return execFileSync(cmd, args, {
    encoding: "utf8",
    stdio: ["ignore", "pipe", "inherit"],
    ...opts,
  });
}

function semverDesc(a, b) {
  const pa = String(a)
    .replace(/^v/, "")
    .split(".")
    .map((n) => parseInt(n, 10) || 0);
  const pb = String(b)
    .replace(/^v/, "")
    .split(".")
    .map((n) => parseInt(n, 10) || 0);
  for (let i = 0; i < 3; i++) {
    if ((pb[i] || 0) !== (pa[i] || 0)) return (pb[i] || 0) - (pa[i] || 0);
  }
  return 0;
}

function listReleases() {
  try {
    const raw = sh("gh", [
      "release",
      "list",
      "--json",
      "tagName,isDraft,isPrerelease,publishedAt",
      "-L",
      "100",
    ]);
    return JSON.parse(raw);
  } catch (err) {
    console.warn(`assemble-versions: gh release list failed: ${err.message}`);
    return [];
  }
}

function downloadAndExtract(tag) {
  const dir = mkdtempSync(join(tmpdir(), `snap-${tag.replace(/[^A-Za-z0-9.-]/g, "_")}-`));
  try {
    sh("gh", [
      "release",
      "download",
      tag,
      "-p",
      "docs-snapshot-*.tar.gz",
      "-D",
      dir,
    ]);
    const files = readdirSync(dir).filter((f) => f.endsWith(".tar.gz"));
    if (files.length === 0) {
      console.warn(`assemble-versions: ${tag} has no docs-snapshot asset; skipping`);
      return false;
    }
    sh("tar", ["-xzf", join(dir, files[0])], { stdio: "inherit" });
    return true;
  } catch (err) {
    console.warn(`assemble-versions: failed to fetch snapshot for ${tag}: ${err.message}`);
    return false;
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
}

const releases = listReleases()
  .filter((r) => !r.isDraft && !r.isPrerelease)
  .map((r) => r.tagName);

if (releases.length === 0) {
  console.log("assemble-versions: no published releases found; nothing to assemble");
  exit(0);
}

const tags = releases.sort(semverDesc).slice(0, MAX_VERSIONS);
const kept = [];
for (const tag of tags) {
  const ver = tag.replace(/^v/, "");
  if (downloadAndExtract(tag)) kept.push(ver);
}

if (kept.length === 0) {
  console.log("assemble-versions: no snapshots successfully extracted");
  exit(0);
}

writeFileSync("api_versions.json", JSON.stringify(kept, null, 2) + "\n");
console.log(`assemble-versions: wrote api_versions.json with ${kept.length} version(s):`);
for (const v of kept) console.log(`  - ${v}`);
