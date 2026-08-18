package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode"
)

// VariantSeparator joins an API path to the variant of its body schema in the
// synthetic paths prep emits. The generator maps a schema per path, so a schema
// that branches needs one path per branch for the branches to reach Terraform as
// separate objects.
const VariantSeparator = "@"

// VariantExtension carries what the runtime needs to tell one variant of an
// object from another: which property selects it, or failing that which
// properties only it declares.
const VariantExtension = "x-terraform-variant"

// VariantBlockDescriptionSuffix is appended to the description of every branch a
// composed schema is split into. It is how the registry recognises the blocks
// again in the generated code specification, the same way an untyped schema is
// recognised by RawJSONDescriptionSuffix.
const VariantBlockDescriptionSuffix = "One of several forms of its parent; set exactly one of them."

// variant is one branch of a composed schema.
type variant struct {
	// Name identifies the branch in generator_config.yml and in the synthetic
	// path prep emits for it.
	Name string
	// Discriminator is the property whose value selects this variant, taken from
	// the spec's own discriminator. Empty when the spec declares none.
	Discriminator string
	// Path is the property path from the object the runtime sees down to the
	// object Discriminator is declared on, empty when the two are the same. A
	// tag's kind is selected by `config.type`, so the branching schema is the
	// `config` a tag holds rather than the tag itself.
	Path []string
	// Value is what Discriminator holds for this variant.
	Value string
	// Match lists the properties only this variant declares. It is how a variant
	// is recognised when the spec has no discriminator: an object carrying `fqdn`
	// is the fqdn variant of a link monitor.
	Match []string
	// Wrapper is the property this variant was lifted out of, empty where it was
	// left where the API declared it. See hoistWrapper.
	Wrapper string
	// Wrapped names the properties that came up out of Wrapper, which is what it
	// takes to put them back: once hoisted they look like any other property of the
	// object holding them.
	Wrapped []string
	// Schema is the branch merged with whatever the composed schema declared
	// alongside the composition, so it stands alone as an object's whole schema.
	Schema map[string]any
	// Bare is the branch by itself, for the case where it becomes one block of the
	// composed schema rather than a schema of its own.
	Bare map[string]any
}

// variantsOf splits a composed schema into its branches, or returns nil when the
// schema does not branch into two or more objects.
//
// Every branch keeps the properties the composed schema declared next to the
// composition — a link monitor's id belongs to all three of its forms — so a
// branch is a complete schema rather than a fragment.
func (p *Prep) variantsOf(schema map[string]any, loc string) []variant {
	key := compositionKey(schema)
	if key == "" {
		return nil
	}

	raw, _ := schema[key].([]any)

	var (
		resolved []map[string]any
		refs     []string
	)

	for i, entry := range raw {
		target := p.resolve(entry, fmt.Sprintf("%s.%s[%d]", loc, key, i))
		if target == nil || isNullSchema(target) {
			continue
		}

		if _, isObject := target["properties"].(map[string]any); !isObject && target["type"] != "object" {
			continue
		}

		resolved = append(resolved, target)

		asMap, _ := entry.(map[string]any)
		refs = append(refs, stringOr(asMap["$ref"]))
	}

	if len(resolved) < 2 {
		return nil
	}

	discriminator, byValue := discriminatorOf(schema)
	names := variantNames(resolved, refs, byValue, discriminator)

	out := make([]variant, 0, len(resolved))

	for i, target := range resolved {
		// The mapping is the only place a branch's discriminator value is written
		// down when no branch declares the property, so the branch is completed
		// from it before anything is derived from the branch.
		branch := withDiscriminator(target, discriminator, names[i], byValue)

		out = append(out, variant{
			Name:          names[i],
			Discriminator: discriminator,
			Value:         discriminatorValue(branch, discriminator),
			Match:         distinctiveProperties(resolved, i),
			Schema:        p.variantSchema(schema, branch, key),
			Bare:          branch,
		})
	}

	return out
}

