package core

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/erc7824/nitrolite/pkg/sign"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// PackChannelKeyStateV1 packs the session key state for signing using ABI encoding.
// This is used to generate a deterministic hash that the user signs when registering/updating a session key.
// The user_sig field is excluded from packing since it is the signature itself.
func PackChannelKeyStateV1(sessionKey string, metadataHash common.Hash) ([]byte, error) {
	args := abi.Arguments{
		{Type: abi.Type{T: abi.AddressTy}},              // session_key
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // hashed metadata
	}

	packed, err := args.Pack(
		common.HexToAddress(sessionKey),
		metadataHash,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pack session key state: %w", err)
	}

	return crypto.Keccak256(packed), nil
}

func GetChannelSessionKeyAuthMetadataHashV1(version uint64, assets []string, expiresAt int64) (common.Hash, error) {
	stringArrayType, err := abi.NewType("string[]", "", nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to create string array type: %w", err)
	}

	metadtataArgs := abi.Arguments{
		{Type: abi.Type{T: abi.UintTy, Size: 64}}, // version
		{Type: stringArrayType},                   // assets
		{Type: abi.Type{T: abi.UintTy, Size: 64}}, // expires_at (unix timestamp)
	}

	packedMetadataArgs, err := metadtataArgs.Pack(
		version,
		assets,
		uint64(expiresAt),
	)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack metadata args: %w", err)
	}

	hashedMetadata := crypto.Keccak256Hash(packedMetadataArgs)
	return hashedMetadata, nil
}

func ValidateChannelSessionKeyAuthSigV1(state ChannelSessionKeyStateV1) error {
	metadataHash, err := GetChannelSessionKeyAuthMetadataHashV1(state.Version, state.Assets, state.ExpiresAt.Unix())
	if err != nil {
		return fmt.Errorf("failed to get metadata hash: %w", err)
	}

	packed, err := PackChannelKeyStateV1(state.SessionKey, metadataHash)
	if err != nil {
		return fmt.Errorf("failed to pack session key state: %w", err)
	}

	authSigBytes, err := hexutil.Decode(state.UserSig)
	if err != nil {
		return fmt.Errorf("failed to decode user signature: %w", err)
	}

	recoverer, err := sign.NewAddressRecoverer(sign.TypeEthereumMsg)
	if err != nil {
		return fmt.Errorf("failed to create address recoverer: %w", err)
	}

	recoveredAddr, err := recoverer.RecoverAddress(packed, authSigBytes)
	if err != nil {
		return fmt.Errorf("failed to recover address from signature: %w", err)
	}

	if !strings.EqualFold(recoveredAddr.String(), state.UserAddress) {
		return fmt.Errorf("invalid signature: recovered address %s does not match wallet %s", recoveredAddr.String(), state.UserAddress)
	}

	return nil
}

type ChannelSessionSignerTypeV1 uint8

const (
	ChannelSessionSignerTypeV1_Wallet     ChannelSessionSignerTypeV1 = 0x01
	ChannelSessionSignerTypeV1_SessionKey ChannelSessionSignerTypeV1 = 0x02
)

type ChannelWalletSignerV1 struct {
	sign.Signer
}

func NewChannelWalletSignerV1(signer sign.Signer) (*ChannelWalletSignerV1, error) {
	return &ChannelWalletSignerV1{
		Signer: signer,
	}, nil
}

func (s *ChannelWalletSignerV1) Sign(data []byte) (sign.Signature, error) {
	sig, err := s.Signer.Sign(data)
	if err != nil {
		return sign.Signature{}, err
	}

	return append([]byte{byte(ChannelSessionSignerTypeV1_Wallet)}, sig...), nil
}

type ChannelSessionKeySignerV1 struct {
	sign.Signer

	metadataHash common.Hash
	authSig      []byte
}

func NewChannelSessionKeySignerV1(signer sign.Signer, metadataHash, authSig string) (*ChannelSessionKeySignerV1, error) {
	authSigBytes, err := hexutil.Decode(authSig)
	if err != nil {
		return nil, fmt.Errorf("failed to decode auth signature: %w", err)
	}

	return &ChannelSessionKeySignerV1{
		Signer:       signer,
		metadataHash: common.HexToHash(metadataHash),
		authSig:      authSigBytes,
	}, nil
}

