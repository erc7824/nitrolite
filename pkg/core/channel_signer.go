package core

import (
	"fmt"
	"strings"

	"github.com/erc7824/nitrolite/pkg/sign"
	"github.com/ethereum/go-ethereum/common"
)

type ChannelSignerType uint8

const (
	ChannelSignerType_Default    ChannelSignerType = 0x00
	ChannelSignerType_SessionKey ChannelSignerType = 0x01
)

var (
	ChannelSignerTypes = []ChannelSignerType{
		ChannelSignerType_Default,
		ChannelSignerType_SessionKey,
	}
)

type ChannelDefaultSigner struct {
	sign.Signer
}

func NewChannelDefaultSigner(signer sign.Signer) (*ChannelDefaultSigner, error) {
	return &ChannelDefaultSigner{
		Signer: signer,
	}, nil
}

func (s *ChannelDefaultSigner) Sign(data []byte) (sign.Signature, error) {
	sig, err := s.Signer.Sign(data)
	if err != nil {
		return sign.Signature{}, err
	}

	return append([]byte{byte(ChannelSignerType_Default)}, sig...), nil
}

type ChannelSigValidator struct {
	recoverer         sign.AddressRecoverer
	verifyPermissions VerifyChannelSessionKePermissionsV1
}

func NewChannelSigValidator(permissionsVerifier VerifyChannelSessionKePermissionsV1) *ChannelSigValidator {
	recoverer, err := sign.NewAddressRecoverer(sign.TypeEthereumMsg)
	if err != nil {
		panic(fmt.Sprintf("failed to create address recoverer: %v", err))
	}

	return &ChannelSigValidator{
		recoverer:         recoverer,
		verifyPermissions: permissionsVerifier,
	}
}

func (s *ChannelSigValidator) Recover(data, sig []byte) (string, error) {
	if len(sig) < 1 {
		return "", fmt.Errorf("invalid signature: too short")
	}

	signerType := ChannelSignerType(sig[0])
	switch signerType {
	case ChannelSignerType_Default:
		addr, err := s.recoverer.RecoverAddress(data, sig[1:])
		if err != nil {
			return "", fmt.Errorf("failed to recover wallet address: %w", err)
		}
		return addr.String(), nil
	case ChannelSignerType_SessionKey:
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

func (s *ChannelSigValidator) Verify(wallet string, data, sig []byte) error {
	address, err := s.Recover(data, sig)
	if err != nil {
		return err
	}

	if !strings.EqualFold(address, wallet) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}
