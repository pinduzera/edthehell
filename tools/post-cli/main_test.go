package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"I still love the Internet": "i-still-love-the-internet",
		"Olá, Mundo!":               "ola-mundo",
		"Don't  Stop":               "dont-stop",
		"  --spaces--  ":            "spaces",
		"Café & Pão":                "cafe-pao",
		"2026":                      "2026",
		"":                          "",
	}
	for input, want := range tests {
		if got := slugify(input); got != want {
			t.Errorf("slugify(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestParseDate(t *testing.T) {
	now := time.Date(2026, 11, 21, 9, 0, 0, 0, time.Local)

	tests := []struct {
		in   string
		want string
	}{
		{"", "2026-11-21T09:00:00"},
		{"now", "2026-11-21T09:00:00"},
		{"NOW", "2026-11-21T09:00:00"},
		{"2026-11-17", "2026-11-17T12:00:00"},
		{"2026-11-17 09:30", "2026-11-17T09:30:00"},
		{"2026-11-17 09:30:15", "2026-11-17T09:30:15"},
		{"2026-11-17T09:30", "2026-11-17T09:30:00"},
	}
	for _, tt := range tests {
		got, err := parseDate(tt.in, now)
		if err != nil {
			t.Fatalf("parseDate(%q): %v", tt.in, err)
		}
		if formatted := got.Format("2006-01-02T15:04:05"); formatted != tt.want {
			t.Errorf("parseDate(%q) = %s, want %s", tt.in, formatted, tt.want)
		}
	}

	if _, err := parseDate("17/11/2026", now); err == nil {
		t.Error("expected an error for an unsupported date format")
	}
}

func TestRunCreatesEnglishPost(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer

	err := run([]string{
		"-root", root,
		"-date", "2026-11-21 10:30",
		"-tags", "go, cli",
		"-categories", "notes",
		"-slug", "my-post",
		"My Post",
	}, &out)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	path := filepath.Join(root, "2026", "11", "my-post", "index.md")
	content := readFile(t, path)

	for _, want := range []string{
		`title: "My Post"`,
		`author: "Ed The Hell"`,
		`date: '2026-11-21T10:30:00`,
		"slug: my-post",
		`categories: ["notes"]`,
		`tags: ["go", "cli"]`,
	} {
		if !strings.Contains(content, want) {
			t.Errorf("index.md missing %q:\n%s", want, content)
		}
	}
	if strings.Contains(content, "draft:") {
		t.Error("index.md should not be a draft by default")
	}

	if _, err := os.Stat(filepath.Join(root, "2026", "11", "my-post", "index.pt.md")); !os.IsNotExist(err) {
		t.Error("index.pt.md should not exist without -both or -lang pt")
	}

	if !strings.Contains(out.String(), "url: /posts/2026/11/my-post/") {
		t.Errorf("missing English URL in output:\n%s", out.String())
	}

	for _, section := range []string{
		filepath.Join(root, "2026", "_index.md"),
		filepath.Join(root, "2026", "_index.pt.md"),
		filepath.Join(root, "2026", "11", "_index.md"),
		filepath.Join(root, "2026", "11", "_index.pt.md"),
	} {
		if _, err := os.Stat(section); err != nil {
			t.Errorf("missing section index %s: %v", section, err)
		}
	}
	if got := readFile(t, filepath.Join(root, "2026", "11", "_index.pt.md")); !strings.Contains(got, `"Novembro 2026"`) {
		t.Errorf("unexpected Portuguese month index:\n%s", got)
	}
}

func TestRunCreatesBothLanguages(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer

	err := run([]string{
		"-root", root,
		"-date", "2026-11-21",
		"-both",
		"-title-pt", "Meu Post",
		"-draft",
		"My Post",
	}, &out)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	english := readFile(t, filepath.Join(root, "2026", "11", "my-post", "index.md"))
	if !strings.Contains(english, `title: "My Post"`) {
		t.Errorf("unexpected English front matter:\n%s", english)
	}
	portuguese := readFile(t, filepath.Join(root, "2026", "11", "my-post", "index.pt.md"))
	if !strings.Contains(portuguese, `title: "Meu Post"`) {
		t.Errorf("unexpected Portuguese front matter:\n%s", portuguese)
	}
	for name, content := range map[string]string{"en": english, "pt": portuguese} {
		if !strings.Contains(content, "draft: true") {
			t.Errorf("%s post should be a draft:\n%s", name, content)
		}
	}

	if !strings.Contains(out.String(), "url: /posts/2026/11/my-post/") ||
		!strings.Contains(out.String(), "url: /pt/posts/2026/11/my-post/") {
		t.Errorf("missing URLs in output:\n%s", out.String())
	}
}

func TestRunPortugueseOnly(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer

	if err := run([]string{"-root", root, "-lang", "pt", "-date", "2027-01-05", "Ação"}, &out); err != nil {
		t.Fatalf("run: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "2027", "01", "acao", "index.pt.md")); err != nil {
		t.Errorf("missing Portuguese post: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "2027", "01", "acao", "index.md")); !os.IsNotExist(err) {
		t.Error("English post should not exist for -lang pt")
	}
}

func TestRunRefusesToOverwrite(t *testing.T) {
	root := t.TempDir()
	args := []string{"-root", root, "-date", "2026-11-21", "My Post"}

	if err := run(args, &bytes.Buffer{}); err != nil {
		t.Fatalf("first run: %v", err)
	}
	err := run(args, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected already exists error, got %v", err)
	}

	if err := run(append(args, "-force"), &bytes.Buffer{}); err != nil {
		t.Fatalf("forced run: %v", err)
	}
}

func TestRunKeepsExistingSectionIndexes(t *testing.T) {
	root := t.TempDir()
	yearDir := filepath.Join(root, "2026")
	if err := os.MkdirAll(yearDir, 0o755); err != nil {
		t.Fatal(err)
	}
	custom := "---\ntitle: Custom\n---\n"
	if err := os.WriteFile(filepath.Join(yearDir, "_index.md"), []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := run([]string{"-root", root, "-date", "2026-11-21", "My Post"}, &bytes.Buffer{}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := readFile(t, filepath.Join(yearDir, "_index.md")); got != custom {
		t.Errorf("section index was overwritten: %q", got)
	}
}

func TestRunHelp(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"-h"}, &out); err != nil {
		t.Fatalf("run -h: %v", err)
	}
	if !strings.Contains(out.String(), "usage: post") {
		t.Errorf("missing usage output:\n%s", out.String())
	}
}

func TestRunRequiresTitle(t *testing.T) {
	root := t.TempDir()
	err := run([]string{"-root", root}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "title is required") {
		t.Fatalf("expected title error, got %v", err)
	}
}

func TestResolveRoot(t *testing.T) {
	root := t.TempDir()
	got, err := resolveRoot(root)
	if err != nil {
		t.Fatalf("resolveRoot: %v", err)
	}
	if got != root {
		t.Errorf("resolveRoot = %q, want %q", got, root)
	}

	if _, err := resolveRoot(filepath.Join(root, "missing")); err == nil {
		t.Error("expected an error for a missing -root")
	}
}

func TestPresenterPlainOutput(t *testing.T) {
	var out bytes.Buffer
	ui := newPresenter(&out)
	if ui.gum {
		t.Fatal("gum should be disabled for a non-terminal writer")
	}

	ui.header("My Post", time.Date(2026, 11, 21, 0, 0, 0, 0, time.Local), "my-post")
	ui.created([]string{"content/posts/2026/11/my-post/index.md"})
	ui.links([]string{"/posts/2026/11/my-post/", "/pt/posts/2026/11/my-post/"})

	want := "created content/posts/2026/11/my-post/index.md\n" +
		"url: /posts/2026/11/my-post/\n" +
		"url: /pt/posts/2026/11/my-post/\n"
	if got := out.String(); got != want {
		t.Errorf("plain output = %q, want %q", got, want)
	}

	out.Reset()
	ui.fail(os.ErrNotExist)
	if got := out.String(); got != "post: "+os.ErrNotExist.Error()+"\n" {
		t.Errorf("plain error output = %q", got)
	}
}

func TestDisplayPath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	inside := filepath.Join(cwd, "content", "posts", "2026", "11", "my-post", "index.md")
	if got, want := displayPath(inside), filepath.Join("content", "posts", "2026", "11", "my-post", "index.md"); got != want {
		t.Errorf("displayPath(%q) = %q, want %q", inside, got, want)
	}
	if got, want := displayPath("relative/path.md"), "relative/path.md"; got != want {
		t.Errorf("displayPath(relative) = %q, want %q", got, want)
	}
	outside := filepath.Join(filepath.Dir(cwd), "elsewhere.md")
	if got := displayPath(outside); got != outside {
		t.Errorf("displayPath(outside) = %q, want %q", got, outside)
	}
}

func TestRunUpdatesPostDate(t *testing.T) {
	root := t.TempDir()
	if err := run([]string{
		"-root", root,
		"-date", "2026-11-21 10:30",
		"-both",
		"-title-pt", "Meu Post",
		"My Post",
	}, &bytes.Buffer{}); err != nil {
		t.Fatalf("create: %v", err)
	}

	var out bytes.Buffer
	if err := run([]string{
		"-root", root,
		"-update-date",
		"-slug", "my-post",
		"-date", "2026-12-05 09:15",
	}, &out); err != nil {
		t.Fatalf("update: %v", err)
	}

	oldDir := filepath.Join(root, "2026", "11", "my-post")
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Errorf("old post directory still exists: %v", err)
	}

	dir := filepath.Join(root, "2026", "12", "my-post")
	english := readFile(t, filepath.Join(dir, "index.md"))
	if !strings.Contains(english, "date: '2026-12-05T09:15:00") {
		t.Errorf("English date was not updated:\n%s", english)
	}
	if !strings.Contains(english, `title: "My Post"`) {
		t.Errorf("English title changed:\n%s", english)
	}
	portuguese := readFile(t, filepath.Join(dir, "index.pt.md"))
	if !strings.Contains(portuguese, "date: '2026-12-05T09:15:00") {
		t.Errorf("Portuguese date was not updated:\n%s", portuguese)
	}
	if !strings.Contains(portuguese, `title: "Meu Post"`) {
		t.Errorf("Portuguese title changed:\n%s", portuguese)
	}

	for _, want := range []string{
		"moved ",
		"updated ",
		"url: /posts/2026/12/my-post/",
		"url: /pt/posts/2026/12/my-post/",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in output:\n%s", want, out.String())
		}
	}

	for _, section := range []string{
		filepath.Join(root, "2026", "12", "_index.md"),
		filepath.Join(root, "2026", "12", "_index.pt.md"),
	} {
		if _, err := os.Stat(section); err != nil {
			t.Errorf("missing section index %s: %v", section, err)
		}
	}
}

func TestRunUpdatesPostTimeWithoutMoving(t *testing.T) {
	root := t.TempDir()
	if err := run([]string{"-root", root, "-date", "2026-11-21 10:30", "My Post"}, &bytes.Buffer{}); err != nil {
		t.Fatalf("create: %v", err)
	}

	var out bytes.Buffer
	if err := run([]string{
		"-root", root,
		"-update-date",
		"-slug", "my-post",
		"-date", "2026-11-21 18:45",
	}, &out); err != nil {
		t.Fatalf("update: %v", err)
	}

	content := readFile(t, filepath.Join(root, "2026", "11", "my-post", "index.md"))
	if !strings.Contains(content, "date: '2026-11-21T18:45:00") {
		t.Errorf("date was not updated:\n%s", content)
	}
	if strings.Contains(out.String(), "moved ") {
		t.Errorf("post should not have moved:\n%s", out.String())
	}
}

func TestRunUpdateDateErrors(t *testing.T) {
	root := t.TempDir()

	err := run([]string{"-root", root, "-update-date", "-slug", "my-post"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "-date") {
		t.Errorf("expected a -date error, got %v", err)
	}

	err = run([]string{"-root", root, "-update-date", "-slug", "missing", "-date", "2026-12-01"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected a not-found error, got %v", err)
	}
}

func TestRunDeletesDraftWithoutConfirmation(t *testing.T) {
	root := t.TempDir()
	if err := run([]string{"-root", root, "-date", "2026-11-21", "-draft", "My Post"}, &bytes.Buffer{}); err != nil {
		t.Fatalf("create: %v", err)
	}

	original := stdin
	stdin = strings.NewReader("n\n")
	t.Cleanup(func() { stdin = original })

	var out bytes.Buffer
	if err := run([]string{"-root", root, "-delete", "-slug", "my-post"}, &out); err != nil {
		t.Fatalf("delete draft: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "2026", "11", "my-post")); !os.IsNotExist(err) {
		t.Errorf("draft post was not deleted: %v", err)
	}
	if !strings.Contains(out.String(), "deleted ") {
		t.Errorf("missing delete output:\n%s", out.String())
	}
}

func TestRunDeleteNonDraftAsksForConfirmation(t *testing.T) {
	root := t.TempDir()
	if err := run([]string{"-root", root, "-date", "2026-11-21", "My Post"}, &bytes.Buffer{}); err != nil {
		t.Fatalf("create: %v", err)
	}
	dir := filepath.Join(root, "2026", "11", "my-post")

	original := stdin
	t.Cleanup(func() { stdin = original })

	stdin = strings.NewReader("n\n")
	var out bytes.Buffer
	if err := run([]string{"-root", root, "-delete", "-slug", "my-post"}, &out); err != nil {
		t.Fatalf("declined delete: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("post should still exist after declining: %v", err)
	}
	if !strings.Contains(out.String(), "cancelled") {
		t.Errorf("missing cancellation message:\n%s", out.String())
	}

	stdin = strings.NewReader("y\n")
	out.Reset()
	if err := run([]string{"-root", root, "-delete", "-slug", "my-post"}, &out); err != nil {
		t.Fatalf("accepted delete: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("post was not deleted after confirming: %v", err)
	}
}

func TestRunDeleteNonDraftRefusesWithoutInput(t *testing.T) {
	root := t.TempDir()
	if err := run([]string{"-root", root, "-date", "2026-11-21", "My Post"}, &bytes.Buffer{}); err != nil {
		t.Fatalf("create: %v", err)
	}
	dir := filepath.Join(root, "2026", "11", "my-post")

	original := stdin
	t.Cleanup(func() { stdin = original })
	stdin = strings.NewReader("")

	err := run([]string{"-root", root, "-delete", "-slug", "my-post"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "-force") {
		t.Fatalf("expected a -force hint, got %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("post should still exist: %v", err)
	}

	if err := run([]string{"-root", root, "-delete", "-slug", "my-post", "-force"}, &bytes.Buffer{}); err != nil {
		t.Fatalf("forced delete: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("post was not deleted with -force: %v", err)
	}
}

func TestFrontMatterDraft(t *testing.T) {
	tests := map[string]bool{
		"---\ndraft: true\n---\n":            true,
		"---\ndraft: yes\n---\n":             true,
		"---\ndraft: \"true\"\n---\n":        true,
		"---\ndraft: false\n---\n":           false,
		"---\ntitle: \"X\"\n---\n":           false,
		"---\ntitle: \"draft: true\"\n---\n": false,
		"no front matter\n":                  false,
	}
	for content, want := range tests {
		if got := frontMatterDraft(content); got != want {
			t.Errorf("frontMatterDraft(%q) = %v, want %v", content, got, want)
		}
	}
}

func TestWithFrontMatterDate(t *testing.T) {
	date := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	formatted := "date: '" + formatFrontMatterDate(date) + "'\n"

	got, err := withFrontMatterDate("---\ntitle: \"X\"\ndraft: true\n---\nbody\n", date)
	if err != nil {
		t.Fatalf("withFrontMatterDate: %v", err)
	}
	want := "---\ntitle: \"X\"\ndraft: true\n" + formatted + "---\nbody\n"
	if got != want {
		t.Errorf("inserted date:\ngot  %q\nwant %q", got, want)
	}

	got, err = withFrontMatterDate("---\ntitle: \"X\"\ndate: '2020-01-01T00:00:00-03:00'\n---\n", date)
	if err != nil {
		t.Fatalf("withFrontMatterDate: %v", err)
	}
	if !strings.Contains(got, formatted) || strings.Contains(got, "2020-01-01") {
		t.Errorf("replaced date incorrectly:\n%s", got)
	}

	if _, err := withFrontMatterDate("no front matter\n", date); err == nil {
		t.Error("expected an error for content without front matter")
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
