package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// RawJSONDescriptionSuffix is appended to the description of every untyped
// OpenAPI schema that gets rewritten into a string attribute. The provider
// runtime treats those attributes as embedded JSON documents.
const RawJSONDescriptionSuffix = "Raw JSON document, encoded as a string."

// EnumDescriptionPrefix introduces the allowed-values list appended to the
// description of every schema carrying an `enum`. The generator already turns
// `enum` into a validator on its own; nothing but the description reaches
// tfplugindocs, so this is the only way a practitioner sees the values a
// generated page does not otherwise restate.
const EnumDescriptionPrefix = "Must be one of:"

// maxRefHops bounds $ref chasing so a self-referential spec cannot hang the
// generator.
const maxRefHops = 16

// schemaChildKeys are the schema keywords holding a single nested schema.
var schemaChildKeys = []string{"items", "not", "additionalProperties", "contains"}

// schemaListKeys are the schema keywords holding a list of nested schemas.
var schemaListKeys = []string{"allOf", "oneOf", "anyOf", "prefixItems"}

// compositionKeys are the "exactly one of" keywords that get collapsed into a
// single object. `allOf` is deliberately absent: it means "all of these apply",
// so it is only reported, never merged.
var compositionKeys = []string{"oneOf", "anyOf"}

// typeBearingKeys are the keywords that tell the generator what an attribute
// looks like. A schema with none of them cannot be mapped.
var typeBearingKeys = []string{"type", "properties", "enum", "const", "format", "$ref"}

// Prep rewrites a bundled OpenAPI document into an equivalent one that the
// Terraform OpenAPI provider spec generator can map in full.
//
// The generator has two relevant gaps: it cannot map schema composition
// (`oneOf`/`anyOf` with more than one non-null subschema) and it drops schemas
// carrying no type information. Both appear in the BWAN spec, so rather than
// hand-writing those attributes in Go the spec is normalised first:
//
//   - composition becomes a single object holding the union of every variant's
//     properties, required only where every variant agrees;
//   - untyped schemas become strings holding a JSON document;
//   - path parameters containing "-" are renamed to "_" so they survive as
//     Terraform attribute names.
type Prep struct {
	doc      map[string]any
	schemas  map[string]any
	done     map[string]bool
	inFlight map[string]bool

	// claimed lists, per API path, the variants generator_config.yml binds to a
	// Terraform type of their own. A composed schema on a claimed path is split
	// across synthetic paths, one per variant; a composed schema nobody claims
	// becomes one property per variant instead.
	claimed map[string][]string

	// renames maps an API field name to the Terraform attribute name it has to be
	// exposed under, for the names Terraform reserves.
	renames map[string]string

	// drops lists the API fields removed from every schema before renames run,
	// so a field superseded by a differently-named replacement can hand its name
	// over without the collision guard in renameReservedProperties refusing it.
	drops []string

	// taken records the branch each synthetic component was built from, so the
	// metadata reaches the path that references it. direct marks the components
	// that are a branch themselves rather than a clone reaching one.
	taken  map[string]*variant
	direct map[string]bool

	// Warnings collects everything a human should look at: lossy merges and
	// constructs left untouched because they could not be collapsed safely.
	Warnings []string
}

func NewPrep(doc map[string]any) *Prep {
	schemas, _ := nestedMap(doc, "components", "schemas")

	return &Prep{
		doc:      doc,
		schemas:  schemas,
		done:     map[string]bool{},
		inFlight: map[string]bool{},
		claimed:  map[string][]string{},
		renames:  map[string]string{},
		taken:    map[string]*variant{},
		direct:   map[string]bool{},
	}
}

// Claim binds a variant of whatever schemas path uses to a Terraform type of its
// own, which is how one API endpoint becomes several resources.
func (p *Prep) Claim(path, name string) {
	if !slices.Contains(p.claimed[path], name) {
		p.claimed[path] = append(p.claimed[path], name)
	}
}

// Rename exposes an API field under a different Terraform attribute name.
func (p *Prep) Rename(from, to string) {
	p.renames[from] = to
}

// Drop removes an API field from every schema, ahead of any renames.
func (p *Prep) Drop(name string) {
	if !slices.Contains(p.drops, name) {
		p.drops = append(p.drops, name)
	}
}

func (p *Prep) Run() {
	p.renameDashedPathParams()
	p.dropProperties()
	p.renameReservedProperties()

	// Claimed paths are cloned before anything is normalised: once a composed
	// schema has been split across properties there is no composition left to
	// take a variant of.
	p.emitVariantPaths()

	for _, name := range slices.Sorted(maps.Keys(p.schemas)) {
		p.transformComponent(name)
	}

	p.transformOperationSchemas()
}

func (p *Prep) warnf(format string, args ...any) {
	p.Warnings = append(p.Warnings, fmt.Sprintf(format, args...))
}

