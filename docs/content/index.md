---
title: gh-docs CLI Reference
description: Reference manual for the gh-docs GitHub CLI extension.
---

`gh-docs` is a GitHub CLI extension for assembling documentation indexes and generating starter site files.

## Use case

It is a helper utility in the grand scheme of building consistent documentation for an entire GitHub or organization, utilizing GitHub Pages built-in sub page feature:

- **GitHub Pages parent/child feature**<br/>
  A repository named `{OWNER}.github.io` (e.g. `mindovermachine-dev.github.io`) is a special repo. If it exists, it's becomes the root of _all_ other github pages under the user/organization — unless they explicitly define a different `CNAME`.
- **GitHub favors Jekyll**<br/>
  GitHub has a special history with Jekyll. If a GitHub page does not use a GitHub Action to generate the static web site then it points to a branch/folder. That folder is expected to contain _either_ valid HTML _or_ a a valid Jekyll site which GitHub will then build and publish.
- **MarkDown is preferred**<br/>
  GitHub pages only supports a small subset of Jekyll features and themes (see the [`github-pages`](https://github.com/github/pages-gem) gem). But the general concept that developers maintain their reference documentation in MarkDown only, and leave all boilerplate code aside is really a beautiful concept.
- **Remote Themes in Jekyll and Astro**<br/>
  If you want to use a more verbose or feature rich jekyll theme or even a different static site generator like Hugo or Astro. Then you need quite a substantial amount of boilerplate code — It's a scope creep. Now you're not just maintaining content. Jekyll has an elegant solution; a [`jekyll-remote-theme`](https://github.com/benbalter/jekyll-remote-theme) gem which allows you to keep _all_ your the boilerplate code for your website in a seperate theme repo and then continue to just keep the actual content files in your repo. The Astro Starlight theme has become the de facto standard for generating manual pages, but unfortunately Astro didn't have an equivalent remote theme package so we built it. [`remote-asto-theme`](https://docs.mindovermachine.dev/remote-astro-theme)

Say you enable the GitHub pages for `my-org/my-org.github.io` and publish it on `https://my-org.com` (you set the `CNAME`). And then on a product repo in organization `my-org/some-cli` also setup GitHub pages, but leave the site field empty. Then it will automatically be published at `https://my-org.com/some-cli`
