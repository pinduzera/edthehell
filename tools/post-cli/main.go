// Command post creates Hugo post bundles for thehellas.com without R.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const defaultAuthor = "Ed The Hell"

var stdin io.Reader = os.Stdin

var portugueseMonths = [...]string{
	"janeiro", "fevereiro", "março", "abril", "maio", "junho",
	"julho", "agosto", "setembro", "outubro", "novembro", "dezembro",
}

var transliterations = map[rune]string{
	'á': "a", 'à': "a", 'â': "a", 'ã': "a", 'ä': "a", 'å': "a",
	'é': "e", 'è': "e", 'ê': "e", 'ë': "e",
	'í': "i", 'ì': "i", 'î': "i", 'ï': "i",
	'ó': "o", 'ò': "o", 'ô': "o", 'õ': "o", 'ö': "o",
	'ú': "u", 'ù': "u", 'û': "u", 'ü': "u",
	'ç': "c", 'ñ': "n", 'ý': "y", 'ÿ': "y",
}

type target struct {
	path    string
	content string
}

type pendingUpdate struct {
	name    string
	content string
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		newPresenter(os.Stderr).fail(err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	ui := newPresenter(stdout)

	fs := flag.NewFlagSet("post", flag.ContinueOnError)
	fs.SetOutput(stdout)
	fs.Usage = func() {
		fmt.Fprintln(stdout, "usage: post [flags] \"Post title\"")
		fmt.Fprintln(stdout, "\nCreates content/posts/YYYY/MM/<slug>/index.md (and optionally index.pt.md).")
		fmt.Fprintln(stdout, "With -update-date, moves an existing post to a new date and rewrites its front matter.")
		fmt.Fprintln(stdout, "With -delete, removes an existing post bundle.")
		fmt.Fprintln(stdout, "\nflags:")
		fs.PrintDefaults()
	}

	title := fs.String("title", "", "post title (alternative to the positional argument)")
	slug := fs.String("slug", "", "URL slug (default: derived from the title)")
	lang := fs.String("lang", "en", "language of the post: en or pt")
	both := fs.Bool("both", false, "create both the English and Portuguese versions")
	titlePT := fs.String("title-pt", "", "Portuguese title (default: same as the English title)")
	dateArg := fs.String("date", "", "post date: now, YYYY-MM-DD, YYYY-MM-DD HH:MM[:SS], or RFC 3339 (default: now)")
	updateDate := fs.Bool("update-date", false, "move an existing post to the -date and rewrite its front matter")
	deletePost := fs.Bool("delete", false, "delete an existing post bundle (requires -slug, or a title)")
	author := fs.String("author", defaultAuthor, "front matter author")
	tagsArg := fs.String("tags", "", "comma-separated tags")
	categoriesArg := fs.String("categories", "", "comma-separated categories")
	draft := fs.Bool("draft", false, "mark the post as a draft")
	rootArg := fs.String("root", "", "posts directory (default: nearest content/posts)")
	force := fs.Bool("force", false, "overwrite existing post files")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	postTitle := strings.TrimSpace(*title)
	if postTitle == "" {
		postTitle = strings.TrimSpace(strings.Join(fs.Args(), " "))
	}

	if *deletePost && *updateDate {
		return errors.New("-delete and -update-date are mutually exclusive")
	}
	if *deletePost {
		return runDelete(ui, postTitle, *slug, *force, *rootArg)
	}
	if *updateDate {
		return runUpdateDate(ui, postTitle, *slug, *dateArg, *rootArg)
	}

	if postTitle == "" {
		fs.Usage()
		return errors.New("a title is required")
	}

	if *lang != "en" && *lang != "pt" {
		return fmt.Errorf("invalid -lang %q: use en or pt", *lang)
	}

	date, err := parseDate(*dateArg, time.Now())
	if err != nil {
		return err
	}

	postSlug := strings.TrimSpace(*slug)
	if postSlug == "" {
		postSlug = slugify(postTitle)
	}
	if postSlug == "" {
		return fmt.Errorf("could not derive a slug from %q; pass -slug", postTitle)
	}

	root, err := resolveRoot(*rootArg)
	if err != nil {
		return err
	}

	dir := filepath.Join(root, date.Format("2006"), date.Format("01"), postSlug)

	var langs []string
	if *both {
		langs = []string{"en", "pt"}
	} else {
		langs = []string{*lang}
	}

	targets := make([]target, 0, len(langs))
	for _, l := range langs {
		postTitleForLang := postTitle
		if l == "pt" && strings.TrimSpace(*titlePT) != "" {
			postTitleForLang = strings.TrimSpace(*titlePT)
		}
		targets = append(targets, target{
			path:    filepath.Join(dir, filenameForLang(l)),
			content: renderPost(postTitleForLang, *author, date, postSlug, splitList(*categoriesArg), splitList(*tagsArg), *draft),
		})
	}

	if !*force {
		for _, t := range targets {
			if _, err := os.Stat(t.path); err == nil {
				return fmt.Errorf("%s already exists (use -force to overwrite)", t.path)
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	created := make([]string, 0, len(targets)+4)
	for _, t := range targets {
		if err := os.WriteFile(t.path, []byte(t.content), 0o644); err != nil {
			return err
		}
		created = append(created, t.path)
	}

	sections, err := ensureSectionIndexes(root, date)
	if err != nil {
		return err
	}
	created = append(created, sections...)

	urls := []string{fmt.Sprintf("/posts/%s/%s/%s/", date.Format("2006"), date.Format("01"), postSlug)}
	if *both || *lang == "pt" {
		urls = append(urls, fmt.Sprintf("/pt/posts/%s/%s/%s/", date.Format("2006"), date.Format("01"), postSlug))
	}

	ui.header(postTitle, date, postSlug)
	ui.created(created)
	ui.links(urls)
	return nil
}

func runDelete(ui *presenter, title, slugArg string, force bool, rootArg string) error {
	postSlug := strings.TrimSpace(slugArg)
	if postSlug == "" {
		postSlug = slugify(title)
	}
	if postSlug == "" {
		return errors.New("-delete needs -slug, or a title to derive it from")
	}

	root, err := resolveRoot(rootArg)
	if err != nil {
		return err
	}

	dir, err := findPostDir(root, postSlug)
	if err != nil {
		return err
	}

	files, err := readPostFiles(dir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("%s has no index.md or index.pt.md", dir)
	}

	draft := true
	for _, file := range files {
		if !frontMatterDraft(file.content) {
			draft = false
			break
		}
	}

	if !draft && !force {
		ok, err := ui.confirm(fmt.Sprintf("Delete post %q? This cannot be undone.", postSlug))
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(ui.out, "cancelled")
			return nil
		}
	}

	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	ui.deleted(dir)
	return nil
}

func frontMatterDraft(content string) bool {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return false
	}
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "---" {
			break
		}
		value, ok := strings.CutPrefix(line, "draft:")
		if !ok {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		switch strings.ToLower(value) {
		case "true", "yes", "on":
			return true
		}
		return false
	}
	return false
}

func runUpdateDate(ui *presenter, title, slugArg, dateArg, rootArg string) error {
	postSlug := strings.TrimSpace(slugArg)
	if postSlug == "" {
		postSlug = slugify(title)
	}
	if postSlug == "" {
		return errors.New("-update-date needs -slug, or a title to derive it from")
	}
	if strings.TrimSpace(dateArg) == "" {
		return errors.New("-update-date needs -date with the new publish date")
	}

	date, err := parseDate(dateArg, time.Now())
	if err != nil {
		return err
	}

	root, err := resolveRoot(rootArg)
	if err != nil {
		return err
	}

	oldDir, err := findPostDir(root, postSlug)
	if err != nil {
		return err
	}

	files, err := readPostFiles(oldDir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("%s has no index.md or index.pt.md", oldDir)
	}

	updatedFiles := make([]pendingUpdate, 0, len(files))
	for _, file := range files {
		content, err := withFrontMatterDate(file.content, date)
		if err != nil {
			return fmt.Errorf("%s: %w", file.path, err)
		}
		updatedFiles = append(updatedFiles, pendingUpdate{name: filepath.Base(file.path), content: content})
	}

	newDir := filepath.Join(root, date.Format("2006"), date.Format("01"), postSlug)
	if newDir != oldDir {
		if _, err := os.Stat(newDir); err == nil {
			return fmt.Errorf("%s already exists", newDir)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(newDir), 0o755); err != nil {
			return err
		}
		if err := os.Rename(oldDir, newDir); err != nil {
			return err
		}
	}

	updated := make([]string, 0, len(updatedFiles))
	for _, file := range updatedFiles {
		path := filepath.Join(newDir, file.name)
		if err := os.WriteFile(path, []byte(file.content), 0o644); err != nil {
			return err
		}
		updated = append(updated, path)
	}

	sections, err := ensureSectionIndexes(root, date)
	if err != nil {
		return err
	}

	var urls []string
	for _, file := range files {
		switch filepath.Base(file.path) {
		case "index.md":
			urls = append(urls, fmt.Sprintf("/posts/%s/%s/%s/", date.Format("2006"), date.Format("01"), postSlug))
		case "index.pt.md":
			urls = append(urls, fmt.Sprintf("/pt/posts/%s/%s/%s/", date.Format("2006"), date.Format("01"), postSlug))
		}
	}

	ui.header(postSlug, date, "")
	if newDir != oldDir {
		ui.moved(oldDir, newDir)
	}
	ui.updated(updated)
	ui.created(sections)
	ui.links(urls)
	return nil
}

func findPostDir(root, slug string) (string, error) {
	years, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}

	var found []string
	for _, year := range years {
		if !year.IsDir() || !digits(year.Name(), 4) {
			continue
		}
		months, err := os.ReadDir(filepath.Join(root, year.Name()))
		if err != nil {
			continue
		}
		for _, month := range months {
			if !month.IsDir() || !digits(month.Name(), 2) {
				continue
			}
			candidate := filepath.Join(root, year.Name(), month.Name(), slug)
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				found = append(found, candidate)
			}
		}
	}

	switch len(found) {
	case 0:
		return "", fmt.Errorf("post %q not found under %s", slug, root)
	case 1:
		return found[0], nil
	default:
		return "", fmt.Errorf("post %q is ambiguous: %s", slug, strings.Join(found, ", "))
	}
}

