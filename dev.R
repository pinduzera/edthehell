library("blogdown")
options(blogdown.method = "markdown")
options(blogdown.author = "Ed The Hell")
options(blogdown.subdir = "posts")
#options(blogdown.new_bundle =  TRUE)
options(blogdown.hugo.version = "0.160.0")

# New post: page bundle at content/posts/<year>/<month>/<slug>/index.md.
# `file` sets the folder layout from the post date and `slug` pins the URL
# segment, so the permalinks config still turns it into posts/2026/11/<slug>/
# while the date stays in front matter. Without `file`, blogdown would create
# YYYY-MM-DD-<slug>/index.md; without `slug`, translated bundles like
# index.pt.md fall back to the post title for the URL.
new_post <- function(title, slug, lang = "en", date = Sys.Date(), ...) {
  date <- as.Date(date)
  file <- file.path(
    "posts", format(date, "%Y"), format(date, "%m"), slug,
    if (lang == "en") "index.md" else paste0("index.", lang, ".md")
  )
  blogdown::new_post(title, file = file, slug = slug, date = date, ...)
}

# new_post("My title", "my-slug")             # English: index.md
# new_post("Meu título", "my-slug", "pt")     # Portuguese: index.pt.md

serve_site()
#build_site()