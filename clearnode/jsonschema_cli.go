package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/invopop/jsonschema"
)

func runExportJsonSchemaCli(logger Logger) {
	logger = logger.NewSystem("jsonschema")
	if len(os.Args) < 3 {
		logger.Fatal("Usage: clearnode jsonschema <out_dir>")
	}

	outDir := os.Args[2]
	requests := []any{
		&GetLedgerTransactionsParams{},
	}
	responses := []any{
		&TransactionResponse{},
	}

	for _, typ := range requests {
		buildSchema(typ, true, outDir, logger)
	}
	for _, typ := range responses {
		buildSchema(typ, false, outDir, logger)
	}
}

func buildSchema(v any, request bool, outDir string, logger Logger) {
	schema := jsonschema.Reflect(v)
	serialized, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		logger.Fatal("Failed to marshal JSON schema", "err", err)
	}

	typeName := strings.Split(schema.Ref, "/")[2]
	fileName := fmt.Sprintf("%s.json", strings.ToLower(typeName))

	var targetDir string
	if request {
		targetDir = filepath.Join(outDir, "request")
	} else {
		targetDir = filepath.Join(outDir, "response")
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		logger.Fatal("Failed to create directory", "dir", targetDir, "err", err)
	}

	filePath := filepath.Join(targetDir, fileName)
	if err := os.WriteFile(filePath, serialized, 0o644); err != nil {
		logger.Fatal("Failed to write schema file", "file", filePath, "err", err)
	}

	logger.Info("Generated schema", "file", filePath)
}
