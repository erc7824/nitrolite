package user_v1

// Handler manages user data operations and provides RPC endpoints.
type Handler struct {
	useStoreInTx StoreTxProvider
}

// NewHandler creates a new Handler instance with the provided dependencies.
func NewHandler(
	useStoreInTx StoreTxProvider,
) *Handler {
	return &Handler{
		useStoreInTx: useStoreInTx,
	}
}
