package sign

import (
	"fmt"
)

type ECDSASigValidator struct {
	recoverer AddressRecoverer
}

func NewECDSASigValidator() *ECDSASigValidator {
	recoverer, _ := NewAddressRecoverer(TypeEthereum)
	return &ECDSASigValidator{
		recoverer: recoverer,
	}
}

func (s *ECDSASigValidator) Recover(data, sig []byte) (string, error) {
	address, err := s.recoverer.RecoverAddress(data, sig)
	if err != nil {
		return "", err
	}
	return address.String(), nil
}

func (s *ECDSASigValidator) Verify(wallet string, data, sig []byte) error {
	address, err := s.Recover(data, sig)
	if err != nil {
		return err
	}

	if address != wallet {
		return fmt.Errorf("invalid signature")
	}
	return nil
}
