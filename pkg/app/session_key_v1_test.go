package app

import (
	"strings"
	"testing"
	"time"

	"github.com/erc7824/nitrolite/pkg/sign"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a signer
func createTestSigner(t *testing.T) (sign.Signer, string) {
	pk, err := crypto.GenerateKey()
	require.NoError(t, err)
	pkHex := hexutil.Encode(crypto.FromECDSA(pk))
	
	rawSigner, err := sign.NewEthereumRawSigner(pkHex)
	require.NoError(t, err)
	
	msgSigner, err := sign.NewEthereumMsgSignerFromRaw(rawSigner)
	require.NoError(t, err)
	
	return msgSigner, rawSigner.PublicKey().Address().String()
}

func TestGenerateSessionKeyStateIDV1(t *testing.T) {
	userAddr := "0x1111111111111111111111111111111111111111"
	sessionKey := "0x2222222222222222222222222222222222222222"
	version := uint64(1)

	id1, err := GenerateSessionKeyStateIDV1(userAddr, sessionKey, version)
	require.NoError(t, err)
	assert.NotEmpty(t, id1)

	id2, err := GenerateSessionKeyStateIDV1(userAddr, sessionKey, version)
	require.NoError(t, err)
	assert.Equal(t, id1, id2)

	id3, err := GenerateSessionKeyStateIDV1(userAddr, sessionKey, version+1)
	require.NoError(t, err)
	assert.NotEqual(t, id1, id3)
}

func TestPackAppSessionKeyStateV1(t *testing.T) {
	state := AppSessionKeyStateV1{
		UserAddress:    "0x1111111111111111111111111111111111111111",
		SessionKey:     "0x2222222222222222222222222222222222222222",
		Version:        1,
		ApplicationIDs: []string{"app1"},
		AppSessionIDs:  []string{"session1"},
		ExpiresAt:      time.Now().Add(time.Hour),
		UserSig:        "0xSig",
	}

	packed, err := PackAppSessionKeyStateV1(state)
	require.NoError(t, err)
	assert.NotEmpty(t, packed)
	assert.Len(t, packed, 32)
}

func TestAppSessionSignerV1(t *testing.T) {
	baseSigner, _ := createTestSigner(t)
	data := []byte("hello")

	t.Run("WalletSigner", func(t *testing.T) {
		signer, err := NewAppSessionWalletSignerV1(baseSigner)
		require.NoError(t, err)
		
		sig, err := signer.Sign(data)
		require.NoError(t, err)
		assert.Equal(t, byte(AppSessionSignerTypeV1_Wallet), sig[0])
	})

	t.Run("SessionKeySigner", func(t *testing.T) {
		signer, err := NewAppSessionKeySignerV1(baseSigner)
		require.NoError(t, err)
		
		sig, err := signer.Sign(data)
		require.NoError(t, err)
		assert.Equal(t, byte(AppSessionSignerTypeV1_SessionKey), sig[0])
	})
	
	t.Run("InvalidType", func(t *testing.T) {
		_, err := newAppSessionSignerV1(0xFF, baseSigner)
		assert.Error(t, err)
	})
}

func TestAppSessionKeyValidatorV1(t *testing.T) {
	userSigner, userAddr := createTestSigner(t)
	sessionSigner, sessionKeyAddr := createTestSigner(t)
	data := []byte("hello")

	// Setup validator
	validator := NewAppSessionKeySigValidatorV1(func(skAddr string) (string, error) {
		if strings.EqualFold(skAddr, sessionKeyAddr) {
			return userAddr, nil
		}
		return "", assert.AnError
	})

	t.Run("VerifyWalletSignature", func(t *testing.T) {
		signer, _ := NewAppSessionWalletSignerV1(userSigner)
		sig, _ := signer.Sign(data)
		
		err := validator.Verify(userAddr, data, sig)
		assert.NoError(t, err)
		
		recovered, err := validator.Recover(data, sig)
		assert.NoError(t, err)
		assert.Equal(t, strings.ToLower(userAddr), strings.ToLower(recovered))
	})

	t.Run("VerifySessionKeySignature", func(t *testing.T) {
		signer, _ := NewAppSessionKeySignerV1(sessionSigner)
		sig, _ := signer.Sign(data)
		
		err := validator.Verify(userAddr, data, sig)
		assert.NoError(t, err)
		
		recovered, err := validator.Recover(data, sig)
		assert.NoError(t, err)
		assert.Equal(t, userAddr, recovered)
	})
	
	t.Run("InvalidSignature", func(t *testing.T) {
		err := validator.Verify(userAddr, data, []byte{0x00}) // Too short
		assert.Error(t, err)
		
		err = validator.Verify(userAddr, data, []byte{0xFF, 0x01}) // Unknown type
		assert.Error(t, err)
	})
	
	t.Run("WrongOwner", func(t *testing.T) {
		signer, _ := NewAppSessionWalletSignerV1(sessionSigner) // Signed by session key but claims to be wallet
		sig, _ := signer.Sign(data)
		
		// Should recover sessionKeyAddr, which != userAddr
		err := validator.Verify(userAddr, data, sig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signature")
	})
}

func TestAppSessionSignerTypeV1_String(t *testing.T) {
	assert.Equal(t, "wallet", AppSessionSignerTypeV1_Wallet.String())
	assert.Equal(t, "session_key", AppSessionSignerTypeV1_SessionKey.String())
	assert.Equal(t, "unknown(255)", AppSessionSignerTypeV1(255).String())
}
