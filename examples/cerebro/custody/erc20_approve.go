package custody

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/erc7824/nitrolite/examples/bridge/unisig"
)

func ApproveAllowance(wallet unisig.Signer, chainID uint32, chainRPC string,
	tokenAddress, spenderAddress common.Address, amount *big.Int) error {
	client, err := ethclient.Dial(chainRPC)
	if err != nil {
		return err
	}

	token, err := NewIERC20(tokenAddress, client)
	if err != nil {
		return err
	}

	txOpts := signerTxOpts(wallet, chainID)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to suggest gas price: %w", err)
	}
	txOpts.GasPrice = gasPrice.Add(gasPrice, gasPrice)

	tx, err := token.Approve(txOpts, spenderAddress, amount)
	if err != nil {
		return err
	}

	if _, err := bind.WaitMined(context.Background(), client, tx.Hash()); err != nil {
		return err
	}

	return nil
}
