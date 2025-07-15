package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RequestGenerator handles request-specific code generation
type RequestGenerator struct {
	requestDefs   map[string]SchemaProperty
	requestTypes  map[string]string // typeName -> rpcMethod
	commonDefs    map[string]SchemaProperty
	zodGenerator  *ZodSchemaGenerator
	sortedDefNames func(map[string]SchemaProperty) []string
	rpcMethodToEnum func(string) string
}

// NewRequestGenerator creates a new request generator
func NewRequestGenerator(requestDefs map[string]SchemaProperty, requestTypes map[string]string, commonDefs map[string]SchemaProperty, sortedDefNames func(map[string]SchemaProperty) []string, rpcMethodToEnum func(string) string) *RequestGenerator {
	return &RequestGenerator{
		requestDefs:     requestDefs,
		requestTypes:    requestTypes,
		commonDefs:      commonDefs,
		zodGenerator:    &ZodSchemaGenerator{},
		sortedDefNames:  sortedDefNames,
		rpcMethodToEnum: rpcMethodToEnum,
	}
}

// GenerateRequestsFile generates the requests_gen.ts file
func (r *RequestGenerator) GenerateRequestsFile(outDir string) error {
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

	// Add request-specific schemas
	sb.WriteString("// Request schemas\n\n")

	definitionNames := r.sortedDefNames(r.requestDefs)
	sb.WriteString(r.zodGenerator.GenerateSchemaDefinitions(definitionNames, r.requestDefs))

	// Generate parser mapping
	sb.WriteString(r.generateRequestParsers())

	// Write to file
	outputPath := filepath.Join(outDir, "requests_gen.ts")
	return os.WriteFile(outputPath, []byte(sb.String()), 0o644)
}

// generateRequestParsers generates the request parser mapping
func (r *RequestGenerator) generateRequestParsers() string {
	var sb strings.Builder
	sb.WriteString("// Request parser mapping\n")
	sb.WriteString("export const requestParsers: Record<string, (params: any) => any> = {\n")

	for typeName, rpcMethod := range r.requestTypes {
		sb.WriteString(fmt.Sprintf("  [RPCMethod.%s]: (params) => %sSchema.parse(params),\n",
			r.rpcMethodToEnum(rpcMethod), typeName))
	}

	sb.WriteString("};\n")
	return sb.String()
}

// Future: GenerateAPIFile will generate sdk/src/rpc/api.ts
// func (r *RequestGenerator) GenerateAPIFile(outDir string) error {
//     // Implementation for generating API file
//     return nil
// }