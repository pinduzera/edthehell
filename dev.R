library("blogdown")
options(blogdown.method = "markdown")
options(blogdown.author = "Ed The Hell")
options(blogdown.subdir = "posts")
#options(blogdown.new_bundle =  TRUE)
options(blogdown.hugo.version = "0.170.0")


new_post <- function(title, slug, lang = "en", date = Sys.time(), ...) {
  if (inherits(date, "Date")) date <- as.POSIXct(format(date, "%Y-%m-%d"))
  stamp <- format(date, "%Y-%m-%dT%H:%M:%S%z")
  stamp <- sub("([+-][0-9]{2})([0-9]{2})$", "\\1:\\2", stamp)
  file <- file.path(
    "posts", format(date, "%Y"), format(date, "%m"), slug,
    if (lang == "en") "index.md" else paste0("index.", lang, ".md")
  )
  args <- list(...)
  args$time <- NULL
  do.call(blogdown::new_post, c(list(title, file = file, slug = slug, date = stamp), args))
}

# new_post("My title", "my-slug")             # English: index.md
# new_post("Meu título", "my-slug", "pt")     # Portuguese: index.pt.md
# new_post("Backdated", "backdated", date = as.POSIXct("2026-11-17 09:30:00"))

serve_site()
#build_site()
