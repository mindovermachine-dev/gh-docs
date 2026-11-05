---
title: index
description: Export repository data or generate repository pages and a Starlight sidebar.
---

**`gh docs index` lists repositories owned by a GitHub user or organization and filters them by repository topic. YAML, JSON, and HTML produce a single-file export; sidebar mode produces one Markdown page per repository and a Starlight sidebar module.**

## Usage

```sh
gh docs index --owner mindovermachine-dev
```

The owner can be set in `.docs.yml` instead of passed as a flag. The command invokes `gh repo list` and therefore requires the GitHub CLI to be installed and authenticated with access to the selected repositories.

## Flags

| Flag                 | Default                 | Description                                                                                      |
| -------------------- | ----------------------- | ------------------------------------------------------------------------------------------------ |
| `-c`, `--config`     | `.docs.yml`             | Path to the configuration file.                                                                  |
| `-o`, `--owner`      | None                    | GitHub user or organization to list. Required unless configured.                                 |
| `-v`, `--visibility` | `public`                | Repository visibility: `public`, `internal`, or `private`.                                       |
| `-t`, `--topic`      | `documentation`         | Repository topic to match. Matching is case-insensitive.                                         |
| `-F`, `--format`     | `yaml`                  | Output format: `yaml`, `json`, `html`, or `sidebar`. Cannot be used with `--file`.               |
| `-f`, `--file`       | `docs-repos.yml`        | Output path. `.mjs` selects paired sidebar/page generation.                                      |
| `--pages-dir`        | `docs/content/repo`     | Directory for generated Markdown pages in sidebar mode.                                          |
| `--template`         | Built-in Starlight page | Markdown template in sidebar mode; with `--starlight`, destination path.                         |
| `--base-url`         | Inferred                | URL prefix for local sidebar links when the filesystem layout does not determine the site route. |
| `--badge-url`        | None                    | Optional status badge URL exposed to the page template.                                          |
| `--starlight`        | `false`                 | Copy the maintained page template and exit without querying GitHub.                              |
| `-l`, `--limit`      | `1000`                  | Maximum number of repositories returned by `gh repo list`.                                       |

Supported output extensions are `.yml`, `.yaml`, `.json`, `.html`, and `.mjs`. YAML, JSON, and HTML are single-file exports. The `.mjs` extension selects the paired sidebar/pages format.

## formats

The output format of the generated index is determined either by the file extension when using the `--file` flag or explicitly by using the `--format` flag. These flags are mutually exclusive to each other.

### html

Select HTML with `--format html` or give `--file` a `.html` extension.

```sh
gh docs index --owner mindovermachine-dev --file docs/index.html
```

The output is a standalone HTML document with a Documentation heading and a bulleted list. A repository name links to its homepage URL when one is set; otherwise, the name is plain text. A non-empty description follows the name.

Names, descriptions, and URLs are escaped by Go's HTML template renderer before they are written into the page.

### json

Select JSON with `--format json` or give `--file` a `.json` extension.

```sh
gh docs index --owner mindovermachine-dev --file docs/docs-repos.json
```

The output is a JSON array. Each object contains `name`, `description`, and `homepageUrl`:

```json
[
  {
    "name": "remote-astro-theme",
    "description": "A remote Astro theme for documentation sites",
    "homepageUrl": "https://docs.mindovermachine.dev/remote-astro-theme/"
  }
]
```

### yaml

YAML is the default `index` output. Use `.yml` or `.yaml` with `--file`, or select `--format yaml`.

```sh
gh docs index --owner mindovermachine-dev --file docs/docs-repos.yml
```

The output is a sequence of repositories. Each entry contains `name`, `description`, and `homepageUrl`:

```yaml
- name: remote-astro-theme
  description: A remote Astro theme for documentation sites
  homepageUrl: https://docs.example.org/remote-astro-theme/
```

An empty GitHub homepage URL is emitted as an empty string. The repository description and homepage URL reflect the metadata returned by `gh repo list`.

### Sidebar (ES module)

The `sidebar` format writes a Markdown page for every matching repository and an ES module exporting local links to those pages. Select it with `--format sidebar` or give `--file` a `.mjs` extension.

See [Starlight's sidebar reference](https://starlight.astro.build/guides/sidebar/)

```sh
gh docs index --owner mindovermachine-dev --file docs/sidebar/repositories.mjs \
  --pages-dir astro/src/content/docs/repositories \
  --template astro/repository-template.md
```

The generated module exports one `Documentation` group with `{ label, link }` items. Links point to the pages generated in the same invocation, not to GitHub or repository homepage URLs. For the default `docs/content/repo` directory the inferred route is `/repo/<slug>/`; for a path under `src/content/docs`, the route is inferred relative to `docs`. Use `--base-url` when you need to specify a route explicitly. Missing directories are created and generated pages are overwritten. Pages for repositories no longer selected are not deleted and must be cleaned up manually.

Run `gh docs index --starlight --template astro/repository-template.md` to copy the maintained starter Markdown template without querying GitHub. If omitted, the destination is `repository-template.md` in the current directory. Existing files are never overwritten. Edit the copied file and pass it to a normal sidebar invocation.

See [Starlight sidebar and page templates](../output-formats/starlight-sidebar.md) for template keywords, delimiters, configuration, and site imports.

## Examples

Generate the default YAML output:

```sh
gh docs index --owner mindovermachine-dev
```

Write a JSON index at a specific path:

```sh
gh docs index --owner mindovermachine-dev --file site/data/docs-repos.json
```

Select an organization and topic from flags, and write repository pages plus a Starlight sidebar module:

```sh
gh docs index --owner mindovermachine-dev --topic documentation --file docs/docs-repos.mjs
```

See [configuration](../configuration.md) for setting command defaults and [output formats](../output-formats/yaml.md) for the serialized structures.
