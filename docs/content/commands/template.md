---
title: template
description: Copy embedded starter files for a GitHub Pages Jekyll site.
---

`gh docs template` copies embedded site templates into an output folder. The current template type is Jekyll, intended for GitHub Pages to build without a separate local Jekyll build step.

## Usage

```sh
gh docs template --jekyll --output docz
```

By default, the output folder is `docs`. Use another folder when that name conflicts with the extension binary or an existing file.

## Flags

| Flag             | Default     | Description                                      |
| ---------------- | ----------- | ------------------------------------------------ |
| `--jekyll`       | Not enabled | Required. Selects the Jekyll template.           |
| `-c`, `--config` | `.docs.yml` | Path to the configuration file.                  |
| `-o`, `--output` | `docs`      | Destination folder.                              |
| `--force`        | Not enabled | Overwrite generated files if they already exist. |

The `template` node in `.docs.yml` supports the `output` setting. The `--output` flag takes precedence over that setting.

## Generated files

The command copies the embedded files `index.md` and `_config.yml`. The index reads `docs-repos.yml` from the site root and renders each entry as a link when it has a homepage URL.

If either generated file already exists, the command stops before writing unless `--force` is supplied. Use `--force` only when replacing those files is intended.

```sh
gh docs template --jekyll --output docz --force
```
