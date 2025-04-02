.PHONY: anvil deploy-counter deploy-custody-lite deploy

# Default anvil private key
PRIVATE_KEY ?= 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80

# Default anvil RPC URL
RPC_URL ?= http://localhost:8545

# Chain ID for local development
CHAIN_ID ?= 31337

localnet:
	@echo "Starting local Anvil chain..."
	anvil --chain-id $(CHAIN_ID)

deploy:
	@echo "All contracts deployed!"
	forge script script/Counter.s.sol:CounterScript --rpc-url $(RPC_URL) --private-key $(PRIVATE_KEY) --broadcast
	forge script script/Custody.s.sol:CustodyScript --rpc-url $(RPC_URL) --private-key $(PRIVATE_KEY) --broadcast

# Helper to verify contract on explorer (for testnet/mainnet)
verify:
	@echo "Verifying contract..."
	forge verify-contract --chain-id $(CHAIN_ID) --compiler-version 0.8.28 $(CONTRACT) $(CONTRACT_NAME)A
