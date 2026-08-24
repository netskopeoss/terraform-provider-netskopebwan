package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Every generated example opens with generatedMarker, and no documentation page
// shows it.
//
// The marker is in the example file because that is the only place the generator
// can both write and read: tfplugindocs embeds an example verbatim, so a marker in
// the file is one it finds again on the next run, and a file without one is
// somebody's own work to leave alone. Embedding verbatim is also what makes it a
// problem — it lands at the top of a code block meant to be pasted into a real
// configuration, where a note about how this repository works has no business
// being.
//
// So it is taken back out of the block, and nothing is put in its place. A page
// saying it was generated would be telling a practitioner nothing they can act on:
// the page is generated too, and says so in tfplugindocs' own comment. What is
// left is the configuration.
//
// Editing the pages afterwards rather than templating them leaves their layout
// upstream's: a copy of tfplugindocs' default templates here would be a fork to
// re-sync on every upgrade of the tool, to delete one line.
func runStrip(args []string) error {
	flags := flag.NewFlagSet("strip", flag.ExitOnError)
	in := flags.String("in", "", "directory of markdown tfplugindocs generated")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *in == "" {
		return errors.New("strip: -in is required")
	}

	var pages, stripped int

	err := filepath.WalkDir(*in, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		out, count, err := strip(string(content))
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}

		pages++
		stripped += count

		if out == string(content) {
			return nil
		}

		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if pages == 0 {
		return fmt.Errorf("no markdown pages under %s", *in)
	}

	fmt.Fprintf(os.Stderr, "tfdocs: strip: took the marker out of %d example(s) across %d page(s)\n", stripped, pages)

	return nil
}

// strip removes the marker from the examples embedded in one page, and reports how
// many it took out.
//
// A marker left behind is an error rather than something to live with: it means
// the line arrived in a shape this does not recognise, which is the notice landing
// on the registry instead. tfplugindocs owns how an example is embedded, so that
// is a thing an upgrade of it can change.
func strip(page string) (string, int, error) {
	lines := strings.Split(page, "\n")
	kept := make([]string, 0, len(lines))
	stripped := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == generatedMarker {
			stripped++

			continue
		}

		kept = append(kept, line)
	}

	out := strings.Join(kept, "\n")

	if strings.Contains(out, generatedMarker) {
		return "", 0, fmt.Errorf("the marker survived stripping, so tfplugindocs embeds an example some other way now")
	}

	return out, stripped, nil
}
