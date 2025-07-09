package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

// ExportOptions contains options for exporting transactions
type ExportOptions struct {
	AccountID   string
	AssetSymbol string
	TxType      *TransactionType
	OutputDir   string
}

// TransactionExporter handles exporting transactions to CSV
type TransactionExporter struct {
	db *gorm.DB
}

// NewTransactionExporter creates a new transaction exporter
func NewTransactionExporter(db *gorm.DB, logger Logger) *TransactionExporter {
	return &TransactionExporter{
		db: db,
	}
}

// ExportToCSV exports transactions to CSV format
func (e *TransactionExporter) ExportToCSV(writer io.Writer, options ExportOptions) error {
	transactions, err := GetLedgerTransactionsWithTags(e.db, NewAccountID(options.AccountID), options.AssetSymbol, options.TxType)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{"ID", "Type", "FromAccount", "FromAccountTag", "ToAccount", "ToAccountTag", "AssetSymbol", "Amount", "CreatedAt"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write header to CSV: %w", err)
	}

	// Write transactions
	for _, tx := range transactions {
		row := []string{
			fmt.Sprintf("%d", tx.ID),
			tx.Type.String(),
			tx.FromAccount,
			tx.FromAccountTag,
			tx.ToAccount,
			tx.ToAccountTag,
			tx.AssetSymbol,
			tx.Amount.String(),
			tx.CreatedAt.String(),
		}
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write row to CSV: %w", err)
		}
	}
	return nil
}

// ExportToFile exports transactions to a CSV file
func (e *TransactionExporter) ExportToFile(options ExportOptions) (string, error) {
	if err := os.MkdirAll(options.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", options.OutputDir, err)
	}

	fileName := filepath.Join(options.OutputDir, fmt.Sprintf("transactions_%s.csv", options.AccountID))
	file, err := os.Create(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to create CSV file %s: %w", fileName, err)
	}
	defer file.Close()

	if err := e.ExportToCSV(file, options); err != nil {
		return "", fmt.Errorf("failed to export to CSV: %w", err)
	}

	return fileName, nil
}

func runExportTransactionsCli(logger Logger) {
	logger = logger.NewSystem("export-transactions")
	if len(os.Args) < 3 || len(os.Args) > 5 {
		logger.Fatal("Usage: clearnode export-transactions <accountID> [asset] [txType]")
	}

	accountID := os.Args[2]

	var assetSymbol string
	var txType *TransactionType

	// Optional asset parameter
	if len(os.Args) > 3 {
		assetSymbol = os.Args[3]
	}

	// Optional transaction type parameter
	if len(os.Args) > 4 {
		parsedType, err := parseLedgerTransactionType(os.Args[4])
		if err != nil {
			logger.Fatal("Invalid transaction type", "type", os.Args[4], "error", err)
		}
		txType = &parsedType
	}

	config, err := LoadConfig(logger)
	if err != nil {
		logger.Fatal("Failed to load configuration", "error", err)
	}

	db, err := ConnectToDB(config.dbConf)
	if err != nil {
		logger.Fatal("Failed to setup database", "error", err)
	}

	exporter := NewTransactionExporter(db, logger)
	options := ExportOptions{
		AccountID:   accountID,
		AssetSymbol: assetSymbol,
		TxType:      txType,
		OutputDir:   "csv_export",
	}

	fileName, err := exporter.ExportToFile(options)
	if err != nil {
		logger.Fatal("Failed to export transactions", "error", err)
	}
	logger.Info("Successfully exported transactions", "file", fileName)
}
