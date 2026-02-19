package channel_v1

import (
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

// GetChannels retrieves all channels for a user with optional status/asset filtering and pagination.
func (h *Handler) GetChannels(c *rpc.Context) {
	var req rpc.ChannelsV1GetChannelsRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse request")
		return
	}

	if req.Wallet == "" {
		c.Fail(rpc.Errorf("wallet is required"), "missing wallet")
		return
	}

	var limit, offset uint32
	if req.Pagination != nil {
		if req.Pagination.Limit != nil {
			limit = *req.Pagination.Limit
		}
		if req.Pagination.Offset != nil {
			offset = *req.Pagination.Offset
		}
	}

	var channels []core.Channel
	var totalCount uint32

	err := h.useStoreInTx(func(tx Store) error {
		var err error
		channels, totalCount, err = tx.GetUserChannels(req.Wallet, req.Status, req.Asset, limit, offset)
		if err != nil {
			return rpc.Errorf("failed to get channels: %v", err)
		}
		return nil
	})

	if err != nil {
		c.Fail(err, "failed to get channels")
		return
	}

	rpcChannels := make([]rpc.ChannelV1, len(channels))
	for i, ch := range channels {
		rpcChannels[i] = coreChannelToRPC(ch)
	}

	if limit == 0 {
		limit = 100
	}

	pageCount := uint32(0)
	if limit > 0 {
		pageCount = (totalCount + limit - 1) / limit
	}

	page := uint32(1)
	if limit > 0 && offset > 0 {
		page = (offset / limit) + 1
	}

	response := rpc.ChannelsV1GetChannelsResponse{
		Channels: rpcChannels,
		Metadata: rpc.PaginationMetadataV1{
			Page:       page,
			PerPage:    limit,
			TotalCount: totalCount,
			PageCount:  pageCount,
		},
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
