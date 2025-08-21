package unisig

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Signer interface {
	Address() common.Address
	Sign(msg []byte) (Signature, error)
}

type Signature []byte

func (s Signature) MarshalJSON() ([]byte, error) {
	return json.Marshal(hexutil.Encode(s))
}

func (s *Signature) UnmarshalJSON(data []byte) error {
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return err
	}
	decoded, err := hexutil.Decode(hexStr)
	if err != nil {
		return err
	}
	*s = decoded
	return nil
}

func (s Signature) String() string {
	return hexutil.Encode(s)
}
