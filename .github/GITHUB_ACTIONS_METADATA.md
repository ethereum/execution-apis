# GitHub Actions Metadata Syntax Reference

This document provides a comprehensive guide to GitHub Actions metadata syntax for the `execution-apis` repository, based on the [official GitHub documentation](https://docs.github.com/en/actions/reference/workflows-and-actions/metadata-syntax).

## Overview

GitHub Actions metadata syntax is used to define custom actions and configure their behavior. There are two main uses:
1. **Action Metadata** – Defining reusable custom actions with `action.yml`
2. **Workflow Metadata** – Configuring workflows in `.github/workflows/*.yml`

---

## 1. Action Metadata (`action.yml`)

If your repository exposes reusable GitHub Actions, use `action.yml` to define their interface.

### Basic Structure

```yaml
name: My Custom Action
author: moonrager13
description: A brief description of what this action does

inputs:
  my-input:
    description: Description of the input parameter
    required: true
    default: default-value

outputs:
  my-output:
    description: Description of the output value
    value: ${{ steps.my-step.outputs.result }}

runs:
  using: 'composite'
  steps:
    - run: echo "Running my action"
      shell: bash
```

### Key Sections

#### `name` (Required)
The display name of your action as shown in GitHub Marketplace.

```yaml
name: Build and Deploy Spec
```

#### `description` (Required)
A short description of what the action does.

```yaml
description: Builds OpenRPC specification and deploys it to GitHub Pages
```

#### `author` (Optional)
The creator's name or organization.

```yaml
author: moonrager13
```

#### `inputs`
Define parameters that the action accepts at runtime.

**Rules:**
- Input IDs must start with a letter or underscore
- Can contain alphanumeric characters, hyphens, or underscores
- Input IDs are case-insensitive

```yaml
inputs:
  version:
    description: Version to build
    required: true
    default: '1.0.0'
  
  deploy-to-pages:
    description: Deploy to GitHub Pages after build
    required: false
    default: 'true'
```

#### `outputs`
Declare data the action produces for use by subsequent actions.

```yaml
outputs:
  build-artifact:
    description: Path to the built artifact
    value: ${{ steps.build.outputs.artifact }}
  
  spec-version:
    description: Version of the generated spec
    value: ${{ steps.extract-version.outputs.version }}
```

#### `runs`
Specifies how the action executes. Varies by action type:

##### JavaScript Action
```yaml
runs:
  using: 'node20'  # or node24
  main: 'index.js'
  pre: 'pre.js'         # optional
  post: 'post.js'       # optional
```

##### Docker Container Action
```yaml
runs:
  using: 'docker'
  image: 'docker://ghcr.io/owner/image:latest'
  entrypoint: '/entrypoint.sh'
  args:
    - 'arg1'
    - 'arg2'
  pre-entrypoint: 'pre.sh'    # optional
  post-entrypoint: 'post.sh'  # optional
  env:
    MY_VAR: value
```

##### Composite Action
```yaml
runs:
  using: 'composite'
  steps:
    - run: echo "Step 1"
      shell: bash
    
    - uses: actions/checkout@v4
    
    - run: make build
      shell: bash
```

#### `branding`
Optional styling for GitHub Marketplace display.

```yaml
branding:
  icon: 'box'              # Feather icon name
  color: 'blue'            # Color name
```

---

## 2. Workflow Metadata

### Structure

All workflows in this repository follow the pattern:

```yaml
name: Workflow Display Name

on:
  # Trigger conditions
  push:
    branches: [main]
  workflow_dispatch:

permissions:
  contents: read
  pages: write

jobs:
  job-name:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: make build
```

### Common Metadata Fields

#### `name`
Display name for the workflow in the Actions tab.

```yaml
name: Deploy to GitHub Pages
```

#### `on`
Triggers that run the workflow.

```yaml
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  workflow_dispatch:
  release:
    types: [published]
```

#### `permissions`
Specifies which GitHub token permissions are required.

```yaml
permissions:
  contents: read
  pages: write
  id-token: write
```

---

## 3. Current Workflows in execution-apis

### Overview

This repository uses multiple specialized workflows:

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| **release.yaml** | Release published or manual dispatch | Full release pipeline: build, version, GitHub Release, deploy |
| **deploy.yaml** | Push to main or manual dispatch | Continuous deployment to GitHub Pages |
| **test.yaml** | Push/PR | Run tests, build checks, spec validation |
| **test-deploy.yaml** | PR to main | Smoke-test site build with version snapshot |
| **sync-release-notes.yaml** | Release published/edited | Mirror release notes to docs via PR |
| **publish-spec.yaml** | Manual dispatch | Recovery: re-push assembled spec from release assets |
| **codeql.yml** | Push/PR | Security scanning |
| **spellcheck.yaml** | Push/PR | Spell checking |

### Key Patterns Used

#### 1. Conditional Permissions
Each workflow declares minimal required permissions:

```yaml
permissions:
  contents: read      # For checkout
  pages: write        # For GitHub Pages
  id-token: write     # For trusted publishing
```

#### 2. Job Dependencies
Jobs declare explicit dependencies with `needs`:

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps: [...]

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps: [...]
```

#### 3. Concurrency Control
Prevents simultaneous runs to a shared resource:

```yaml
concurrency:
  group: "pages"
  cancel-in-progress: false
```

#### 4. Artifact Sharing
Jobs pass data via artifacts:

```yaml
# Upload artifact
- name: Upload spec
  uses: actions/upload-artifact@v4
  with:
    name: openrpc-spec
    path: |
      openrpc.json
      refs-openrpc.json

# Download artifact
- name: Download spec
  uses: actions/download-artifact@v4
  with:
    name: openrpc-spec
```

#### 5. Environment Variables & Secrets
Workflows use GitHub-provided and custom variables:

```yaml
- name: Build
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    BUILD_VERSION: ${{ github.ref_name }}
  run: |
    echo "Building version: $BUILD_VERSION"
```

---

## 4. Best Practices

### For Custom Actions

✅ **Do:**
- Provide clear `description` for each input and output
- Use composite actions for shell-based reusable workflows
- Pin action versions: `actions/checkout@v4` not `@latest`
- Use `pre` and `post` scripts for setup/cleanup when needed

❌ **Don't:**
- Use generic input IDs like `input1`, `input2`
- Forget to document required vs. optional inputs
- Hard-code credentials or secrets in action metadata

### For Workflows

✅ **Do:**
- Set explicit `permissions` (principle of least privilege)
- Name steps clearly: `- name: "Build spec"`
- Use `concurrency` to prevent race conditions
- Declare `needs` for job dependencies
- Pin action versions
- Use matrix strategies for multiple OS/version testing

❌ **Don't:**
- Use broad permissions like `write-all`
- Rely on implicit job ordering
- Hard-code credentials in workflows
- Use `@latest` for action versions

### For This Repository

1. **Spec Generation**: All spec generation happens in `build` and `deploy-spec-main` jobs
2. **Artifact Contracts**: Respect the artifact naming convention:
   - `openrpc-spec` — generated OpenRPC JSON files
   - `docs-snapshot` — documentation tar.gz
   - `github-pages` — site build artifact
3. **Version Management**: 
   - Stamped versions go to `assembled-spec` branch (releases)
   - Unstamped versions go to `assembled-spec-main` branch (main)

---

## 5. Resources

- [GitHub Actions: Metadata syntax for workflows](https://docs.github.com/en/actions/reference/workflows-and-actions/metadata-syntax)
- [GitHub Actions: Workflow syntax for GitHub Actions](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
- [GitHub Actions: Using actions in a workflow](https://docs.github.com/en/actions/using-workflows/using-actions)
- [GitHub Actions: Creating a Docker container action](https://docs.github.com/en/actions/creating-actions/creating-a-docker-container-action)
- [GitHub Actions: Creating a composite action](https://docs.github.com/en/actions/creating-actions/creating-a-composite-action)

---

## 6. Quick Reference: Workflow Triggers

| Trigger | Example | Use Case |
|---------|---------|----------|
| `push` | `branches: [main]` | Run on push to specific branches |
| `pull_request` | `types: [opened, synchronize]` | Run on PR events |
| `release` | `types: [published]` | Run when release is published |
| `workflow_dispatch` | Manual trigger | Ad-hoc manual runs |
| `schedule` | `- cron: '0 0 * * *'` | Scheduled runs |

---

## 7. Troubleshooting

### Issue: Workflow doesn't trigger
- Check trigger conditions (`on:`) match your event
- Verify branch name in `branches:` filter
- Check if workflow file is on the correct branch

### Issue: Artifact not found in downstream job
- Verify upstream job actually uploaded the artifact
- Check artifact name matches exactly (case-sensitive)
- Ensure downstream job has `needs: upstream-job`

### Issue: Insufficient permissions
- Check workflow `permissions:` section
- Verify GitHub App token has required scopes
- Check branch protection rules

---

*Last updated: 2026-06-18*
*For questions, see `.github/workflows/README.md`*
