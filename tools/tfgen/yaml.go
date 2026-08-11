package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func readYAML(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	return doc, nil
}

func writeYAML(path string, doc map[string]any) error {
	encoded, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", path, err)
	}

	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}
