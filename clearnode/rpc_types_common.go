package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateCommonFile generates the common_gen.ts file with shared schemas
func (g *ZodGenerator) generateCommonFile(sdkRootDir string) error {
	var sb strings.Builder
	zodGen := &ZodSchemaGenerator{}

	// Add imports
	sb.WriteString("import { z } from 'zod';\n")
	sb.WriteString("import { Address } from 'viem';\n\n")

	// Add common schemas
	sb.WriteString("// Common schemas used by both requests and responses\n\n")

	// Add built-in common schemas
	sb.WriteString(zodGen.GenerateBuiltinSchemas())

	// Generate common definitions
	definitionNames := g.getSortedDefinitionNames(g.commonDefs)
	sb.WriteString(zodGen.GenerateSchemaDefinitions(definitionNames, g.commonDefs))

	// Ensure directory exists
	outputDir := filepath.Join(sdkRootDir, "src", "rpc", "parse")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", outputDir, err)
	}

	// Write to file
	outputPath := filepath.Join(outputDir, "common_gen.ts")
	return os.WriteFile(outputPath, []byte(sb.String()), 0o644)
}