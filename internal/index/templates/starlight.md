---
title: [% .Title | yaml %]
description: [% .Description | yaml %]
---

[% if .BadgeURL %]
<img src="[% .BadgeURL %]" alt="Build status">
[% end %]

## [% .Title %]

[% .Description %]

## Repository

- Owner: [% .Owner %]
- Name: [% .Name %]
- Page slug: `[% .Slug %]`
- Source: <[% .RepoURL %]>
  [% if .HomepageURL %]
- Documentation: <a href="[% .HomepageURL %]">[% .HomepageURL %]</a>
  [% end %]
  [% if .DefaultBranch %]- Default branch: `[% .DefaultBranch %]`[% end %]
- Stars: [% .Stars %]
  [% if .License %]
- License: [% .License %]
  [% end %]
- Last updated: [% .UpdatedAt %]

## Topics

[% if .Topics %]
[% range .Topics %]- [% . %]
[% end %]
[% else %]No topics listed.
[% end %]

## Activity

[% if .LatestVersion %]

- Latest stable version: [% .LatestVersion %]

[% end %]
[% if .LatestCommit.SHA %]

- Latest commit: <a href="[% .RepoURL %]/commit/[% .LatestCommit.SHA %]">[% .LatestCommit.SHA %]</a>

[% end %]
[% if .LatestCommit.Date %]

- Commit date: [% .LatestCommit.Date %]

[% end %]
[% if .LatestCommit.Message %]

- Commit message: [% .LatestCommit.Message %]

[% end %]
