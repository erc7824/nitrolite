package sign

import (
	"fmt"
	"strings"
)

type SigValidator struct {
	recoverer AddressRecoverer
}

func NewSigValidator(sigType Type) (*SigValidator, error) {
	recoverer, err := NewAddressRecoverer(sigType)
	if err != nil {
		return nil, err
	}

	return &SigValidator{
		recoverer: recoverer,
	}, nil
}

func (s *SigValidator) Recover(data, sig []byte) (string, error) {
	address, err := s.recoverer.RecoverAddress(data, sig)
	if err != nil {
		return "", err
	}
	return address.String(), nil
}

func (s *SigValidator) Verify(wallet string, data, sig []byte) error {
	address, err := s.Recover(data, sig)
	if err != nil {
		return err
	}

	if !strings.EqualFold(address, wallet) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}
