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

	// Map RPC methods to their request/response types
	rpcSchemas := map[RPCMethod]RPCSchemaMapping{
		RPCMethodGetLedgerTransactions: {
			Request:  &GetLedgerTransactionsParams{},
			Response: &TransactionResponse{},
		},
		// Add more RPC methods here as needed
	}

	for method, mapping := range rpcSchemas {
		if mapping.Request != nil {
			buildSchemaWithMethod(mapping.Request, method, true, outDir, logger)
		}
		if mapping.Response != nil {
			buildSchemaWithMethod(mapping.Response, method, false, outDir, logger)
		}
	}
}

type RPCSchemaMapping struct {
	Request  any
	Response any
}

func buildSchemaWithMethod(v any, method RPCMethod, request bool, outDir string, logger Logger) {
	schema := jsonschema.Reflect(v)

	// Add method metadata to the schema
	if schema.Extras == nil {
		schema.Extras = make(map[string]any)
	}
	schema.Extras["rpc_method"] = method.String()

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

	logger.Info("Generated schema", "file", filePath, "method", method)
}
