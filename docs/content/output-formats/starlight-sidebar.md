---
title: Starlight sidebar output
description: Generate an importable Starlight sidebar module from matching repositories.
---

The `sidebar` format writes an ES module exporting a Starlight-compatible sidebar array. Select it with `--format sidebar` or give `--file` a `.mjs` extension.

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

Import the named export into the site's Astro configuration and pass it to Starlight:

```js
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import { sidebar } from "./docs/docs-repos.mjs";

export default defineConfig({
  integrations: [starlight({ title: "Documentation", sidebar })],
});
```

See [Starlight's sidebar reference](https://starlight.astro.build/reference/configuration/#sidebar) for the supported item schema, and [YAML](./yaml.md), [JSON](./json.md), and [HTML](./html.md) for other CLI output formats.
