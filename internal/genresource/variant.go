package genresource

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Variant identifies one kind of object on an endpoint that serves several.
//
// The API creates, reads and lists all four kinds of tag through /overlay-tags,
// telling them apart by a `config.type` field. Each kind is its own Terraform
// type, so each one has to recognise its own objects and ignore the rest:
// without that, listing wanlink tags would return overlay tags with every
// wanlink field null, and reading one by id would quietly adopt an object of the
// wrong kind.
type Variant struct {
	// Name is the branch as the spec names it, e.g. "wanlink".
	Name string
	// Discriminator is the field whose value selects the kind, where the API
	// declares one, dotted where the field is nested: a tag's kind is at
	// `config.type`.
	Discriminator string
	// Value is what Discriminator holds for this kind.
	Value string
	// Match lists the fields only this kind has, used when the API declares no
	// discriminator.
	Match []string
	// Wrapper is the field the API nests this kind's own fields under, which the
	// schema does not have: a tag's fields sit beside its name rather than inside a
	// `config`, because the Terraform type already says which kind it is. Empty
	// where the API nests nothing.
	Wrapper string
	// Wrapped names the fields that belong under Wrapper, which is what it takes to
	// build a request: hoisted into the object, they look like any other field of it.
	Wrapped []string
}

// Matches reports whether document is the kind of object v stands for. A nil
// Variant matches everything, which is what an endpoint serving one kind needs.
func (v *Variant) Matches(document any) bool {
	if v == nil {
		return true
	}

	object, ok := document.(map[string]any)
	if !ok {
		return false
	}

	if v.Discriminator != "" {
		return discriminant(object, v.Discriminator) == v.Value
	}

	for _, name := range v.Match {
		if value, ok := object[name]; ok && value != nil {
			return true
		}
	}

	return false
}

// Keep narrows a collection to the elements of this kind.
func (v *Variant) Keep(elements []any) []any {
	if v == nil {
		return elements
	}

	out := make([]any, 0, len(elements))

	for _, element := range elements {
		if v.Matches(element) {
			out = append(out, element)
		}
	}

	return out
}

// Flatten reshapes a document the API sent into the shape the schema declares:
// whatever Wrapper held comes up to the object itself, and the wrapper goes with the
// discriminator inside it.
//
// Everything in the wrapper is lifted, not only the fields Wrapped names. A field
// the API grew since the spec was last read has nowhere to go either way — Apply
// only looks for what the schema declares — and lifting it keeps this from being one
// more place that has to be regenerated to stay correct.
func (v *Variant) Flatten(document any) any {
	object, ok := document.(map[string]any)
	if v == nil || v.Wrapper == "" || !ok {
		return document
	}

	out := make(map[string]any, len(object))
	maps.Copy(out, object)

	wrapped, ok := out[v.Wrapper].(map[string]any)
	if !ok {
		// The API sent no wrapper, so this kind contributed nothing of its own.
		// Deleting the key regardless would turn "the API said nothing" into "the
		// API said null", which AfterRead reads as the field having been cleared.
		return out
	}

	delete(out, v.Wrapper)

	for name, value := range wrapped {
		if name == v.wrappedDiscriminator() {
			continue
		}

		out[name] = value
	}

	return out
}

// Nest is the inverse, for a request body: the fields Wrapped names go back under
// Wrapper, and the discriminator the schema does not carry is written beside them.
//
// The wrapper is written even when nothing goes into it. A topology tag is
// `{"config": {"type": "topology"}}` and nothing more, and the API requires that
// much; the schema, having neither the wrapper nor the discriminator, cannot say it.
func (v *Variant) Nest(body map[string]any) map[string]any {
	if v == nil || v.Wrapper == "" {
		return body
	}

	out := make(map[string]any, len(body)+1)
	wrapped := map[string]any{v.wrappedDiscriminator(): v.Value}

	for name, value := range body {
		if slices.Contains(v.Wrapped, name) {
			wrapped[name] = value

			continue
		}

		out[name] = value
	}

	out[v.Wrapper] = wrapped

	return out
}

// wrappedDiscriminator names the discriminator within the wrapper. A wrapper is only
// hoisted where the discriminator is one of the fields inside it, so what is left of
// its path once the wrapper is trimmed off is the field itself.
func (v *Variant) wrappedDiscriminator() string {
	return strings.TrimPrefix(v.Discriminator, v.Wrapper+".")
}

// FlattenAll reshapes every element of a collection, which is how a collection is
// recognised too.
func (v *Variant) FlattenAll(elements []any) []any {
	if v == nil || v.Wrapper == "" {
		return elements
	}

	out := make([]any, 0, len(elements))

	for _, element := range elements {
		out = append(out, v.Flatten(element))
	}

	return out
}

// Mismatch explains that an object exists but is not of this kind, naming what it
// is instead where the API says so. A practitioner hitting this has usually given
// one Terraform type the id of an object belonging to another.
func (v *Variant) Mismatch(typeName string, document any) string {
	if v.Discriminator == "" {
		return fmt.Sprintf(
			"The object exists but is not a %s: %s manages the %q kind, recognised by %v, and this one carries none of those fields.",
			typeName, typeName, v.Name, v.Match)
	}

	object, _ := document.(map[string]any)

	found := discriminant(object, v.Discriminator)
	if found == "" {
		found = "unset"
	}

	return fmt.Sprintf(
		"The object exists but its %s is %q, and %s manages the %q kind. Use the Terraform type for %q instead.",
		v.Discriminator, found, typeName, v.Value, found)
}

// discriminant reads the field a dotted path names, walking the objects on the
// way to it. Anything missing or of another type reads as no value at all, which
// is what an object of a kind this Variant does not manage looks like.
func discriminant(object map[string]any, path string) string {
	names := strings.Split(path, ".")

	for _, name := range names[:len(names)-1] {
		next, ok := object[name].(map[string]any)
		if !ok {
			return ""
		}

		object = next
	}

	value, _ := object[names[len(names)-1]].(string)

	return value
}
