package genresource

import "fmt"

// Variant identifies one kind of object on an endpoint that serves several.
//
// The API creates, reads and lists all four kinds of tag through /tags, telling
// them apart by a `type` field. Each kind is its own Terraform type, so each one
// has to recognise its own objects and ignore the rest: without that, listing
// wanlink tags would return overlay tags with every wanlink field null, and
// reading one by id would quietly adopt an object of the wrong kind.
type Variant struct {
	// Name is the branch as the spec names it, e.g. "wanlink".
	Name string
	// Discriminator is the field whose value selects the kind, where the API
	// declares one.
	Discriminator string
	// Value is what Discriminator holds for this kind.
	Value string
	// Match lists the fields only this kind has, used when the API declares no
	// discriminator.
	Match []string
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
		value, _ := object[v.Discriminator].(string)

		return value == v.Value
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
	found, _ := object[v.Discriminator].(string)

	if found == "" {
		found = "unset"
	}

	return fmt.Sprintf(
		"The object exists but its %s is %q, and %s manages the %q kind. Use the Terraform type for %q instead.",
		v.Discriminator, found, typeName, v.Value, found)
}
