import {
  readFileSync,
  writeFileSync,
  appendFileSync,
  mkdirSync,
  existsSync,
} from "node:fs";
import { join } from "node:path";
import { argv, env, exit } from "node:process";

const flags = {};
for (let i = 2; i < argv.length; i++) {
  const a = argv[i];
  if (a.startsWith("--")) flags[a.slice(2)] = argv[++i];
}

const outDir = flags["out-dir"] || env.OUT_DIR || "docs-releases";

function slugify(s) {
  return String(s).replace(/[^A-Za-z0-9.+-]/g, "-");
}

// this puts the newest releases first
function sidebarPos(t) {
  const m = String(t).match(/^v?(\d+)\.(\d+)\.(\d+)/);
  if (!m) return 0;
  const [, maj, min, pat] = m.map(Number);
  return -(maj * 1_000_000 + min * 1_000 + pat);
}

function escapeFrontmatter(s) {
  return String(s).replace(/"/g, '\\"');
}

async function fetchReleaseByTag(repo, tag, token) {
  if (!repo) throw new Error("missing GITHUB_REPOSITORY");
  const headers = {
    accept: "application/vnd.github+json",
    "x-github-api-version": "2022-11-28",
    "user-agent": "execution-apis-release-sync",
  };
  if (token) headers.authorization = `Bearer ${token}`;
  const url = `https://api.github.com/repos/${repo}/releases/tags/${encodeURIComponent(tag)}`;
  const res = await fetch(url, { headers });
  if (!res.ok) {
    throw new Error(`${url} -> ${res.status} ${res.statusText}`);
  }
  return res.json();
}

function normalize(r) {
  return {
    tag: r.tag_name,
    title: (r.name && r.name.trim()) || r.tag_name,
    htmlUrl: r.html_url || "",
    publishedAt: r.published_at || r.created_at || "",
    body: (r.body || "").trim(),
  };
}

// This gets the release note from the GitHub
async function resolveRelease() {
  if (env.GITHUB_EVENT_NAME === "release" && env.GITHUB_EVENT_PATH) {
    const evt = JSON.parse(readFileSync(env.GITHUB_EVENT_PATH, "utf8"));
    if (!evt.release) throw new Error("event payload has no .release");
    return normalize(evt.release);
  }

  if (flags["event-file"]) {
    const evt = JSON.parse(readFileSync(flags["event-file"], "utf8"));
    const r = evt.release || evt;
    return normalize(r);
  }

  const tag = flags.tag || env.INPUT_TAG;
  if (!tag) {
    throw new Error(
      "no release source: set GITHUB_EVENT_NAME=release+GITHUB_EVENT_PATH, " +
        "or pass --event-file, or pass --tag (with optional --body-file for offline mode, " +
        "else GITHUB_REPOSITORY+GITHUB_TOKEN to fetch from the API)",
    );
  }

  if (flags["body-file"]) {
    return {
      tag,
      title: flags.title || tag,
      htmlUrl: flags["html-url"] || "",
      publishedAt: flags["published-at"] || "",
      body: readFileSync(flags["body-file"], "utf8").trim(),
    };
  }

  const repo = flags.repo || env.GITHUB_REPOSITORY;
  const token = env.GITHUB_TOKEN || env.GH_TOKEN;
  return normalize(await fetchReleaseByTag(repo, tag, token));
}

// this builds the markdown for the release note
function buildMarkdown(r) {
  const fm = [
    "---",
    `title: "${escapeFrontmatter(r.title)}"`,
    `sidebar_label: "${escapeFrontmatter(r.tag)}"`,
    `sidebar_position: ${sidebarPos(r.tag)}`,
    `slug: /${slugify(r.tag)}`,
    r.publishedAt ? `date: ${r.publishedAt}` : null,
    "---",
  ]
    .filter(Boolean)
    .join("\n");
  const dateStr = r.publishedAt ? r.publishedAt.slice(0, 10) : "unreleased";
  const link = r.htmlUrl
    ? `_Released ${dateStr}._ · [View on GitHub](${r.htmlUrl})\n\n`
    : `_Released ${dateStr}._\n\n`;
  const body = r.body || "_No release notes._";
  return `${fm}\n\n# ${r.title}\n\n${link}${body}\n`;
}

// this writes outputs to be used for creating the PR for the release note into
// docs
function emitOutputs(r) {
  const slug = slugify(r.tag);
  const lines = [
    `tag=${r.tag}`,
    `slug=${slug}`,
    `title=${r.title}`,
    `html_url=${r.htmlUrl}`,
    `published_at=${r.publishedAt}`,
  ];
  if (env.GITHUB_OUTPUT) {
    appendFileSync(env.GITHUB_OUTPUT, lines.join("\n") + "\n");
  }
  for (const l of lines) console.log(l);
}

try {
  const release = await resolveRelease();
  if (!existsSync(outDir)) mkdirSync(outDir, { recursive: true });
  const file = join(outDir, `${slugify(release.tag)}.md`);
  writeFileSync(file, buildMarkdown(release));
  console.log(`wrote ${file}`);
  emitOutputs(release);
} catch (err) {
  console.error(`sync-release-note: ${err.message}`);
  exit(1);
}
