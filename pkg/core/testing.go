package core

// mockAssetStore is a simple mock implementation of AssetStore for testing
type mockAssetStore struct {
	decimalsMap map[string]uint8
}

func newMockAssetStore() *mockAssetStore {
	return &mockAssetStore{
		decimalsMap: make(map[string]uint8),
	}
}

func (m *mockAssetStore) AddToken(blockchainID uint64, tokenAddress string, decimals uint8) {
	key := m.makeKey(blockchainID, tokenAddress)
	m.decimalsMap[key] = decimals
}

func (m *mockAssetStore) makeKey(blockchainID uint64, tokenAddress string) string {
	return string(rune(blockchainID)) + ":" + tokenAddress
}

func (m *mockAssetStore) GetAssetDecimals(asset string) (uint8, error) {
	// Not used in PackState, but required by interface
	return 0, nil
}

func (m *mockAssetStore) GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error) {
	key := m.makeKey(blockchainID, tokenAddress)
	if decimals, ok := m.decimalsMap[key]; ok {
		return decimals, nil
	}
	return 0, nil
}
