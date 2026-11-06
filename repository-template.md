---
title: [% .Title | yaml %]
description: [% .Description | yaml %]
tableOfContents: false
hero:
  tagline: [% .Description | yaml %]
  actions:
    - text: Documentation
      link: [% .HomepageURL %]
      icon: right-arrow
    - text: Github
      link: [% .RepoURL %]
      icon: right-arrow
      variant: minimal
---

**`[[% .Owner %]/[% .Name %]]([% .RepoURL %])`**

[% if index .Workflows "release.yml" -%]
[![Release]([% .RepoURL %]/actions/workflows/release.yml/badge.svg)]([% .RepoURL %]/actions/workflows/release.yml)
[% end -%]
[% if index .Workflows "release-to-prod.yml" -%]
[![Release]([% .RepoURL %]/actions/workflows/release-to-prod.yml/badge.svg)]([% .RepoURL %]/actions/workflows/release-to-prod.yml)
[% end -%]
[% if index .Workflows "on_semver.yml" -%]
[![Release]([% .RepoURL %]/actions/workflows/on_semver.yml/badge.svg)]([% .RepoURL %]/actions/workflows/on_semver.yml)
[% end -%]
[% if .LatestVersion -%]
🏷️ [`[% .LatestVersion %]`]([% .RepoURL %]/releases/tag/[% .LatestVersion %])<br/>
[% end -%]
[% if .License -%]
📜 License: <br/>[% .License %]
[% end -%]
💬 [Discussions]([% .RepoURL %]/discussions/)<br/>
🐛 [Issues]([% .RepoURL %]/issues/)
🏆 [Releases]([% .RepoURL %]/releases/)
[% if .Topics %]
**Topics**

[% range .Topics %]- [% . %]
[% end %]
[% end %]
