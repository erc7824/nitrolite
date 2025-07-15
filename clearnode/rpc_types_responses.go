package main

import (
	"fmt"
	"os"
	"path/filepath"
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
func (r *ResponseGenerator) GenerateResponsesFile(outDir string) error {
	var sb strings.Builder

	// Add imports
	sb.WriteString("import { z } from 'zod';\n")
	sb.WriteString("import { RPCMethod } from '../sdk/src/rpc/types';\n")
	sb.WriteString("import { addressSchema, hexSchema } from './common_gen';\n")

	// Import common schemas
	commonNames := r.sortedDefNames(r.commonDefs)
	if len(commonNames) > 0 {
		sb.WriteString(r.zodGenerator.GenerateCommonSchemaImports(commonNames))
	}
	sb.WriteString("\n")

	// Add response-specific schemas
	sb.WriteString("// Response schemas\n\n")

	definitionNames := r.sortedDefNames(r.responseDefs)
	sb.WriteString(r.zodGenerator.GenerateSchemaDefinitions(definitionNames, r.responseDefs))

	// Generate parser mapping
	sb.WriteString(r.generateResponseParsers())

	// Write to file
	outputPath := filepath.Join(outDir, "response_gen.ts")
	return os.WriteFile(outputPath, []byte(sb.String()), 0o644)
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

// Future: GenerateResponseTypesFile will generate sdk/src/rpc/types/response.ts
// func (r *ResponseGenerator) GenerateResponseTypesFile(outDir string) error {
//     // Implementation for generating response types file
//     return nil
// }