// withDiscriminator returns target with the discriminator property declared, so
// that the branch says which kind it is.
//
// OpenAPI lets a discriminator name a property no branch declares — the BWAN
// spec maps `type` to the four kinds of tag config without any of them having a
// `type` — and a branch like that leaves its own kind unsayable: nothing in the
// generated schema carries it, so neither could a request built from that
// schema. The value comes from the mapping, which is where the spec does say
// which kind the branch is.
//
// A branch that declares the property already is returned untouched, and so is
// every branch of a schema whose kinds are told apart by their fields rather
// than by a discriminator.
func withDiscriminator(target map[string]any, discriminator, name string, byValue map[string]string) map[string]any {
	if discriminator == "" || discriminatorValue(target, discriminator) != "" {
		return target
	}

	// The branch is named after the mapping entry that points at it, so the name
	// is the value wherever the mapping is what named it. Anything else named it
	// after something that is not a discriminator value, and inventing one from
	// that name would be putting words in the spec's mouth.
	if !slices.Contains(mappedValues(byValue), name) {
		return target
	}

	out, _ := deepCopy(target).(map[string]any)

	props, ok := out["properties"].(map[string]any)
	if !ok {
		props = map[string]any{}
		out["properties"] = props
	}

	props[discriminator] = map[string]any{
		"type": "string",
		"enum": []any{name},
	}

	setRequired(out, union(stringList(out["required"]), []string{discriminator}))

	return out
}

// splitIntoProperties rewrites a composed schema into one property per branch,
// each holding that branch's own schema.
//
// Terraform has no type for "exactly one of these shapes", but it can say
// "exactly one of these attributes", which is the same constraint moved down a
// level: a link monitor keeps its id and gains an fqdn, an ipv4 and an ipv6
// block, of which one is set. The runtime splices the set block back out when it
// builds a request, so the API still sees the shape it declared.
//
// The branch is inlined rather than referenced, because the marker has to go on
// this copy of it: the component it came from is shared with everything else in
// the document that uses the same shape.
func (p *Prep) splitIntoProperties(schema map[string]any, variants []variant, key string) {
	delete(schema, key)
	delete(schema, "discriminator")
	delete(schema, "additionalProperties")
	schema["type"] = "object"

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		props = map[string]any{}
		schema["properties"] = props
	}

	for _, entry := range variants {
		block, _ := deepCopy(entry.Bare).(map[string]any)
		block["description"] = strings.TrimSpace(stringOr(block["description"]) + " " + VariantBlockDescriptionSuffix)

		// A branch cannot be required: which one is set is the choice being made.
		props[entry.Name] = block
	}
}

// emitVariantPaths clones every claimed path once per variant, rewriting the
// schemas it reaches so that each clone describes exactly one branch.
//
// The generator maps one schema per path, so this is what lets four kinds of tag
// on /tags become four Terraform resources: each gets a path of its own that only
// the generator ever sees, while the runtime keeps calling /tags.
func (p *Prep) emitVariantPaths() {
	paths, ok := p.doc["paths"].(map[string]any)
	if !ok {
		return
	}

	for _, path := range slices.Sorted(maps.Keys(p.claimed)) {
		item, ok := paths[path].(map[string]any)
		if !ok {
			p.warnf("generator_config claims a variant of %s, which the spec does not have", path)

			continue
		}

		for _, name := range p.claimed[path] {
			clone, _ := deepCopy(item).(map[string]any)

			found := p.rewriteToVariant(clone, name, map[string]bool{}, nil)

			// The metadata describes how to recognise this kind in a response, so
			// it can only come from a response. A create body that branches does
			// not make the objects the endpoint lists two different kinds: a
			// gateway is one object whichever body made it.
			switch responded := p.respondedVariant(clone); {
			case responded != nil:
				clone[VariantExtension] = map[string]any{
					"name":          responded.Name,
					"discriminator": responded.discriminatorPath(),
					"value":         responded.Value,
					"match":         toAnyList(responded.Match),
					"wrapper":       responded.Wrapper,
					"wrapped":       toAnyList(responded.Wrapped),
				}
			case found != nil:
			case len(p.variantNamesOn(item)) > 0:
				p.warnf("%s has no variant %q to take; the spec offers %s", path, name, quoteList(p.variantNamesOn(item)))

				continue
			}

			// A path whose schemas do not branch is cloned unchanged: a gateway's
			// create body has two forms but the object it reads back has one, and
			// both Terraform types need the same read.
			paths[path+VariantSeparator+name] = clone
		}
	}
}

