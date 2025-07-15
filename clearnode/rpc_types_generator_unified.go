package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UnifiedGenerator centralizes common generation logic
type UnifiedGenerator struct {
	codeBuilder      *CodeBuilder
	zodGenerator     *ZodSchemaGenerator
	propertySorter   *PropertySorter
	dependencies     *GeneratorDependencies
	errorCollector   *ErrorCollector
}

// NewUnifiedGenerator creates a new unified generator with all dependencies
func NewUnifiedGenerator(deps *GeneratorDependencies) (*UnifiedGenerator, error) {
	codeBuilder, err := NewCodeBuilder()
	if err != nil {
		return nil, fmt.Errorf("failed to create code builder: %w", err)
	}
	
	zodGenerator, err := NewZodSchemaGenerator()
	if err != nil {
		// Fallback to basic generator
		zodGenerator = &ZodSchemaGenerator{}
	}
	
	return &UnifiedGenerator{
		codeBuilder:    codeBuilder,
		zodGenerator:   zodGenerator,
		propertySorter: NewPropertySorter(),
		dependencies:   deps,
		errorCollector: NewErrorCollector(),
	}, nil
}

// GenerateCommonSchemaFile generates common schema definitions
func (ug *UnifiedGenerator) GenerateCommonSchemaFile(config *GenerationConfig) error {
	content := ug.buildCommonSchemaContent()
	return ug.writeFileWithDirectoryCreation(
		filepath.Join(config.ParseOutputPath, "common_gen.ts"),
		content,
	)
}

// GenerateRequestSchemaFile generates request schema definitions
func (ug *UnifiedGenerator) GenerateRequestSchemaFile(config *GenerationConfig) error {
	content := ug.buildRequestSchemaContent()
	return ug.writeFileWithDirectoryCreation(
		filepath.Join(config.ParseOutputPath, "requests_gen.ts"),
		content,
	)
}

// GenerateResponseSchemaFile generates response schema definitions
func (ug *UnifiedGenerator) GenerateResponseSchemaFile(config *GenerationConfig) error {
	content := ug.buildResponseSchemaContent()
	return ug.writeFileWithDirectoryCreation(
		filepath.Join(config.ParseOutputPath, "response_gen.ts"),
		content,
	)
}

// GenerateTypeScriptTypesFile generates TypeScript interface definitions
func (ug *UnifiedGenerator) GenerateTypeScriptTypesFile(config *GenerationConfig) error {
	content := ug.buildTypeScriptTypesContent()
	return ug.writeFileWithDirectoryCreation(
		filepath.Join(config.TypesOutputPath, "response.ts"),
		content,
	)
}

// buildCommonSchemaContent builds the content for common schema file
func (ug *UnifiedGenerator) buildCommonSchemaContent() string {
	var contentBuilder strings.Builder
	
	// Add imports
	contentBuilder.WriteString("import { z } from 'zod';\n")
	contentBuilder.WriteString("import { Address } from 'viem';\n\n")
	contentBuilder.WriteString("// Common schemas used by both requests and responses\n\n")
	
	// Add built-in schemas
	contentBuilder.WriteString(ug.zodGenerator.GenerateBuiltinSchemas())
	
	// Add common definitions
	sortedDefinitions := ug.dependencies.DefinitionSorter(ug.dependencies.CommonDefinitions)
	contentBuilder.WriteString(ug.zodGenerator.GenerateSchemaDefinitions(sortedDefinitions, ug.dependencies.CommonDefinitions))
	
	return contentBuilder.String()
}