func (s *ChannelSessionKeySignerV1) Sign(data []byte) (sign.Signature, error) {
	sessionKeySig, err := s.Signer.Sign(data)
	if err != nil {
		return sign.Signature{}, err
	}

	fullSig, err := encodeChannelSessionKeySignature(
		channelSessionKeyAuthorization{
			SessionKey:    common.HexToAddress(s.Signer.PublicKey().Address().String()),
			MetadataHash:  s.metadataHash,
			AuthSignature: s.authSig,
		},
		sessionKeySig,
	)
	if err != nil {
		return sign.Signature{}, fmt.Errorf("failed to encode session key signature: %w", err)
	}

	return append([]byte{byte(ChannelSessionSignerTypeV1_SessionKey)}, fullSig...), nil
}

type ChannelSigValidatorV1 struct {
	recoverer         sign.AddressRecoverer
	verifyPermissions VerifyChannelSessionKePermissionsV1
}

type VerifyChannelSessionKePermissionsV1 func(walletAddr, sessionKeyAddr, metadataHash string) (bool, error)

func NewChannelSigValidatorV1(permissionsVerifier VerifyChannelSessionKePermissionsV1) *ChannelSigValidatorV1 {
	recoverer, err := sign.NewAddressRecoverer(sign.TypeEthereumMsg)
	if err != nil {
		panic(fmt.Sprintf("failed to create address recoverer: %v", err))
	}

	return &ChannelSigValidatorV1{
		recoverer:         recoverer,
		verifyPermissions: permissionsVerifier,
	}
}

func (s *ChannelSigValidatorV1) Recover(data, sig []byte) (string, error) {
	if len(sig) < 1 {
		return "", fmt.Errorf("invalid signature: too short")
	}

	signerType := ChannelSessionSignerTypeV1(sig[0])
	switch signerType {
	case ChannelSessionSignerTypeV1_Wallet:
		addr, err := s.recoverer.RecoverAddress(data, sig[1:])
		if err != nil {
			return "", fmt.Errorf("failed to recover wallet address: %w", err)
		}
		return addr.String(), nil
	case ChannelSessionSignerTypeV1_SessionKey:
		// Decode: (SessionKeyAuthorization memory skAuth, bytes memory skSignature) =
		//     abi.decode(signature, (SessionKeyAuthorization, bytes));
		skAuth, skSignature, err := decodeChannelSessionKeySignature(sig[1:])
		if err != nil {
			return "", fmt.Errorf("failed to decode session key signature: %w", err)
		}

		// Step 1: Verify participant authorized this session key
		// authMessage = _toSigningData(skAuth) = abi.encode(skAuth.sessionKey, skAuth.metadataHash)
		packedAuth, err := PackChannelKeyStateV1(skAuth.SessionKey.Hex(), skAuth.MetadataHash)
		if err != nil {
			return "", fmt.Errorf("failed to pack auth data: %w", err)
		}

		walletAddr, err := s.recoverer.RecoverAddress(packedAuth, skAuth.AuthSignature)
		if err != nil {
			return "", fmt.Errorf("failed to recover wallet address from auth signature: %w", err)
		}

		// Step 2: Verify session key signed the state data
		sessionKeyAddr, err := s.recoverer.RecoverAddress(data, skSignature)
		if err != nil {
			return "", fmt.Errorf("failed to recover session key address: %w", err)
		}

		if !strings.EqualFold(sessionKeyAddr.String(), skAuth.SessionKey.Hex()) {
			return "", fmt.Errorf("session key mismatch: recovered %s, expected %s", sessionKeyAddr.String(), skAuth.SessionKey.Hex())
		}

		ok, err := s.verifyPermissions(walletAddr.String(), sessionKeyAddr.String(), common.Hash(skAuth.MetadataHash).String())
		if err != nil {
			return "", err
		}

		if !ok {
			return "", fmt.Errorf("session key does not have permission to sign for this data")
		}
		// VerifyChannelSessionKeyPermissions(walletAddr, sessionKey, asset, metadataHash) (bool, error)

		return walletAddr.String(), nil
	default:
		return "", fmt.Errorf("invalid signature: unknown signer type %d", signerType)
	}
}

