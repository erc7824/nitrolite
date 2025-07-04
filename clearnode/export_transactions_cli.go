package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func runExportTransactionsCli(logger Logger) {
	logger = logger.NewSystem("export-transactions")
	if len(os.Args) < 3 {
		logger.Fatal("Usage: clearnode export-transactions <accountID>")
	}

	accountID := os.Args[2]

	config, err := LoadConfig(logger)
	if err != nil {
		logger.Fatal("Failed to load configuration", "error", err)
	}

	db, err := ConnectToDB(config.dbConf)
	if err != nil {
		logger.Fatal("Failed to setup database", "error", err)
	}

	transactions, err := GetLedgerTransactions(db, NewAccountID(accountID), "", nil)
	if err != nil {
		logger.Fatal("Failed to get transactions", "error", err)
	}

	if err := os.MkdirAll("csv_export", 0755); err != nil {
		logger.Fatal("Failed to create directory", "error", err)
	}
	fileName := fmt.Sprintf("csv_export/transactions_%s.csv", accountID)
	file, err := os.Create(fileName)
	if err != nil {
		logger.Fatal("Failed to create CSV file", "error", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "Type", "FromAccount", "ToAccount", "AssetSymbol", "Amount", "CreatedAt"}
	if err := writer.Write(header); err != nil {
		logger.Fatal("Failed to write header to CSV", "error", err)
	}

	// Write transactions
	for _, tx := range transactions {
		row := []string{
			fmt.Sprintf("%d", tx.ID),
			tx.Type.String(),
			tx.FromAccount,
			tx.ToAccount,
			tx.AssetSymbol,
			tx.Amount.String(),
			tx.CreatedAt.String(),
		}
		if err := writer.Write(row); err != nil {
			logger.Fatal("Failed to write row to CSV", "error", err)
		}
	}

	logger.Info("Successfully exported transactions", "file", fileName)
}
