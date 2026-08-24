package main

import (
	"fmt"
	"maps"
	"os"
	"slices"

	"gopkg.in/yaml.v3"
)

// The keys of the provider's own extension to the generator configuration.
const (
	terraformKey = "x_terraform"
	variantKey   = "variant"
	// elementKey marks a data source that stands for one element of the collection
	// its read path names, for an object the API serves no single-object read for.
	elementKey = "element"
	// renamesKey holds the document-wide field renames, at the top level of the
	// file rather than on an object, because the fields they name are shared.
	renamesKey = "x_terraform_renames"
	// dropsKey holds the document-wide field removals that have to run before
	// renames, for the same reason renames live at the top level: the field
	// being dropped is shared with whatever schema is taking its name over.
	dropsKey = "x_terraform_drops"
)

// readRenames reads the fields whose Terraform name has to differ from the name
// the API uses.
//
// Terraform reserves a handful of attribute names at the root of a resource —
// `provider`, `count`, `lifecycle` and friends — and a schema declaring one is
// rejected outright, taking the whole provider with it. A cloud account really
// does have a field called `provider`, and the API requires it on create, so
// dropping it is not an option either: the only way to expose the field is under
// another name, with the runtime translating.
//
// The same map also carries a field superseded by a differently-named
// replacement — `definitions_v2` taking over `definitions`, say — once
// `x_terraform_drops` has cleared the old name out of the way.
func readRenames(path string) (map[string]string, error) {
	doc, err := readYAML(path)
	if err != nil {
		return nil, err
	}

	raw, _ := doc[renamesKey].(map[string]any)
	out := make(map[string]string, len(raw))

	for from, to := range raw {
		if name := stringOr(to); name != "" {
			out[from] = name
		}
	}

	return out, nil
}

// readDrops reads the fields removed from every schema before renames run. A
// field superseded by a differently-named replacement is dropped so the
// replacement can be renamed onto the vacated name without the collision guard
// in renameReservedProperties refusing the move.
func readDrops(path string) ([]string, error) {
	doc, err := readYAML(path)
	if err != nil {
		return nil, err
	}

	raw, _ := doc[dropsKey].([]any)
	out := make([]string, 0, len(raw))

	for _, entry := range raw {
		if name := stringOr(entry); name != "" {
			out = append(out, name)
		}
	}

	return out, nil
}

// readElementClaims lists the collections a data source stands for one element
// of, which prep needs because it has to lift the element's schema onto a path of
// its own for the generator to map.
//
// The claim is explicit rather than worked out from the paths, so that the
// configuration goes on naming endpoints the API has: a collection read is what
// the API offers for one of these objects, and an entry saying "one element of
// that" says what the provider does with it. Nothing has to be believed about a
// path that is not in the spec.
func readElementClaims(path string) ([]string, error) {
	doc, err := readYAML(path)
	if err != nil {
		return nil, err
	}

	objects, _ := doc["data_sources"].(map[string]any)

	var out []string

	for _, name := range slices.Sorted(maps.Keys(objects)) {
		object, ok := objects[name].(map[string]any)
		if !ok || !objectElement(object) {
			continue
		}

		read, ok := object["read"].(map[string]any)
		if !ok {
			continue
		}

		if readPath := stringOr(read["path"]); readPath != "" {
			out = append(out, readPath)
		}
	}

	return out, nil
}

// variantClaims lists, per API path, the branches generator_config.yml binds to a
// Terraform type of their own.
type variantClaims map[string][]string

// readVariantClaims finds every object that claims a branch of the schemas behind
// its endpoints. An object without a claim takes whatever the schema is, with any
// branching turned into one attribute per branch instead.
func readVariantClaims(path string) (variantClaims, error) {
	doc, err := readYAML(path)
	if err != nil {
		return nil, err
	}

	claims := variantClaims{}

	for _, kind := range []string{"resources", "data_sources"} {
		objects, _ := doc[kind].(map[string]any)

		for _, name := range slices.Sorted(maps.Keys(objects)) {
			object, ok := objects[name].(map[string]any)
			if !ok {
				continue
			}

			variant := objectVariant(object)
			if variant == "" {
				continue
			}

			for _, operation := range operationPaths(object) {
				if !slices.Contains(claims[operation], variant) {
					claims[operation] = append(claims[operation], variant)
				}
			}
		}
	}

	return claims, nil
}

func objectVariant(object map[string]any) string {
	extension, _ := object[terraformKey].(map[string]any)

	return stringOr(extension[variantKey])
}

func objectElement(object map[string]any) bool {
	extension, _ := object[terraformKey].(map[string]any)
	element, _ := extension[elementKey].(bool)

	return element
}

func operationPaths(object map[string]any) []string {
	var out []string

	for _, key := range slices.Sorted(maps.Keys(object)) {
		operation, ok := object[key].(map[string]any)
		if !ok {
			continue
		}

		if path := stringOr(operation["path"]); path != "" {
			out = append(out, path)
		}
	}

	return out
}

// writeGeneratorConfig renders the configuration the HashiCorp generator takes.
//
// The generator maps one schema per path, so an object whose schema prep lifted
// onto a path of its own has to point at that path: a branch of a composed schema,
// or one element of a collection. Nothing else in the file changes, and the
// provider's own extension is dropped because the generator has no use for it.
func writeGeneratorConfig(configPath, outPath string) error {
	doc, err := readYAML(configPath)
	if err != nil {
		return err
	}

	delete(doc, renamesKey)
	delete(doc, dropsKey)

	for _, kind := range []string{"resources", "data_sources"} {
		objects, _ := doc[kind].(map[string]any)

		for _, name := range slices.Sorted(maps.Keys(objects)) {
			object, ok := objects[name].(map[string]any)
			if !ok {
				continue
			}

			variant := objectVariant(object)
			element := objectElement(object)
			delete(object, terraformKey)

			if variant == "" && !element {
				continue
			}

			for _, key := range slices.Sorted(maps.Keys(object)) {
				operation, ok := object[key].(map[string]any)
				if !ok {
					continue
				}

				path := stringOr(operation["path"])
				if path == "" {
					continue
				}

				// Only one of the two applies: `tfgen registry` refuses an object that
				// claims both, because prep takes a variant of the collection rather
				// than of the element lifted out of it.
				if element {
					path += ElementSuffix
				} else {
					path += VariantSeparator + variant
				}

				operation["path"] = path
			}
		}
	}

	encoded, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", outPath, err)
	}

	if err := os.WriteFile(outPath, encoded, 0o600); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}

	return nil
}