func digits(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func readPostFiles(dir string) ([]target, error) {
	var files []target
	for _, name := range []string{"index.md", "index.pt.md"} {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		files = append(files, target{path: path, content: string(data)})
	}
	return files, nil
}

func withFrontMatterDate(content string, date time.Time) (string, error) {
	lines := strings.SplitAfter(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", errors.New("missing YAML front matter")
	}

	newline := "\n"
	if strings.HasSuffix(lines[0], "\r\n") {
		newline = "\r\n"
	}

	closing := -1
	replaced := false
	for i := 1; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r\n")
		if line == "---" {
			closing = i
			break
		}
		if !replaced && strings.HasPrefix(line, "date:") {
			lines[i] = "date: '" + formatFrontMatterDate(date) + "'" + newline
			replaced = true
		}
	}
	if closing == -1 {
		return "", errors.New("unterminated YAML front matter")
	}
	if !replaced {
		insert := "date: '" + formatFrontMatterDate(date) + "'" + newline
		lines = append(lines[:closing], append([]string{insert}, lines[closing:]...)...)
	}
	return strings.Join(lines, ""), nil
}

func filenameForLang(lang string) string {
	if lang == "pt" {
		return "index.pt.md"
	}
	return "index.md"
}

func renderPost(title, author string, date time.Time, slug string, categories, tags []string, draft bool) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %s\n", yamlString(title))
	fmt.Fprintf(&b, "author: %s\n", yamlString(author))
	fmt.Fprintf(&b, "date: '%s'\n", formatFrontMatterDate(date))
	fmt.Fprintf(&b, "slug: %s\n", slug)
	fmt.Fprintf(&b, "categories: %s\n", yamlList(categories))
	fmt.Fprintf(&b, "tags: %s\n", yamlList(tags))
	if draft {
		b.WriteString("draft: true\n")
	}
	b.WriteString("---\n")
	return b.String()
}

