# GitHub Actions Setup & Workflow Guide

This guide covers the GitHub Actions infrastructure in the `execution-apis` repository, including workflow organization, metadata best practices, and common patterns.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Workflow Organization](#workflow-organization)
3. [Adding New Workflows](#adding-new-workflows)
4. [Using Custom Actions](#using-custom-actions)
5. [Troubleshooting](#troubleshooting)

---

## Quick Start

### View Workflow Status

1. Go to the **Actions** tab: https://github.com/moonrager13/execution-apis/actions
2. Select a workflow to see recent runs
3. Click a run to view detailed logs

### Manually Trigger a Workflow

```bash
# List available workflows
gh workflow list

# Run a workflow with parameters
gh workflow run release.yaml -f tag=vX.Y.Z

# Run deploy workflow
gh workflow run deploy.yaml --ref main
```

### View Workflow Logs

```bash
# List recent runs
gh run list --workflow deploy.yaml

# View logs for a specific run
gh run view <run-id> --log
```

---

## Workflow Organization

### Directory Structure

```
.github/
├── workflows/          # Workflow definitions
│   ├── README.md      # Workflow documentation
│   ├── deploy.yaml    # Deploy to GitHub Pages
│   ├── release.yaml   # Release pipeline
│   ├── test.yaml      # Run tests
│   └── ...
├── ACTIONS_GUIDE.md   # This file
└── (future: actions/) # For custom actions
```

### Workflow Categories

#### 1. **Release Pipeline** (`release.yaml`)
- **Trigger**: Release published or manual dispatch with tag
- **Steps**:
  1. Build spec and docs
  2. Create GitHub Release with assets
  3. Publish stamped spec to `assembled-spec` branch
  4. Trigger deployment
- **Artifacts**: `openrpc.json`, `refs-openrpc.json`, docs snapshot

#### 2. **Continuous Deployment** (`deploy.yaml`)
- **Trigger**: Push to main or manual dispatch
- **Steps**:
  1. Build spec using Go tools
  2. Build docs site using Node.js
  3. Deploy to GitHub Pages
  4. Push unstamped spec to `assembled-spec-main`
- **Concurrency**: Only one deployment at a time to Pages

#### 3. **Testing** (`test.yaml`, `test-deploy.yaml`)
- **test.yaml**: Runs on push/PR
  - Execute `make build`
  - Validate specs with `speccheck`
  - Check for broken links
  - Run spell check
- **test-deploy.yaml**: Runs on PR to main
  - Smoke-test site build with synthetic version

#### 4. **Release Notes** (`sync-release-notes.yaml`)
- **Trigger**: Release published or edited
- **Purpose**: Mirror GitHub Release notes to `docs-releases/` via PR

#### 5. **Spec Recovery** (`publish-spec.yaml`)
- **Trigger**: Manual dispatch with tag input
- **Purpose**: Re-publish spec from existing release assets

---

## Adding New Workflows

### Step 1: Create Workflow File

Create `.github/workflows/my-workflow.yaml`:

```yaml
name: My New Workflow

on:
  push:
    branches: [main]
  pull_request:
  workflow_dispatch:

permissions:
  contents: read

jobs:
  my-job:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up tools
        uses: actions/setup-go@v6
        with:
          go-version: ^1.26
      
      - name: Run task
        run: |
          make my-target
```

### Step 2: Test the Workflow

```bash
# Dry-run: check workflow syntax
gh workflow view my-workflow.yaml

# Run manually to test
gh workflow run my-workflow.yaml --ref main
```

### Step 3: Document in README.md

Add to `.github/workflows/README.md`:

```markdown
- [my-workflow.yaml](my-workflow.yaml) — `trigger: condition`. Brief description.
```

### Step 4: Update Mermaid Diagram

If your workflow is part of the release pipeline, update the flow diagram in `.github/workflows/README.md`.

---

## Using Custom Actions

### Option 1: Use a Published Action

```yaml
- uses: actions/checkout@v4
- uses: actions/setup-go@v6
  with:
    go-version: ^1.26
```

### Option 2: Create a Custom Action

If you need reusable logic across workflows:

1. **Create `action.yml`** in the action directory:

```yaml
name: My Custom Action
description: Does something useful

inputs:
  input-param:
    description: Input description
    required: true

outputs:
  output-result:
    description: Output description
    value: ${{ steps.my-step.outputs.result }}

runs:
  using: composite
  steps:
    - run: echo "Hello"
      shell: bash
```

2. **Reference in workflow**:

```yaml
- uses: ./path/to/action
  with:
    input-param: value
```

### Action Types

| Type | Use Case | File Extension |
|------|----------|----------------|
| **Composite** | Reusable shell/multi-step scripts | `.sh`, `.yaml` |
| **JavaScript** | Node.js scripts | `.js` |
| **Docker** | Containerized tools | `Dockerfile` |

For this repository, **composite actions** are recommended for spec generation and deployment steps.

---

## Environment Variables & Secrets

### GitHub-Provided Variables

```yaml
steps:
  - run: |
      echo "Repository: ${{ github.repository }}"
      echo "Branch: ${{ github.ref }}"
      echo "Commit: ${{ github.sha }}"
      echo "Actor: ${{ github.actor }}"
```

### Using Secrets

```yaml
steps:
  - name: Deploy
    env:
      GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    run: gh release upload $TAG artifacts/*
```

**Available secrets** (configure in repository settings):
- `GITHUB_TOKEN` — Provided automatically
- Custom secrets added in Settings → Secrets and variables

---

## Permissions Best Practices

Always use minimal required permissions:

```yaml
permissions:
  # Explicitly set each permission
  contents: read        # For checkout
  pages: write          # For GitHub Pages deployment
  id-token: write       # For trusted publishing
```

**Don't use:**
```yaml
permissions: write-all  # Too permissive!
```

---

## Artifact Sharing Between Jobs

### Upload Artifact

```yaml
- name: Upload specs
  uses: actions/upload-artifact@v4
  with:
    name: openrpc-spec
    path: |
      openrpc.json
      refs-openrpc.json
    retention-days: 1  # Clean up after 1 day
```

### Download Artifact

```yaml
- name: Download specs
  uses: actions/download-artifact@v4
  with:
    name: openrpc-spec
    path: ./specs

- run: ls -la specs/
```

### Key Points

- Artifact **names are case-sensitive**
- Use `needs: upstream-job` for explicit dependency
- Default retention is 90 days (adjust as needed)
- Large artifacts may impact performance

---

## Concurrency Control

Prevent simultaneous runs that conflict (e.g., deploying to the same Pages environment):

```yaml
concurrency:
  group: "pages"               # Name of concurrency group
  cancel-in-progress: false    # Cancel previous runs
```

Set `cancel-in-progress: true` for fast-failing workflows, `false` for critical deployments.

---

## Troubleshooting

### Workflow File Not Running

**Check:**
1. Is the file in `.github/workflows/`?
2. Does it have valid YAML syntax? → `gh workflow validate my-workflow.yaml`
3. Does the trigger condition match your event?
4. Is the file on the branch that triggered it?

### Job Times Out

**Solutions:**
- Increase `timeout-minutes` if needed
- Optimize build steps (e.g., use `cache` for dependencies)
- Run expensive steps in parallel using matrix strategy

### Artifact Not Found

**Check:**
1. Did the upstream job actually upload it? → View job logs
2. Is the name exact match (case-sensitive)?
3. Does downstream job have `needs: upstream-job`?
4. Has artifact retention expired?

### Permission Denied Errors

**Fix:**
1. Check workflow `permissions:` section
2. Verify GitHub App/token has correct scopes
3. Check branch protection rules
4. Ensure secrets are properly configured

### Workflow Dispatch Parameters Not Working

```bash
# Correct syntax for inputs
gh workflow run my-workflow.yaml -f param1=value1 -f param2=value2

# View available inputs
gh workflow view my-workflow.yaml
```

---

## Workflow State & Caching

### Cache Dependencies

```yaml
- uses: actions/setup-node@v4
  with:
    node-version: 24
    cache: npm  # Automatically cache npm dependencies

- uses: actions/setup-go@v6
  with:
    go-version: ^1.26
    # Go cache is automatic
```

### Manual Cache Management

```yaml
- uses: actions/cache@v4
  with:
    path: |
      ~/.cache/my-tool
      ./vendor
    key: ${{ runner.os }}-build-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-build-
```

---

## Matrix Strategies

Run a job across multiple configurations:

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.25', '1.26']
        node-version: ['20', '24']
    steps:
      - uses: actions/setup-go@v6
        with:
          go-version: ${{ matrix.go-version }}
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node-version }}
      - run: make test
```

---

## References

- [Workflow documentation](./.github/workflows/README.md)
- [GitHub Actions metadata syntax](./workflows/GITHUB_ACTIONS_METADATA.md)
- [Official GitHub Actions docs](https://docs.github.com/en/actions)

---

*Last updated: 2026-06-18*