// respondedVariant finds the branch the operations on a rewritten path answer
// with, which is the only branch the runtime can recognise an object by.
func (p *Prep) respondedVariant(item map[string]any) *variant {
	for _, method := range slices.Sorted(maps.Keys(item)) {
		operation, ok := item[method].(map[string]any)
		if !ok {
			continue
		}

		responses, ok := operation["responses"].(map[string]any)
		if !ok {
			continue
		}

		if found := p.variantBehind(responses); found != nil {
			return found
		}
	}

	return nil
}

// variantBehind reports the branch reached through the references in node. A
// component cloned for a branch records the branch it was cloned for, so a branch
// wrapped in a collection envelope is found through the envelope's clone.
func (p *Prep) variantBehind(node any) *variant {
	switch typed := node.(type) {
	case map[string]any:
		if ref, ok := typed["$ref"].(string); ok {
			name, _ := componentName(ref)

			return p.taken[name]
		}

		for _, key := range slices.Sorted(maps.Keys(typed)) {
			if found := p.variantBehind(typed[key]); found != nil {
				return found
			}
		}
	case []any:
		for _, entry := range typed {
			if found := p.variantBehind(entry); found != nil {
				return found
			}
		}
	}

	return nil
}

// rewriteToVariant points every `$ref` in node at a component holding only the
// named branch, following references so a branch nested inside a collection
// envelope is reached too. It reports the branch it found, for the metadata the
// runtime discriminates with.
//
// path is where node sits inside the object the branch will be recognised on,
// which is what tells the runtime a tag's kind is at `config.type` rather than
// at `type`. It grows by a property name on the way into a property, and resets
// on the way into an array's items: a collection is recognised element by
// element, so an element's path starts again from the element.
func (p *Prep) rewriteToVariant(node any, name string, inFlight map[string]bool, path []string) *variant {
	var found *variant

	switch typed := node.(type) {
	case map[string]any:
		if ref, ok := typed["$ref"].(string); ok {
			component, replacement, direct := p.variantComponent(ref, name, inFlight)
			if component != "" {
				typed["$ref"] = component
			}

			return pickVariant(found, replacement.under(path, direct), direct)
		}

		if props, ok := typed["properties"].(map[string]any); ok {
			for _, property := range slices.Sorted(maps.Keys(props)) {
				found = pickVariant(found, p.rewriteToVariant(props[property], name, inFlight, append(slices.Clone(path), property)), false)
			}
		}

		for _, key := range slices.Sorted(maps.Keys(typed)) {
			if key == "properties" {
				continue
			}

			within := path
			if key == "items" {
				within = nil
			}

			found = pickVariant(found, p.rewriteToVariant(typed[key], name, inFlight, within), false)
		}
	case []any:
		for _, entry := range typed {
			found = pickVariant(found, p.rewriteToVariant(entry, name, inFlight, path), false)
		}
	}

	return found
}

// under places v's discriminator inside the object at path. A branch reached
// straight through a reference is the schema the discriminator is declared on,
// so the discriminator is wherever that reference was; a branch reached through
// a component that merely leads to one is already placed inside that component,
// so the two paths join.
func (v *variant) under(path []string, direct bool) *variant {
	if v == nil {
		return nil
	}

	out := *v

	if direct {
		out.Path = slices.Clone(path)
	} else {
		out.Path = append(slices.Clone(path), v.Path...)
	}

	return &out
}

// discriminatorPath spells out where the discriminator is on the object the
// runtime recognises, as the dotted path the registry hands the runtime.
func (v *variant) discriminatorPath() string {
	if v.Discriminator == "" {
		return ""
	}

	return strings.Join(append(slices.Clone(v.Path), v.Discriminator), ".")
}

// pickVariant keeps the branch worth reporting: one taken straight from a
// branching schema beats one found through an indirection.
func pickVariant(current, candidate *variant, preferred bool) *variant {
	switch {
	case candidate == nil:
		return current
	case current == nil, preferred:
		return candidate
	default:
		return current
	}
}

