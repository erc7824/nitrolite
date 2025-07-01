# Clearnet Bridge

This document describes how we can use Clearnode, to bridge assets from one chain to another.

## Flow

1. User opens channels on chains A and B.
2. User deposits `100` USDC on chain A.
3. Clearnode handles `Created`:
- `ledger_user_balance=0`
- `channel_a_token_amount=100`
- `ledger_channel_a_balance=100`
- `channel_b_token_amount=0`
- `ledger_channel_b_balance=0`
1. Clearnode joins the channel on chain A.
2. Clearnode handles `Joined`:
- `ledger_user_balance=100`
- `channel_a_token_amount=100`
- `ledger_channel_a_balance=0`
- `channel_b_token_amount=0`
- `ledger_channel_b_balance=0`
3. User requests resize on chain A with args `allocate-amount=-100` and `resize_amount=0` .
4. Clearnode handles `Resize`:
- `ledger_user_balance=100`
- `channel_a_token_amount=0`
- `ledger_channel_a_balance=0`
- `channel_b_token_amount=0`
- `ledger_channel_b_balance=0`
5.  User requests resize on chain B with args `allocate-amount=100` and `resize_amount=-100`.
6. Clearnode handles `Resize`:
- `ledger_user_balance=0`
- `channel_a_token_amount=0`
- `ledger_channel_a_balance=0`
- `channel_b_token_amount=0`
- `ledger_channel_b_balance=0`

## Golang CLI

1) Use go-prompt
2) implement method connect