// renameDashedPathParams rewrites `{group-id}` to `{group_id}` in path
// templates and in the matching parameter definitions. Terraform attribute
// names cannot contain "-", so the generator strips the dash and produces
// `groupid`, which no longer matches the placeholder it came from: the runtime
// fills a path by attribute name, and would find nothing to put in `{group-id}`.
func (p *Prep) renameDashedPathParams() {
	paths, ok := p.doc["paths"].(map[string]any)
	if !ok {
		return
	}

	for path, item := range paths {
		if renamed := renamePathTemplate(path); renamed != path {
			delete(paths, path)
			paths[renamed] = item
		}

		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}

		for _, op := range itemMap {
			if opMap, ok := op.(map[string]any); ok {
				renamePathParams(opMap)
			}
		}
	}
}

// renameReservedProperties renames the fields Terraform will not accept as
// attribute names. Every schema is rewritten, not only the ones that end up at a
// resource root, so the attribute is called the same thing wherever it turns up.
func (p *Prep) renameReservedProperties() {
	if len(p.renames) == 0 {
		return
	}

	for _, name := range slices.Sorted(maps.Keys(p.schemas)) {
		schema, ok := p.schemas[name].(map[string]any)
		if !ok {
			continue
		}

		props, ok := schema["properties"].(map[string]any)
		if !ok {
			continue
		}

		for from, to := range p.renames {
			value, ok := props[from]
			if !ok {
				continue
			}

			if _, taken := props[to]; taken {
				p.warnf("components.schemas.%s already has a %q to rename %q to", name, to, from)

				continue
			}

			delete(props, from)
			props[to] = value

			schema["required"] = toAnyList(renamedList(stringList(schema["required"]), from, to))
			if len(stringList(schema["required"])) == 0 {
				delete(schema, "required")
			}
		}
	}
}

func renamedList(values []string, from, to string) []string {
	out := make([]string, 0, len(values))

	for _, value := range values {
		if value == from {
			value = to
		}

		out = append(out, value)
	}

	return out
}

// dropProperties removes a field superseded by a differently-named replacement
// from every schema, before renameReservedProperties runs. Every schema is
// rewritten, not only the ones a resource claims, for the same reason renames
// are: the field is named the same thing wherever it turns up.
func (p *Prep) dropProperties() {
	if len(p.drops) == 0 {
		return
	}

	for _, name := range slices.Sorted(maps.Keys(p.schemas)) {
		schema, ok := p.schemas[name].(map[string]any)
		if !ok {
			continue
		}

		props, ok := schema["properties"].(map[string]any)
		if !ok {
			continue
		}

		for _, drop := range p.drops {
			if _, ok := props[drop]; !ok {
				continue
			}

			delete(props, drop)
			schema["required"] = toAnyList(removedFromList(stringList(schema["required"]), drop))
		}

		if len(stringList(schema["required"])) == 0 {
			delete(schema, "required")
		}
	}
}

func removedFromList(values []string, name string) []string {
	out := make([]string, 0, len(values))

	for _, value := range values {
		if value != name {
			out = append(out, value)
		}
	}

	return out
}

func renamePathTemplate(path string) string {
	var out strings.Builder

	for {
		open := strings.Index(path, "{")
		if open < 0 {
			out.WriteString(path)

			return out.String()
		}

		end := strings.Index(path[open:], "}")
		if end < 0 {
			out.WriteString(path)

			return out.String()
		}

		end += open

		out.WriteString(path[:open+1])
		out.WriteString(strings.ReplaceAll(path[open+1:end], "-", "_"))
		out.WriteString("}")

		path = path[end+1:]
	}
}

func renamePathParams(op map[string]any) {
	params, ok := op["parameters"].([]any)
	if !ok {
		return
	}

	for _, param := range params {
		paramMap, ok := param.(map[string]any)
		if !ok || paramMap["in"] != "path" {
			continue
		}

		if name, ok := paramMap["name"].(string); ok {
			paramMap["name"] = strings.ReplaceAll(name, "-", "_")
		}
	}
}

// transformComponent normalises a named component schema exactly once. Named
// schemas are transformed on demand because collapsing composition needs its
// variants already normalised.
func (p *Prep) transformComponent(name string) {
	if p.done[name] {
		return
	}

	if p.inFlight[name] {
		p.warnf("components.schemas.%s: recursive $ref, left as-is", name)

		return
	}

	schema, ok := p.schemas[name].(map[string]any)
	if !ok {
		p.done[name] = true

		return
	}

	p.inFlight[name] = true
	p.transformSchema(schema, "components.schemas."+name)
	delete(p.inFlight, name)

	p.done[name] = true
}

