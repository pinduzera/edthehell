package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type presenter struct {
	out io.Writer
	gum bool
}

func newPresenter(out io.Writer) *presenter {
	return &presenter{out: out, gum: gumEnabled(out)}
}

func gumEnabled(out io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	if !isTerminal(out) {
		return false
	}
	_, err := exec.LookPath("gum")
	return err == nil
}

func isTerminal(out io.Writer) bool {
	f, ok := out.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func isTerminalReader(in io.Reader) bool {
	f, ok := in.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func (p *presenter) render(text string, args ...string) {
	cmd := exec.Command("gum", append([]string{"style"}, args...)...)
	cmd.Stdin = strings.NewReader(text)
	cmd.Stdout = p.out
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(p.out, text)
	}
}

func (p *presenter) header(title string, date time.Time, slug string) {
	if !p.gum {
		return
	}
	p.render(title,
		"--border", "rounded",
		"--border-foreground", "212",
		"--foreground", "212",
		"--bold",
		"--padding", "0 2",
		"--margin", "0 0 1 0",
	)
	meta := date.Format("2006-01-02")
	if slug != "" {
		meta += " · " + slug
	}
	p.render(meta,
		"--foreground", "240",
		"--margin", "0 0 1 2",
	)
}

func (p *presenter) created(paths []string) {
	if len(paths) == 0 {
		return
	}
	if !p.gum {
		for _, path := range paths {
			fmt.Fprintf(p.out, "created %s\n", path)
		}
		return
	}
	lines := make([]string, len(paths))
	for i, path := range paths {
		lines[i] = "✓ " + displayPath(path)
	}
	p.render(strings.Join(lines, "\n"),
		"--foreground", "114",
		"--margin", "0 0 1 2",
	)
}

func (p *presenter) updated(paths []string) {
	if len(paths) == 0 {
		return
	}
	if !p.gum {
		for _, path := range paths {
			fmt.Fprintf(p.out, "updated %s\n", path)
		}
		return
	}
	lines := make([]string, len(paths))
	for i, path := range paths {
		lines[i] = "↻ " + displayPath(path)
	}
	p.render(strings.Join(lines, "\n"),
		"--foreground", "114",
		"--margin", "0 0 1 2",
	)
}

func (p *presenter) moved(from, to string) {
	if !p.gum {
		fmt.Fprintf(p.out, "moved %s -> %s\n", from, to)
		return
	}
	p.render("→ "+displayPath(from)+" → "+displayPath(to),
		"--foreground", "240",
		"--margin", "0 0 1 2",
	)
}

func (p *presenter) links(urls []string) {
	if len(urls) == 0 {
		return
	}
	if !p.gum {
		for _, url := range urls {
			fmt.Fprintf(p.out, "url: %s\n", url)
		}
		return
	}
	lines := make([]string, len(urls))
	for i, url := range urls {
		lines[i] = "→ " + url
	}
	p.render(strings.Join(lines, "\n"),
		"--foreground", "81",
		"--underline",
		"--margin", "0 0 0 2",
	)
}

func (p *presenter) confirm(prompt string) (bool, error) {
	if p.gum && isTerminalReader(stdin) {
		cmd := exec.Command("gum", "confirm", "--default=false", prompt)
		cmd.Stdin = stdin
		cmd.Stdout = p.out
		cmd.Stderr = p.out
		err := cmd.Run()
		if err == nil {
			return true, nil
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}

	fmt.Fprintf(p.out, "%s [y/N] ", prompt)
	line, err := bufio.NewReader(stdin).ReadString('\n')
	if errors.Is(err, io.EOF) && strings.TrimSpace(line) == "" {
		return false, errors.New("no confirmation available; pass -force to delete a non-draft post")
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

func (p *presenter) deleted(path string) {
	if !p.gum {
		fmt.Fprintf(p.out, "deleted %s\n", path)
		return
	}
	p.render("✗ "+displayPath(path),
		"--foreground", "203",
		"--margin", "0 0 1 2",
	)
}

func (p *presenter) fail(err error) {
	if !p.gum {
		fmt.Fprintln(p.out, "post:", err)
		return
	}
	p.render("✗ "+err.Error(),
		"--foreground", "203",
		"--bold",
		"--margin", "1 0 0 0",
	)
}

func displayPath(path string) string {
	if !filepath.IsAbs(path) {
		return path
	}
	wd, err := os.Getwd()
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(wd, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}
	return rel
}
