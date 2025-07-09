package main

import (
	"bytes"
	"encoding/csv"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestTransactionExporter_ExportToCSV(t *testing.T) {
	db, cleanup := setupTestDB(t)
	t.Cleanup(cleanup)

	logger := NewLoggerIPFS("test")
	exporter := NewTransactionExporter(db, logger)

	// Create test data
	account1 := "0x1234567890123456789012345678901234567890"
	account2 := "0x0987654321098765432109876543210987654321"
	account3 := "app-session-123"

	// Create user tags for wallet accounts
	tag1, err := GenerateOrRetrieveUserTag(db, account1)
	require.NoError(t, err)
	tag2, err := GenerateOrRetrieveUserTag(db, account2)
	require.NoError(t, err)

	// Create test transactions
	_, err = RecordLedgerTransaction(db, TransactionTypeTransfer, NewAccountID(account1), NewAccountID(account2), "usdc", decimal.NewFromInt(100))
	require.NoError(t, err)

	_, err = RecordLedgerTransaction(db, TransactionTypeDeposit, NewAccountID(account2), NewAccountID(account1), "eth", decimal.NewFromInt(50))
	require.NoError(t, err)

	_, err = RecordLedgerTransaction(db, TransactionTypeAppDeposit, NewAccountID(account1), NewAccountID(account3), "usdc", decimal.NewFromInt(25))
	require.NoError(t, err)

	t.Run("Export", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		options := ExportOptions{
			AccountID: account1,
		}

		err := exporter.ExportToCSV(&buf, options)
		require.NoError(t, err)

		// Parse CSV output
		reader := csv.NewReader(&buf)
		records, err := reader.ReadAll()
		require.NoError(t, err)

		// Should have header + 3 transactions (account1 is involved in all)
		require.Len(t, records, 4)

		// Check header
		expectedHeader := []string{"ID", "Type", "FromAccount", "FromAccountTag", "ToAccount", "ToAccountTag", "AssetSymbol", "Amount", "CreatedAt"}
		require.Equal(t, expectedHeader, records[0])

		// Verify transaction data
		foundTx1, foundTx2, foundTx3 := false, false, false
		for i, record := range records[1:] {
			t.Logf("Row %d: %v", i+1, record)

			switch record[1] { // Type column
			case "transfer":
				require.Equal(t, account1, record[2]) // FromAccount
				require.Equal(t, tag1.Tag, record[3]) // FromAccountTag
				require.Equal(t, account2, record[4]) // ToAccount
				require.Equal(t, tag2.Tag, record[5]) // ToAccountTag
				require.Equal(t, "usdc", record[6])   // AssetSymbol
				require.Equal(t, "100", record[7])    // Amount
				foundTx1 = true
			case "deposit":
				require.Equal(t, account2, record[2]) // FromAccount
				require.Equal(t, tag2.Tag, record[3]) // FromAccountTag
				require.Equal(t, account1, record[4]) // ToAccount
				require.Equal(t, tag1.Tag, record[5]) // ToAccountTag
				require.Equal(t, "eth", record[6])    // AssetSymbol
				require.Equal(t, "50", record[7])     // Amount
				foundTx2 = true
			case "app_deposit":
				require.Equal(t, account1, record[2]) // FromAccount
				require.Equal(t, tag1.Tag, record[3]) // FromAccountTag
				require.Equal(t, account3, record[4]) // ToAccount
				require.Empty(t, record[5])           // ToAccountTag (app account has no tag)
				require.Equal(t, "usdc", record[6])   // AssetSymbol
				require.Equal(t, "25", record[7])     // Amount
				foundTx3 = true
			}
		}

		require.True(t, foundTx1, "Transfer transaction should be present")
		require.True(t, foundTx2, "Deposit transaction should be present")
		require.True(t, foundTx3, "App deposit transaction should be present")
	})

	t.Run("ExportWithAssetFilter", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		options := ExportOptions{
			AccountID:   account1,
			AssetSymbol: "usdc",
		}

		err := exporter.ExportToCSV(&buf, options)
		require.NoError(t, err)

		// Parse CSV output
		reader := csv.NewReader(&buf)
		records, err := reader.ReadAll()
		require.NoError(t, err)

		// Should have header + 2 USDC transactions
		require.Len(t, records, 3)

		// All transactions should be USDC
		for _, record := range records[1:] {
			require.Equal(t, "usdc", record[6])
		}
	})

	t.Run("ExportWithTypeFilter", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		txType := TransactionTypeTransfer
		options := ExportOptions{
			AccountID: account1,
			TxType:    &txType,
		}

		err := exporter.ExportToCSV(&buf, options)
		require.NoError(t, err)

		// Parse CSV output
		reader := csv.NewReader(&buf)
		records, err := reader.ReadAll()
		require.NoError(t, err)

		// Should have header + 1 transfer transaction
		require.Len(t, records, 2)

		// Should be transfer type
		require.Equal(t, "transfer", records[1][1])
	})

	t.Run("ExportNoTransactions", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		options := ExportOptions{
			AccountID: "0xNonExistentAccount",
		}

		err := exporter.ExportToCSV(&buf, options)
		require.NoError(t, err)

		// Parse CSV output
		reader := csv.NewReader(&buf)
		records, err := reader.ReadAll()
		require.NoError(t, err)

		// Should have only header
		require.Len(t, records, 1)
		expectedHeader := []string{"ID", "Type", "FromAccount", "FromAccountTag", "ToAccount", "ToAccountTag", "AssetSymbol", "Amount", "CreatedAt"}
		require.Equal(t, expectedHeader, records[0])
	})
}