// buildRequestSchemaContent builds the content for request schema file
func (ug *UnifiedGenerator) buildRequestSchemaContent() string {
	var contentBuilder strings.Builder
	
	// Add imports
	contentBuilder.WriteString("import { z } from 'zod';\n")
	contentBuilder.WriteString("import { RPCMethod } from '../types';\n")
	contentBuilder.WriteString("import { addressSchema, hexSchema } from './common_gen';\n")
	
	// Add common schema imports
	commonDefinitions := ug.dependencies.DefinitionSorter(ug.dependencies.CommonDefinitions)
	if len(commonDefinitions) > 0 {
		contentBuilder.WriteString(ug.zodGenerator.GenerateCommonSchemaImports(commonDefinitions))
	}
	contentBuilder.WriteString("\n// Request schemas\n\n")
	
	// Add request-specific definitions
	requestDefinitions := ug.dependencies.DefinitionSorter(ug.dependencies.RequestDefinitions)
	contentBuilder.WriteString(ug.zodGenerator.GenerateSchemaDefinitions(requestDefinitions, ug.dependencies.RequestDefinitions))
	
	// Add parser mapping
	contentBuilder.WriteString(ug.buildParserMapping("request", ug.dependencies.RequestTypeMappings))
	
	return contentBuilder.String()
}

// buildResponseSchemaContent builds the content for response schema file
func (ug *UnifiedGenerator) buildResponseSchemaContent() string {
	var contentBuilder strings.Builder
	
	// Add imports
	contentBuilder.WriteString("import { z } from 'zod';\n")
	contentBuilder.WriteString("import { RPCMethod } from '../types';\n")
	contentBuilder.WriteString("import { addressSchema, hexSchema } from './common_gen';\n")
	
	// Add TypeScript type imports
	contentBuilder.WriteString("import type {\n")
	
	// Import common types
	commonDefinitions := ug.dependencies.DefinitionSorter(ug.dependencies.CommonDefinitions)
	for _, name := range commonDefinitions {
		contentBuilder.WriteString(fmt.Sprintf("  %s,\n", name))
	}
	
	// Import response-specific types
	responseDefinitions := ug.dependencies.DefinitionSorter(ug.dependencies.ResponseDefinitions)
	for _, name := range responseDefinitions {
		contentBuilder.WriteString(fmt.Sprintf("  %s,\n", name))
	}
	
	contentBuilder.WriteString("} from '../types/response';\n")
	
	// Add common schema imports
	if len(commonDefinitions) > 0 {
		contentBuilder.WriteString(ug.zodGenerator.GenerateCommonSchemaImports(commonDefinitions))
	}
	contentBuilder.WriteString("\n// Response schemas with camelCase transforms\n\n")
	
	// Add response schemas with transforms
	contentBuilder.WriteString(ug.buildResponseSchemasWithTransforms(responseDefinitions))
	
	// Add parser mapping
	contentBuilder.WriteString(ug.buildParserMapping("response", ug.dependencies.ResponseTypeMappings))
	
	return contentBuilder.String()
}

// buildTypeScriptTypesContent builds the content for TypeScript types file
func (ug *UnifiedGenerator) buildTypeScriptTypesContent() string {
	var contentBuilder strings.Builder
	
	// Add header
	contentBuilder.WriteString("// Auto-generated TypeScript response types with camelCase field names\n")
	contentBuilder.WriteString("// Generated from JSON schemas\n\n")
	contentBuilder.WriteString("import type { Address, Hex } from 'viem';\n")
	contentBuilder.WriteString("import {RPCMethod, GenericRPCMessage} from '.';\n\n")
	
	// Add common type interfaces
	contentBuilder.WriteString(ug.buildTypeScriptInterfaces(ug.dependencies.CommonDefinitions, false))
	
	// Add response type interfaces
	contentBuilder.WriteString(ug.buildTypeScriptInterfaces(ug.dependencies.ResponseDefinitions, true))
	
	// Add union type and helpers
	contentBuilder.WriteString(ug.buildUnionTypeAndHelpers())
	
	return contentBuilder.String()
}

// buildResponseSchemasWithTransforms builds response schemas with camelCase transforms
func (ug *UnifiedGenerator) buildResponseSchemasWithTransforms(definitions []string) string {
	var contentBuilder strings.Builder
	
	for _, name := range definitions {
		definition := ug.dependencies.ResponseDefinitions[name]
		if definition.Type == "object" {
			zodSchema := ug.zodGenerator.GenerateObjectSchemaWithTransform(definition, name)
			contentBuilder.WriteString(fmt.Sprintf("export const %sSchema = %s;\n\n", name, zodSchema))
		} else {
			zodSchema := ug.zodGenerator.GenerateZodSchema(definition)
			contentBuilder.WriteString(fmt.Sprintf("export const %sSchema = %s;\n\n", name, zodSchema))
		}
	}
	
	return contentBuilder.String()
}

