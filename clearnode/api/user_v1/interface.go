package user_v1

import (
	"github.com/erc7824/nitrolite/pkg/core"
)

// Store defines the persistence layer interface for user data management.
// All methods should be implemented to work within database transactions.
type Store interface {
	// GetUserBalances retrieves the balances for a user's wallet.
	GetUserBalances(wallet string) ([]core.BalanceEntry, error)

	// GetUserTransactions retrieves transaction history for a user with optional filters.
	GetUserTransactions(Wallet string,
		Asset *string,
		TxType *core.TransactionType,
		FromTime *uint64,
		ToTime *uint64,
		Paginate *core.PaginationParams) ([]core.Transaction, core.PaginationMetadata, error)
}
