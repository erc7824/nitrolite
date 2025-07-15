package main

import (
	"os"
	"path/filepath"
)

func runZodGeneratorCli(logger Logger) {
	logger = logger.NewSystem("zod-generator")
	if len(os.Args) < 3 {
		logger.Fatal("Usage: clearnode zod-generator <out_dir>")
	}

	outDir := os.Args[2]
	generator := NewZodGenerator()

	// Load schemas from request and response directories
	requestDir := filepath.Join(outDir, "request")
	responseDir := filepath.Join(outDir, "response")

	if err := generator.LoadSchemas(requestDir, responseDir); err != nil {
		logger.Fatal("Failed to load schemas", "err", err)
	}

	// Categorize definitions
	generator.CategorizeDefinitions()

	// Generate TypeScript files
	if err := generator.GenerateAllFiles(outDir); err != nil {
		logger.Fatal("Failed to generate TypeScript files", "err", err)
	}

	logger.Info("Generated Zod TypeScript files", "dir", outDir)
}