// buildTypeScriptInterfaces builds TypeScript interfaces for definitions
func (ug *UnifiedGenerator) buildTypeScriptInterfaces(definitions map[string]SchemaProperty, includeRequestStructures bool) string {
	var contentBuilder strings.Builder
	
	sortedDefinitions := ug.dependencies.DefinitionSorter(definitions)
	
	for _, name := range sortedDefinitions {
		definition := definitions[name]
		
		// Skip interface generation for special types
		if ug.shouldSkipInterfaceGeneration(name) {
			continue
		}
		
		// Generate Request structure first for responses
		if includeRequestStructures {
			if rpcMethod, exists := ug.dependencies.ResponseTypeMappings[name]; exists {
				requestInterface := ug.buildRequestInterface(name, rpcMethod)
				contentBuilder.WriteString(requestInterface)
			}
		}
		
		// Generate Params interface
		interfaceCode := ug.buildTypeScriptInterface(name, definition)
		contentBuilder.WriteString(interfaceCode)
	}
	
	return contentBuilder.String()
}

// buildRequestInterface builds a request interface for RPC responses
func (ug *UnifiedGenerator) buildRequestInterface(name string, rpcMethod string) string {
	enumValue := ug.dependencies.EnumNameConverter(rpcMethod)
	jsDocComment := fmt.Sprintf("Represents the response structure for the {@link RPCMethod.%s} RPC method.", enumValue)
	
	requestInterface, err := ug.codeBuilder.BuildRequestInterface(name, enumValue, jsDocComment)
	if err != nil {
		// Fallback to manual construction
		return fmt.Sprintf("export interface %sResponse extends GenericRPCMessage {\n    method: RPCMethod.%s;\n    params: %sResponseParams;\n}\n\n", 
			strings.TrimSuffix(name, "Response"), enumValue, name)
	}
	
	return requestInterface
}

// buildTypeScriptInterface builds a TypeScript interface from schema property
func (ug *UnifiedGenerator) buildTypeScriptInterface(name string, property SchemaProperty) string {
	switch property.Type {
	case "object":
		return ug.buildObjectInterface(name, property)
	case "enum":
		return ug.buildEnumInterface(name, property)
	default:
		return ""
	}
}

// buildObjectInterface builds a TypeScript interface for object types
func (ug *UnifiedGenerator) buildObjectInterface(name string, property SchemaProperty) string {
	properties := ug.createPropertyDataListForInterfaces(property.Properties, property.Required)
	
	interfaceCode, err := ug.codeBuilder.BuildTypeScriptInterface(name, properties)
	if err != nil {
		// Fallback to manual construction
		return fmt.Sprintf("export interface %sParams {\n  // Interface generation failed\n}\n\n", name)
	}
	
	return interfaceCode
}

// createPropertyDataListForInterfaces creates PropertyData for TypeScript interfaces
func (ug *UnifiedGenerator) createPropertyDataListForInterfaces(properties map[string]SchemaProperty, requiredFields []string) []PropertyData {
	return ug.propertySorter.CreatePropertyDataList(
		properties,
		requiredFields,
		ug.zodGenerator,
		ug.generateTypeScriptType,
	)
}

