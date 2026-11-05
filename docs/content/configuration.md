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
  format: sidebar
  file: docs/sidebar/repositories.mjs
  pages-dir: astro/src/content/docs/repositories
  base-url: /repositories
  limit: 1000

template:
  output: docs

version:
  file: dist/version.txt
```

## `index` settings

| Key          | Default               | Description                                                                                         |
| ------------ | --------------------- | --------------------------------------------------------------------------------------------------- |
| `owner`      | None                  | GitHub user or organization. Required unless provided with `--owner`.                               |
| `visibility` | `public`              | One of `public`, `internal`, or `private`.                                                          |
| `topic`      | `documentation`       | Repository topic used to select repositories.                                                       |
| `format`     | `yaml`                | `yaml`, `json`, and `html` are single-file exports; `sidebar` generates pages and a sidebar module. |
| `file`       | Derived from `format` | Output path. `.mjs` selects paired sidebar/page generation.                                         |
| `pages-dir`  | `docs/content/repo`   | Markdown or MDX page output directory (sidebar mode).                                               |
| `base-url`   | Inferred              | URL prefix for local sidebar links when not inferable from the content path.                        |
| `badge-url`  | None                  | Optional status badge URL exposed to templates.                                                     |
| `limit`      | `1000`                | Maximum repositories fetched from GitHub.                                                           |

Do not set both `index.file` and `index.format`; the configuration is rejected. Likewise, `--file` and `--format` cannot be passed together. When a flag supplies one of these settings, it overrides the config value for the other and selects the output accordingly.

Flags override `.docs.yml`, which overrides built-in defaults. Relative output paths are resolved from the current working directory. `pages-dir`, `base-url`, and `badge-url` apply to sidebar mode; they are ignored for YAML, JSON, and HTML exports. Sidebar mode reads `docs/repository-template.md` or `docs/repository-template.mdx`, preferring `.md` if both exist. Generated pages use the selected template's extension. Create the starter Markdown template with `gh docs template --starlight --output docs`. Generated pages are replaced on later runs, but stale pages are not deleted automatically.

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
