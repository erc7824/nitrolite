package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
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

type ZodGenerator struct {
	schemas map[string]JSONSchema
	defs    map[string]SchemaProperty
}

func runZodGeneratorCli(logger Logger) {
	logger = logger.NewSystem("zod-generator")
	if len(os.Args) < 3 {
		logger.Fatal("Usage: clearnode zod-generator <out_dir>")
	}

	outDir := os.Args[2]
	generator := NewZodGenerator()

	// Load schemas from request and response directories
	requestDir := filepath.Join(outDir, "request")
	responseDir := filepath.Join(outDir, "response")

	if err := generator.LoadSchemas(requestDir, responseDir); err != nil {
		logger.Fatal("Failed to load schemas", "err", err)
	}

	// Generate TypeScript code
	tsCode, err := generator.GenerateTypeScript()
	if err != nil {
		logger.Fatal("Failed to generate TypeScript", "err", err)
	}

	// Write to output file
	outputPath := filepath.Join(outDir, "generated.ts")
	if err := os.WriteFile(outputPath, []byte(tsCode), 0o644); err != nil {
		logger.Fatal("Failed to write output file", "path", outputPath, "err", err)
	}

	logger.Info("Generated Zod TypeScript file", "path", outputPath)
}

func NewZodGenerator() *ZodGenerator {
	return &ZodGenerator{
		schemas: make(map[string]JSONSchema),
		defs:    make(map[string]SchemaProperty),
	}
}

func (g *ZodGenerator) LoadSchemas(requestDir, responseDir string) error {
	dirs := []string{requestDir, responseDir}

	for _, dir := range dirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %w", dir, err)
		}

		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".json") {
				continue
			}

			path := filepath.Join(dir, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}

			var schema JSONSchema
			if err := json.Unmarshal(data, &schema); err != nil {
				return fmt.Errorf("failed to parse JSON schema %s: %w", path, err)
			}

			g.schemas[file.Name()] = schema

			// Merge definitions
			for name, def := range schema.Defs {
				g.defs[name] = def
			}
		}
	}

	return nil
}

func (g *ZodGenerator) GenerateTypeScript() (string, error) {
	var sb strings.Builder

	// Add imports
	sb.WriteString("import { z } from 'zod';\n")
	sb.WriteString("import { Address } from 'viem';\n")
	sb.WriteString("import { addressSchema, hexSchema } from './common';\n\n")

	// Generate schemas for each definition in dependency order
	definitionNames := g.getSortedDefinitionNames()

	for _, name := range definitionNames {
		def := g.defs[name]
		zodSchema := g.generateZodSchema(def)
		sb.WriteString(fmt.Sprintf("export const %sSchema = %s;\n\n", name, zodSchema))
	}

	return sb.String(), nil
}

func (g *ZodGenerator) generateZodSchema(prop SchemaProperty) string {
	switch prop.Type {
	case "string":
		return g.generateStringSchema(prop)
	case "integer":
		return "z.number()"
	case "object":
		return g.generateObjectSchema(prop)
	case "enum":
		return g.generateEnumSchema(prop)
	default:
		if prop.Ref != "" {
			return g.generateRefSchema(prop.Ref)
		}
		return "z.unknown()"
	}
}

func (g *ZodGenerator) generateStringSchema(prop SchemaProperty) string {
	switch prop.Format {
	case "address":
		return "addressSchema"
	case "hex":
		return "hexSchema"
	case "date-time":
		return "z.union([z.string(), z.date()]).transform((v) => new Date(v))"
	case "bignumber":
		return "z.string().transform((v) => BigInt(v))"
	default:
		return "z.string()"
	}
}

func (g *ZodGenerator) generateObjectSchema(prop SchemaProperty) string {
	if len(prop.Properties) == 0 {
		return "z.object({})"
	}

	var sb strings.Builder
	sb.WriteString("z.object({\n")

	propertyNames := make([]string, 0, len(prop.Properties))
	for name := range prop.Properties {
		propertyNames = append(propertyNames, name)
	}
	sort.Strings(propertyNames)

	for i, name := range propertyNames {
		propDef := prop.Properties[name]
		zodSchema := g.generateZodSchema(propDef)

		// Check if property is required
		isRequired := slices.Contains(prop.Required, name)
		if !isRequired {
			zodSchema += ".optional()"
		}

		sb.WriteString(fmt.Sprintf("  %s: %s", name, zodSchema))
		if i < len(propertyNames)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("})")
	return sb.String()
}

func (g *ZodGenerator) generateEnumSchema(prop SchemaProperty) string {
	if len(prop.Enum) == 0 {
		return "z.string()"
	}

	enumValues := make([]string, len(prop.Enum))
	for i, val := range prop.Enum {
		enumValues[i] = fmt.Sprintf("\"%s\"", val)
	}

	return fmt.Sprintf("z.enum([%s])", strings.Join(enumValues, ", "))
}

func (g *ZodGenerator) generateRefSchema(ref string) string {
	// Extract the definition name from the reference
	parts := strings.Split(ref, "/")
	if len(parts) < 3 {
		return "z.unknown()"
	}

	defName := parts[len(parts)-1]
	return fmt.Sprintf("%sSchema", defName)
}

func (g *ZodGenerator) getSortedDefinitionNames() []string {
	// Simple topological sort to handle dependencies
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	result := make([]string, 0, len(g.defs))

	var visit func(string) bool
	visit = func(name string) bool {
		if visiting[name] {
			// Circular dependency, just continue
			return true
		}
		if visited[name] {
			return true
		}

		visiting[name] = true
		def := g.defs[name]

		// Visit dependencies first
		deps := g.getDependencies(def)
		for _, dep := range deps {
			if _, exists := g.defs[dep]; exists {
				visit(dep)
			}
		}

		visiting[name] = false
		visited[name] = true
		result = append(result, name)
		return true
	}

	definitionNames := make([]string, 0, len(g.defs))
	for name := range g.defs {
		definitionNames = append(definitionNames, name)
	}
	sort.Strings(definitionNames)

	for _, name := range definitionNames {
		visit(name)
	}

	return result
}

func (g *ZodGenerator) getDependencies(prop SchemaProperty) []string {
	var deps []string

	if prop.Ref != "" {
		parts := strings.Split(prop.Ref, "/")
		if len(parts) >= 3 {
			deps = append(deps, parts[len(parts)-1])
		}
	}

	for _, subProp := range prop.Properties {
		deps = append(deps, g.getDependencies(subProp)...)
	}

	return deps
}
