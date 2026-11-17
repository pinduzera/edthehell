# Markdown reference

Notes for writing posts. This file sits at the repository root, outside
`content/`, so Hugo never builds or publishes it.

## Referring to another post

Post bundles live under `content/posts/<year>/<month>/<slug>/`, while the public
URL comes from the `permalinks` config:

```yaml
permalinks:
  posts: "/posts/:year/:month/:slug/"
```

Always reference pages by their content path, not by URL, so links survive
permalink or slug changes. Paths start at `content/` and may point at the bundle
directory or at its `index.md`.

### Built-in `ref` and `relref`

```md
[Placeholder post]({{< relref "/posts/2026/11/placeholder" >}})
[Placeholder post]({{< ref "/posts/2026/11/placeholder/index.md" >}})
```

| Shortcode | Output | Use when |
| --- | --- | --- |
| `relref` | Root-relative URL, `/posts/2026/11/placeholder/` | Normal in-site links |
| `ref` | Absolute URL using `baseURL` | You need a full URL (RSS, embeds) |

- Both are language-aware: on a Portuguese page they resolve to `/pt/...` when the translation exists.
- They only search the current language. If the target has no translation there, the build fails with `REF_NOT_FOUND`, even when you point at `index.pt.md`.
- Append `#fragment` to the path to link to a heading, e.g. `{{< relref "/posts/2026/11/placeholder#some-heading" >}}`.
- Shortcodes are expanded before Markdown, so they work inside link destinations as shown above.

### Cross-language `ref-lang`

When the two pages are not translations of each other, use the site shortcode in
`layouts/shortcodes/ref-lang.html`:

```md
[PT-only post]({{< ref-lang "/posts/2026/11/pt-only" "pt" >}})
[EN-only post]({{< ref-lang "/posts/2026/11/en-only" "en" >}})
```

- Arguments: content path, language code (`en` or `pt`), both required.
- Emits the target page's `RelPermalink` from that language's site, including the `/pt/` prefix when needed.
- Fails the build with a clear message if the page does not exist in that language.
- It takes a plain path only; use `ref`/`relref` when you also need a `#fragment`.

## Things to avoid

```md
[Post](/posts/2026/11/placeholder/)
```

This renders, but nothing validates it and it breaks whenever the permalink
pattern or slug changes. Likewise, a bare `{{< ref "placeholder" >}}` resolves
today but becomes ambiguous once more posts exist, so prefer the full path.

The theme's link render hook opens only external hosts in a new tab, so internal
references stay in the same tab.
