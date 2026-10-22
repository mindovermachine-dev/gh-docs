---
title: index
description: List documentation repositories and write their metadata to an output file.
---

**`gh docs index` lists repositories owned by a GitHub user or organization, filters them by repository topic, and writes each matching repository's name, description, and homepage URL.**

## Usage

```sh
gh docs index --owner mindovermachine-dev
```

The owner can be set in `.docs.yml` instead of passed as a flag. The command invokes `gh repo list` and therefore requires the GitHub CLI to be installed and authenticated with access to the selected repositories.

## Flags

| Flag                 | Default          | Description                                                                            |
| -------------------- | ---------------- | -------------------------------------------------------------------------------------- |
| `-c`, `--config`     | `.docs.yml`      | Path to the configuration file.                                                        |
| `-o`, `--owner`      | None             | GitHub user or organization to list. Required unless configured.                       |
| `-v`, `--visibility` | `public`         | Repository visibility: `public`, `internal`, or `private`.                             |
| `-t`, `--topic`      | `documentation`  | Repository topic to match. Matching is case-insensitive.                               |
| `-F`, `--format`     | `yaml`           | Output format: `yaml`, `json`, `html`, or `sidebar`. Cannot be used with `--file`.     |
| `-f`, `--file`       | `docs-repos.yml` | Output path. Its extension selects the format. Missing parent directories are created. |
| `-l`, `--limit`      | `1000`           | Maximum number of repositories returned by `gh repo list`.                             |

Supported output extensions are `.yml`, `.yaml`, `.json`, `.html`, and `.mjs`. The `.mjs` extension selects the Starlight sidebar format.

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

The `sidebar` format writes an ES module exporting an Astro Starlight theme compatible sidebar array. Select it with `--format sidebar` or give `--file` a `.mjs` extension.

See [Starlight's sidebar reference](https://starlight.astro.build/guides/sidebar/)

```sh
gh docs index --owner mindovermachine-dev --file docs/docs-repos.mjs
```

The export contains a `Documentation` group. Each matching repository becomes a `{ label, link }` item, the form supported for external links by Starlight. `label` is the repository name. `link` uses the GitHub homepage URL when configured; if it is empty, the CLI uses `https://github.com/<owner>/<repo>`.

```js
export const sidebar = [
  {
    label: "Documentation",
    items: [{ label: "remote-astro-theme", link: "https://docs.example.org/remote-astro-theme/" }],
  },
];
```

Import the named export into the site's Astro configuration and use it in Starlight's main configuration site. The following is an example of how it is used in combination with a [`remote-astro-theme`](https://docs.mindovermachine.dev/remote-astro-theme)

```js
// from mindovermachine-dev/mindovermachine-dev.github.io
import { sidebar } from "./docs-repos.mjs";

export default {
  theme: "mindovermachine-dev/mom-doc-theme@main",
  astro: {
    site: "https://docs.mindovermachine.dev",
    base: ".",
    server: { host: true },
  },
  title: "Mind over Machine Dev Docs",
  social: [
    {
      icon: "github",
      label: "GitHub",
      href: "https://github.com/mindovermachine-dev",
    },
  ],
  sidebar,
};
```

## Examples

Generate the default YAML output:

```sh
gh docs index --owner mindovermachine-dev
```

Write a JSON index at a specific path:

```sh
gh docs index --owner mindovermachine-dev --file site/data/docs-repos.json
```

Select an organization and topic from flags, and write a Starlight sidebar module:

```sh
gh docs index --owner mindovermachine-dev --topic documentation --file docs/docs-repos.mjs
```

See [configuration](../configuration.md) for setting command defaults and [output formats](../output-formats/yaml.md) for the serialized structures.
