package main

import (
	"encoding/json"
	"fmt"
	"go/format"
	"maps"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Where the framework generator writes its packages, relative to the module
// root. These have to match the outputs declared in //mgmt/tf-provider:BUILD.
const (
	resourcesDir   = "internal/gen/resources"
	dataSourcesDir = "internal/gen/datasources"
)

// generatorConfig is the part of generator_config.yml the registry needs. The
// file is the HashiCorp generator's own configuration format; everything else in
// it describes schemas rather than behaviour.
type generatorConfig struct {
	Resources   map[string]resourceConfig   `yaml:"resources"`
	DataSources map[string]dataSourceConfig `yaml:"data_sources"`
}

type resourceConfig struct {
	Create    *operationConfig `yaml:"create"`
	Read      *operationConfig `yaml:"read"`
	Update    *operationConfig `yaml:"update"`
	Delete    *operationConfig `yaml:"delete"`
	Terraform terraformConfig  `yaml:"x_terraform"`
}

type dataSourceConfig struct {
	Read      *operationConfig `yaml:"read"`
	Terraform terraformConfig  `yaml:"x_terraform"`
}

type operationConfig struct {
	Path   string `yaml:"path"`
	Method string `yaml:"method"`
}

// terraformConfig is the provider's own extension to the generator config. The
// HashiCorp generator ignores unknown keys, so it lives in the same file.
type terraformConfig struct {
	// RawFeature names the opt-in guarding an object the provider does not model
	// in full, usually because its configuration is an opaque JSON document.
	RawFeature string `yaml:"raw_feature"`
	// Variant binds this object to one branch of the schemas behind its endpoints,
	// which is how one endpoint serving several kinds of object becomes several
	// Terraform types.
	Variant string `yaml:"variant"`
}

// rawAttributeSuffix marks an attribute holding an object's whole configuration
// as an opaque document: the BWAN API names those device_config_raw,
// policy_config_raw and so on. An opaque *value* — a credentials blob, an audit
// diff — does not carry the suffix and does not make an object raw, because
// there is no typed replacement to roll out for it.
const rawAttributeSuffix = "_raw"

// Names the runtime and the API agree on.
const (
	idAttribute     = "id"
	filterParameter = "filter"
)

func generateRegistry(configPath, specPath, openapiPath, modulePath string) ([]byte, error) {
	config, err := readGeneratorConfig(configPath)
	if err != nil {
		return nil, err
	}

	rawJSON, variantBlocks, err := readMarkedAttributes(specPath)
	if err != nil {
		return nil, err
	}

	if err := checkGenerated(config, rawJSON); err != nil {
		return nil, err
	}

	if err := checkRawGating(config, rawJSON); err != nil {
		return nil, err
	}

	searchable, err := readSearchablePaths(openapiPath)
	if err != nil {
		return nil, err
	}

	variants, err := readVariantMetadata(openapiPath)
	if err != nil {
		return nil, err
	}

	renames, err := readRenames(configPath)
	if err != nil {
		return nil, err
	}

	return renderRegistry(config, rawJSON, variantBlocks, searchable, variants, renames, modulePath)
}

// variantMeta is how the runtime tells one kind of object from another on an
// endpoint that serves several. prep works it out from the spec and records it on
// the synthetic path it emits for the branch.
type variantMeta struct {
	Name          string   `yaml:"name"`
	Discriminator string   `yaml:"discriminator"`
	Value         string   `yaml:"value"`
	Match         []string `yaml:"match"`
}

// readVariantMetadata collects what prep recorded on each synthetic path.
//
// Absence is meaningful: a path cloned for a variant whose schemas turned out not
// to branch has no metadata, and needs none. Two Terraform types over one
// unbranching endpoint — a typed gateway and a raw one — are the same objects
// shown differently, not different objects to be told apart.
func readVariantMetadata(path string) (map[string]*variantMeta, error) {
	if path == "" {
		return map[string]*variantMeta{}, nil
	}

	doc, err := readYAML(path)
	if err != nil {
		return nil, err
	}

	paths, _ := doc["paths"].(map[string]any)
	out := map[string]*variantMeta{}

	for template, item := range paths {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}

		raw, ok := itemMap[VariantExtension].(map[string]any)
		if !ok {
			continue
		}

		out[template] = &variantMeta{
			Name:          stringOr(raw["name"]),
			Discriminator: stringOr(raw["discriminator"]),
			Value:         stringOr(raw["value"]),
			Match:         anyStrings(raw["match"]),
		}
	}

	return out, nil
}

