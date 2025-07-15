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
	Extras map[string]any            `json:",inline"`
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
	RPCMethod string
	IsRequest bool
	MainType  string
}

type ZodGenerator struct {
	schemas       map[string]SchemaInfo
	allDefs       map[string]SchemaProperty
	commonDefs    map[string]SchemaProperty
	requestDefs   map[string]SchemaProperty
	responseDefs  map[string]SchemaProperty
	requestTypes  map[string]string // typeName -> rpcMethod
	responseTypes map[string]string // typeName -> rpcMethod
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

	// Categorize definitions
	generator.CategorizeDefinitions()

	// Generate TypeScript files
	if err := generator.GenerateAllFiles(outDir); err != nil {
		logger.Fatal("Failed to generate TypeScript files", "err", err)
	}

	logger.Info("Generated Zod TypeScript files", "dir", outDir)
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
			return fmt.Errorf("failed to read directory %s: %w", dir.path, err)
		}

		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".json") {
				continue
			}

			path := filepath.Join(dir.path, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}

			var schema JSONSchema
			if err := json.Unmarshal(data, &schema); err != nil {
				return fmt.Errorf("failed to parse JSON schema %s: %w", path, err)
			}

			// Extract RPC method from schema extras
			rpcMethod := ""

			// First, try to unmarshal into a map to get the rpc_method
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

			g.schemas[file.Name()] = SchemaInfo{
				Schema:    schema,
				RPCMethod: rpcMethod,
				IsRequest: dir.isRequest,
				MainType:  mainType,
			}

			// Merge definitions
			for name, def := range schema.Defs {
				g.allDefs[name] = def
			}
		}
	}

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
			deps := g.getDependencies(def)
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

func (g *ZodGenerator) generateCommonFile(outDir string) error {
	var sb strings.Builder

	// Add imports
	sb.WriteString("import { z } from 'zod';\n")
	sb.WriteString("import { Address } from 'viem';\n\n")

	// Add common schemas
	sb.WriteString("// Common schemas used by both requests and responses\n\n")

	// Add built-in common schemas
	sb.WriteString("export const addressSchema = z.string().refine((val) => /^0x[0-9a-fA-F]{40}$/.test(val), {\n")
	sb.WriteString("  message: 'Must be a 0x-prefixed hex string of 40 hex chars (EVM address)',\n")
	sb.WriteString("});\n\n")

	sb.WriteString("export const hexSchema = z.string().refine((val) => /^0x[0-9a-fA-F]*$/.test(val), {\n")
	sb.WriteString("  message: 'Must be a 0x-prefixed hex string',\n")
	sb.WriteString("});\n\n")

	// Generate common definitions
	definitionNames := g.getSortedDefinitionNames(g.commonDefs)
	for _, name := range definitionNames {
		def := g.commonDefs[name]
		zodSchema := g.generateZodSchema(def)
		sb.WriteString(fmt.Sprintf("export const %sSchema = %s;\n\n", name, zodSchema))
	}

	// Write to file
	outputPath := filepath.Join(outDir, "common_gen.ts")
	return os.WriteFile(outputPath, []byte(sb.String()), 0o644)
}

func (g *ZodGenerator) generateRequestFile(outDir string) error {
	var sb strings.Builder

	// Add imports
	sb.WriteString("import { z } from 'zod';\n")
	sb.WriteString("import { RPCMethod } from '../sdk/src/rpc/types';\n")
	sb.WriteString("import { addressSchema, hexSchema } from './common_gen';\n")

	// Import common schemas
	commonNames := g.getSortedDefinitionNames(g.commonDefs)
	if len(commonNames) > 0 {
		sb.WriteString("import {\n")
		for i, name := range commonNames {
			sb.WriteString(fmt.Sprintf("  %sSchema", name))
			if i < len(commonNames)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("} from './common_gen';\n")
	}
	sb.WriteString("\n")

	// Add request-specific schemas
	sb.WriteString("// Request schemas\n\n")

	definitionNames := g.getSortedDefinitionNames(g.requestDefs)
	for _, name := range definitionNames {
		def := g.requestDefs[name]
		zodSchema := g.generateZodSchema(def)
		sb.WriteString(fmt.Sprintf("export const %sSchema = %s;\n\n", name, zodSchema))
	}

	// Generate parser mapping
	sb.WriteString("// Request parser mapping\n")
	sb.WriteString("export const requestParsers: Record<string, (params: any) => any> = {\n")

	for typeName, rpcMethod := range g.requestTypes {
		sb.WriteString(fmt.Sprintf("  [RPCMethod.%s]: (params) => %sSchema.parse(params),\n",
			g.rpcMethodToEnumName(rpcMethod), typeName))
	}

	sb.WriteString("};\n")

	// Write to file
	outputPath := filepath.Join(outDir, "requests_gen.ts")
	return os.WriteFile(outputPath, []byte(sb.String()), 0o644)
}

func (g *ZodGenerator) generateResponseFile(outDir string) error {
	var sb strings.Builder

	// Add imports
	sb.WriteString("import { z } from 'zod';\n")
	sb.WriteString("import { RPCMethod } from '../sdk/src/rpc/types';\n")
	sb.WriteString("import { addressSchema, hexSchema } from './common_gen';\n")

	// Import common schemas
	commonNames := g.getSortedDefinitionNames(g.commonDefs)
	if len(commonNames) > 0 {
		sb.WriteString("import {\n")
		for i, name := range commonNames {
			sb.WriteString(fmt.Sprintf("  %sSchema", name))
			if i < len(commonNames)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("} from './common_gen';\n")
	}
	sb.WriteString("\n")

	// Add response-specific schemas
	sb.WriteString("// Response schemas\n\n")

	definitionNames := g.getSortedDefinitionNames(g.responseDefs)
	for _, name := range definitionNames {
		def := g.responseDefs[name]
		zodSchema := g.generateZodSchema(def)
		sb.WriteString(fmt.Sprintf("export const %sSchema = %s;\n\n", name, zodSchema))
	}

	// Generate parser mapping
	sb.WriteString("// Response parser mapping\n")
	sb.WriteString("export const responseParsers: Record<string, (params: any) => any> = {\n")

	for typeName, rpcMethod := range g.responseTypes {
		sb.WriteString(fmt.Sprintf("  [RPCMethod.%s]: (params) => %sSchema.parse(params),\n",
			g.rpcMethodToEnumName(rpcMethod), typeName))
	}

	sb.WriteString("};\n")

	// Write to file
	outputPath := filepath.Join(outDir, "response_gen.ts")
	return os.WriteFile(outputPath, []byte(sb.String()), 0o644)
}

func (g *ZodGenerator) rpcMethodToEnumName(method string) string {
	// Convert snake_case to PascalCase
	parts := strings.Split(method, "_")
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}
	return strings.Join(parts, "")
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
		deps := g.getDependencies(def)
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

