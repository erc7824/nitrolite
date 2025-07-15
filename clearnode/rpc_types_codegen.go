package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

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