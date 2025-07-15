package main

import (
	"bytes"
	"slices"
	"sort"
	"strings"
	"text/template"
)

// TemplateData holds data for template execution
type TemplateData struct {
	Properties     []PropertyData
	TypeName       string
	ImportStatements []string
	EnumValues     []string
	UnionTypes     []string
	RPCMethod      string
	JSDocComment   string
	TransformLogic string
}

// PropertyData represents a single property for template generation
type PropertyData struct {
	Name         string
	CamelName    string
	ZodSchema    string
	TypeScriptType string
	IsRequired   bool
	IsLast       bool
}

// CodeTemplates holds all the templates used for code generation
type CodeTemplates struct {
	TypeScriptInterface *template.Template
	ZodObjectSchema     *template.Template
	ZodSchemaWithTransform *template.Template
	ImportStatements    *template.Template
	UnionType          *template.Template
	RequestInterface   *template.Template
	ResponseParsers    *template.Template
}

// NewCodeTemplates creates and parses all templates
func NewCodeTemplates() (*CodeTemplates, error) {
	templates := &CodeTemplates{}
	
	// TypeScript interface template
	tsInterfaceTemplate := `export interface {{.TypeName}}Params {
{{range .Properties}}  {{.CamelName}}{{if not .IsRequired}}?{{end}}: {{.TypeScriptType}}{{if not .IsLast}},{{end}}
{{end}}}

`
	
	// Zod object schema template
	zodObjectTemplate := `z.object({
{{range .Properties}}  {{.Name}}: {{.ZodSchema}}{{if not .IsRequired}}.optional(){{end}}{{if not .IsLast}},{{end}}
{{end}}})`
	
	// Zod schema with transform template
	zodTransformTemplate := `z.object({
{{range .Properties}}  {{.Name}}: {{.ZodSchema}}{{if not .IsRequired}}.optional(){{end}}{{if not .IsLast}},{{end}}
{{end}}})
    .transform((raw) => ({
{{range .Properties}}      {{.CamelName}}: raw.{{.Name}}{{if not .IsLast}},{{end}}
{{end}}    }) as {{.TypeName}})`
	
	// Import statements template
	importTemplate := `{{range .ImportStatements}}{{.}}
{{end}}`
	
	// Union type template
	unionTemplate := `export type RPCResponse =
{{range .UnionTypes}}    | {{.}}
{{end}};`
	
	// Request interface template
	requestTemplate := `/**
 * {{.JSDocComment}}
 */
export interface {{.TypeName}}Response extends GenericRPCMessage {
    method: RPCMethod.{{.RPCMethod}};
    params: {{.TypeName}}ResponseParams;
}

`
	
	// Response parsers template
	parsersTemplate := `export const responseParsers: Record<string, (params: any) => any> = {
{{range .Properties}}  [RPCMethod.{{.RPCMethod}}]: (params) => {{.Name}}Schema.parse(params),
{{end}}};`
	
	var err error
	
	if templates.TypeScriptInterface, err = template.New("typescript").Parse(tsInterfaceTemplate); err != nil {
		return nil, err
	}
	
	if templates.ZodObjectSchema, err = template.New("zodObject").Parse(zodObjectTemplate); err != nil {
		return nil, err
	}
	
	if templates.ZodSchemaWithTransform, err = template.New("zodTransform").Parse(zodTransformTemplate); err != nil {
		return nil, err
	}
	
	if templates.ImportStatements, err = template.New("imports").Parse(importTemplate); err != nil {
		return nil, err
	}
	
	if templates.UnionType, err = template.New("union").Parse(unionTemplate); err != nil {
		return nil, err
	}
	
	if templates.RequestInterface, err = template.New("request").Parse(requestTemplate); err != nil {
		return nil, err
	}
	
	if templates.ResponseParsers, err = template.New("parsers").Parse(parsersTemplate); err != nil {
		return nil, err
	}
	
	return templates, nil
}

// CodeBuilder provides a clean interface for building generated code
type CodeBuilder struct {
	templates *CodeTemplates
}

// NewCodeBuilder creates a new code builder with templates
func NewCodeBuilder() (*CodeBuilder, error) {
	templates, err := NewCodeTemplates()
	if err != nil {
		return nil, err
	}
	
	return &CodeBuilder{templates: templates}, nil
}

