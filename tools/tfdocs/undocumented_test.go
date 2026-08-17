package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/genresource"
)

// TestUndocumentedAttributesReachNoPage is the test the whole arrangement rests
// on. The attribute is in the schema the provider serves — it has to be, or
// Terraform would reject a configuration using it — so this command is the only
// thing standing between it and the registry. Everything published comes out of
// these two subcommands: the schema tfplugindocs renders the pages from, and the
// examples those pages embed.
func TestUndocumentedAttributesReachNoPage(t *testing.T) {
	require.NotEmpty(t, undocumented)

	directory := t.TempDir()
	schema := filepath.Join(directory, "schema.json")

	require.NoError(t, run([]string{"schema", "-out", schema}))

	encoded, err := os.ReadFile(schema)
	require.NoError(t, err)

	for name := range undocumented {
		require.NotContains(t, string(encoded), name,
			"%s is served by the provider but must not be documented", name)
	}

	examples := filepath.Join(directory, "examples")
	require.NoError(t, run([]string{"examples", "-out", examples}))

	var written int

	require.NoError(t, filepath.Walk(examples, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		written++

		for name := range undocumented {
			require.NotContains(t, string(content), name, path)
		}

		return nil
	}))

	require.NotZero(t, written, "the examples subcommand wrote nothing to check")
}

// TestTheSchemaIsOtherwiseWhole guards against the filter taking too much with
// it: the pages are rendered from this file, so an attribute dropped by mistake
// is an attribute nobody can read about.
func TestTheSchemaIsOtherwiseWhole(t *testing.T) {
	directory := t.TempDir()
	schema := filepath.Join(directory, "schema.json")

	require.NoError(t, run([]string{"schema", "-out", schema}))

	encoded, err := os.ReadFile(schema)
	require.NoError(t, err)

	document := string(encoded)

	for _, name := range []string{"netskopebwan_segment", "endpoint", "\"id\"", "enable_pre_release"} {
		require.True(t, strings.Contains(document, name), "%s is missing from the generated schema", name)
	}
}

func TestDocumentedIsTheOnlyFilter(t *testing.T) {
	require.False(t, documented(genresource.OperatingTenantAttribute))
	require.True(t, documented("id"))
	require.True(t, documented("name"))
}
