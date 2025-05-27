package main

// The mnemonic DEADCLIC is used to help remember the effect of debit or credit transactions on the relevant accounts.
// DEAD: Debit to increase Expense, Asset and Drawing accounts and CLIC: Credit to increase Liability, Income and Capital accounts.

//                Debit	Credit
// Asset	    Increase	Decrease
// Liability	Decrease	Increase
// Capital	    Decrease	Increase
// Revenue	    Decrease	Increase
// Expense	    Increase	Decrease

// AccountType represents the type of account in the ledger system
type AccountType uint16

const (
	// Assets (1000-1999)
	AssetDefault AccountType = 1000

	// Liabilities (2000-2999)
	LiabilityDefault AccountType = 2000

	// Equity/Capital (3000-3999)
	EquityDefault AccountType = 3000

	// Revenue (4000-4999)
	RevenueDefault AccountType = 4000

	// Expenses (5000-5999)
	ExpenseDefault AccountType = 5000
)
