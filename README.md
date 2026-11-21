# Using this site without R

The site is a plain [Hugo](https://gohugo.io/) project. R and blogdown are only a
convenience wrapper around Hugo, so everything except knitting `.Rmd` files can be
done with the Hugo CLI alone.

## Setup

Install Hugo extended, version 0.170.x (same as CI and `hugo.toml`):

```sh
hugo version   # must report v0.170.x
```

The theme is vendored in `themes/github.com/omegion/hugo-omegion`, and JavaScript
is bundled by Hugo's built-in esbuild. No Go modules, npm install, or R needed.

## Preview locally

```sh
hugo server -D          # http://localhost:1313/, -D includes drafts
```

`hugo server` rebuilds on save.

## Create a post with the CLI

`tools/post-cli` builds the date folder, slug folder, front matter, and the
year/month section indexes for you:

```sh
cd tools/post-cli
go build -o post .

# English post
./post -date "2026-11-21 10:30" -tags "go,cli" "My title"

# English and Portuguese together
./post -both -title-pt "Meu título" -tags "go" "My title"

# Move an existing post to a new publish date
./post -update-date -slug my-title -date "2026-12-05 09:15"

# Delete a post (asks first unless it is a draft or -force is passed)
./post -delete -slug my-title

# Other options
./post -h
```

Run it from the repository root or anywhere inside it: it looks upwards for
`content/posts`. Pass `-root path/to/content/posts` to override, `-draft` to
mark it as a draft, and `-force` to overwrite an existing post. Existing
section indexes are never overwritten. On a terminal the result is rendered
with [gum](https://github.com/charmbracelet/gum) when it is installed, and
falls back to plain output otherwise. Run its tests with
`cd tools/post-cli && go test ./...`.

## Write or edit a post by hand

Posts are bundles. Create the folder and file yourself:

```
content/posts/2026/11/my-slug/index.md      -> /posts/2026/11/my-slug/  (English)
content/posts/2026/11/my-slug/index.pt.md   -> /pt/posts/2026/11/my-slug/ (Portuguese)
```

The folder name is not the URL: `config/_default/config.yaml` builds URLs from
the front matter `date` (year and month) and `slug`. Copy the front matter shape
used by existing posts:

```yaml
---
title: "My title"
author: "Ed The Hell"
date: '2026-11-21T10:00:00-03:00'
slug: my-slug
categories: []
tags: [tag1, tag2]
---
```

Add `draft: true` to keep a post out of production builds. Images live in the
same bundle folder and are referenced relatively: `![alt](picture.webp)`.

## Build

```sh
hugo --gc --minify       # output in public/
```

Pushing to `main` triggers `.github/workflows/hugo.yaml`, which deploys the
built site to GitHub Pages. Committing is enough; no manual deploy step.

## What still needs R

- `.Rmd` sources and anything produced by knitr: figures, computed values,
  inline `` `r code` ``. Knit with blogdown as before.
- R Markdown image attributes such as `{width=60%}` inside the title. In plain
  Markdown, put the attribute list on its own line after a standalone image:

  ```markdown
  ![alt](picture.webp)

  {width=60%}
  ```

  (allowed by the goldmark settings in `config/_default/markup.toml`), or use
  plain HTML: `<img src="picture.webp" width="60%">`.
- The `blogdown/postref` shortcode is available at build time from
  `layouts/shortcodes/blogdown/postref.html`, so it keeps working without R.

See `TRANSLATING.md` for the bilingual content rules.
