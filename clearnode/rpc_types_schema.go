package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

type JSONSchema struct {
	Schema string                    `json:"$schema"`
	Ref    string                    `json:"$ref"`
	Defs   map[string]SchemaProperty `json:"$defs"`
}

type SchemaProperty struct {
	Type                 string                    `json:"type"`
	Format               string                    `json:"format"`
	Properties           map[string]SchemaProperty `json:"properties"`
	Required             []string                  `json:"required"`
	AdditionalProperties bool                      `json:"additionalProperties"`
	Ref                  string                    `json:"$ref"`
	Enum                 []string                  `json:"enum"`
}

type SchemaInfo struct {
	Schema    JSONSchema
	RPCMethod RPCMethod
	IsRequest bool
	MainType  string
}

// LoadSchemas loads JSON schemas from request and response directories
func LoadSchemas(requestDir, responseDir string) (map[string]SchemaInfo, map[string]SchemaProperty, error) {
	schemas := make(map[string]SchemaInfo)
	allDefs := make(map[string]SchemaProperty)

	dirs := []struct {
		path      string
		isRequest bool
	}{
		{requestDir, true},
		{responseDir, false},
	}

	for _, dir := range dirs {
		files, err := os.ReadDir(dir.path)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read directory %s: %w", dir.path, err)
		}

		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".json") {
				continue
			}

			path := filepath.Join(dir.path, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to read file %s: %w", path, err)
			}

			var schema JSONSchema
			if err := json.Unmarshal(data, &schema); err != nil {
				return nil, nil, fmt.Errorf("failed to parse JSON schema %s: %w", path, err)
			}

			// Extract RPC method from schema extras
			var rpcMethod string
			var schemaMap map[string]any
			if err := json.Unmarshal(data, &schemaMap); err == nil {
				if method, ok := schemaMap["rpc_method"].(string); ok {
					rpcMethod = method
				}
			}

			// Extract main type from schema reference
			mainType := ""
			if schema.Ref != "" {
				parts := strings.Split(schema.Ref, "/")
				if len(parts) >= 3 {
					mainType = parts[len(parts)-1]
				}
			}

			schemas[file.Name()] = SchemaInfo{
				Schema:    schema,
				RPCMethod: RPCMethod(rpcMethod),
				IsRequest: dir.isRequest,
				MainType:  mainType,
			}

			// Merge definitions
			maps.Copy(allDefs, schema.Defs)
		}
	}

	return schemas, allDefs, nil
}

// GetDependencies recursively finds all dependencies of a schema property
func GetDependencies(prop SchemaProperty) []string {
	var deps []string

	if prop.Ref != "" {
		parts := strings.Split(prop.Ref, "/")
		if len(parts) >= 3 {
			deps = append(deps, parts[len(parts)-1])
		}
	}

	for _, subProp := range prop.Properties {
		deps = append(deps, GetDependencies(subProp)...)
	}

	return deps
}
