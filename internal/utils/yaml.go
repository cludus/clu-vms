package utils

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ParseYaml parses YAML content into the provided out object.
func ParseYaml(content string, out any) error {
	if err := yaml.Unmarshal([]byte(content), out); err != nil {
		return fmt.Errorf("failed to parse yaml: %w", err)
	}
	return nil
}

// SerializeYaml serializes the provided object into YAML content.
func SerializeYaml(in any) (string, error) {
	bytes, err := yaml.Marshal(in)
	if err != nil {
		return "", fmt.Errorf("failed to serialize yaml: %w", err)
	}
	return string(bytes), nil
}
