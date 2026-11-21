# post manual

`post` creates Hugo post bundles for thehellas.com without R. It writes the
front matter, the date/slug folder structure, and the bilingual year/month
section indexes, then prints the resulting URLs.

```
content/posts/2026/11/my-slug/index.md      -> /posts/2026/11/my-slug/
content/posts/2026/11/my-slug/index.pt.md   -> /pt/posts/2026/11/my-slug/
```

## Build

Requires Go 1.22 or newer.

```sh
cd tools/post-cli
go build -o post .
```

The rest of this manual assumes `./post` is run from `tools/post-cli`. From
anywhere else, use the full path to the binary.

You can also add this `~/tools/post-cli` folder to PATH and call `post` directly.

## Usage

```
post [flags] "Post title"
```

The title may be given as one or more positional arguments (they are joined with
spaces) or with `-title`. `-title` wins when both are present. A title is
required; with no title the tool prints usage and exits with an error.

Flags must come before the title, because Go's `flag` package stops parsing at
the first non-flag argument.

## Pretty output

On an interactive terminal, `post` renders the result with
[gum](https://github.com/charmbracelet/gum), the same tool `manual.sh` uses: a
boxed title, the list of created files, and the public URLs.

```sh
go install github.com/charmbracelet/gum@latest
```

Pretty output is optional. Without `gum`, or when `NO_COLOR` is set or the
output is redirected, the tool falls back to plain `created ...` and
`url: ...` lines, so scripts, editors, and tests stay stable.

## Flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `-title` | | Post title, alternative to the positional argument. |
| `-slug` | derived from title | URL slug and folder name. |
| `-lang` | `en` | Language of the post: `en` or `pt`. |
| `-both` | `false` | Create both English and Portuguese versions. |
| `-title-pt` | same as English | Portuguese title; only used when a Portuguese version is created. |
| `-date` | now | Post date. See [Date formats](#date-formats). |
| `-update-date` | `false` | Move an existing post to `-date` and rewrite its front matter date. |
| `-delete` | `false` | Delete an existing post bundle. |
| `-author` | `Ed The Hell` | Front matter author. |
| `-tags` | `[]` | Comma-separated tags. |
| `-categories` | `[]` | Comma-separated categories. |
| `-draft` | `false` | Add `draft: true` to the front matter. |
| `-root` | nearest `content/posts` | Posts directory. Must already exist. |
| `-force` | `false` | Overwrite existing post files, or skip the delete confirmation. |
| `-h` | | Print usage and exit successfully. |

`-both` implies both `en` and `pt`; `-lang pt` creates only the Portuguese file.
`-title-pt` is ignored unless a Portuguese version is produced.

## Updating a post's date

After drafting, set the publish date without recreating the post:

```sh
./post -update-date -slug my-post -date "2026-12-05 09:15"

# or derive the slug from the title
./post -update-date -date "2026-12-05 09:15" "My Post"
```

`-date` is required. Use `-date now` to stamp the current time. The tool finds
`YYYY/MM/<slug>/` under the posts root,
rewrites the `date:` field in `index.md` and/or `index.pt.md`, and moves the
bundle to the year and month of the new date. Missing month section indexes are
created. Title, slug, tags, categories, and draft status are left unchanged.

## Deleting a post

Remove a post bundle once it is no longer wanted:

```sh
./post -delete -slug my-post

# or derive the slug from the title
./post -delete "My Post"
```

The whole `YYYY/MM/<slug>/` directory is removed. Draft posts are deleted
without prompting. If any language file is not a draft, the tool asks for
confirmation first; pass `-force` to skip the prompt in scripts. Section
indexes and other posts are left untouched.

## Where files go

`-root` must point at an existing directory. Without it, `post` walks up
from the current directory looking for `content/posts`, so it can be run from
the repository root or any subdirectory:

```sh
./post "My title"                       # finds ../../content/posts
./post -root ../../content/posts "My title"
```

Inside the root it creates:

```
<root>/YYYY/MM/<slug>/index.md      (English)
<root>/YYYY/MM/<slug>/index.pt.md   (Portuguese)
```

Year and month come from the resolved date, not the current date.

`-force` overwrites existing **post** files only. Section indexes are never
overwritten, even with `-force`.

## Section indexes

For every run, missing section indexes are created next to the post:

| File | Title |
| --- | --- |
| `<root>/YYYY/_index.md` | `2026` |
| `<root>/YYYY/_index.pt.md` | `2026` |
| `<root>/YYYY/MM/_index.md` | `November 2026` |
| `<root>/YYYY/MM/_index.pt.md` | `Novembro 2026` |

Existing files are left untouched, so hand-edited descriptions survive. The
English and Portuguese indexes are created for every run, even for a
single-language post, because the section is shared.

Every index gets a `description` in the matching language, for example
`Everything I've written in November 2026.` and
`Tudo o que eu escrevi em novembro de 2026.`

## Front matter

The generated file contains only front matter; the body is left for you to
write:

```yaml
---
title: "My title"
author: "Ed The Hell"
date: '2026-11-21T10:30:00-03:00'
slug: my-slug
categories: []
tags: ["go", "cli"]
---
```

- Strings are double-quoted.
- `date` is RFC 3339 in the local time zone, so the site's URL date matches the
  post date.
- Lists are inline arrays; empty lists render as `[]`.
- `draft: true` appears only with `-draft`.

## Date formats

| Input | Result |
| --- | --- |
| omitted or `now` | current time, truncated to the second |
| `2026-11-21` | `2026-11-21T12:00:00` local time (noon avoids time zone date shifts) |
| `2026-11-21 10:30` | local time |
| `2026-11-21 10:30:15` | local time |
| `2026-11-21T10:30` | local time |
| `2026-11-21T10:30:15-03:00` | RFC 3339, as given |

Unsupported formats such as `21/11/2026` are rejected with an error.

## Slugs

Without `-slug`, the slug is derived from the title:

| Title | Slug |
| --- | --- |
| `I still love the Internet` | `i-still-love-the-internet` |
| `Olá, Mundo!` | `ola-mundo` |
| `Don't  Stop` | `dont-stop` |
| `Café & Pão` | `cafe-pao` |
| `  --spaces--  ` | `spaces` |

Rules: lowercase, accents transliterated to ASCII, apostrophes removed, runs of
other characters collapsed into a single hyphen, leading and trailing hyphens
trimmed. If nothing usable remains (for example an emoji-only title), the tool
stops and asks for `-slug`.

The slug is written to the front matter, and `config/_default/config.yaml`
builds the URL from the front matter `date` and `slug`, so renaming the folder
breaks links. Pick the slug first.

## Examples

English post with tags and categories:

```sh
./post -date "2026-11-21 10:30" -tags "go,cli" -categories "notes" "My title"
```

English and Portuguese together:

```sh
./post -both -title-pt "Meu título" -tags "go" "My title"
```

Portuguese only, with an explicit slug:

```sh
./post -lang pt -slug acao -date 2027-01-05 "Ação"
```

Draft in a non-default posts directory:

```sh
./post -root /tmp/posts -draft "Work in progress"
```

## Output

On success it prints each created file and the resulting URL(s):

```
created content/posts/2026/11/my-title/index.md
created content/posts/2026/_index.md
created content/posts/2026/_index.pt.md
created content/posts/2026/11/_index.md
created content/posts/2026/11/_index.pt.md
url: /posts/2026/11/my-title/
```

With `-both` or `-lang pt` a `/pt/...` URL follows. Indexes that already existed
are not listed.

Errors go to stderr prefixed with `post:` and exit status is 1. Successful
runs, including `-h`, exit 0.

## Errors you may hit

| Message | Cause |
| --- | --- |
| `a title is required` | No positional arguments and no `-title`. |
| `could not derive a slug from ...; pass -slug` | Title has no letters or digits. |
| `invalid -lang ...: use en or pt` | Unsupported language. |
| `invalid -date ...` | Unrecognized date format. |
| `-update-date needs -date ...` | Update mode without a new date. |
| `post "..." not found under ...` | No bundle with that slug. |
| `post "..." is ambiguous: ...` | The slug exists under more than one month. |
| `no confirmation available; pass -force ...` | Non-draft delete with no stdin to answer the prompt. |
| `... already exists (use -force to overwrite)` | Post file exists; rerun with `-force`. |
| `content/posts not found; run from the repository or pass -root` | Not inside the repo and no `-root`. |
| `-root ...: no such file or directory` / `not a directory` | `-root` must be an existing directory. |

The overwrite check runs before anything is written, so a refused run leaves
the tree untouched.

## Tests

```sh
cd tools/post
go test ./...
```

The suite covers slug generation, date parsing, English/Portuguese/bilingual
output, overwrite protection, section-index preservation, help, and root
discovery.

## Related reading

- `README.md` at the repository root: previewing and building the site.
- `TRANSLATING.md`: the bilingual content rules.