// GenerateSessionKeyStateIDV1 generates a deterministic ID from user_address, session_key, and version.
func GenerateSessionKeyStateIDV1(userAddress, sessionKey string, version uint64) (string, error) {
	args := abi.Arguments{
		{Type: abi.Type{T: abi.AddressTy}},        // user_address
		{Type: abi.Type{T: abi.AddressTy}},        // session_key
		{Type: abi.Type{T: abi.UintTy, Size: 64}}, // version
	}

	packed, err := args.Pack(
		common.HexToAddress(userAddress),
		common.HexToAddress(sessionKey),
		version,
	)
	if err != nil {
		return "", fmt.Errorf("failed to pack session key state ID: %w", err)
	}

	return crypto.Keccak256Hash(packed).Hex(), nil
}
func (s *ChannelSigValidatorV1) Verify(wallet string, data, sig []byte) error {
	address, err := s.Recover(data, sig)
	if err != nil {
		return err
	}

	if !strings.EqualFold(address, wallet) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

// channelSessionKeyAuthorization matches the Solidity SessionKeyAuthorization struct.
type channelSessionKeyAuthorization struct {
	SessionKey    common.Address
	MetadataHash  [32]byte
	AuthSignature []byte
}

func encodeChannelSessionKeySignature(skAuth channelSessionKeyAuthorization, skSignature []byte) ([]byte, error) {
	args, err := getChannelSessionKeyArgsV1()
	if err != nil {
		return nil, fmt.Errorf("failed to get channel session key args: %w", err)
	}

	packed, err := args.Pack(skAuth, skSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to pack session key signature: %w", err)
	}

	return packed, nil
}

func decodeChannelSessionKeySignature(data []byte) (skAuth channelSessionKeyAuthorization, skSignature []byte, err error) {
	args, err := getChannelSessionKeyArgsV1()
	if err != nil {
		return skAuth, nil, fmt.Errorf("failed to get channel session key args: %w", err)
	}

	values, err := args.Unpack(data)
	if err != nil {
		return skAuth, nil, fmt.Errorf("failed to unpack session key signature: %w", err)
	}

	if len(values) != 2 {
		return skAuth, nil, fmt.Errorf("expected 2 values from unpack, got %d", len(values))
	}

	// The tuple unpacks to an anonymous struct; use reflect to access fields
	authStruct := reflect.ValueOf(values[0])
	if authStruct.Kind() != reflect.Struct || authStruct.NumField() != 3 {
		return skAuth, nil, fmt.Errorf("unexpected skAuth structure: kind=%s, fields=%d", authStruct.Kind(), authStruct.NumField())
	}

	sessionKey, ok := authStruct.Field(0).Interface().(common.Address)
	if !ok {
		return skAuth, nil, fmt.Errorf("unexpected type for sessionKey: %T", authStruct.Field(0).Interface())
	}
	metadataHash, ok := authStruct.Field(1).Interface().([32]byte)
	if !ok {
		return skAuth, nil, fmt.Errorf("unexpected type for metadataHash: %T", authStruct.Field(1).Interface())
	}
	authSignature, ok := authStruct.Field(2).Interface().([]byte)
	if !ok {
		return skAuth, nil, fmt.Errorf("unexpected type for authSignature: %T", authStruct.Field(2).Interface())
	}

	skAuth = channelSessionKeyAuthorization{
		SessionKey:    sessionKey,
		MetadataHash:  metadataHash,
		AuthSignature: authSignature,
	}

	skSignature, ok = values[1].([]byte)
	if !ok {
		return skAuth, nil, fmt.Errorf("unexpected type for skSignature: %T", values[1])
	}

	return skAuth, skSignature, nil
}

func getChannelSessionKeyArgsV1() (*abi.Arguments, error) {
	skAuthType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "sessionKey", Type: "address"},
		{Name: "metadataHash", Type: "bytes32"},
		{Name: "authSignature", Type: "bytes"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create skAuth type: %w", err)
	}

	bytesType, err := abi.NewType("bytes", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create bytes type: %w", err)
	}

	return &abi.Arguments{
		{Type: skAuthType},
		{Type: bytesType},
	}, nil
}
