package node_v1

import (
	"github.com/erc7824/nitrolite/pkg/rpc"
)

// GetAssets retrieves the current assets of the Node.
func (h *Handler) GetAssets(c *rpc.Context) {
	var req rpc.NodeV1GetAssetsRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	assets, err := h.memoryStore.GetAssets(req.ChainID)
	if err != nil {
		c.Fail(err, "failed to retrieve assets")
		return
	}

	response := rpc.NodeV1GetAssetsResponse{
		Assets: []rpc.AssetV1{},
	}

	for _, asset := range assets {
		response.Assets = append(response.Assets, mapAssetV1(asset))
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
