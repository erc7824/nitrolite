package main

import (
	"fmt"
	"slices"
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
type ZodSchemaGenerator struct {
	codeBuilder    *CodeBuilder
	propertySorter *PropertySorter
	stringUtils    *StringUtils
}

// NewZodSchemaGenerator creates a new Zod schema generator with utilities
func NewZodSchemaGenerator() (*ZodSchemaGenerator, error) {
	codeBuilder, err := NewCodeBuilder()
	if err != nil {
		return nil, err
	}
	
	return &ZodSchemaGenerator{
		codeBuilder:    codeBuilder,
		propertySorter: NewPropertySorter(),
		stringUtils:    NewStringUtils(),
	}, nil
}

// GenerateZodSchema converts a SchemaProperty to a Zod schema string
func (z *ZodSchemaGenerator) GenerateZodSchema(prop SchemaProperty) string {
	switch prop.Type {
	case "string":
		return z.generateZodStringSchema(prop)
	case "integer":
		return "z.number()"
	case "object":
		return z.generateZodObjectSchema(prop)
	case "enum":
		return z.generateZodEnumSchema(prop)
	default:
		if prop.Ref != "" {
			return z.generateZodRefSchema(prop.Ref)
		}
		return "z.unknown()"
	}
}

// generateZodStringSchema handles string type with various formats
func (z *ZodSchemaGenerator) generateZodStringSchema(prop SchemaProperty) string {
	if zodSchema := z.getMappedZodSchemaForFormat(prop.Format); zodSchema != "" {
		return zodSchema
	}

	return z.getSpecialFormatZodSchema(prop.Format)
}

// getMappedZodSchemaForFormat retrieves Zod schema for mapped type formats
func (z *ZodSchemaGenerator) getMappedZodSchemaForFormat(format string) string {
	typeMappings := getTypeMappings()
	for typeName, mapping := range typeMappings {
		if strings.ToLower(typeName) == format {
			return mapping.ZodSchemaForFormat
		}
	}
	return ""
}

// getSpecialFormatZodSchema handles special formats not in type mappings
func (z *ZodSchemaGenerator) getSpecialFormatZodSchema(format string) string {
	switch format {
	case "date-time":
		return "z.union([z.string(), z.date()]).transform((v) => new Date(v))"
	default:
		return "z.string()"
	}
}

// generateZodObjectSchema handles object type with properties and required fields
func (z *ZodSchemaGenerator) generateZodObjectSchema(prop SchemaProperty) string {
	if len(prop.Properties) == 0 {
		return "z.object({})"
	}

	properties := z.createPropertyDataForZodSchema(prop)
	zodSchema, err := z.codeBuilder.BuildZodObjectSchema(properties)
	if err != nil {
		// Fallback to basic object schema if template fails
		return "z.object({})"
	}
	
	return zodSchema
}

// createPropertyDataForZodSchema creates PropertyData list for Zod schema generation
func (z *ZodSchemaGenerator) createPropertyDataForZodSchema(prop SchemaProperty) []PropertyData {
	sortedNames := z.propertySorter.SortPropertyNames(prop.Properties)
	properties := make([]PropertyData, 0, len(sortedNames))
	
	for i, name := range sortedNames {
		propDef := prop.Properties[name]
		zodSchema := z.GenerateZodSchema(propDef)
		
		propertyData := PropertyData{
			Name:       name,
			ZodSchema:  zodSchema,
			IsRequired: slices.Contains(prop.Required, name),
			IsLast:     i == len(sortedNames)-1,
		}
		
		properties = append(properties, propertyData)
	}
	
	return properties
}

// GenerateObjectSchemaWithTransform generates object schema with camelCase transform
func (z *ZodSchemaGenerator) GenerateObjectSchemaWithTransform(prop SchemaProperty, typeName string) string {
	if len(prop.Properties) == 0 {
		return "z.object({})"
	}

	properties := z.createPropertyDataForZodTransform(prop)
	zodSchema, err := z.codeBuilder.BuildZodSchemaWithTransform(typeName, properties)
	if err != nil {
		// Fallback to basic object schema if template fails
		return z.generateZodObjectSchema(prop)
	}
	
	return zodSchema
}

// createPropertyDataForZodTransform creates PropertyData list for Zod transform generation
func (z *ZodSchemaGenerator) createPropertyDataForZodTransform(prop SchemaProperty) []PropertyData {
	sortedNames := z.propertySorter.SortPropertyNames(prop.Properties)
	properties := make([]PropertyData, 0, len(sortedNames))
	
	for i, name := range sortedNames {
		propDef := prop.Properties[name]
		zodSchema := z.GenerateZodSchema(propDef)
		
		propertyData := PropertyData{
			Name:       name,
			CamelName:  z.stringUtils.ToCamelCase(name),
			ZodSchema:  zodSchema,
			IsRequired: slices.Contains(prop.Required, name),
			IsLast:     i == len(sortedNames)-1,
		}
		
		properties = append(properties, propertyData)
	}
	
	return properties
}

// generateZodEnumSchema handles enum type with proper validation
func (z *ZodSchemaGenerator) generateZodEnumSchema(prop SchemaProperty) string {
	if len(prop.Enum) == 0 {
		return "z.string()"
	}

	return z.buildZodEnumSchema(prop.Enum)
}

// buildZodEnumSchema creates a Zod enum schema from string values
func (z *ZodSchemaGenerator) buildZodEnumSchema(enumValues []string) string {
	quotedValues := make([]string, len(enumValues))
	for i, val := range enumValues {
		quotedValues[i] = fmt.Sprintf("\"%s\"", val)
	}

	return fmt.Sprintf("z.enum([%s])", strings.Join(quotedValues, ", "))
}

// generateZodRefSchema handles reference type with mapped type support
func (z *ZodSchemaGenerator) generateZodRefSchema(ref string) string {
	defName := z.extractDefinitionNameFromRef(ref)
	if defName == "" {
		return "z.unknown()"
	}

	if zodSchema := z.getMappedZodSchemaForRef(defName); zodSchema != "" {
		return zodSchema
	}

	return fmt.Sprintf("%sSchema", defName)
}

// extractDefinitionNameFromRef extracts the definition name from a JSON schema reference
func (z *ZodSchemaGenerator) extractDefinitionNameFromRef(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) < 3 {
		return ""
	}
	return parts[len(parts)-1]
}

// getMappedZodSchemaForRef retrieves Zod schema for mapped reference types
func (z *ZodSchemaGenerator) getMappedZodSchemaForRef(defName string) string {
	typeMappings := getTypeMappings()
	if mapping, exists := typeMappings[defName]; exists {
		return mapping.ZodSchemaForRef
	}
	return ""
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

// toCamelCase converts snake_case to camelCase (backward compatibility)
func toCamelCase(s string) string {
	stringUtils := NewStringUtils()
	return stringUtils.ToCamelCase(s)
}

