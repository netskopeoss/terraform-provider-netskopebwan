package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
)

// page is one documentation page.
type page struct {
	// Path is the page's location under the output root, e.g.
	// "resources/link_monitor.html".
	Path string
	// Section groups pages in the sidebar: "Resources", "Data Sources".
	Section string
	// Title is what the page and the sidebar call it.
	Title  string
	Body   string
	Weight int
}

func runHTML(args []string) error {
	flags := flag.NewFlagSet("html", flag.ExitOnError)
	in := flags.String("in", "", "directory of markdown tfplugindocs generated")
	out := flags.String("out", "", "directory to write the HTML site to")
	title := flags.String("title", "bwan provider", "the site's title")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *in == "" || *out == "" {
		return errors.New("html: both -in and -out are required")
	}

	pages, err := readPages(*in)
	if err != nil {
		return err
	}

	if len(pages) == 0 {
		return fmt.Errorf("no markdown pages under %s", *in)
	}

	for _, current := range pages {
		target := filepath.Join(*out, filepath.FromSlash(current.Path))

		if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
			return fmt.Errorf("creating %s: %w", filepath.Dir(target), err)
		}

		if err := os.WriteFile(target, []byte(render(current, pages, *title)), 0o600); err != nil {
			return fmt.Errorf("writing %s: %w", target, err)
		}
	}

	return nil
}

// readPages renders every markdown file under root, in the order the sidebar
// should list them.
func readPages(root string) ([]page, error) {
	markdown := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		// The registry pages carry their own anchors as raw HTML, which is how their
		// cross references between nested schemas work.
		goldmark.WithRendererOptions(goldmarkhtml.WithUnsafe()),
	)

	// The input is usually a build output linked into the tree, and a walk does not
	// descend into a symlinked root.
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("resolving the input directory: %w", err)
	}

	var pages []page

	err = filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(name) != ".md" {
			return err
		}

		source, err := os.ReadFile(name)
		if err != nil {
			return fmt.Errorf("reading %s: %w", name, err)
		}

		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}

		front, body := splitFrontMatter(string(source))

		var buffer bytes.Buffer
		if err := markdown.Convert([]byte(body), &buffer); err != nil {
			return fmt.Errorf("rendering %s: %w", name, err)
		}

		slashed := filepath.ToSlash(relative)
		section, weight := sectionOf(slashed)

		pages = append(pages, page{
			Path:    strings.TrimSuffix(slashed, ".md") + ".html",
			Section: section,
			Title:   titleOf(front, slashed),
			Body:    linkToHTML(buffer.String()),
			Weight:  weight,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	slices.SortFunc(pages, func(a, b page) int {
		if a.Weight != b.Weight {
			return a.Weight - b.Weight
		}

		return strings.Compare(a.Title, b.Title)
	})

	return pages, nil
}

// sectionOf names the sidebar group a page belongs to, and how early that group
// appears. The provider's own page comes first; anything the generator adds later
// lands in a group of its own rather than being dropped.
func sectionOf(relative string) (string, int) {
	directory := path.Dir(relative)

	switch directory {
	case ".":
		return "Provider", 0
	case "resources":
		return "Resources", 1
	case "data-sources":
		return "Data Sources", 2
	default:
		return titleCase(strings.ReplaceAll(directory, "-", " ")), 3
	}
}

func titleOf(front, relative string) string {
	if value := frontMatterValue(front, "page_title"); value != "" {
		// A registry page title reads "bwan_segment Resource - bwan"; the sidebar
		// already says which section it is in.
		if cut, _, ok := strings.Cut(value, " Resource - "); ok {
			return cut
		}

		if cut, _, ok := strings.Cut(value, " Data Source - "); ok {
			return cut
		}

		return value
	}

	return strings.TrimSuffix(path.Base(relative), ".md")
}

// splitFrontMatter separates the YAML block the generator puts at the top of every
// page from the body.
func splitFrontMatter(source string) (string, string) {
	const fence = "---"

	rest, ok := strings.CutPrefix(strings.TrimLeft(source, "\n"), fence+"\n")
	if !ok {
		return "", source
	}

	front, body, ok := strings.Cut(rest, "\n"+fence)
	if !ok {
		return "", source
	}

	return front, body
}

func frontMatterValue(front, key string) string {
	for line := range strings.SplitSeq(front, "\n") {
		name, value, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(name) != key {
			continue
		}

		return strings.Trim(strings.TrimSpace(value), `"`)
	}

	return ""
}

// markdownLink matches a link to another page, which has to point at the rendered
// file rather than the source.
var markdownLink = regexp.MustCompile(`(href=")([^"]+)\.md(#[^"]*)?(")`)

func linkToHTML(body string) string {
	return markdownLink.ReplaceAllString(body, "${1}${2}.html${3}${4}")
}

func titleCase(value string) string {
	parts := strings.Split(value, " ")

	for i, part := range parts {
		if part != "" {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}

	return strings.Join(parts, " ")
}