func (p *Prep) transformOperationSchemas() {
	paths, ok := p.doc["paths"].(map[string]any)
	if !ok {
		return
	}

	for _, path := range slices.Sorted(maps.Keys(paths)) {
		itemMap, ok := paths[path].(map[string]any)
		if !ok {
			continue
		}

		for _, method := range slices.Sorted(maps.Keys(itemMap)) {
			if opMap, ok := itemMap[method].(map[string]any); ok {
				p.transformOperation(opMap, path+"."+method)
			}
		}
	}
}

func (p *Prep) transformOperation(op map[string]any, loc string) {
	if params, ok := op["parameters"].([]any); ok {
		for i, param := range params {
			paramMap, ok := param.(map[string]any)
			if !ok {
				continue
			}

			if schema, ok := paramMap["schema"].(map[string]any); ok {
				p.transformSchema(schema, fmt.Sprintf("%s.parameters[%d]", loc, i))
			}
		}
	}

	if body, ok := op["requestBody"].(map[string]any); ok {
		p.transformContent(body, loc+".requestBody")
	}

	responses, ok := op["responses"].(map[string]any)
	if !ok {
		return
	}

	for _, code := range slices.Sorted(maps.Keys(responses)) {
		if responseMap, ok := responses[code].(map[string]any); ok {
			p.transformContent(responseMap, loc+".responses."+code)
		}
	}
}

func (p *Prep) transformContent(holder map[string]any, loc string) {
	content, ok := holder["content"].(map[string]any)
	if !ok {
		return
	}

	for _, mediaType := range slices.Sorted(maps.Keys(content)) {
		mediaMap, ok := content[mediaType].(map[string]any)
		if !ok {
			continue
		}

		if schema, ok := mediaMap["schema"].(map[string]any); ok {
			p.transformSchema(schema, loc+".content."+mediaType)
		}
	}
}

// transformSchema normalises schema in place, depth first: children are
// normalised before the node itself so a collapsed variant is never re-visited.
func (p *Prep) transformSchema(schema map[string]any, loc string) {
	if _, isRef := schema["$ref"]; isRef {
		// The target is a named component and gets normalised through
		// transformComponent; rewriting it here would duplicate the work.
		return
	}

	for _, key := range schemaChildKeys {
		if child, ok := schema[key].(map[string]any); ok {
			p.transformSchema(child, loc+"."+key)
		}
	}

	if props, ok := schema["properties"].(map[string]any); ok {
		for _, name := range slices.Sorted(maps.Keys(props)) {
			if child, ok := props[name].(map[string]any); ok {
				p.transformSchema(child, loc+"."+name)
			}
		}
	}

	for _, key := range schemaListKeys {
		list, ok := schema[key].([]any)
		if !ok {
			continue
		}

		for i, entry := range list {
			if child, ok := entry.(map[string]any); ok {
				p.transformSchema(child, fmt.Sprintf("%s.%s[%d]", loc, key, i))
			}
		}
	}

	if _, ok := schema["allOf"]; ok {
		p.warnf("%s: allOf is not collapsed, the generator will skip it", loc)
	}

	p.collapseComposition(schema, loc)
	p.typeUntyped(schema, loc)
	p.documentEnum(schema, loc)
}

// collapseComposition rewrites a composed schema into something the generator can
// map: one property per branch, each holding that branch's schema.
//
// A branch generator_config.yml claimed has already been lifted onto a synthetic
// path of its own, so what reaches here is the branching nobody wanted separate
// Terraform types for.
func (p *Prep) collapseComposition(schema map[string]any, loc string) {
	key := compositionKey(schema)
	if key == "" {
		return
	}

	objects, present := p.objectVariants(schema, key, loc)

	// A single variant next to `null` is how the spec spells "nullable"; the
	// generator maps that natively.
	if present < 2 {
		return
	}

	// A union of an object and a scalar is the API accepting a shorthand ("an IP"
	// instead of "{ip: ...}"). Terraform cannot type that, so only the object form
	// is exposed; it can always express the shorthand too.
	if dropped := present - len(objects); dropped > 0 {
		p.warnf("%s.%s: dropped %d non-object variant(s), only the object form is exposed", loc, key, dropped)
	}

	if len(objects) == 0 {
		p.warnf("%s.%s: no object variant to merge into, left as-is", loc, key)

		return
	}

	// Two or more forms become a block each; one form left after the scalars were
	// dropped is not a choice, so the schema simply becomes that form.
	if variants := p.variantsOf(schema, loc); variants != nil {
		p.splitIntoProperties(schema, variants, key)

		return
	}

	merged := p.variantSchema(schema, objects[0], key)

	clear(schema)
	maps.Copy(schema, merged)
}

