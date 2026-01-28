package sign

import (
	"fmt"
)

type SigValidator struct {
	recoverer AddressRecoverer
}

func NewSigValidator(sigType Type) *SigValidator {
	recoverer, _ := NewAddressRecoverer(sigType)
	return &SigValidator{
		recoverer: recoverer,
	}
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

	if address != wallet {
		return fmt.Errorf("invalid signature")
	}
	return nil
}