// renderVariant emits the variant an object is pinned to, if the endpoint behind
// it serves more than one kind.
func renderVariant(readPath, variant string, metadata map[string]*variantMeta) string {
	if variant == "" {
		return ""
	}

	meta, ok := metadata[readPath+VariantSeparator+variant]
	if !ok {
		return ""
	}

	return fmt.Sprintf(
		"&genresource.Variant{Name: %q, Discriminator: %q, Value: %q, Match: %s}",
		meta.Name, meta.Discriminator, meta.Value, renderStrings(meta.Match))
}

// searchablePaths holds the collection endpoints that can be filtered, so a data
// source addressing one object can offer a lookup by filter as well as by id.
type searchablePaths map[string]bool

func readSearchablePaths(path string) (searchablePaths, error) {
	if path == "" {
		return searchablePaths{}, nil
	}

	doc, err := readYAML(path)
	if err != nil {
		return nil, err
	}

	paths, _ := doc["paths"].(map[string]any)
	out := searchablePaths{}

	for template, item := range paths {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}

		operation, ok := itemMap["get"].(map[string]any)
		if !ok {
			continue
		}

		if hasQueryParameter(operation, filterParameter) {
			out[template] = true
		}
	}

	return out, nil
}

func hasQueryParameter(operation map[string]any, name string) bool {
	parameters, _ := operation["parameters"].([]any)

	for _, parameter := range parameters {
		asMap, ok := parameter.(map[string]any)
		if !ok {
			continue
		}

		if asMap["in"] == "query" && asMap["name"] == name {
			return true
		}
	}

	return false
}

// searchPath returns the filterable collection a single-object read path belongs
// to, if the API has one.
func (s searchablePaths) searchPath(readPath string) string {
	const suffix = "/{" + idAttribute + "}"

	if !strings.HasSuffix(readPath, suffix) {
		return ""
	}

	collection := strings.TrimSuffix(readPath, suffix)
	if !s[collection] {
		return ""
	}

	return collection
}

func readGeneratorConfig(path string) (*generatorConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var config generatorConfig
	if err := yaml.Unmarshal(raw, &config); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	if len(config.Resources) == 0 && len(config.DataSources) == 0 {
		return nil, fmt.Errorf("%s declares no resources or data sources", path)
	}

	return &config, nil
}

// markedAttributes maps an object name to the dot-separated attribute paths prep
// marked in its description: the strings holding an embedded JSON document, and
// the blocks one of which stands for the form its parent takes.
type markedAttributes struct {
	resources   map[string][]string
	dataSources map[string][]string
}

