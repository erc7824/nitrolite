package unisig

import (
	"github.com/ethereum/go-ethereum/common"
)

type Signer interface {
	Address() common.Address
	Sign(data []byte) ([]byte, error)
}
