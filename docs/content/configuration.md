---
title: Configuration
description: Configure gh-docs commands with the shared .docs.yml file.
---

The CLI reads command settings from `.docs.yml` in the current working directory. Each command has its own top-level YAML node, so settings can grow independently as more commands are added.

```yaml
index:
  owner: mindovermachine-dev
  visibility: public
  topic: documentation
  format: yaml
  limit: 1000

template:
  output: docs

version:
  file: dist/version.txt
```

## `index` settings

| Key          | Default               | Description                                                                          |
| ------------ | --------------------- | ------------------------------------------------------------------------------------ |
| `owner`      | None                  | GitHub user or organization. Required unless provided with `--owner`.                |
| `visibility` | `public`              | One of `public`, `internal`, or `private`.                                           |
| `topic`      | `documentation`       | Repository topic used to select repositories.                                        |
| `format`     | `yaml`                | One of `yaml`, `json`, `html`, or `sidebar`.                                         |
| `file`       | Derived from `format` | Output path. Supported extensions are `.yml`, `.yaml`, `.json`, `.html`, and `.mjs`. |
| `limit`      | `1000`                | Maximum repositories fetched from GitHub.                                            |

Do not set both `index.file` and `index.format`; the configuration is rejected. Likewise, `--file` and `--format` cannot be passed together. When a flag supplies one of these settings, it overrides the config value for the other and selects the output accordingly.

## `template` settings

| Key      | Default | Description                                 |
| -------- | ------- | ------------------------------------------- |
| `output` | `docs`  | Folder where the template files are copied. |

## `version` settings

| Key    | Default            | Description                      |
| ------ | ------------------ | -------------------------------- |
| `file` | `dist/version.txt` | Output path of the version file. |

## Precedence and config paths

For settings provided by multiple sources, command-line flags take precedence over the command's node in `.docs.yml`, which takes precedence over built-in defaults. The index owner has no built-in default and must be supplied by flag or config.

Use `--config` (or `-c`) to select another config file. A missing default `.docs.yml` is allowed; a missing file explicitly selected with `--config` is an error.
