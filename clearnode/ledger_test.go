package main

import (
	"testing"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestLedgerOperations tests basic ledger operations
func TestLedgerOperations(t *testing.T) {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create tables
	err = db.AutoMigrate(&Entry{})
	if err != nil {
		t.Fatalf("Failed to migrate tables: %v", err)
	}

	// Test cases
	testCases := []struct {
		name        string
		accountID   string
		assetSymbol string
		amount      decimal.Decimal
		expectError bool
	}{
		{
			name:        "Record credit entry",
			accountID:   "test-account",
			assetSymbol: "TEST",
			amount:      decimal.NewFromInt(100),
			expectError: false,
		},
		{
			name:        "Record debit entry",
			accountID:   "test-account",
			assetSymbol: "TEST",
			amount:      decimal.NewFromInt(-50),
			expectError: false,
		},
		{
			name:        "Record zero entry",
			accountID:   "test-account",
			assetSymbol: "TEST",
			amount:      decimal.Zero,
			expectError: false,
		},
		{
			name:        "Record credit to different account",
			accountID:   "other-account",
			assetSymbol: "TEST",
			amount:      decimal.NewFromInt(200),
			expectError: false,
		},
		{
			name:        "Record with different asset",
			accountID:   "test-account",
			assetSymbol: "OTHER",
			amount:      decimal.NewFromInt(300),
			expectError: false,
		},
	}

	// Create a wallet ledger
	walletLedger := &WalletLedger{
		wallet: "test-wallet",
		db:     db,
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := walletLedger.Record(tc.accountID, tc.assetSymbol, tc.amount)
			
			if tc.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			
			// If not zero amount, verify the entry was created
			if !tc.amount.IsZero() {
				var entries []Entry
				result := db.Where("account_id = ? AND asset_symbol = ? AND wallet = ?", 
					tc.accountID, tc.assetSymbol, walletLedger.wallet).
					Find(&entries)
				
				if result.Error != nil {
					t.Errorf("Failed to retrieve entries: %v", result.Error)
					return
				}
				
				found := false
				for _, entry := range entries {
					if (tc.amount.IsPositive() && entry.Credit.Equal(tc.amount)) ||
					   (tc.amount.IsNegative() && entry.Debit.Equal(tc.amount.Abs())) {
						found = true
						break
					}
				}
				
				if !found {
					t.Errorf("Entry with amount %s not found", tc.amount)
				}
			}
		})
	}

	// Test Balance
	balance, err := walletLedger.Balance("test-account", "TEST")
	if err != nil {
		t.Errorf("Failed to get balance: %v", err)
	}
	
	// 100 - 50 + 0 = 50
	expectedBalance := decimal.NewFromInt(50)
	if !balance.Equal(expectedBalance) {
		t.Errorf("Expected balance %s, got %s", expectedBalance, balance)
	}
	
	// Test GetBalances
	balances, err := walletLedger.GetBalances("test-account")
	if err != nil {
		t.Errorf("Failed to get balances: %v", err)
	}
	
	if len(balances) != 2 { // TEST and OTHER
		t.Errorf("Expected 2 balances, got %d", len(balances))
	}
	
	// Test GetEntries
	entries, err := walletLedger.GetEntries("test-account", "TEST")
	if err != nil {
		t.Errorf("Failed to get entries: %v", err)
	}
	
	// Should be 2 entries for test-account/TEST (credit and debit, zero entry is not recorded)
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

// TestLedgerPublisherIntegration tests that the ledger properly integrates with the publisher
func TestLedgerPublisherIntegration(t *testing.T) {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create tables
	err = db.AutoMigrate(&Entry{})
	if err != nil {
		t.Fatalf("Failed to migrate tables: %v", err)
	}
	
	// Create a test signer
	testPrivateKey := "1111111111111111111111111111111111111111111111111111111111111111"
	signer, err := NewSigner(testPrivateKey)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}
	
	// Create a publisher
	publisher := NewLedgerPublisher(signer)
	defer publisher.Stop()
	
	// Set as global publisher
	SetPublisher(publisher)
	
	// Get a wallet ledger instance
	walletLedger := GetWalletLedger(db, "test-wallet")
	
	// Verify the publisher was set
	if walletLedger.publisher == nil {
		t.Errorf("Expected wallet ledger to have a publisher")
	}
	
	// Test making an entry (we can't easily verify the publish happened without mocks)
	err = walletLedger.Record("test-account", "TEST", decimal.NewFromInt(100))
	if err != nil {
		t.Fatalf("Failed to record entry: %v", err)
	}
	
	// Verify the entry was created
	var entries []Entry
	if err := db.Where("account_id = ? AND asset_symbol = ? AND wallet = ? AND credit = ?", 
		"test-account", "TEST", "test-wallet", decimal.NewFromInt(100)).Find(&entries).Error; err != nil {
		t.Fatalf("Failed to retrieve entries: %v", err)
	}
	
	// Check for at least 1 matching entry
	if len(entries) < 1 {
		t.Fatalf("Entry with credit 100 not found")
	}
	
	// Test the global publisher functions
	globalPub := GetPublisher()
	if globalPub == nil {
		t.Errorf("Expected to get global publisher")
	}
	
	// Test stopping and unsetting the publisher
	SetPublisher(nil)
	
	// Verify it was unset
	globalPub = GetPublisher()
	if globalPub != nil {
		t.Errorf("Expected global publisher to be nil after unsetting")
	}
	
	// Get a new wallet ledger and verify no publisher
	newLedger := GetWalletLedger(db, "test-wallet")
	if newLedger.publisher != nil {
		t.Errorf("Expected new wallet ledger to have no publisher")
	}
	
	// Test recording without publisher (should still work)
	err = newLedger.Record("test-account", "TEST2", decimal.NewFromInt(200))
	if err != nil {
		t.Fatalf("Failed to record entry without publisher: %v", err)
	}
	
	// Verify the entry was created
	var entries2 []Entry
	if err := db.Where("account_id = ? AND asset_symbol = ? AND wallet = ?", 
		"test-account", "TEST2", "test-wallet").Find(&entries2).Error; err != nil {
		t.Fatalf("Failed to retrieve entries: %v", err)
	}
	
	if len(entries2) < 1 {
		t.Fatalf("Expected at least 1 entry, got %d", len(entries2))
	}
	
	// Find the specific entry we just created
	var found bool
	for _, e := range entries2 {
		if e.Credit.Equal(decimal.NewFromInt(200)) {
			found = true
			break
		}
	}
	
	if !found {
		t.Fatalf("Entry with credit 200 not found")
	}
}