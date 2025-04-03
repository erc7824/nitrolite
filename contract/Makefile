.PHONY: anvil deploy-counter deploy-custody-lite deploy

# Default anvil private key
PRIVATE_KEY ?= 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80

# Default anvil RPC URL
RPC_URL ?= http://localhost:8545

# Chain ID for local development
CHAIN_ID ?= 1337

localnet:
	@echo "Starting local Anvil chain..."
	anvil --chain-id $(CHAIN_ID) --accounts 15 --balance 300

deploy-counter:
	@echo "Deploying Counter... 0x5fbdb2315678afecb367f032d93f642f64180aa3"
	forge script script/Counter.s.sol:CounterScript --rpc-url $(RPC_URL) --private-key $(PRIVATE_KEY) --broadcast

deploy-erc-20:
	@echo "Deploying ERC20... 0xe7f1725e7734ce288f8367e1bb143e90bb3f0512"
	forge script ./script/DeployFundERC20.s.sol --broadcast --rpc-url $(RPC_URL) --sig "run(string)" "test test test test test test test test test test test junk"

deploy: deploy-counter deploy-erc-20
	@echo "All contracts deployed!"

# Helper to verify contract on explorer (for testnet/mainnet)
verify:
	@echo "Verifying contract..."
	forge verify-contract --chain-id $(CHAIN_ID) --compiler-version 0.8.28 $(CONTRACT) $(CONTRACT_NAME)A
