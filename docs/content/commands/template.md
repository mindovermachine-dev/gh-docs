---
title: template
description: Copy starter files for Jekyll sites or Starlight repository pages.
---

`gh docs template` copies starter files into an output folder. Use Jekyll for a simple GitHub Pages site or Starlight to create the Markdown template used by `gh docs index --format sidebar`.

## Usage

```sh
gh docs template --jekyll --output docz
```

By default, the output folder is `docs`. Use another folder when that name conflicts with the extension binary or an existing file.

Create the Starlight repository page template with:

```sh
gh docs template --starlight
```

This writes `docs/repository-template.md`. Edit this standalone Markdown file, then use `gh docs index --format sidebar`; sidebar mode reads exactly that path.

## Flags

| Flag             | Default     | Description                                                       |
| ---------------- | ----------- | ----------------------------------------------------------------- |
| `--jekyll`       | Not enabled | Select the Jekyll starter. Mutually exclusive with `--starlight`. |
| `--starlight`    | Not enabled | Copy `repository-template.md` from the Starlight starter source.  |
| `-c`, `--config` | `.docs.yml` | Path to the configuration file.                                   |
| `-o`, `--output` | `docs`      | Destination folder.                                               |
| `--force`        | Not enabled | Overwrite generated files if they already exist.                  |

The `template` node in `.docs.yml` supports the `output` setting. The `--output` flag takes precedence over that setting.

## Generated files

`--jekyll` copies the embedded files `index.md` and `_config.yml`. The index reads `docs-repos.yml` from the site root and renders each entry as a link when it has a homepage URL.

If either generated file already exists, the command stops before writing unless `--force` is supplied. Use `--force` only when replacing those files is intended.

```sh
gh docs template --jekyll --output docz --force
```

`--starlight` copies the maintained Markdown source from `internal/index/templates/starlight.md` to `<output>/repository-template.md`. The default output folder therefore creates `docs/repository-template.md`; an alternate `--output` directory is honored. If the file exists, the command refuses to overwrite it unless `--force` is provided.
