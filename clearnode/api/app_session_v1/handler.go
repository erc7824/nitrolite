package app_session_v1

import (
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/sign"
)

// Handler manages app session operations and provides RPC endpoints for app session management.
type Handler struct {
	useStoreInTx  StoreTxProvider
	signer        sign.Signer
	stateAdvancer core.StateAdvancer
	sigValidator  map[SigType]SigValidator
	nodeAddress   string // Node's wallet address
}

// NewHandler creates a new Handler instance with the provided dependencies.
func NewHandler(
	useStoreInTx StoreTxProvider,
	signer sign.Signer,
	stateAdvancer core.StateAdvancer,
	sigValidators map[SigType]SigValidator,
	nodeAddress string,
) *Handler {
	return &Handler{
		useStoreInTx:  useStoreInTx,
		signer:        signer,
		stateAdvancer: stateAdvancer,
		sigValidator:  sigValidators,
		nodeAddress:   nodeAddress,
	}
}