// generateTypeScriptType converts a schema property to TypeScript type
func (ug *UnifiedGenerator) generateTypeScriptType(prop SchemaProperty) string {
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

// buildEnumInterface builds a TypeScript type union for enum types
func (ug *UnifiedGenerator) buildEnumInterface(name string, property SchemaProperty) string {
	if len(property.Enum) == 0 {
		return ""
	}
	
	var enumValues []string
	for _, val := range property.Enum {
		enumValues = append(enumValues, fmt.Sprintf("\"%s\"", val))
	}
	
	return fmt.Sprintf("export type %s = %s;\n\n", name, strings.Join(enumValues, " | "))
}

// buildParserMapping builds parser mapping for requests or responses
func (ug *UnifiedGenerator) buildParserMapping(generationType string, typeMappings map[string]string) string {
	var contentBuilder strings.Builder
	
	contentBuilder.WriteString(fmt.Sprintf("// %s parser mapping\n", strings.Title(generationType)))
	contentBuilder.WriteString(fmt.Sprintf("export const %sParsers: Record<string, (params: any) => any> = {\n", generationType))
	
	for typeName, rpcMethod := range typeMappings {
		enumValue := ug.dependencies.EnumNameConverter(rpcMethod)
		contentBuilder.WriteString(fmt.Sprintf("  [RPCMethod.%s]: (params) => %sSchema.parse(params),\n", enumValue, typeName))
	}
	
	contentBuilder.WriteString("};\n")
	return contentBuilder.String()
}

// buildUnionTypeAndHelpers builds the RPCResponse union type and helper types
func (ug *UnifiedGenerator) buildUnionTypeAndHelpers() string {
	var contentBuilder strings.Builder
	
	// Generate RPCResponse union type
	contentBuilder.WriteString("/**\n")
	contentBuilder.WriteString(" * Union type for all possible RPC response types.\n")
	contentBuilder.WriteString(" * This allows for type-safe handling of different response structures.\n")
	contentBuilder.WriteString(" */\n")
	contentBuilder.WriteString("export type RPCResponse =\n")
	
	// Get all response types with RPC methods
	var unionTypes []string
	for name := range ug.dependencies.ResponseDefinitions {
		if !ug.shouldSkipInterfaceGeneration(name) {
			if _, hasRPCMethod := ug.dependencies.ResponseTypeMappings[name]; hasRPCMethod {
				requestName := strings.TrimSuffix(name, "Response") + "Response"
				unionTypes = append(unionTypes, requestName)
			}
		}
	}
	
	if len(unionTypes) > 0 {
		for i, unionType := range unionTypes {
			if i == 0 {
				contentBuilder.WriteString(fmt.Sprintf("    | %s\n", unionType))
			} else {
				contentBuilder.WriteString(fmt.Sprintf("    | %s\n", unionType))
			}
		}
	} else {
		contentBuilder.WriteString("    | never\n")
	}
	contentBuilder.WriteString(";\n\n")
	
	// Add helper types
	contentBuilder.WriteString("/**\n")
	contentBuilder.WriteString(" * Maps RPC methods to their corresponding parameter types.\n")
	contentBuilder.WriteString(" */\n")
	contentBuilder.WriteString("// Helper type to extract the response type for a given method\n")
	contentBuilder.WriteString("export type ExtractResponseByMethod<M extends RPCMethod> = Extract<RPCResponse, { method: M }>;\n\n")
	contentBuilder.WriteString("export type RPCResponseParams = ExtractResponseByMethod<RPCMethod>['params'];\n\n")
	contentBuilder.WriteString("export type RPCResponseParamsByMethod = {\n")
	contentBuilder.WriteString("    [M in RPCMethod]: ExtractResponseByMethod<M>['params'];\n")
	contentBuilder.WriteString("};\n\n")
	
	return contentBuilder.String()
}

// shouldSkipInterfaceGeneration checks if interface generation should be skipped
func (ug *UnifiedGenerator) shouldSkipInterfaceGeneration(name string) bool {
	typeMappings := getTypeMappings()
	_, exists := typeMappings[name]
	return exists
}

// writeFileWithDirectoryCreation writes content to a file, creating directories if needed
func (ug *UnifiedGenerator) writeFileWithDirectoryCreation(filePath string, content string) error {
	directoryPath := filepath.Dir(filePath)
	if err := os.MkdirAll(directoryPath, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", directoryPath, err)
	}
	
	return os.WriteFile(filePath, []byte(content), 0o644)
}