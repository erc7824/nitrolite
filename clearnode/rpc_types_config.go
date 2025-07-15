package main

import (
	"errors"
	"fmt"
	"path/filepath"
)

// GenerationConfig holds all configuration for code generation
type GenerationConfig struct {
	SchemaDirectoryPath  string
	SdkRootDirectoryPath string
	RequestSchemaPath    string
	ResponseSchemaPath   string
	ParseOutputPath      string
	TypesOutputPath      string
}

// NewGenerationConfig creates a configuration with validated paths
func NewGenerationConfig(schemaDir, sdkRootDir string) (*GenerationConfig, error) {
	if schemaDir == "" {
		return nil, fmt.Errorf("schema directory path cannot be empty")
	}
	if sdkRootDir == "" {
		return nil, fmt.Errorf("SDK root directory path cannot be empty")
	}

	return &GenerationConfig{
		SchemaDirectoryPath:  schemaDir,
		SdkRootDirectoryPath: sdkRootDir,
		RequestSchemaPath:    filepath.Join(schemaDir, "request"),
		ResponseSchemaPath:   filepath.Join(schemaDir, "response"),
		ParseOutputPath:      filepath.Join(sdkRootDir, "src", "rpc", "parse"),
		TypesOutputPath:      filepath.Join(sdkRootDir, "src", "rpc", "types"),
	}, nil
}

// GeneratorDependencies holds shared dependencies for generators
type GeneratorDependencies struct {
	RequestDefinitions   map[string]SchemaProperty
	ResponseDefinitions  map[string]SchemaProperty
	CommonDefinitions    map[string]SchemaProperty
	RequestTypeMappings  map[string]RPCMethod // typeName -> rpcMethod
	ResponseTypeMappings map[string]RPCMethod // typeName -> rpcMethod
	DefinitionSorter     func(map[string]SchemaProperty) []string
	EnumNameConverter    func(string) string
}

// FileGenerator provides a unified interface for file generation
type FileGenerator interface {
	GenerateFile(config *GenerationConfig) error
}

// ErrorCollector aggregates multiple errors during generation
type ErrorCollector struct {
	errors []error
}

// NewErrorCollector creates a new error collector
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{errors: make([]error, 0)}
}

// Add adds an error to the collector
func (ec *ErrorCollector) Add(err error) {
	if err != nil {
		ec.errors = append(ec.errors, err)
	}
}

// HasErrors returns true if any errors were collected
func (ec *ErrorCollector) HasErrors() bool {
	return len(ec.errors) > 0
}

// CombinedError returns all errors combined into a single error
func (ec *ErrorCollector) CombinedError() error {
	if !ec.HasErrors() {
		return nil
	}

	if len(ec.errors) == 1 {
		return ec.errors[0]
	}

	errorMessage := "multiple errors occurred:"
	for _, err := range ec.errors {
		errorMessage += fmt.Sprintf("\n  - %s", err.Error())
	}

	return errors.New(errorMessage)
}
