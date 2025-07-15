package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func runZodGeneratorCli(cliLogger Logger) {
	systemLogger := cliLogger.NewSystem("zod-generator")
	
	config, err := parseCommandLineArguments()
	if err != nil {
		systemLogger.Fatal("Invalid command line arguments", "err", err)
	}

	codeGenerator := NewZodGenerator()

	// Load schemas from request and response directories
	requestSchemaPath := filepath.Join(config.SchemaDirectoryPath, "request")
	responseSchemaPath := filepath.Join(config.SchemaDirectoryPath, "response")

	if err := codeGenerator.LoadSchemas(requestSchemaPath, responseSchemaPath); err != nil {
		systemLogger.Fatal("Failed to load schemas", "err", err)
	}

	// Categorize definitions
	codeGenerator.CategorizeDefinitions()

	// Generate TypeScript files
	if err := codeGenerator.GenerateAllFiles(config.SchemaDirectoryPath, config.SdkRootDirectoryPath); err != nil {
		systemLogger.Fatal("Failed to generate TypeScript files", "err", err)
	}

	systemLogger.Info("Generated Zod TypeScript files", 
		"schema_directory", config.SchemaDirectoryPath, 
		"sdk_root_directory", config.SdkRootDirectoryPath)
}

// parseCommandLineArguments parses and validates command line arguments
func parseCommandLineArguments() (*GenerationConfig, error) {
	if len(os.Args) < 4 {
		return nil, fmt.Errorf("usage: clearnode zod-generator <schema_directory> <sdk_root_directory>")
	}

	schemaDirectoryPath := os.Args[2]
	sdkRootDirectoryPath := os.Args[3]

	return NewGenerationConfig(schemaDirectoryPath, sdkRootDirectoryPath)
}