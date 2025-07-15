package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// TypeMapping defines how special types should be handled across different generators
type TypeMapping struct {
	ZodSchemaForFormat string // Zod schema for format-based types (e.g., "bignumber")
	ZodSchemaForRef    string // Zod schema for reference-based types (e.g., "#/$defs/BigNumber")
	TypeScriptType     string // TypeScript type in interfaces
}

// getTypeMappings returns the centralized type mappings
func getTypeMappings() map[string]TypeMapping {
	return map[string]TypeMapping{
		"BigNumber": {
			ZodSchemaForFormat: "z.string().transform((v) => BigInt(v))",
			ZodSchemaForRef:    "BigNumberSchema",
			TypeScriptType:     "bigint",
		},
		"Address": {
			ZodSchemaForFormat: "addressSchema",
			ZodSchemaForRef:    "AddressSchema",
			TypeScriptType:     "Address",
		},
		"Hex": {
			ZodSchemaForFormat: "hexSchema",
			ZodSchemaForRef:    "HexSchema",
			TypeScriptType:     "Hex",
		},
	}
}

// ZodSchemaGenerator provides common Zod schema generation utilities
type ZodSchemaGenerator struct{}

// GenerateZodSchema converts a SchemaProperty to a Zod schema string
func (z *ZodSchemaGenerator) GenerateZodSchema(prop SchemaProperty) string {
	switch prop.Type {
	case "string":
		return z.generateStringSchema(prop)
	case "integer":
		return "z.number()"
	case "object":
		return z.generateObjectSchema(prop)
	case "enum":
		return z.generateEnumSchema(prop)
	default:
		if prop.Ref != "" {
			return z.generateRefSchema(prop.Ref)
		}
		return "z.unknown()"
	}
}

// generateStringSchema handles string type with various formats
func (z *ZodSchemaGenerator) generateStringSchema(prop SchemaProperty) string {
	// Check if this is a mapped type based on format
	typeMappings := getTypeMappings()
	for typeName, mapping := range typeMappings {
		if strings.ToLower(typeName) == prop.Format {
			return mapping.ZodSchemaForFormat
		}
	}

	// Handle special formats not in type mappings
	switch prop.Format {
	case "date-time":
		return "z.union([z.string(), z.date()]).transform((v) => new Date(v))"
	default:
		return "z.string()"
	}
}

// generateObjectSchema handles object type with properties and required fields
func (z *ZodSchemaGenerator) generateObjectSchema(prop SchemaProperty) string {
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
		zodSchema := z.GenerateZodSchema(propDef)

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

// GenerateObjectSchemaWithTransform generates object schema with camelCase transform
func (z *ZodSchemaGenerator) GenerateObjectSchemaWithTransform(prop SchemaProperty, typeName string) string {
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
		zodSchema := z.GenerateZodSchema(propDef)

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

	// Add transform to convert snake_case to camelCase
	if typeName != "" {
		sb.WriteString("\n    .transform((raw) => ({\n")
		for i, name := range propertyNames {
			camelCaseName := toCamelCase(name)
			sb.WriteString(fmt.Sprintf("      %s: raw.%s", camelCaseName, name))
			if i < len(propertyNames)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("    }) as %s)", typeName))
	}

	return sb.String()
}

// generateEnumSchema handles enum type
func (z *ZodSchemaGenerator) generateEnumSchema(prop SchemaProperty) string {
	if len(prop.Enum) == 0 {
		return "z.string()"
	}

	enumValues := make([]string, len(prop.Enum))
	for i, val := range prop.Enum {
		enumValues[i] = fmt.Sprintf("\"%s\"", val)
	}

	return fmt.Sprintf("z.enum([%s])", strings.Join(enumValues, ", "))
}

// generateRefSchema handles reference type
func (z *ZodSchemaGenerator) generateRefSchema(ref string) string {
	// Extract the definition name from the reference
	parts := strings.Split(ref, "/")
	if len(parts) < 3 {
		return "z.unknown()"
	}

	defName := parts[len(parts)-1]

	// Check if this is a mapped type
	typeMappings := getTypeMappings()
	if mapping, exists := typeMappings[defName]; exists {
		return mapping.ZodSchemaForRef
	}

	return fmt.Sprintf("%sSchema", defName)
}

// GenerateCommonImports generates common import statements for Zod files
func (z *ZodSchemaGenerator) GenerateCommonImports() string {
	var sb strings.Builder
	sb.WriteString("import { z } from 'zod';\n")
	sb.WriteString("import { addressSchema, hexSchema } from './common_gen';\n")
	return sb.String()
}

// GenerateCommonSchemaImports generates import statements for common schemas
func (z *ZodSchemaGenerator) GenerateCommonSchemaImports(commonNames []string) string {
	if len(commonNames) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("import {\n")
	for i, name := range commonNames {
		sb.WriteString(fmt.Sprintf("  %sSchema", name))
		if i < len(commonNames)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("} from './common_gen';\n")
	return sb.String()
}

// GenerateBuiltinSchemas generates built-in common schemas (address, hex)
func (z *ZodSchemaGenerator) GenerateBuiltinSchemas() string {
	var sb strings.Builder
	sb.WriteString("export const addressSchema = z.string().refine((val) => /^0x[0-9a-fA-F]{40}$/.test(val), {\n")
	sb.WriteString("  message: 'Must be a 0x-prefixed hex string of 40 hex chars (EVM address)',\n")
	sb.WriteString("});\n\n")

	sb.WriteString("export const hexSchema = z.string().refine((val) => /^0x[0-9a-fA-F]*$/.test(val), {\n")
	sb.WriteString("  message: 'Must be a 0x-prefixed hex string',\n")
	sb.WriteString("});\n\n")

	return sb.String()
}

// GenerateSchemaDefinitions generates schema definitions for a set of definitions
func (z *ZodSchemaGenerator) GenerateSchemaDefinitions(definitionNames []string, defs map[string]SchemaProperty) string {
	var sb strings.Builder
	for _, name := range definitionNames {
		def := defs[name]
		zodSchema := z.GenerateZodSchema(def)
		sb.WriteString(fmt.Sprintf("export const %sSchema = %s;\n\n", name, zodSchema))
	}
	return sb.String()
}

// toCamelCase converts snake_case to camelCase
func toCamelCase(s string) string {
	if s == "" {
		return s
	}

	parts := strings.Split(s, "_")
	if len(parts) == 1 {
		return s
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return result
}