// BuildTypeScriptInterface generates a TypeScript interface using templates
func (cb *CodeBuilder) BuildTypeScriptInterface(typeName string, properties []PropertyData) (string, error) {
	data := TemplateData{
		TypeName:   typeName,
		Properties: properties,
	}
	
	var buffer bytes.Buffer
	if err := cb.templates.TypeScriptInterface.Execute(&buffer, data); err != nil {
		return "", err
	}
	
	return buffer.String(), nil
}

// BuildZodObjectSchema generates a Zod object schema using templates
func (cb *CodeBuilder) BuildZodObjectSchema(properties []PropertyData) (string, error) {
	data := TemplateData{Properties: properties}
	
	var buffer bytes.Buffer
	if err := cb.templates.ZodObjectSchema.Execute(&buffer, data); err != nil {
		return "", err
	}
	
	return buffer.String(), nil
}

// BuildZodSchemaWithTransform generates a Zod schema with camelCase transform
func (cb *CodeBuilder) BuildZodSchemaWithTransform(typeName string, properties []PropertyData) (string, error) {
	data := TemplateData{
		TypeName:   typeName,
		Properties: properties,
	}
	
	var buffer bytes.Buffer
	if err := cb.templates.ZodSchemaWithTransform.Execute(&buffer, data); err != nil {
		return "", err
	}
	
	return buffer.String(), nil
}

// BuildUnionType generates a union type using templates
func (cb *CodeBuilder) BuildUnionType(unionTypes []string) (string, error) {
	data := TemplateData{UnionTypes: unionTypes}
	
	var buffer bytes.Buffer
	if err := cb.templates.UnionType.Execute(&buffer, data); err != nil {
		return "", err
	}
	
	return buffer.String(), nil
}

// BuildRequestInterface generates a request interface using templates
func (cb *CodeBuilder) BuildRequestInterface(typeName string, rpcMethod string, jsDocComment string) (string, error) {
	data := TemplateData{
		TypeName:     typeName,
		RPCMethod:    rpcMethod,
		JSDocComment: jsDocComment,
	}
	
	var buffer bytes.Buffer
	if err := cb.templates.RequestInterface.Execute(&buffer, data); err != nil {
		return "", err
	}
	
	return buffer.String(), nil
}

// PropertySorter provides utilities for consistent property ordering
type PropertySorter struct{}

// NewPropertySorter creates a new property sorter
func NewPropertySorter() *PropertySorter {
	return &PropertySorter{}
}

// SortPropertyNames sorts property names consistently
func (ps *PropertySorter) SortPropertyNames(props map[string]SchemaProperty) []string {
	names := make([]string, 0, len(props))
	for name := range props {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// CreatePropertyDataList creates a sorted list of PropertyData from schema properties
func (ps *PropertySorter) CreatePropertyDataList(
	props map[string]SchemaProperty,
	requiredFields []string,
	zodGenerator *ZodSchemaGenerator,
	responseGenerator *ResponseGenerator,
) []PropertyData {
	sortedNames := ps.SortPropertyNames(props)
	propertyList := make([]PropertyData, 0, len(sortedNames))
	
	for i, name := range sortedNames {
		propDef := props[name]
		
		propertyData := PropertyData{
			Name:           name,
			CamelName:      toCamelCase(name),
			ZodSchema:      zodGenerator.GenerateZodSchema(propDef),
			TypeScriptType: responseGenerator.generateTypeScriptType(propDef),
			IsRequired:     slices.Contains(requiredFields, name),
			IsLast:         i == len(sortedNames)-1,
		}
		
		propertyList = append(propertyList, propertyData)
	}
	
	return propertyList
}

// StringUtils provides utility functions for string manipulation
type StringUtils struct{}

// NewStringUtils creates a new string utils instance
func NewStringUtils() *StringUtils {
	return &StringUtils{}
}

// ToCamelCase converts snake_case to camelCase with proper error handling
func (su *StringUtils) ToCamelCase(input string) string {
	if input == "" {
		return input
	}
	
	parts := strings.Split(input, "_")
	if len(parts) == 1 {
		return input
	}
	
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return result
}