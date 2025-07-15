package main

import (
	"os"
	"path/filepath"
)

func runZodGeneratorCli(logger Logger) {
	logger = logger.NewSystem("zod-generator")
	if len(os.Args) < 4 {
		logger.Fatal("Usage: clearnode zod-generator <schemas_dir> <sdk_root_dir>")
	}

	schemasDir := os.Args[2]
	sdkRootDir := os.Args[3]
	generator := NewZodGenerator()

	// Load schemas from request and response directories
	requestDir := filepath.Join(schemasDir, "request")
	responseDir := filepath.Join(schemasDir, "response")

	if err := generator.LoadSchemas(requestDir, responseDir); err != nil {
		logger.Fatal("Failed to load schemas", "err", err)
	}

	// Categorize definitions
	generator.CategorizeDefinitions()

	// Generate TypeScript files
	if err := generator.GenerateAllFiles(schemasDir, sdkRootDir); err != nil {
		logger.Fatal("Failed to generate TypeScript files", "err", err)
	}

	logger.Info("Generated Zod TypeScript files", "schemas_dir", schemasDir, "sdk_root_dir", sdkRootDir)
}