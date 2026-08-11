// Command tfgen runs the two code generation steps that surround the HashiCorp
// OpenAPI and framework generators.
//
//	tfgen prep      normalises the bundled OpenAPI document into something the
//	                provider spec generator can map in full
//	tfgen registry  turns generator_config.yml plus the generated provider code
//	                specification into the provider's resource registry
//
// Both steps live in one binary so they share the marker the first one writes
// and the second one reads.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "tfgen:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("expected a subcommand: prep or registry")
	}

	switch args[0] {
	case "prep":
		return runPrep(args[1:])
	case "registry":
		return runRegistry(args[1:])
	default:
		return fmt.Errorf("unknown subcommand %q, expected prep or registry", args[0])
	}
}

func runPrep(args []string) error {
	flags := flag.NewFlagSet("prep", flag.ExitOnError)
	in := flags.String("in", "", "path to the bundled OpenAPI document")
	config := flags.String("config", "", "path to generator_config.yml, read for the variants it claims")
	out := flags.String("out", "", "path to write the normalised OpenAPI document to")
	outConfig := flags.String("out-config", "", "path to write the generated HashiCorp generator configuration to")
	strict := flags.Bool("strict", false, "fail when a construct could not be normalised")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *in == "" || *out == "" {
		return errors.New("prep: both -in and -out are required")
	}

	doc, err := readYAML(*in)
	if err != nil {
		return err
	}

	prep := NewPrep(doc)

	if *config != "" {
		claims, err := readVariantClaims(*config)
		if err != nil {
			return err
		}

		for path, variants := range claims {
			for _, variant := range variants {
				prep.Claim(path, variant)
			}
		}

		renames, err := readRenames(*config)
		if err != nil {
			return err
		}

		for from, to := range renames {
			prep.Rename(from, to)
		}
	}

	prep.Run()

	for _, warning := range prep.Warnings {
		fmt.Fprintln(os.Stderr, "tfgen: prep:", warning)
	}

	if *strict && len(prep.Warnings) > 0 {
		return fmt.Errorf("prep: %d construct(s) could not be normalised", len(prep.Warnings))
	}

	if err := writeYAML(*out, doc); err != nil {
		return err
	}

	if *outConfig != "" {
		return writeGeneratorConfig(*config, *outConfig)
	}

	return nil
}

func runRegistry(args []string) error {
	flags := flag.NewFlagSet("registry", flag.ExitOnError)
	config := flags.String("config", "", "path to generator_config.yml")
	spec := flags.String("spec", "", "path to the generated provider code specification")
	openapi := flags.String("openapi", "", "path to the normalised OpenAPI document, used to find the filterable collections")
	module := flags.String("module", "", "module path the generated packages live under")
	out := flags.String("out", "", "path to write the generated registry to")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *config == "" || *spec == "" || *openapi == "" || *module == "" || *out == "" {
		return errors.New("registry: -config, -spec, -openapi, -module and -out are all required")
	}

	source, err := generateRegistry(*config, *spec, *openapi, *module)
	if err != nil {
		return err
	}

	if err := os.WriteFile(*out, source, 0o600); err != nil {
		return fmt.Errorf("writing %s: %w", *out, err)
	}

	return nil
}
