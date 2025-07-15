package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// ResponseGenerator handles response-specific code generation
type ResponseGenerator struct {
	responseDefs    map[string]SchemaProperty
	responseTypes   map[string]string // typeName -> rpcMethod
	commonDefs      map[string]SchemaProperty
	zodGenerator    *ZodSchemaGenerator
	sortedDefNames  func(map[string]SchemaProperty) []string
	rpcMethodToEnum func(string) string
}

// NewResponseGenerator creates a new response generator
func NewResponseGenerator(responseDefs map[string]SchemaProperty, responseTypes map[string]string, commonDefs map[string]SchemaProperty, sortedDefNames func(map[string]SchemaProperty) []string, rpcMethodToEnum func(string) string) *ResponseGenerator {
	return &ResponseGenerator{
		responseDefs:    responseDefs,
		responseTypes:   responseTypes,
		commonDefs:      commonDefs,
		zodGenerator:    &ZodSchemaGenerator{},
		sortedDefNames:  sortedDefNames,
		rpcMethodToEnum: rpcMethodToEnum,
	}
}

// GenerateResponsesFile generates the response_gen.ts file
func (r *ResponseGenerator) GenerateResponsesFile(sdkRootDir string) error {
	var sb strings.Builder

	// Add imports
	sb.WriteString("import { z } from 'zod';\n")
	sb.WriteString("import { RPCMethod } from '../types';\n")
	sb.WriteString("import { addressSchema, hexSchema } from './common_gen';\n")

	// Import TypeScript types for transforms
	sb.WriteString("import type {\n")

	// Import common types
	commonNames := r.sortedDefNames(r.commonDefs)
	for _, name := range commonNames {
		sb.WriteString(fmt.Sprintf("  %s,\n", name))
	}

	// Import response-specific types
	definitionNames := r.sortedDefNames(r.responseDefs)
	for _, name := range definitionNames {
		sb.WriteString(fmt.Sprintf("  %s,\n", name))
	}

	sb.WriteString("} from '../types/response';\n")

	// Import common schemas
	if len(commonNames) > 0 {
		sb.WriteString(r.zodGenerator.GenerateCommonSchemaImports(commonNames))
	}
	sb.WriteString("\n")

	// Add response-specific schemas
	sb.WriteString("// Response schemas with camelCase transforms\n\n")

	sb.WriteString(r.generateResponseSchemasWithTransform(definitionNames))

	// Generate parser mapping
	sb.WriteString(r.generateResponseParsers())

	// Ensure directory exists
	outputDir := filepath.Join(sdkRootDir, "src", "rpc", "parse")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", outputDir, err)
	}

	// Write to file
	outputPath := filepath.Join(outputDir, "response_gen.ts")
	return os.WriteFile(outputPath, []byte(sb.String()), 0o644)
}

// generateResponseSchemasWithTransform generates response schemas with camelCase transforms
func (r *ResponseGenerator) generateResponseSchemasWithTransform(definitionNames []string) string {
	var sb strings.Builder

	for _, name := range definitionNames {
		def := r.responseDefs[name]
		if def.Type == "object" {
			zodSchema := r.zodGenerator.GenerateObjectSchemaWithTransform(def, name)
			sb.WriteString(fmt.Sprintf("export const %sSchema = %s;\n\n", name, zodSchema))
		} else {
			zodSchema := r.zodGenerator.GenerateZodSchema(def)
			sb.WriteString(fmt.Sprintf("export const %sSchema = %s;\n\n", name, zodSchema))
		}
	}

	return sb.String()
}

// generateResponseParsers generates the response parser mapping
func (r *ResponseGenerator) generateResponseParsers() string {
	var sb strings.Builder
	sb.WriteString("// Response parser mapping\n")
	sb.WriteString("export const responseParsers: Record<string, (params: any) => any> = {\n")

	for typeName, rpcMethod := range r.responseTypes {
		sb.WriteString(fmt.Sprintf("  [RPCMethod.%s]: (params) => %sSchema.parse(params),\n",
			r.rpcMethodToEnum(rpcMethod), typeName))
	}

	sb.WriteString("};\n")
	return sb.String()
}

