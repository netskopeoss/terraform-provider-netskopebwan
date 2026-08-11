// Package genresource implements Terraform resources and data sources directly
// from the schemas generated out of the BWAN OpenAPI document.
//
// Everything a resource does is derived from its generated schema plus the API
// operations behind it, so adding an object to the provider means adding it to
// generator_config.yml — no per-resource Go code.
package genresource

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/bwanclient"
)

// idAttribute names the attribute holding the identity the API assigns.
const idAttribute = "id"

// Operation is one API call: a method and a path template whose "{name}"
// placeholders are filled from attributes of the same name.
type Operation struct {
	Method string
	Path   string
}

// Supported reports whether the API offers the operation at all.
func (o Operation) Supported() bool {
	return o.Path != "" && o.Method != ""
}

// placeholders lists the "{name}" segments of a path template, in order.
func placeholders(template string) []string {
	var out []string

	for {
		open := strings.Index(template, "{")
		if open < 0 {
			return out
		}

		end := strings.Index(template[open:], "}")
		if end < 0 {
			return out
		}

		end += open

		out = append(out, template[open+1:end])
		template = template[end+1:]
	}
}

// parentPlaceholders lists the path placeholders identifying the object a
// resource lives under, which is every placeholder except its own id.
func parentPlaceholders(templates ...string) []string {
	var out []string

	for _, template := range templates {
		for _, name := range placeholders(template) {
			if name != idAttribute && !slices.Contains(out, name) {
				out = append(out, name)
			}
		}
	}

	return out
}

// resolvePath fills a path template from the string attributes of value.
func resolvePath(template string, value tftypes.Value) (string, error) {
	names := placeholders(template)
	if len(names) == 0 {
		return template, nil
	}

	members, err := stringMembers(value, names)
	if err != nil {
		return "", err
	}

	out := template

	for _, name := range names {
		out = strings.ReplaceAll(out, "{"+name+"}", bwanclient.EscapePathSegment(members[name]))
	}

	return out, nil
}

// stringMembers reads the named top-level attributes of an object value as
// strings.
func stringMembers(value tftypes.Value, names []string) (map[string]string, error) {
	if value.IsNull() || !value.IsKnown() {
		return nil, fmt.Errorf("no value to read %s from", strings.Join(names, ", "))
	}

	var members map[string]tftypes.Value
	if err := value.As(&members); err != nil {
		return nil, err
	}

	out := make(map[string]string, len(names))

	for _, name := range names {
		member, ok := members[name]
		if !ok {
			return nil, fmt.Errorf("%s is not an attribute of this resource", name)
		}

		if member.IsNull() || !member.IsKnown() {
			return nil, fmt.Errorf("%s is not set", name)
		}

		var text string
		if err := member.As(&text); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}

		out[name] = text
	}

	return out, nil
}

// skipSet lists the attributes that must not appear in a request body or query
// string because they are already part of the path.
func skipSet(names []string) map[string]bool {
	out := make(map[string]bool, len(names))

	for _, name := range names {
		out[name] = true
	}

	return out
}
