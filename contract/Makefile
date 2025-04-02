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

deploy-counter:
	@echo "Deploying Counter..."
	forge script script/Counter.s.sol:CounterScript --rpc-url $(RPC_URL) --private-key $(PRIVATE_KEY) --broadcast

deploy-custody-lite:
	@echo "Deploying CustodyLite..."
	forge script script/CustodyLite.s.sol:CustodyLiteScript --rpc-url $(RPC_URL) --private-key $(PRIVATE_KEY) --broadcast

deploy: deploy-counter deploy-custody-lite
	@echo "All contracts deployed!"

# Helper to verify contract on explorer (for testnet/mainnet)
verify:
	@echo "Verifying contract..."
	forge verify-contract --chain-id $(CHAIN_ID) --compiler-version 0.8.13 $(CONTRACT) $(CONTRACT_NAME)