// objectVariants resolves a composition's branches, returning the object-shaped
// ones and how many branches there were beside `null`. A single branch next to
// `null` is how the spec spells nullable, which the generator maps natively.
func (p *Prep) objectVariants(schema map[string]any, key, loc string) ([]map[string]any, int) {
	raw, _ := schema[key].([]any)

	var (
		objects []map[string]any
		present int
	)

	for i, entry := range raw {
		target := p.resolve(entry, fmt.Sprintf("%s.%s[%d]", loc, key, i))
		if target == nil || isNullSchema(target) {
			continue
		}

		present++

		if _, isObject := target["properties"].(map[string]any); isObject || target["type"] == "object" {
			objects = append(objects, target)
		}
	}

	return objects, present
}

// resolve follows `$ref` to a normalised component schema.
func (p *Prep) resolve(entry any, loc string) map[string]any {
	schema, ok := entry.(map[string]any)
	if !ok {
		return nil
	}

	for range maxRefHops {
		ref, ok := schema["$ref"].(string)
		if !ok {
			return schema
		}

		name, ok := componentName(ref)
		if !ok {
			p.warnf("%s: unsupported $ref %q", loc, ref)

			return schema
		}

		p.transformComponent(name)

		target, ok := p.schemas[name].(map[string]any)
		if !ok {
			p.warnf("%s: dangling $ref %q", loc, ref)

			return schema
		}

		schema = target
	}

	p.warnf("%s: $ref chain too deep", loc)

	return schema
}

// typeUntyped turns a schema carrying no type information into a string holding
// a JSON document. The generator drops such schemas, and Terraform has no
// dynamic-shape attribute that round-trips arbitrary JSON safely.
func (p *Prep) typeUntyped(schema map[string]any, loc string) {
	for _, key := range typeBearingKeys {
		if _, ok := schema[key]; ok {
			return
		}
	}

	for _, key := range schemaListKeys {
		if _, ok := schema[key]; ok {
			return
		}
	}

	schema["type"] = "string"
	schema["description"] = strings.TrimSpace(stringOr(schema["description"]) + " " + RawJSONDescriptionSuffix)
}

// documentEnum appends the allowed values to the description of a schema
// carrying an `enum`, so the list a validator already enforces also reaches the
// generated docs. Non-string members are left out of the list: an enum mixing
// scalar types is not something the BWAN spec does today, and stringifying a
// number or bool loses whether it round-trips as one.
func (p *Prep) documentEnum(schema map[string]any, loc string) {
	values := stringList(schema["enum"])
	if len(values) == 0 {
		return
	}

	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = "`" + value + "`"
	}

	suffix := EnumDescriptionPrefix + " " + strings.Join(quoted, ", ") + "."
	schema["description"] = strings.TrimSpace(stringOr(schema["description"]) + " " + suffix)
}

// isNullSchema reports whether a composition variant only exists to make its
// sibling variants nullable. Unquoted in YAML, `type: null` parses as the null
// value rather than the string, so both spellings count.
func isNullSchema(schema map[string]any) bool {
	value, ok := schema["type"]

	return ok && (value == nil || value == "null")
}

func setRequired(schema map[string]any, required []string) {
	if len(required) == 0 {
		delete(schema, "required")

		return
	}

	schema["required"] = toAnyList(required)
}

func componentName(ref string) (string, bool) {
	const prefix = "#/components/schemas/"

	if !strings.HasPrefix(ref, prefix) {
		return "", false
	}

	name := strings.TrimPrefix(ref, prefix)
	if name == "" || strings.Contains(name, "/") {
		return "", false
	}

	return name, true
}

func nestedMap(root map[string]any, keys ...string) (map[string]any, bool) {
	current := root

	for _, key := range keys {
		next, ok := current[key].(map[string]any)
		if !ok {
			return nil, false
		}

		current = next
	}

	return current, true
}

func stringList(value any) []string {
	return anyStrings(value)
}

func anyStrings(value any) []string {
	list, ok := value.([]any)
	if !ok {
		return nil
	}

	out := make([]string, 0, len(list))

	for _, entry := range list {
		if s, ok := entry.(string); ok {
			out = append(out, s)
		}
	}

	return out
}

func toAnyList(values []string) []any {
	out := make([]any, 0, len(values))

	for _, value := range values {
		out = append(out, value)
	}

	return out
}

// union keeps first-seen order so the output is stable across runs.
func union(lists ...[]string) []string {
	var out []string

	for _, list := range lists {
		for _, value := range list {
			if !slices.Contains(out, value) {
				out = append(out, value)
			}
		}
	}

	return out
}

func stringOr(value any) string {
	if s, ok := value.(string); ok {
		return s
	}

	return ""
}

func deepCopy(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, entry := range typed {
			out[key] = deepCopy(entry)
		}

		return out
	case []any:
		out := make([]any, 0, len(typed))
		for _, entry := range typed {
			out = append(out, deepCopy(entry))
		}

		return out
	default:
		return value
	}
}
