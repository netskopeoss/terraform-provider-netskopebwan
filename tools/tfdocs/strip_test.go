package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// resourcePage is a page as tfplugindocs renders one, with the marker still in
// each block it embedded an example into.
const resourcePage = `---
page_title: "netskopebwan_thing Resource - netskopebwan"
---

# netskopebwan_thing (Resource)

## Example Usage

` + "```terraform" + `
` + generatedMarker + `
resource "netskopebwan_thing" "example" {
  name = "example"
}
` + "```" + `

## Schema

### Required

- ` + "`name`" + ` (String)

## Import

Import is supported using the following syntax:

` + "```shell" + `
` + generatedMarker + `
terraform import netskopebwan_thing.example <id>
` + "```" + `
`

func TestStripTakesTheMarkerOutOfEveryBlock(t *testing.T) {
	out, stripped, err := strip(resourcePage)

	require.NoError(t, err)
	require.Equal(t, 2, stripped, "a resource page embeds a usage example and an import command")
	require.NotContains(t, out, generatedMarker)

	// Each block now opens on the configuration, and nothing else about the page
	// moved.
	require.Contains(t, out, "```terraform\nresource \"netskopebwan_thing\" \"example\" {")
	require.Contains(t, out, "```shell\nterraform import netskopebwan_thing.example <id>")
	require.Contains(t, out, "## Schema")
	require.Contains(t, out, "- `name` (String)")
}

// TestStripIsIdempotent matters because docs/ is committed and compared against
// git: a step that changed a page on every run would fail docs-check.
func TestStripIsIdempotent(t *testing.T) {
	once, _, err := strip(resourcePage)
	require.NoError(t, err)

	twice, stripped, err := strip(once)
	require.NoError(t, err)

	require.Equal(t, once, twice)
	require.Zero(t, stripped, "there is nothing left to take out")
}

// TestStripLeavesAHandWrittenExampleAlone covers the escape hatch from the page's
// side: an example nobody generated carries no marker, so there is nothing to take
// out of the block it was embedded into.
func TestStripLeavesAHandWrittenExampleAlone(t *testing.T) {
	page := `# netskopebwan Provider

## Example Usage

` + "```terraform" + `
provider "netskopebwan" {}
` + "```" + `
`

	out, stripped, err := strip(page)

	require.NoError(t, err)
	require.Zero(t, stripped)
	require.Equal(t, page, out)
}

// TestStripReportsAMarkerItCannotTakeOut covers tfplugindocs embedding an example
// some other way — indented into a list, say. The marker would reach the registry
// at the top of a code block, which is the whole thing this step exists to prevent,
// so it fails instead.
func TestStripReportsAMarkerItCannotTakeOut(t *testing.T) {
	page := "## Example Usage\n\n```terraform\n  " + generatedMarker + " and something after it\n```\n"

	_, _, err := strip(page)

	require.ErrorContains(t, err, "the marker survived stripping")
}

// TestNoCommittedPageShowsTheMarker is the check on the committed output: docs/ is
// what the registry serves, so the marker being in one of those pages is the
// notice being published.
func TestNoCommittedPageShowsTheMarker(t *testing.T) {
	pages, err := filepath.Glob(filepath.Join("..", "..", "docs", "*", "*.md"))
	require.NoError(t, err)
	require.NotEmpty(t, pages)

	for _, page := range pages {
		content, err := os.ReadFile(page)
		require.NoError(t, err)

		require.NotContains(t, string(content), generatedMarker, "%s; run make docs", page)

		// And nothing put a notice back in its place: the page already says it was
		// generated, in tfplugindocs' own comment.
		require.NotContains(t, string(content), "Generated from the provider's schema", page)
	}

	// The examples themselves still carry it, which is what makes them the
	// generator's own.
	example, err := os.ReadFile(filepath.Join("..", "..", "examples", "resources",
		"netskopebwan_address_group", "resource.tf"))
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(example), generatedMarker))
}
