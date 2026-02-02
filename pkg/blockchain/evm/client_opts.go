package evm

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

type ClientAllowanceCheck struct {
	RequireAllowanceCheck bool
}

func (ch ClientAllowanceCheck) apply(c *Client) {
	c.requireCheckAllowance = ch.RequireAllowanceCheck
}

type ClientBalanceCheck struct {
	RequireBalanceCheck bool
}

func (ch ClientBalanceCheck) apply(c *Client) {
	c.requireCheckBalance = ch.RequireBalanceCheck
}

type ClientFeeCheck struct {
	RequirePositiveNativeBalance bool
}

func (ch ClientFeeCheck) apply(c *Client) {
	if !ch.RequirePositiveNativeBalance {
		c.checkFeeFn = func(ctx context.Context, account common.Address) error {
			return nil
		}
	} else {
		c.checkFeeFn = getDefaultCheckFeeFn(c.evmClient)
	}
}

func getDefaultCheckFeeFn(evmClient EVMClient) func(ctx context.Context, account common.Address) error {
	return func(ctx context.Context, account common.Address) error {
		balance, err := evmClient.BalanceAt(ctx, account, nil)
		if err != nil {
			return err
		}

		if balance.Sign() <= 0 {
			return fmt.Errorf("insufficient balance for fee")
		}

		return nil
	}
}
