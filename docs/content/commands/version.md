---
title: version
description: Write a build provenance file with the repository, ref and commit of the current build.
---

`gh docs version` writes a small text file describing where a build came from. It is meant to be published alongside a site so a deployed version can be traced back to a commit.

## Usage

```sh
gh docs version
gh docs version --file astro/public/version.txt
```

The default output is `dist/version.txt`. Missing parent directories are created.

## Output

```text
Build from https://github.com/mindovermachine-dev/gh-docs
Ref: main
Commit: 0123456789abcdef0123456789abcdef01234567
```

## Flags

| Flag             | Default            | Description                      |
| ---------------- | ------------------ | -------------------------------- |
| `-c`, `--config` | `.docs.yml`        | Path to the configuration file.  |
| `-f`, `--file`   | `dist/version.txt` | Output path of the version file. |

## Where the values come from

Each value is taken from the first source that defines it, so the command works unchanged locally and in GitHub Actions.

| Value  | 1. Explicit  | 2. GitHub Actions                         | 3. git                      |
| ------ | ------------ | ----------------------------------------- | --------------------------- |
| Repo   | `BUILD_REPO` | `GITHUB_SERVER_URL` + `GITHUB_REPOSITORY` | `git remote get-url origin` |
| Ref    | `BUILD_REF`  | `GITHUB_REF_NAME`                         | `git branch --show-current` |
| Commit | `BUILD_SHA`  | `GITHUB_SHA`                              | `git rev-parse HEAD`        |

On a detached HEAD with no variables set, the ref is written as `detached`. If a value cannot be resolved from any source, for example outside a git repository, the command fails.

## Configuration

The `version` node in `.docs.yml` supports `file`. The `--file` flag takes precedence. See [configuration](../configuration.md).