// variantComponent returns the component describing one branch of ref, creating it
// the first time it is asked for. A component that does not branch but reaches one
// that does is cloned too, so the branch survives the indirection.
func (p *Prep) variantComponent(ref, name string, inFlight map[string]bool) (string, *variant, bool) {
	source, ok := componentName(ref)
	if !ok || inFlight[source] {
		return "", nil, false
	}

	schema, ok := p.schemas[source].(map[string]any)
	if !ok {
		return "", nil, false
	}

	target := source + pascal(name)
	if _, exists := p.schemas[target]; exists {
		return componentRef(target), p.taken[target], p.direct[target]
	}

	if variants := p.variantsOf(schema, "components.schemas."+source); variants != nil {
		for _, entry := range variants {
			if entry.Name != name {
				continue
			}

			p.schemas[target] = entry.Schema
			p.taken[target] = &entry
			p.direct[target] = true

			return componentRef(target), &entry, true
		}

		return "", nil, false
	}

	clone, _ := deepCopy(schema).(map[string]any)

	inFlight[source] = true
	found := p.rewriteToVariant(clone, name, inFlight, nil)
	delete(inFlight, source)

	if found == nil {
		return "", nil, false
	}

	// The clone now points at one branch, so whatever the API wrapped that branch
	// in has exactly one thing left inside it and can go.
	found = p.hoistWrapper(clone, found, "components.schemas."+target)

	p.schemas[target] = clone
	p.taken[target] = found

	return componentRef(target), found, false
}

// hoistWrapper lifts a branch out of the property the API nests it under, and
// drops the discriminator with it.
//
// Both only exist to make the shape sayable on the wire. A Terraform type that is
// one kind of tag already says which kind it is, so requiring `config = { type =
// "gateway" }` on top of that asks a practitioner to write down what the resource
// name has already established — and for the two kinds whose branch declares no
// fields of its own, that is the whole of what they would write. The runtime puts
// both back when it builds a request; see genresource.Variant.Nest.
//
// Only a top-level branch is hoisted: the wrapper has to be a property of this very
// object, so that what comes up out of it can only meet the object's own fields.
// The variant returned records what moved, because nothing in the schema can say so
// afterwards — a hoisted property looks like any other.
func (p *Prep) hoistWrapper(schema map[string]any, found *variant, loc string) *variant {
	// A discriminator with no value is one the spec never spelled out for this
	// branch, which leaves the runtime nothing to write back in its place.
	if len(found.Path) != 1 || found.Discriminator == "" || found.Value == "" {
		return found
	}

	wrapper := found.Path[0]

	props, _ := schema["properties"].(map[string]any)

	entry, ok := props[wrapper].(map[string]any)
	if !ok {
		// This object reaches the branch rather than holding it — a collection
		// envelope around it, say — so there is nothing here to hoist. The variant
		// keeps whatever the object it reaches recorded, so the metadata describes
		// an element of the collection.
		return found
	}

	branch := p.branchBehind(entry)
	if branch == nil {
		return found
	}

	hoisted, required := hoistedProperties(branch, found.Discriminator)

	for _, name := range slices.Sorted(maps.Keys(hoisted)) {
		if _, clash := props[name]; clash {
			// Leaving it wrapped keeps the schema and the runtime agreeing on where
			// the field lives, which matters more than the tidier shape: the variant
			// goes back unchanged, so no wrapper is advertised and nothing tries to
			// unwrap one.
			p.warnf("%s: cannot hoist %q, it and its parent both declare %q; left wrapped", loc, wrapper, name)

			return found
		}
	}

	delete(props, wrapper)
	maps.Copy(props, hoisted)
	setRequired(schema, removedFromList(union(stringList(schema["required"]), required), wrapper))

	out := *found
	out.Wrapper = wrapper
	out.Wrapped = slices.Sorted(maps.Keys(hoisted))

	return &out
}

// branchBehind returns the component a hoisted wrapper was narrowed to. Only a
// reference is followed, and only into the components prep itself built: the
// wrapper's schema was rewritten to point at one branch, so a wrapper still holding
// anything else is not one this can hoist.
func (p *Prep) branchBehind(entry map[string]any) map[string]any {
	ref, ok := entry["$ref"].(string)
	if !ok {
		return nil
	}

	name, ok := componentName(ref)
	if !ok {
		return nil
	}

	schema, _ := p.schemas[name].(map[string]any)

	return schema
}

// hoistedProperties returns what a branch contributes to the object it is lifted
// into: its properties and its required names, both without the discriminator.
func hoistedProperties(branch map[string]any, discriminator string) (map[string]any, []string) {
	props, _ := branch["properties"].(map[string]any)

	out := make(map[string]any, len(props))

	for name, value := range props {
		if name == discriminator {
			continue
		}

		out[name] = deepCopy(value)
	}

	return out, removedFromList(stringList(branch["required"]), discriminator)
}

