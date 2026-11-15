# Translations

The site is bilingual: English at the site root (`/`) and Portuguese under
`/pt/`. English is the default language (`defaultContentLanguage: "en"`), and
because `defaultContentLanguageInSubdir: false`, English URLs have no prefix.

## Writing a translated page

Hugo pairs translations by **filename suffix** or by an explicit
`translationKey`. The suffix approach is what blogdown and this site use:

```
content/posts/my-post/index.md       → /posts/my-post/       (English)
content/posts/my-post/index.pt.md    → /pt/posts/my-post/    (Portuguese)
```

`index.en.md` also projects and is what blogdown writes when knitting
`index.en.Rmd`. The language code in the suffix must match the language key in
`config/_default/languages.<lang>.yaml` (here: `en` and `pt`).

If you'd rather not use suffixes — useful when a Portuguese page is a rewrite
rather than a translation — set the same `translationKey` in both files' front
matter:

```yaml
---
title: "A Second Post"
translationKey: "second-post"
---
```

A page with no translation still renders in its own language. The language
switcher then points to the section index in the other language (and shows a
tooltip saying so), rather than 404ing.

## R Markdown / blogdown

blogdown writes the language suffix into the filename, so knit
`index.en.Rmd` for English and `index.pt.Rmd` for Portuguese. Figures land in
matching directories (`index.en_files/`, `index.pt_files/`), and the
`blogdown/postref` shortcode already resolves per page, so nothing else needs
changing.

## Where the strings live

| What | File |
| --- | --- |
| Site titles, descriptions, bios, date format | `config/_default/languages.<lang>.yaml` |
| Untranslated values shared by all languages | `config/_default/params.yaml` |
| UI strings | theme: `themes/github.com/omegion/hugo-omegion/i18n/en.yaml` (fallback), site: `i18n/pt.yaml` |
| Section index titles ("Posts", "Work") | `content/posts/_index.<lang>.md` |

Templates call `i18n "key"`. Any key missing from a language file falls back to
the theme's `en.yaml`, so adding a language means translating only what differs.

Dates go through `time.Format` rather than `.Date.Format`, because only the
former localizes month names.

## Cool Links

`content/cool-links.md` uses `layout: cool-links`, which renders the groups
declared in its front matter as a two-column grid. To add a group, append one:

```yaml
sections:
  - title: Board games
    links:
      - title: BoardGameGeek
        url: https://boardgamegeek.com
        description: The reference database — ratings, rules, and forums.
```

`title` and `description` support emoji, like the rest of the theme. Relative
URLs (starting with `/`) stay in the same tab and get no external arrow;
anything else opens in a new one. The Portuguese version is
`content/cool-links.pt.md` — same structure, translated strings.
