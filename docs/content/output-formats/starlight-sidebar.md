---
title: Starlight sidebar and page templates
description: Generate repository pages from Markdown or MDX templates and link them in a Starlight sidebar.
---

YAML, JSON, and HTML formats write one data or index file. `sidebar` mode writes a Markdown or MDX page for every matching repository plus an ES module containing the Starlight sidebar entries. Use `--format sidebar` or a `.mjs` output path:

```sh
gh docs index --owner mindovermachine-dev \
  --file docs/sidebar/repositories.mjs \
  --pages-dir astro/src/content/docs/repositories \
  --template docs/repository-template.mdx
```

Sidebar mode defaults to `docs/repository-template.md` or `docs/repository-template.mdx`; it does not use an embedded fallback. If both exist, `.md` is preferred. Use `--template <path>` or `index.template` to select a file elsewhere; the template location is independent of `--pages-dir`. Generated pages use the same extension as the selected template. Create the standalone Markdown starter from the maintained source without contacting GitHub:

```sh
gh docs template --starlight --output docs
```

The default output folder is `docs`, so this creates `docs/repository-template.md`. Edit that file and run the index command. For another output folder, pass `--output <folder>`; an existing template is never replaced unless `--force` is supplied.

## Template syntax and fields

Templates are regular Markdown files parsed with Go `text/template`. The gh-docs-specific delimiters are `[%` and `%]`, not Astro, Liquid, or Jekyll syntax. Markdown is rendered as-is (it is not HTML-escaped). The `yaml` helper safely encodes values for YAML front matter; use `[% yaml .Title %]` or `[% .Title | yaml %]`.

The template receives a `Repo` with these fields. Optional metadata is an empty string, empty list, or zero when unavailable and can be checked with `if`.

When placing a description in YAML frontmatter, pipe it through the `yaml` helper (for example, `description: [% .Description | yaml %]`). Repository descriptions are arbitrary strings; without YAML serialization, quotes, colons, or newlines can make the frontmatter invalid. In page body content, use `[% .Description %]` directly.

| Expression                    | Meaning                                                         |
| ----------------------------- | --------------------------------------------------------------- |
| `[% .Owner %]`                | GitHub owner or organization.                                   |
| `[% .Name %]`                 | Repository name.                                                |
| `[% .Title %]`                | Display title (repository name by default).                     |
| `[% .Description %]`          | Description, if available.                                      |
| `[% .Slug %]`                 | Safe generated page slug.                                       |
| `[% .RepoURL %]`              | Canonical GitHub repository URL.                                |
| `[% .HomepageURL %]`          | Configured documentation/homepage URL.                          |
| `[% .Topics %]`               | Repository topic names; iterate with `range`.                   |
| `[% .Workflows %]`            | Map of active workflow filenames to `true`; check with `index`. |
| `[% .DefaultBranch %]`        | Default branch, if available.                                   |
| `[% .LatestCommit.SHA %]`     | Short latest default-branch commit SHA.                         |
| `[% .LatestCommit.Date %]`    | Commit date in RFC 3339 format.                                 |
| `[% .LatestCommit.Message %]` | First line of the latest commit message.                        |
| `[% .LatestVersion %]`        | Highest stable SemVer tag, excluding prereleases.               |
| `[% .Stars %]`                | Star count, when collected.                                     |
| `[% .License %]`              | SPDX license identifier, if available.                          |
| `[% .UpdatedAt %]`            | Repository update timestamp in RFC 3339 format.                 |

In sidebar mode, the command fetches each repository's active GitHub Actions workflow paths and exposes their filenames through `.Workflows`. Use Go's built-in `index` function to render repository-specific badges only when the corresponding workflow exists:

```markdown
[% if index .Workflows "release.yml" %]
[![Release]([% .RepoURL %]/actions/workflows/release.yml/badge.svg)]([% .RepoURL %]/actions/workflows/release.yml)
[% end %]
```

Workflow lookup failures are warnings; the repository's workflow map is empty, so conditional badge blocks are omitted.

For example, this is a complete user-authored Markdown page template:

```markdown
---
title: [% yaml .Title %]
description: [% yaml .Description %]
---

# [% .Title %]

[% if .Description %][% .Description %][% end %]

Source: [% .RepoURL %]

[% if .HomepageURL %]
Documentation: [% .HomepageURL %]
[% end %]

[% if .Topics %]
Topics:
[% range .Topics %]- [% . %]
[% end %]
[% end %]

[% if .LatestCommit.SHA %]
Latest commit: [% .LatestCommit.SHA %] — [% .LatestCommit.Message %]
[% end %]
```

The complete maintained starter demonstrates every supported field, including optional commit, version, license, stars, and badge values. Near the top, it displays a documentation action link when `.HomepageURL` is set and always displays a GitHub repository action link. Template parse and execution errors identify the template and (for execution errors) the repository. Pages and sidebar content are rendered before output files are written. Metadata lookup failures are warnings; discovery, rendering, and required output write failures stop the command.

## Sidebar links and site setup

The generated module exports one `Documentation` group with one `{ label, link }` item per page. Labels and local links use JSON-safe encoding. Links never point to GitHub or repository homepages: they point to the pages created in the same run.

For the default `docs/content/repo` directory, links have the form `/repo/<slug>/`. When a path includes `content/docs`, the route is inferred relative to that `docs` directory; for other layouts, the last directory name is used. Set `base-url` / `--base-url` if the inferred route is not the public Starlight route. For example, `astro/src/content/docs/repositories` generates `/repositories/<slug>/`.

Import the generated export into the site's configuration. Existing groups can be combined with a spread:

```js
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import { sidebar } from "./docs/sidebar/repositories.mjs";

export default defineConfig({
  integrations: [
    starlight({
      title: "Documentation",
      sidebar: [
        ...sidebar,
        { label: "Guides", items: [{ label: "Getting started", link: "/guides/start/" }] },
      ],
    }),
  ],
});
```

Generated page files are overwritten on later runs. Pages for repositories that are no longer selected are not deleted; remove stale pages manually. YAML, JSON, and HTML modes do not generate pages or fetch sidebar-only metadata.
