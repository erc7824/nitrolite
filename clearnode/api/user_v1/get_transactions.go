package user_v1

import (
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

// GetTransactions retrieves transaction history for a user with optional filters.
func (h *Handler) GetTransactions(c *rpc.Context) {
	var req rpc.UserV1GetTransactionsRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	var paginationParams core.PaginationParams
	if req.Pagination != nil {
		paginationParams.Offset = req.Pagination.Offset
		paginationParams.Limit = req.Pagination.Limit
		paginationParams.Sort = req.Pagination.Sort
	}

	var transactions []core.Transaction
	var metadata core.PaginationMetadata

	err := h.useStoreInTx(func(store Store) error {
		var err error
		transactions, metadata, err = store.GetUserTransactions(req.Wallet, req.Asset, req.TxType, req.FromTime, req.ToTime, &paginationParams)
		if err != nil {
			return err
		}

		return nil
	})

	response := rpc.UserV1GetTransactionsResponse{
		Transactions: []rpc.TransactionV1{},
		Metadata:     *mapPaginationMetadataV1(metadata),
	}

	for _, tx := range transactions {
		response.Transactions = append(response.Transactions, mapTransactionV1(tx))
	}

	if err != nil {
		c.Fail(err, "failed to retrieve transactions")
		return
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
