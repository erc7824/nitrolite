# Nitrolite V1 Clearnode Specifications

This directory introduces new Clearnode architecture, models and communication flows to facilitate communication between user, SDK client, Node and Blockchains that will become the core off-chain engine for the Nitrolite V1 Protocol.

## Contents

- **[api.yaml](api.yaml)** - API definitions including types, state transitions, and RPC methods
- **[data_models.mmd](data_models.mmd)** - Data model diagrams
- **[rpc_message.md](rpc_message.md)** - Standardized RPC message format for communication with a Clearnode via WebSocket

### Communication Flows

- **[transfer.mmd](communication_flows/transfer.mmd)** - Off-chain transfer flow
- **[app_session_deposit.mmd](communication_flows/app_session_deposit.mmd)** - Application session deposit
- **[escrow_chan_deposit.mmd](communication_flows/escrow_chan_deposit.mmd)** - Escrow channel deposit
- **[escrow_chan_withdrawal.mmd](communication_flows/escrow_chan_withdrawal.mmd)** - Escrow channel withdrawal
- **[home_chan_creation_from_scratch.mmd](communication_flows/home_chan_creation_from_scratch.mmd)** - Home channel creation
- **[home_chan_withdraw.mmd](communication_flows/home_chan_withdraw.mmd)** - Home channel withdrawal
- **[home_chan_withdraw_on_create_from_state.mmd](communication_flows/home_chan_withdraw_on_create_from_state.mmd)** - State-based channel creation with withdrawal

---

**Note:**  This directory contains ongoing work on Nitrolite V1 protocol architecture.
