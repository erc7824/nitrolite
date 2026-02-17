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

#### Remaining Flows

The following communication flows are not yet documented but will be added in future iterations:

- **Remaining app session endpoints** are not affected and will be added here later. The only new requirement includes creating app sessions with 0 allocations, and participants depositing one by one. Now app session deposits are limited to one participant deposit at a time.

- **home channel deposit** - Similar to home channel creation with deposit, but for existing channels
- **home chain migration** - Cross-chain state migration between home channels
- **off-chain transfer to a non-existing user** - Handles receiver account creation during transfer

---

**Note:**  This directory contains ongoing work on Nitrolite V1 protocol architecture.

## Project Structure

The following is a suggested project structure that may change as the implementation evolves:

```t
cerebro/
clearnode/
    api/ # AppSessionService
        app_session/
        channel/
        user/
        node/
    config/
        migrations/ # database migration files
            postgres/
            sqlite/
    metric/
        prometheus/ # Prometheus metrics exporter
    store/
        db/ # struct Database implements Store interface
        memory/ # may include in-memory store for Asset's, Blockchain's etc.
    blockchain_worker.go # service: BlockchainWorker, BWStore
    config.go
    event_handler.go # service: EventHandler
    eth_listener.go # service: SmartContractListener, SCLStore (TBD)
    main.go # 1st - monolithic clearnode implementation; then - refactor into microservices
    rpc_router.go # RPC Router binding RPC methods to handlers
contract/
docs/
pkg/
    amm/
    app_session/
    blockchain/
        evm/ # Client implementations for EVM-based blockchains
    core/ # Client interface (Create, Checkpoint, Challenge etc.), PackState, UnpackState, TransitionValidator, functions related to State build
    rpc/ # Node, Client, Requests, Responses, Events, Errors
sdk/
    go/
    ts/ # should include implementations for everything inside /pkg/
test/ # integration test scenarios executed by all SDKs inside sdk/ directory
go.mod
```