// GenerateResponseTypesFile generates sdk/src/rpc/types/response.ts with TypeScript interfaces
func (r *ResponseGenerator) GenerateResponseTypesFile(sdkRootDir string) error {
	var sb strings.Builder

	// Add header comment
	sb.WriteString("// Auto-generated TypeScript response types with camelCase field names\n")
	sb.WriteString("// Generated from JSON schemas\n\n")

	// Add viem imports
	sb.WriteString("import type { Address, Hex } from 'viem';\n")
	sb.WriteString("import {RPCMethod, GenericRPCMessage} from '.';\n\n")

	// Generate common type interfaces
	commonNames := r.sortedDefNames(r.commonDefs)
	for _, name := range commonNames {
		def := r.commonDefs[name]
		// Skip generating interfaces for special types that should be handled as primitives
		if r.shouldSkipInterfaceGeneration(name) {
			continue
		}
		tsInterface := r.generateTypeScriptInterface(name, def)
		sb.WriteString(tsInterface)
	}

	// Generate response-specific type interfaces
	definitionNames := r.sortedDefNames(r.responseDefs)
	for _, name := range definitionNames {
		def := r.responseDefs[name]
		// Skip generating interfaces for special types that should be handled as primitives
		if r.shouldSkipInterfaceGeneration(name) {
			continue
		}

		// Generate Request structure first
		if rpcMethod, exists := r.responseTypes[name]; exists {
			requestInterface := r.generateRequestInterface(name, rpcMethod)
			sb.WriteString(requestInterface)
		}

		// Then generate Params interface
		tsInterface := r.generateTypeScriptInterface(name, def)
		sb.WriteString(tsInterface)
	}

	// Generate RPCResponse union type and helper types
	sb.WriteString(r.generateRPCResponseUnionType())

	// Ensure directory exists
	outputDir := filepath.Join(sdkRootDir, "src", "rpc", "types")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", outputDir, err)
	}

	// Write to file
	outputPath := filepath.Join(outputDir, "response.ts")
	return os.WriteFile(outputPath, []byte(sb.String()), 0o644)
}

// shouldSkipInterfaceGeneration checks if a type should skip interface generation
func (r *ResponseGenerator) shouldSkipInterfaceGeneration(name string) bool {
	typeMappings := getTypeMappings()
	_, exists := typeMappings[name]
	return exists
}

// generateRequestInterface generates a Request structure for RPC responses
func (r *ResponseGenerator) generateRequestInterface(name string, rpcMethod string) string {
	var sb strings.Builder

	// Generate JSDoc comment
	enumValue := r.rpcMethodToEnum(rpcMethod)
	sb.WriteString(fmt.Sprintf("/**\n"))
	sb.WriteString(fmt.Sprintf(" * Represents the response structure for the {@link RPCMethod.%s} RPC method.\n", enumValue))
	sb.WriteString(fmt.Sprintf(" */\n"))

	// Generate the Request interface
	requestName := strings.TrimSuffix(name, "Response") + "Response"
	paramsName := name + "Params"

	sb.WriteString(fmt.Sprintf("export interface %s extends GenericRPCMessage {\n", requestName))
	sb.WriteString(fmt.Sprintf("    method: RPCMethod.%s;\n", enumValue))
	sb.WriteString(fmt.Sprintf("    params: %s;\n", paramsName))
	sb.WriteString("}\n\n")

	return sb.String()
}

