package custody

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
)

func GetTokenBalance(chainID uint32, chainRPC string,
	tokenAddress, walletAddress common.Address) (*big.Int, error) {
	client, err := ethclient.Dial(chainRPC)
	if err != nil {
		return nil, err
	}

	token, err := NewIERC20(tokenAddress, client)
	if err != nil {
		return nil, err
	}

	balance, err := token.BalanceOf(&bind.CallOpts{}, walletAddress)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func ApproveAllowance(wallet sign.Signer, chainID uint32, chainRPC string,
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
