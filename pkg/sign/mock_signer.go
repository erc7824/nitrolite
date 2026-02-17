// Package sign provides mock implementations for testing signature operations.
package sign

import (
	"fmt"
)

var _ Signer = (*MockSigner)(nil)

// MockSigner is a mock implementation of the Signer interface for testing purposes.
// It generates predictable signatures by appending a suffix to the data.
type MockSigner struct {
	publicKey PublicKey
}

// NewMockSigner creates a new MockSigner with the given ID.
// The ID is used to create the underlying mock public key.
func NewMockSigner(id string) *MockSigner {
	return &MockSigner{publicKey: NewMockPublicKey(id)}
}

// Sign generates a mock signature by appending a suffix containing the signer's address.
// This creates predictable signatures useful for testing.
func (m *MockSigner) Sign(data []byte) (Signature, error) {
	sigBytes := append(data, []byte(
		fmt.Sprintf("-signed-by-%s", m.publicKey.Address().String()),
	)...)
	sig := Signature(sigBytes)
	return sig, nil
}

// PublicKey returns the mock public key associated with this signer.
func (m *MockSigner) PublicKey() PublicKey {
	return m.publicKey
}

var _ PublicKey = (*MockPublicKey)(nil)

// MockPublicKey is a mock implementation of the PublicKey interface for testing.
// It stores an ID string that is used as both the key data and address.
type MockPublicKey struct {
	id string
}

// NewMockPublicKey creates a new MockPublicKey with the given ID.
func NewMockPublicKey(id string) *MockPublicKey {
	return &MockPublicKey{id: id}
}

// Address returns a mock address based on the public key's ID.
func (m *MockPublicKey) Address() Address {
	return NewMockAddress(m.id)
}

// Bytes returns the ID as a byte slice.
func (m *MockPublicKey) Bytes() []byte {
	return []byte(m.id)
}

var _ Address = (*MockAddress)(nil)

// MockAddress is a mock implementation of the Address interface for testing.
// It uses a simple string ID as the address representation.
type MockAddress struct {
	id string
}

// NewMockAddress creates a new MockAddress with the given ID.
func NewMockAddress(id string) *MockAddress {
	return &MockAddress{id: id}
}

// String returns the ID as the string representation of the address.
func (m *MockAddress) String() string {
	return m.id
}

// Equals compares this address with another by comparing their string representations.
func (m *MockAddress) Equals(other Address) bool {
	return m.id == other.String()
}
