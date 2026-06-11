# GitHub Actions workflows

This directory drives three kinds of automation: the **release pipeline** (publish → build → GitHub Release → stamped spec branch → dispatch deploy pages to refresh), **continuous deploy from `main`**, and **PR gating** (spec tests, docs smoke-build, spellcheck).

## Workflows

- [release.yaml](release.yaml) — `release: published` or `workflow_dispatch`. Full release: build, version snapshot, GitHub Release assets, stamped spec branch; dispatches [deploy.yaml](deploy.yaml) to refresh Pages (does not deploy Pages itself).

- [sync-release-notes.yaml](sync-release-notes.yaml) — `release: published|edited` or `workflow_dispatch`. Mirrors release notes into `docs-releases/` via PR.
- [publish-spec.yaml](publish-spec.yaml) — `workflow_dispatch` only. Manual recovery: re-push `assembled-spec` from existing release assets. The automatic stamped-spec publish is the `publish-spec` job inside [release.yaml](release.yaml) (`needs: github-release`).
- [deploy.yaml](deploy.yaml) — `push: main` or dispatch (`/next` always tracks main; spec `info.version` stays unstamped `0.0.0`) + pushes `assembled-spec-main` (unstamped).
- [test.yaml](test.yaml) — push/PR. `make build`, speccheck, test filling, lint.
- [test-deploy.yaml](test-deploy.yaml) — PR to `main`. Smoke-builds the site with and without a synthesized version snapshot.
- [spellcheck.yaml](spellcheck.yaml) — push/PR. `rojopolis/spellcheck-github-actions`.

## Release happy path

```mermaid
flowchart TD
  releasePublish["release: published"] --> releaseWf
  dispatch["workflow_dispatch (tag input)"] --> releaseWf

  subgraph releaseWf [release.yaml]
    buildRelease[build-release]
    ghRelease[github-release]
    publishSpec[publish-spec]
    triggerDeploy[trigger-deploy]
    buildRelease --> ghRelease
    ghRelease --> publishSpec
    ghRelease --> triggerDeploy
  end

  buildRelease -->|"docs-snapshot, openrpc.json, refs-openrpc.json"| artifacts[("Actions artifacts")]
  ghRelease -->|"upload assets (additive)"| ghReleasePage[("GitHub Release")]
  publishSpec -->|"force push"| assembledSpec[("assembled-spec branch")]
  triggerDeploy -->|"gh workflow run deploy.yaml"| deployWf[deploy.yaml]
  deployWf --> pages[("GitHub Pages")]

  ghReleasePage -->|"release: published"| syncNotes[sync-release-notes.yaml]
  syncNotes -->|"PR"| notesPR[("docs-releases/ PR")]
```

The `github-release` job attaches assets **additively** to the already-published release. If `gh release view <tag>` finds the release, it runs `gh release upload <tag> --clobber` for `openrpc.json`, `refs-openrpc.json`, and `docs-snapshot-<tag>.tar.gz` (re-runs replace those same assets). If no release exists for the tag, the job logs a notice and skips the upload. Because the release is already published before this job runs, [sync-release-notes.yaml](sync-release-notes.yaml) fires off the same `release: published` event in parallel; it is not gated on assets being present.

## Branch and version lifecycle

Versioned API docs on Pages are assembled from past release snapshots plus the current build. [scripts/assemble-versions.js](../../scripts/assemble-versions.js) downloads `docs-snapshot-*.tar.gz` from published GitHub Releases (up to 10 versions) and writes `api_versions.json` plus `api_versioned_docs/version-X.Y.Z/`.

```mermaid
flowchart LR
  mainPush["push to main"] --> deployWf[deploy.yaml]
  deployWf --> pagesMain[("Pages: /next from main + latest landing")]
  deployWf -->|"force push"| assembledMain[("assembled-spec-main (unstamped)")]

  releasePublish2["release: published"] --> releaseWf2[release.yaml]
  releaseWf2 --> assembledSpec2[("assembled-spec (stamped)")]
  releaseWf2 --> ghRel[("GitHub Release + docs-snapshot-vX.Y.Z.tar.gz")]
  ghRel -.->|"gh release download"| assembleScript["assemble-versions.js (MAX_VERSIONS=10)"]
  assembleScript -->|"writes"| apiVersions[("api_versions.json + api_versioned_docs/version-X.Y.Z/")]
  apiVersions -.-> deployWf
  releaseWf2 -->|"trigger-deploy"| deployWf
```

## Maintainer runbook

### Cut a release

Cut a release `vX.Y.Z` (UI **Draft a new release**, or `gh release create vX.Y.Z`). Publishing the release triggers [release.yaml](release.yaml) automatically; no manual steps are required for the GitHub Release, `assembled-spec`, or Pages (the `trigger-deploy` job dispatches [deploy.yaml](deploy.yaml) after assets are attached).

### Automated Release Notes PR

NOTE: The release also triggers a release notes PR. [release.yaml](release.yaml) dispatches [deploy.yaml](deploy.yaml) after attaching the snapshot asset. The subsequent automated PR that follows, forces the version and the release notes to invalidate the github pages CDN, when merged. This will guarantee that the latest release invalidates the github pages cache and deploys.

### Re-run a release

```bash
gh workflow run release.yaml -f tag=vX.Y.Z
```

The workflow is idempotent: it rebuilds and re-uploads (`--clobber`) the same assets onto the existing release. During `assemble-versions.js`, the pending tag is not downloaded from GitHub—the local snapshot from `docusaurus docs:version:api` is used instead (see [assemble-versions.js lines 110–121](../../scripts/assemble-versions.js)).

### Recover just `assembled-spec`

Use [publish-spec.yaml](publish-spec.yaml) with the tag input. It downloads `openrpc.json` and `refs-openrpc.json` from the existing GitHub Release and force-pushes `assembled-spec`. If those assets are missing, re-run [release.yaml](release.yaml) instead—do not use this recovery workflow.

### Recover release notes PR

```bash
gh workflow run sync-release-notes.yaml -f tag=vX.Y.Z
```

Reopens or updates the `release-notes/<slug>` PR that mirrors the GitHub Release into `docs-releases/`.

### Version dropdown missing a release

Confirm `docs-snapshot-vX.Y.Z.tar.gz` exists on the GitHub Release for that tag. [assemble-versions.js](../../scripts/assemble-versions.js) silently skips releases without that asset.

## Contracts between jobs

- `build-release` uploads two artifacts: `openrpc-spec` (`openrpc.json` + `refs-openrpc.json`) and `docs-snapshot` (the `.tar.gz`). Both `github-release` and `publish-spec` download `openrpc-spec`; `github-release` also downloads `docs-snapshot`.
- [deploy.yaml](deploy.yaml) After `github-release` attaches assets, `trigger-deploy` runs `gh workflow run deploy.yaml --ref main`. `/next` is always built from main HEAD (unstamped `0.0.0`);
- `concurrency.group: "pages"` applies only to [deploy.yaml](deploy.yaml) (`cancel-in-progress: false`).
- `assembled-spec` is **stamped** (version baked in via `npm run spec:set-version`); `assembled-spec-main` is **unstamped** (rolling head of `main`). Consumers pin to one or the other deliberately.
- [sync-release-notes.yaml](sync-release-notes.yaml) keys its PR branch off `steps.sync.outputs.slug`—repeated edits to the same release update the same PR rather than spawning new ones.
- `release: published` (and `edited`) fires the sync-release-notes workflow in parallel with `release.yaml`; `github-release` attaches assets additively to the already-published release and does not draft/publish it.
