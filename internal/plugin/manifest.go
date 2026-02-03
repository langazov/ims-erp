package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ManifestParser handles parsing of plugin manifest files
type ManifestParser struct {
	schemaValidator SchemaValidator
}

// SchemaValidator validates manifest against a schema
type SchemaValidator interface {
	Validate(manifest *PluginManifest) error
}

// NewManifestParser creates a new manifest parser
func NewManifestParser() *ManifestParser {
	return &ManifestParser{
		schemaValidator: &defaultSchemaValidator{},
	}
}

// ParseFile parses a manifest from a file path
func (p *ManifestParser) ParseFile(path string) (*PluginManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	return p.Parse(data)
}

// Parse parses manifest data
func (p *ManifestParser) Parse(data []byte) (*PluginManifest, error) {
	var manifest PluginManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Validate
	if err := p.schemaValidator.Validate(&manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	return &manifest, nil
}

// ParseDirectory scans a directory for plugin manifests
func (p *ManifestParser) ParseDirectory(dir string) ([]*PluginManifest, error) {
	manifests := []*PluginManifest{}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == "plugin.yaml" || info.Name() == "plugin.yml" {
			manifest, err := p.ParseFile(path)
			if err != nil {
				return fmt.Errorf("failed to parse manifest at %s: %w", path, err)
			}
			manifests = append(manifests, manifest)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return manifests, nil
}

type defaultSchemaValidator struct{}

func (v *defaultSchemaValidator) Validate(manifest *PluginManifest) error {
	// Required fields
	if manifest.Name == "" {
		return fmt.Errorf("plugin name is required")
	}

	if manifest.Version == "" {
		return fmt.Errorf("plugin version is required")
	}

	if manifest.EntryPoint == "" {
		return fmt.Errorf("plugin entry point is required")
	}

	// Validate version format (semver-like)
	if !isValidVersion(manifest.Version) {
		return fmt.Errorf("invalid version format: %s", manifest.Version)
	}

	// Validate permissions
	for _, perm := range manifest.Permissions {
		if perm.Resource == "" || perm.Action == "" {
			return fmt.Errorf("permission must have resource and action")
		}
	}

	// Validate dependencies
	for _, dep := range manifest.Dependencies {
		if dep.Name == "" {
			return fmt.Errorf("dependency name is required")
		}
	}

	// Validate routes
	for _, route := range manifest.Routes {
		if route.Path == "" || route.Method == "" {
			return fmt.Errorf("route must have path and method")
		}
		if route.Handler == "" {
			return fmt.Errorf("route must have handler")
		}
	}

	return nil
}

func isValidVersion(version string) bool {
	// Simple semver validation: major.minor.patch
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return false
	}

	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}

	return true
}