// mappedValues lists the discriminator values a mapping spells out.
func mappedValues(byValue map[string]string) []string {
	out := make([]string, 0, len(byValue))

	for _, value := range byValue {
		out = append(out, value)
	}

	return out
}

// variantNamesOn lists the branches a path offers, for the error raised when the
// configuration asks for one that is not there.
func (p *Prep) variantNamesOn(item map[string]any) []string {
	var names []string

	var walk func(node any, inFlight map[string]bool)

	walk = func(node any, inFlight map[string]bool) {
		switch typed := node.(type) {
		case map[string]any:
			if ref, ok := typed["$ref"].(string); ok {
				source, ok := componentName(ref)
				if !ok || inFlight[source] {
					return
				}

				schema, ok := p.schemas[source].(map[string]any)
				if !ok {
					return
				}

				for _, entry := range p.variantsOf(schema, source) {
					if !slices.Contains(names, entry.Name) {
						names = append(names, entry.Name)
					}
				}

				inFlight[source] = true
				walk(schema, inFlight)
				delete(inFlight, source)

				return
			}

			for _, key := range slices.Sorted(maps.Keys(typed)) {
				walk(typed[key], inFlight)
			}
		case []any:
			for _, entry := range typed {
				walk(entry, inFlight)
			}
		}
	}

	walk(item, map[string]bool{})
	slices.Sort(names)

	return names
}

func componentRef(name string) string {
	return "#/components/schemas/" + name
}

func quoteList(values []string) string {
	if len(values) == 0 {
		return "none"
	}

	quoted := make([]string, 0, len(values))

	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("%q", value))
	}

	return strings.Join(quoted, ", ")
}

func compositionKey(schema map[string]any) string {
	for _, key := range compositionKeys {
		if list, ok := schema[key].([]any); ok && len(list) > 0 {
			return key
		}
	}

	return ""
}

// variantSchema builds a branch's standalone schema: everything the composed
// schema declared outside the composition, plus the branch's own properties.
func (p *Prep) variantSchema(schema, target map[string]any, key string) map[string]any {
	out := map[string]any{}

	for name, value := range schema {
		if name == key || name == "discriminator" {
			continue
		}

		out[name] = deepCopy(value)
	}

	out["type"] = "object"

	props, ok := out["properties"].(map[string]any)
	if !ok {
		props = map[string]any{}
		out["properties"] = props
	}

	for name, value := range targetProperties(target) {
		props[name] = deepCopy(value)
	}

	setRequired(out, union(stringList(out["required"]), stringList(target["required"])))

	if description := stringOr(target["description"]); description != "" {
		out["description"] = description
	}

	return out
}

func targetProperties(target map[string]any) map[string]any {
	props, _ := target["properties"].(map[string]any)

	return props
}

// discriminatorOf reads the spec's discriminator, returning the selecting
// property and the component-to-value mapping inverted for lookup.
func discriminatorOf(schema map[string]any) (string, map[string]string) {
	asMap, ok := schema["discriminator"].(map[string]any)
	if !ok {
		return "", nil
	}

	property := stringOr(asMap["propertyName"])
	if property == "" {
		return "", nil
	}

	byValue := map[string]string{}

	mapping, _ := asMap["mapping"].(map[string]any)
	for value, ref := range mapping {
		if name, ok := componentName(stringOr(ref)); ok {
			byValue[name] = value
		}
	}

	return property, byValue
}

// discriminatorValue reads the single value a variant's discriminator property
// is pinned to, which is how the spec spells "this branch is the wanlink one".
func discriminatorValue(target map[string]any, property string) string {
	if property == "" {
		return ""
	}

	props, _ := target["properties"].(map[string]any)

	entry, ok := props[property].(map[string]any)
	if !ok {
		return ""
	}

	values := anyStrings(entry["enum"])
	if len(values) != 1 {
		return ""
	}

	return values[0]
}