func readMarkedAttributes(path string) (*markedAttributes, *markedAttributes, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var spec struct {
		Resources   []specObject `json:"resources"`
		DataSources []specObject `json:"datasources"`
	}

	if err := json.Unmarshal(content, &spec); err != nil {
		return nil, nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	raw := &markedAttributes{resources: map[string][]string{}, dataSources: map[string][]string{}}
	variants := &markedAttributes{resources: map[string][]string{}, dataSources: map[string][]string{}}

	for _, object := range spec.Resources {
		raw.resources[object.Name] = object.marked(RawJSONDescriptionSuffix)
		variants.resources[object.Name] = object.marked(VariantBlockDescriptionSuffix)
	}

	for _, object := range spec.DataSources {
		raw.dataSources[object.Name] = object.marked(RawJSONDescriptionSuffix)
		variants.dataSources[object.Name] = object.marked(VariantBlockDescriptionSuffix)
	}

	return raw, variants, nil
}

type specObject struct {
	Name   string `json:"name"`
	Schema struct {
		Attributes []map[string]any `json:"attributes"`
	} `json:"schema"`
}

// marked collects the dot-separated paths of every attribute whose description
// ends with suffix.
func (o specObject) marked(suffix string) []string {
	var out []string

	collectMarked(o.Schema.Attributes, "", suffix, &out)

	return out
}

func collectMarked(attributes []map[string]any, prefix, suffix string, out *[]string) {
	for _, attribute := range attributes {
		name, _ := attribute["name"].(string)
		if name == "" {
			continue
		}

		path := name
		if prefix != "" {
			path = prefix + "." + name
		}

		for kind, raw := range attribute {
			body, ok := raw.(map[string]any)
			if !ok {
				continue
			}

			if description, _ := body["description"].(string); strings.HasSuffix(strings.TrimSpace(description), suffix) {
				*out = append(*out, path)
			}

			switch kind {
			case "single_nested":
				collectMarked(nestedAttributes(body), path, suffix, out)
			case "list_nested", "set_nested", "map_nested":
				object, _ := body["nested_object"].(map[string]any)
				collectMarked(nestedAttributes(object), path, suffix, out)
			}
		}
	}
}

func nestedAttributes(body map[string]any) []map[string]any {
	list, ok := body["attributes"].([]any)
	if !ok {
		return nil
	}

	out := make([]map[string]any, 0, len(list))

	for _, entry := range list {
		if attribute, ok := entry.(map[string]any); ok {
			out = append(out, attribute)
		}
	}

	return out
}

// checkGenerated fails the build when the schema generator skipped something the
// configuration asked for, which otherwise shows up as a resource silently
// missing from the provider.
func checkGenerated(config *generatorConfig, rawJSON *markedAttributes) error {
	var missing []string

	for _, name := range slices.Sorted(maps.Keys(config.Resources)) {
		if _, ok := rawJSON.resources[name]; !ok {
			missing = append(missing, "resource "+name)
		}
	}

	for _, name := range slices.Sorted(maps.Keys(config.DataSources)) {
		if _, ok := rawJSON.dataSources[name]; !ok {
			missing = append(missing, "data source "+name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("the schema generator produced nothing for: %s", strings.Join(missing, ", "))
	}

	return nil
}

// checkRawGating keeps the two halves of the opaque-configuration policy honest:
// an object whose schema ends up carrying its configuration as an opaque document
// has to be gated behind an opt-in and named so a typed replacement can take the
// plain name later, and a gate with nothing to guard is stale.
func checkRawGating(config *generatorConfig, rawJSON *markedAttributes) error {
	var problems []string

	check := func(kind, name, feature string, paths []string) {
		opaque := hasOpaqueConfig(paths)

		switch {
		case opaque && feature == "":
			problems = append(problems, fmt.Sprintf(
				"%s %s carries its configuration as an opaque document (%s): rename it to %s%s and set x_terraform.raw_feature",
				kind, name, strings.Join(opaqueConfigPaths(paths), ", "), strings.TrimSuffix(name, rawAttributeSuffix), rawAttributeSuffix))
		case opaque && !strings.HasSuffix(name, rawAttributeSuffix):
			problems = append(problems, fmt.Sprintf(
				"%s %s is gated behind raw_feature %q but is not named with a %q suffix, so a typed replacement could not take its place",
				kind, name, feature, rawAttributeSuffix))
		// An object can be gated for a reason other than an opaque document — a
		// shorthand the provider cannot express, say — so a gate with no opaque
		// attribute behind it is fine. What it still has to do is leave the plain
		// name free for the version that models the object in full.
		case !opaque && feature != "" && !strings.HasSuffix(name, rawAttributeSuffix):
			problems = append(problems, fmt.Sprintf(
				"%s %s is gated behind raw_feature %q, so it has to be named %s%s to leave the plain name free",
				kind, name, feature, name, rawAttributeSuffix))
		}
	}

	for _, name := range slices.Sorted(maps.Keys(config.Resources)) {
		check("resource", name, config.Resources[name].Terraform.RawFeature, rawJSON.resources[name])
	}

	for _, name := range slices.Sorted(maps.Keys(config.DataSources)) {
		check("data source", name, config.DataSources[name].Terraform.RawFeature, rawJSON.dataSources[name])
	}

	if len(problems) > 0 {
		return fmt.Errorf("opaque configuration is not gated correctly:\n\t%s", strings.Join(problems, "\n\t"))
	}

	return nil
}

func hasOpaqueConfig(paths []string) bool {
	return len(opaqueConfigPaths(paths)) > 0
}

func opaqueConfigPaths(paths []string) []string {
	var out []string

	for _, path := range paths {
		segments := strings.Split(path, ".")

		if strings.HasSuffix(segments[len(segments)-1], rawAttributeSuffix) {
			out = append(out, path)
		}
	}

	return out
}

// rawFeatures lists the opt-ins the provider has to offer, deduplicated.
func rawFeatures(config *generatorConfig) []string {
	features := map[string]struct{}{}

	for _, resource := range config.Resources {
		if resource.Terraform.RawFeature != "" {
			features[resource.Terraform.RawFeature] = struct{}{}
		}
	}

	for _, dataSource := range config.DataSources {
		if dataSource.Terraform.RawFeature != "" {
			features[dataSource.Terraform.RawFeature] = struct{}{}
		}
	}

	return slices.Sorted(maps.Keys(features))
}

func renderRegistry(config *generatorConfig, rawJSON, variantBlocks *markedAttributes, searchable searchablePaths, variants map[string]*variantMeta, renames map[string]string, modulePath string) ([]byte, error) {
	resourceNames := slices.Sorted(maps.Keys(config.Resources))
	dataSourceNames := slices.Sorted(maps.Keys(config.DataSources))

	var out strings.Builder

	out.WriteString("// Code generated by //mgmt/tf-provider/tools/tfgen. DO NOT EDIT.\n\n")
	out.WriteString("// Package registry lists the resources and data sources the provider exposes.\n")
	out.WriteString("//\n")
	out.WriteString("// It is generated from generator_config.yml, the same file that drives the\n")
	out.WriteString("// schema generators, so the two cannot drift apart.\n")
	out.WriteString("package registry\n\n")

	out.WriteString("import (\n")
	fmt.Fprintf(&out, "\t%q\n\n", modulePath+"/internal/genresource")

	for _, name := range resourceNames {
		fmt.Fprintf(&out, "\t%s %q\n", resourceAlias(name), fmt.Sprintf("%s/%s/resource_%s", modulePath, resourcesDir, name))
	}

	out.WriteString("\n")

	for _, name := range dataSourceNames {
		fmt.Fprintf(&out, "\t%s %q\n", dataSourceAlias(name), fmt.Sprintf("%s/%s/datasource_%s", modulePath, dataSourcesDir, name))
	}

	out.WriteString(")\n\n")

	out.WriteString("// Resources lists every managed resource the provider exposes.\n")
	out.WriteString("func Resources() []genresource.Definition {\n")
	out.WriteString("\treturn []genresource.Definition{\n")

	for _, name := range resourceNames {
		entry, err := renderResource(name, config.Resources[name], rawJSON.resources[name], variantBlocks.resources[name], variants, len(renames) > 0)
		if err != nil {
			return nil, err
		}

		out.WriteString(entry)
	}

	out.WriteString("\t}\n}\n\n")

	out.WriteString("// DataSources lists every data source the provider exposes.\n")
	out.WriteString("func DataSources() []genresource.DataSourceDefinition {\n")
	out.WriteString("\treturn []genresource.DataSourceDefinition{\n")

	for _, name := range dataSourceNames {
		entry, err := renderDataSource(name, config.DataSources[name], rawJSON.dataSources[name], variantBlocks.dataSources[name], searchable, variants, len(renames) > 0)
		if err != nil {
			return nil, err
		}

		out.WriteString(entry)
	}

	out.WriteString("\t}\n}\n\n")

	if len(renames) > 0 {
		out.WriteString("// fieldNames maps an API field to the attribute name it is exposed under.\n")
		out.WriteString("// Terraform reserves a few attribute names outright, so a field the API calls\n")
		out.WriteString("// `provider` has to reach a practitioner under another name and be translated\n")
		out.WriteString("// back on the wire.\n")
		out.WriteString("var fieldNames = map[string]string{\n")

		for _, from := range slices.Sorted(maps.Keys(renames)) {
			fmt.Fprintf(&out, "\t%q: %q,\n", from, renames[from])
		}

		out.WriteString("}\n\n")
	}

	out.WriteString("// RawFeatures lists the objects whose configuration is still an opaque JSON\n")
	out.WriteString("// document. Each one becomes an `enable_raw_<feature>` provider\n")
	out.WriteString("// argument that the objects behind it refuse to work without.\n")
	out.WriteString("func RawFeatures() []string {\n")
	fmt.Fprintf(&out, "\treturn %s\n}\n", renderStrings(rawFeatures(config)))

	formatted, err := format.Source([]byte(out.String()))
	if err != nil {
		return nil, fmt.Errorf("formatting the generated registry: %w", err)
	}

	return formatted, nil
}

func renderResource(name string, config resourceConfig, rawJSON, variantBlocks []string, variants map[string]*variantMeta, renamed bool) (string, error) {
	required := []struct {
		label     string
		operation *operationConfig
	}{
		{"create", config.Create},
		{"read", config.Read},
		{"delete", config.Delete},
	}

	for _, entry := range required {
		if entry.operation == nil || entry.operation.Path == "" || entry.operation.Method == "" {
			return "", fmt.Errorf("resource %s: a %s operation with a path and a method is required", name, entry.label)
		}
	}

	var out strings.Builder

	out.WriteString("\t\t{\n")
	fmt.Fprintf(&out, "\t\t\tName:   %q,\n", name)
	fmt.Fprintf(&out, "\t\t\tSchema: %s.%sResourceSchema,\n", resourceAlias(name), pascal(name))
	fmt.Fprintf(&out, "\t\t\tCreate: %s,\n", renderOperation(config.Create))
	fmt.Fprintf(&out, "\t\t\tRead:   %s,\n", renderOperation(config.Read))

	if config.Update != nil {
		fmt.Fprintf(&out, "\t\t\tUpdate: %s,\n", renderOperation(config.Update))
	}

	fmt.Fprintf(&out, "\t\t\tDelete: %s,\n", renderOperation(config.Delete))

	if len(rawJSON) > 0 {
		fmt.Fprintf(&out, "\t\t\tRawJSONAttributes: %s,\n", renderStrings(rawJSON))
	}

	if len(variantBlocks) > 0 {
		fmt.Fprintf(&out, "\t\t\tVariantBlocks: %s,\n", renderStrings(variantBlocks))
	}

	if config.Terraform.RawFeature != "" {
		fmt.Fprintf(&out, "\t\t\tRawFeature: %q,\n", config.Terraform.RawFeature)
	}

	if variant := renderVariant(config.Read.Path, config.Terraform.Variant, variants); variant != "" {
		fmt.Fprintf(&out, "\t\t\tVariant: %s,\n", variant)
	}

	if renamed {
		out.WriteString("\t\t\tFieldNames: fieldNames,\n")
	}

	out.WriteString("\t\t},\n")

	return out.String(), nil
}

func renderDataSource(name string, config dataSourceConfig, rawJSON, variantBlocks []string, searchable searchablePaths, variants map[string]*variantMeta, renamed bool) (string, error) {
	if config.Read == nil || config.Read.Path == "" || config.Read.Method == "" {
		return "", fmt.Errorf("data source %s: a read operation with a path and a method is required", name)
	}

	var out strings.Builder

	out.WriteString("\t\t{\n")
	fmt.Fprintf(&out, "\t\t\tName:   %q,\n", name)
	fmt.Fprintf(&out, "\t\t\tSchema: %s.%sDataSourceSchema,\n", dataSourceAlias(name), pascal(name))
	fmt.Fprintf(&out, "\t\t\tRead:   %s,\n", renderOperation(config.Read))

	// A data source addressing one object can also find it by filter, as long as
	// the API lets its collection be filtered.
	if search := searchable.searchPath(config.Read.Path); search != "" {
		fmt.Fprintf(&out, "\t\t\tSearch: genresource.Operation{Method: \"GET\", Path: %q},\n", search)
	}

	if len(rawJSON) > 0 {
		fmt.Fprintf(&out, "\t\t\tRawJSONAttributes: %s,\n", renderStrings(rawJSON))
	}

	if len(variantBlocks) > 0 {
		fmt.Fprintf(&out, "\t\t\tVariantBlocks: %s,\n", renderStrings(variantBlocks))
	}

	if config.Terraform.RawFeature != "" {
		fmt.Fprintf(&out, "\t\t\tRawFeature: %q,\n", config.Terraform.RawFeature)
	}

	if variant := renderVariant(config.Read.Path, config.Terraform.Variant, variants); variant != "" {
		fmt.Fprintf(&out, "\t\t\tVariant: %s,\n", variant)
	}

	if renamed {
		out.WriteString("\t\t\tFieldNames: fieldNames,\n")
	}

	out.WriteString("\t\t},\n")

	return out.String(), nil
}

func renderOperation(operation *operationConfig) string {
	return fmt.Sprintf("genresource.Operation{Method: %q, Path: %q}", strings.ToUpper(operation.Method), operation.Path)
}

func renderStrings(values []string) string {
	quoted := make([]string, 0, len(values))

	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("%q", value))
	}

	return "[]string{" + strings.Join(quoted, ", ") + "}"
}

func resourceAlias(name string) string {
	return "res" + pascal(name)
}

func dataSourceAlias(name string) string {
	return "ds" + pascal(name)
}

// pascal mirrors how the framework generator names the schema function for an
// object: every "_"-separated part capitalised, e.g. ntp_config becomes
// NtpConfig.
func pascal(name string) string {
	parts := strings.Split(name, "_")

	for i, part := range parts {
		if part == "" {
			continue
		}

		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}

	return strings.Join(parts, "")
}