// generateRPCResponseUnionType generates the RPCResponse union type and helper types
func (r *ResponseGenerator) generateRPCResponseUnionType() string {
	var sb strings.Builder
	
	// Generate RPCResponse union type
	sb.WriteString("/**\n")
	sb.WriteString(" * Union type for all possible RPC response types.\n")
	sb.WriteString(" * This allows for type-safe handling of different response structures.\n")
	sb.WriteString(" */\n")
	sb.WriteString("export type RPCResponse =\n")
	
	// Get all generated response types sorted by name
	definitionNames := r.sortedDefNames(r.responseDefs)
	var responseNames []string
	for _, name := range definitionNames {
		if !r.shouldSkipInterfaceGeneration(name) {
			// Only include if it has an associated RPC method (meaning it's actually a response type)
			if _, hasRPCMethod := r.responseTypes[name]; hasRPCMethod {
				requestName := strings.TrimSuffix(name, "Response") + "Response"
				responseNames = append(responseNames, requestName)
			}
		}
	}
	
	// Also include common types that might be generated (if they have RPC methods)
	commonNames := r.sortedDefNames(r.commonDefs)
	for _, name := range commonNames {
		if !r.shouldSkipInterfaceGeneration(name) {
			// Only include if it has an associated RPC method
			if _, hasRPCMethod := r.responseTypes[name]; hasRPCMethod {
				requestName := strings.TrimSuffix(name, "Response") + "Response"
				responseNames = append(responseNames, requestName)
			}
		}
	}
	
	// Remove duplicates and sort
	uniqueResponseTypes := make(map[string]bool)
	for _, name := range responseNames {
		uniqueResponseTypes[name] = true
	}
	
	var unionTypes []string
	for name := range uniqueResponseTypes {
		unionTypes = append(unionTypes, name)
	}
	sort.Strings(unionTypes)
	
	// Generate union type - only if we have types to generate
	if len(unionTypes) > 0 {
		for i, name := range unionTypes {
			if i == 0 {
				sb.WriteString(fmt.Sprintf("    | %s\n", name))
			} else {
				sb.WriteString(fmt.Sprintf("    | %s\n", name))
			}
		}
	} else {
		// Fallback if no types are generated
		sb.WriteString("    | never\n")
	}
	sb.WriteString(";\n\n")
	
	// Generate helper types
	sb.WriteString("/**\n")
	sb.WriteString(" * Maps RPC methods to their corresponding parameter types.\n")
	sb.WriteString(" */\n")
	sb.WriteString("// Helper type to extract the response type for a given method\n")
	sb.WriteString("export type ExtractResponseByMethod<M extends RPCMethod> = Extract<RPCResponse, { method: M }>;\n\n")
	sb.WriteString("export type RPCResponseParams = ExtractResponseByMethod<RPCMethod>['params'];\n\n")
	sb.WriteString("export type RPCResponseParamsByMethod = {\n")
	sb.WriteString("    [M in RPCMethod]: ExtractResponseByMethod<M>['params'];\n")
	sb.WriteString("};\n\n")
	
	return sb.String()
}

// generateTypeScriptInterface generates a TypeScript interface from a schema property
func (r *ResponseGenerator) generateTypeScriptInterface(name string, prop SchemaProperty) string {
	var sb strings.Builder

	switch prop.Type {
	case "object":
		sb.WriteString(fmt.Sprintf("export interface %sParams {\n", name))

		// Sort property names for consistent output
		var propertyNames []string
		for propName := range prop.Properties {
			propertyNames = append(propertyNames, propName)
		}
		sort.Strings(propertyNames)

		for i, propName := range propertyNames {
			propDef := prop.Properties[propName]
			camelCaseName := toCamelCase(propName)
			tsType := r.generateTypeScriptType(propDef)

			optional := ""
			if !slices.Contains(prop.Required, propName) {
				optional = "?"
			}

			sb.WriteString(fmt.Sprintf("  %s%s: %s", camelCaseName, optional, tsType))
			if i < len(propertyNames)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}

		sb.WriteString("}\n\n")
	case "enum":
		// Generate type union for enums
		if len(prop.Enum) > 0 {
			sb.WriteString(fmt.Sprintf("export type %s = ", name))
			for i, val := range prop.Enum {
				sb.WriteString(fmt.Sprintf("\"%s\"", val))
				if i < len(prop.Enum)-1 {
					sb.WriteString(" | ")
				}
			}
			sb.WriteString(";\n\n")
		}
	}

	return sb.String()
}

// generateTypeScriptType converts a schema property to TypeScript type
func (r *ResponseGenerator) generateTypeScriptType(prop SchemaProperty) string {
	switch prop.Type {
	case "string":
		// Check if this is a mapped type based on format
		typeMappings := getTypeMappings()
		for typeName, mapping := range typeMappings {
			if strings.ToLower(typeName) == prop.Format {
				return mapping.TypeScriptType
			}
		}

		// Handle special formats not in type mappings
		switch prop.Format {
		case "date-time":
			return "Date"
		default:
			return "string"
		}
	case "integer":
		return "number"
	case "object":
		return "object"
	case "enum":
		if len(prop.Enum) > 0 {
			var enumValues []string
			for _, val := range prop.Enum {
				enumValues = append(enumValues, fmt.Sprintf("\"%s\"", val))
			}
			return strings.Join(enumValues, " | ")
		}
		return "string"
	default:
		if prop.Ref != "" {
			// Extract the definition name from the reference
			parts := strings.Split(prop.Ref, "/")
			if len(parts) >= 3 {
				refName := parts[len(parts)-1]
				// Check if this is a mapped type
				typeMappings := getTypeMappings()
				if mapping, exists := typeMappings[refName]; exists {
					return mapping.TypeScriptType
				}
				return refName
			}
		}
		return "any"
	}
}