// variantNames works out what to call each branch. The spec names them itself in
// two ways — a discriminator mapping, or one component per branch — and where it
// does neither the branch is named after the properties it requires, which is
// the only thing that distinguishes it and is stable if the spec reorders.
func variantNames(variants []map[string]any, refs []string, byValue map[string]string, discriminator string) []string {
	out := make([]string, len(variants))

	if discriminator != "" {
		named := 0

		for i, ref := range refs {
			name, _ := componentName(ref)

			switch {
			case byValue[name] != "":
				out[i] = byValue[name]
				named++
			case discriminatorValue(variants[i], discriminator) != "":
				out[i] = discriminatorValue(variants[i], discriminator)
				named++
			}
		}

		if named == len(variants) && unique(out) {
			return out
		}
	}

	if componentNames := trimmedComponentNames(refs); componentNames != nil {
		return componentNames
	}

	// Nothing in the spec names this branch, so it is named after what it requires:
	// the only thing telling it apart, and stable if the spec reorders the branches.
	for i, target := range variants {
		out[i] = strings.Join(slices.Sorted(slices.Values(stringList(target["required"]))), "_")
		if out[i] == "" {
			out[i] = fmt.Sprintf("variant_%d", i)
		}
	}

	return out
}

// trimmedComponentNames names branches after their components with whatever every
// component has in common removed, so LinkMonitorFqdn/Ipv4/Ipv6 become fqdn, ipv4
// and ipv6 and WanlinkTag/OverlayTag become wanlink and overlay.
func trimmedComponentNames(refs []string) []string {
	names := make([]string, 0, len(refs))

	for _, ref := range refs {
		name, ok := componentName(ref)
		if !ok {
			return nil
		}

		names = append(names, name)
	}

	trimmed := make([]string, len(names))

	prefix := commonPrefix(names)
	suffix := commonSuffix(names)

	for i, name := range names {
		short := name[len(prefix) : len(name)-len(suffix)]
		if short == "" {
			return nil
		}

		trimmed[i] = snake(short)
	}

	if !unique(trimmed) {
		return nil
	}

	return trimmed
}

// commonPrefix returns the longest shared prefix that ends on a word boundary, so
// trimming it never cuts a name mid-word.
func commonPrefix(names []string) string {
	shared := names[0]

	for _, name := range names[1:] {
		limit := min(len(shared), len(name))
		i := 0

		for i < limit && shared[i] == name[i] {
			i++
		}

		shared = shared[:i]
	}

	for len(shared) > 0 && !startsWord(names, len(shared)) {
		shared = shared[:len(shared)-1]
	}

	return shared
}

// startsWord reports whether every name has a word boundary at offset, which for
// the spec's PascalCase component names means an upper-case letter.
func startsWord(names []string, offset int) bool {
	for _, name := range names {
		if offset >= len(name) || !unicode.IsUpper(rune(name[offset])) {
			return false
		}
	}

	return true
}

func commonSuffix(names []string) string {
	shared := names[0]

	for _, name := range names[1:] {
		limit := min(len(shared), len(name))
		i := 0

		for i < limit && shared[len(shared)-1-i] == name[len(name)-1-i] {
			i++
		}

		shared = shared[len(shared)-i:]
	}

	for len(shared) > 0 && !unicode.IsUpper(rune(shared[0])) {
		shared = shared[1:]
	}

	// Every name has to keep something once the suffix is gone.
	for _, name := range names {
		if len(name) <= len(shared) {
			return ""
		}
	}

	return shared
}

// distinctiveProperties lists what only this branch declares, which is how a
// branch is recognised in a response when the spec has no discriminator.
func distinctiveProperties(variants []map[string]any, index int) []string {
	var out []string

	others := map[string]bool{}

	for i, target := range variants {
		if i == index {
			continue
		}

		for name := range targetProperties(target) {
			others[name] = true
		}
	}

	for _, name := range slices.Sorted(maps.Keys(targetProperties(variants[index]))) {
		if !others[name] {
			out = append(out, name)
		}
	}

	return out
}

func unique(values []string) bool {
	seen := map[string]bool{}

	for _, value := range values {
		if value == "" || seen[value] {
			return false
		}

		seen[value] = true
	}

	return true
}

// snake converts a PascalCase component name to the snake_case Terraform uses.
func snake(name string) string {
	var out strings.Builder

	runes := []rune(name)

	for i, r := range runes {
		if unicode.IsUpper(r) {
			// A run of capitals is one word: IPv4Address breaks before "Address",
			// not between the capitals.
			if i > 0 && (!unicode.IsUpper(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1]))) {
				out.WriteByte('_')
			}

			out.WriteRune(unicode.ToLower(r))

			continue
		}

		out.WriteRune(r)
	}

	return out.String()
}