func parseDate(value string, now time.Time) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "now") {
		return now.Truncate(time.Second), nil
	}

	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	for _, layout := range []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	} {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t, nil
		}
	}
	if t, err := time.ParseInLocation("2006-01-02", value, time.Local); err == nil {
		return time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.Local), nil
	}
	return time.Time{}, fmt.Errorf("invalid -date %q: use now, YYYY-MM-DD, YYYY-MM-DD HH:MM, or RFC 3339", value)
}

func formatFrontMatterDate(t time.Time) string {
	return t.Format("2006-01-02T15:04:05-07:00")
}

func slugify(s string) string {
	var b strings.Builder
	pendingHyphen := false
	for _, r := range strings.ToLower(s) {
		if r == '\'' || r == '\u2019' {
			continue
		}
		if mapped, ok := transliterations[r]; ok {
			if pendingHyphen && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteString(mapped)
			pendingHyphen = false
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if pendingHyphen && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			pendingHyphen = false
			continue
		}
		if b.Len() > 0 {
			pendingHyphen = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func splitList(value string) []string {
	var items []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func yamlString(s string) string {
	return strconv.Quote(s)
}

func yamlList(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = yamlString(item)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func resolveRoot(root string) (string, error) {
	if root != "" {
		info, err := os.Stat(root)
		if err != nil {
			return "", fmt.Errorf("-root %s: %w", root, err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("-root %s: not a directory", root)
		}
		return root, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, "content", "posts")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("content/posts not found; run from the repository or pass -root")
		}
		dir = parent
	}
}

func ensureSectionIndexes(root string, date time.Time) ([]string, error) {
	year := date.Format("2006")
	yearDir := filepath.Join(root, year)
	monthDir := filepath.Join(yearDir, date.Format("01"))
	if err := os.MkdirAll(monthDir, 0o755); err != nil {
		return nil, err
	}

	englishMonth := date.Format("January")
	portugueseMonth := portugueseMonths[date.Month()-1]
	capitalizedPT := strings.ToUpper(portugueseMonth[:1]) + portugueseMonth[1:]

	sections := []target{
		{filepath.Join(yearDir, "_index.md"),
			renderSection(year, fmt.Sprintf("Everything I've written in %s.", year))},
		{filepath.Join(yearDir, "_index.pt.md"),
			renderSection(year, fmt.Sprintf("Tudo o que eu escrevi em %s.", year))},
		{filepath.Join(monthDir, "_index.md"),
			renderSection(englishMonth+" "+year, fmt.Sprintf("Everything I've written in %s %s.", englishMonth, year))},
		{filepath.Join(monthDir, "_index.pt.md"),
			renderSection(capitalizedPT+" "+year, fmt.Sprintf("Tudo o que eu escrevi em %s de %s.", portugueseMonth, year))},
	}

	created := make([]string, 0, len(sections))
	for _, section := range sections {
		file, err := os.OpenFile(section.path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return created, err
		}
		if _, err := file.WriteString(section.content); err != nil {
			file.Close()
			return created, err
		}
		if err := file.Close(); err != nil {
			return created, err
		}
		created = append(created, section.path)
	}
	return created, nil
}

func renderSection(title, description string) string {
	return fmt.Sprintf("---\ntitle: %s\ndescription: %s\n---\n", yamlString(title), yamlString(description))
}
