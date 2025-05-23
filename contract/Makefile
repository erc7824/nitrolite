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

deploy-flag-adjudicator:
	# @echo "Deploying FlagAdjudicator... 0x5fbdb2315678afecb367f032d93f642f64180aa3"
	forge create --broadcast --optimizer-runs 1000000 --via-ir ./test/mocks/FlagAdjudicator.sol:FlagAdjudicator --rpc-url http://localhost:8545 --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 --constructor-args true

deploy-erc-20:
	@echo "Deploying ERC20... 0xe7f1725e7734ce288f8367e1bb143e90bb3f0512"
	forge script ./script/DeployFundERC20.s.sol --broadcast --rpc-url $(RPC_URL) --sig "run(string)" "test test test test test test test test test test test junk"

deploy-custody:
	@echo "Deploying Custody... 0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE"
	forge script script/DeployCustody.s.sol:DeployCustodyScript --rpc-url $(RPC_URL) --private-key $(PRIVATE_KEY) --broadcast

deploy: deploy-flag-adjudicator deploy-erc-20 deploy-custody
	@echo "All contracts deployed!"

setup-channel:
	@echo "Setting up channel..."
	forge script script/setupChannel.s.sol:SetupChannelScript --rpc-url $(RPC_URL) --private-key $(PRIVATE_KEY) --broadcast --sig "run(string,address,address,address,uint256)" "test test test test test test test test test test test junk" 0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE 0x5fbdb2315678afecb367f032d93f642f64180aa3 0xe7f1725e7734ce288f8367e1bb143e90bb3f0512 100

setup: setup-channel

deploy-setup: deploy setup

# Helper to verify contract on explorer (for testnet/mainnet)
verify:
	@echo "Verifying contract..."
	forge verify-contract --chain-id $(CHAIN_ID) --compiler-version 0.8.28 $(CONTRACT) $(CONTRACT_NAME)
