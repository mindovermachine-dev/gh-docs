---
title: YAML output
description: YAML structure and file selection for the index command.
---

YAML is the default `index` output. Use `.yml` or `.yaml` with `--file`, or select `--format yaml`.

```sh
gh docs index --owner mindovermachine-dev --file docs-repos.yml
```

The output is a sequence of repositories. Each entry contains `name`, `description`, and `homepageUrl`:

```yaml
- name: remote-astro-theme
  description: A remote Astro theme for documentation sites
  homepageUrl: https://docs.example.org/remote-astro-theme/
```

An empty GitHub homepage URL is emitted as an empty string. The repository description and homepage URL reflect the metadata returned by `gh repo list`.

See [JSON](./json.md), [HTML](./html.md), and [Starlight sidebar](./starlight-sidebar.md) for other formats.
