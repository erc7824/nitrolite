package main

import (
	"sort"
	"strings"
)

type ZodGenerator struct {
	schemas       map[string]SchemaInfo
	allDefs       map[string]SchemaProperty
	commonDefs    map[string]SchemaProperty
	requestDefs   map[string]SchemaProperty
	responseDefs  map[string]SchemaProperty
	requestTypes  map[string]string // typeName -> rpcMethod
	responseTypes map[string]string // typeName -> rpcMethod
}

func NewZodGenerator() *ZodGenerator {
	return &ZodGenerator{
		schemas:       make(map[string]SchemaInfo),
		allDefs:       make(map[string]SchemaProperty),
		commonDefs:    make(map[string]SchemaProperty),
		requestDefs:   make(map[string]SchemaProperty),
		responseDefs:  make(map[string]SchemaProperty),
		requestTypes:  make(map[string]string),
		responseTypes: make(map[string]string),
	}
}

func (g *ZodGenerator) LoadSchemas(requestDir, responseDir string) error {
	schemas, allDefs, err := LoadSchemas(requestDir, responseDir)
	if err != nil {
		return err
	}

	g.schemas = schemas
	g.allDefs = allDefs
	return nil
}

func (g *ZodGenerator) CategorizeDefinitions() {
	// Track which definitions are used by requests vs responses
	requestUsage := make(map[string]bool)
	responseUsage := make(map[string]bool)

	for _, schemaInfo := range g.schemas {
		usage := requestUsage
		if !schemaInfo.IsRequest {
			usage = responseUsage
		}

		// Mark main type as used
		if schemaInfo.MainType != "" {
			usage[schemaInfo.MainType] = true
		}

		// Track dependencies
		for _, def := range schemaInfo.Schema.Defs {
			deps := GetDependencies(def)
			for _, dep := range deps {
				usage[dep] = true
			}
		}
	}

	// Categorize definitions
	for name, def := range g.allDefs {
		inRequest := requestUsage[name]
		inResponse := responseUsage[name]

		if inRequest && inResponse {
			g.commonDefs[name] = def
		} else if inRequest {
			g.requestDefs[name] = def
		} else if inResponse {
			g.responseDefs[name] = def
		} else {
			// Default to common if unclear
			g.commonDefs[name] = def
		}
	}

	// Build type to RPC method mappings
	for _, schemaInfo := range g.schemas {
		if schemaInfo.MainType != "" && schemaInfo.RPCMethod != "" {
			if schemaInfo.IsRequest {
				g.requestTypes[schemaInfo.MainType] = schemaInfo.RPCMethod
			} else {
				g.responseTypes[schemaInfo.MainType] = schemaInfo.RPCMethod
			}
		}
	}
}

func (g *ZodGenerator) GenerateAllFiles(outDir string) error {
	// Generate common definitions
	if err := g.generateCommonFile(outDir); err != nil {
		return err
	}

	// Generate request definitions
	if err := g.generateRequestFile(outDir); err != nil {
		return err
	}

	// Generate response definitions
	if err := g.generateResponseFile(outDir); err != nil {
		return err
	}

	return nil
}

func (g *ZodGenerator) getSortedDefinitionNames(defs map[string]SchemaProperty) []string {
	// Simple topological sort to handle dependencies
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	result := make([]string, 0, len(defs))

	var visit func(string) bool
	visit = func(name string) bool {
		if visiting[name] {
			return true // circular dependency, just continue
		}
		if visited[name] {
			return true
		}

		visiting[name] = true
		def := defs[name]

		// Visit dependencies first
		deps := GetDependencies(def)
		for _, dep := range deps {
			if _, exists := defs[dep]; exists {
				visit(dep)
			}
		}

		visiting[name] = false
		visited[name] = true
		result = append(result, name)
		return true
	}

	definitionNames := make([]string, 0, len(defs))
	for name := range defs {
		definitionNames = append(definitionNames, name)
	}
	sort.Strings(definitionNames)

	for _, name := range definitionNames {
		visit(name)
	}

	return result
}

func (g *ZodGenerator) rpcMethodToEnumName(method string) string {
	// Convert snake_case to PascalCase
	parts := strings.Split(method, "_")
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}
	return strings.Join(parts, "")
}