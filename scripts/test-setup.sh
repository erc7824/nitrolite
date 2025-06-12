#!/bin/bash

# Test setup script for Nitrolite SDK integration tests
set -e

echo "🚀 Setting up Nitrolite SDK test environment..."

# Configuration
ANVIL_PORT=8545
ANVIL_HOST="127.0.0.1"
ACCOUNTS=10
BALANCE=10000

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if a port is in use
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 0  # Port is in use
    else
        return 1  # Port is free
    fi
}

# Function to kill process on port
kill_port() {
    local port=$1
    if check_port $port; then
        print_warning "Port $port is in use. Killing existing process..."
        lsof -ti:$port | xargs kill -9 2>/dev/null || true
        sleep 2
    fi
}

# Function to start Anvil
start_anvil() {
    print_status "Starting Anvil on port $ANVIL_PORT..."
    
    # Kill any existing process on the port
    kill_port $ANVIL_PORT
    
    # Start Anvil in background
    anvil \
        --port $ANVIL_PORT \
        --host $ANVIL_HOST \
        --accounts $ACCOUNTS \
        --balance $BALANCE \
        --gas-limit 12000000 \
        --gas-price 20000000000 \
        --block-time 1 \
        --silent > anvil.log &
    
    ANVIL_PID=$!
    echo $ANVIL_PID > anvil.pid
    
    # Wait for Anvil to start
    print_status "Waiting for Anvil to start..."
    for i in {1..30}; do
        if curl -s -X POST \
            -H "Content-Type: application/json" \
            -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
            http://$ANVIL_HOST:$ANVIL_PORT >/dev/null 2>&1; then
            print_success "Anvil is running on http://$ANVIL_HOST:$ANVIL_PORT"
            return 0
        fi
        sleep 1
        print_status "Waiting... ($i/30)"
    done
    
    print_error "Failed to start Anvil"
    return 1
}

# Function to deploy contracts
deploy_contracts() {
    print_status "Deploying contracts..."
    
    cd "$(dirname "$0")/../contract"
    
    # Check if forge is installed
    if ! command -v forge &> /dev/null; then
        print_error "Forge not found. Please install Foundry."
        return 1
    fi
    
    # Install dependencies
    print_status "Installing contract dependencies..."
    forge install --no-commit 2>/dev/null || true
    
    # Build contracts
    print_status "Building contracts..."
    forge build
    
    # Deploy contracts
    print_status "Deploying contracts to local network..."
    forge script script/Deploy.s.sol \
        --rpc-url http://$ANVIL_HOST:$ANVIL_PORT \
        --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
        --broadcast \
        --legacy || {
        print_error "Contract deployment failed!"
        return 1
    }
    
    # Extract deployed addresses from forge broadcast output
    print_status "Extracting deployed contract addresses..."
    
    # Get the latest broadcast file
    BROADCAST_DIR="broadcast/Deploy.s.sol/31337"
    if [ -d "$BROADCAST_DIR" ]; then
        LATEST_RUN=$(ls -t "$BROADCAST_DIR"/run-*.json | head -1)
        if [ -f "$LATEST_RUN" ]; then
            print_status "Found deployment data in: $LATEST_RUN"
            
            # Extract addresses using multiple methods for robustness
            CUSTODY_ADDR=$(jq -r '.transactions[] | select(.contractName == "Custody" and .transactionType == "CREATE2") | .contractAddress' "$LATEST_RUN" 2>/dev/null)
            ADJUDICATOR_ADDR=$(jq -r '.transactions[] | select(.contractName == "Dummy" and .transactionType == "CREATE2") | .contractAddress' "$LATEST_RUN" 2>/dev/null)
            TOKEN_ADDR=$(jq -r '.transactions[] | select(.contractName == "TestERC20" and .transactionType == "CREATE2") | .contractAddress' "$LATEST_RUN" 2>/dev/null)
            
            # Fallback: try different JSON structure
            if [ -z "$CUSTODY_ADDR" ] || [ "$CUSTODY_ADDR" = "null" ]; then
                CUSTODY_ADDR=$(jq -r '.transactions[] | select(.transactionType == "CREATE") | .contractAddress' "$LATEST_RUN" 2>/dev/null | head -1)
            fi
            if [ -z "$ADJUDICATOR_ADDR" ] || [ "$ADJUDICATOR_ADDR" = "null" ]; then
                ADJUDICATOR_ADDR=$(jq -r '.transactions[] | select(.transactionType == "CREATE") | .contractAddress' "$LATEST_RUN" 2>/dev/null | sed -n '2p')
            fi
            if [ -z "$TOKEN_ADDR" ] || [ "$TOKEN_ADDR" = "null" ]; then
                TOKEN_ADDR=$(jq -r '.transactions[] | select(.transactionType == "CREATE") | .contractAddress' "$LATEST_RUN" 2>/dev/null | sed -n '3p')
            fi
            
            # Set environment variables
            export CUSTODY_CONTRACT_ADDRESS="$CUSTODY_ADDR"
            export ADJUDICATOR_CONTRACT_ADDRESS="$ADJUDICATOR_ADDR"
            export TEST_TOKEN_CONTRACT_ADDRESS="$TOKEN_ADDR"
        fi
    fi
    
    # Fallback to deterministic addresses if extraction failed
    if [ -z "$CUSTODY_CONTRACT_ADDRESS" ] || [ "$CUSTODY_CONTRACT_ADDRESS" = "null" ]; then
        print_warning "Could not extract contract addresses from deployment, using deterministic addresses"
        export CUSTODY_CONTRACT_ADDRESS="0x5FbDB2315678afecb367f032d93F642f64180aa3"
        export ADJUDICATOR_CONTRACT_ADDRESS="0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512"
        export TEST_TOKEN_CONTRACT_ADDRESS="0x9fE46736679d2D9a65F0992F2272dE9f3c7fa6e0"
    fi
    
    # Set RPC URLs
    export ANVIL_RPC_URL="http://$ANVIL_HOST:$ANVIL_PORT"
    export ETH_RPC_URL="http://$ANVIL_HOST:$ANVIL_PORT"
    
    # Save addresses to a file for easy access
    cat > contract_addresses.test.env << EOF
ANVIL_RPC_URL=$ANVIL_RPC_URL
ETH_RPC_URL=$ETH_RPC_URL
CUSTODY_CONTRACT_ADDRESS=$CUSTODY_CONTRACT_ADDRESS
ADJUDICATOR_CONTRACT_ADDRESS=$ADJUDICATOR_CONTRACT_ADDRESS
TEST_TOKEN_CONTRACT_ADDRESS=$TEST_TOKEN_CONTRACT_ADDRESS
EOF
    
    print_success "Contract addresses saved to contract_addresses.test.env"
    print_status "📋 Deployed Contract Addresses:"
    print_status "   Custody: $CUSTODY_CONTRACT_ADDRESS"
    print_status "   Adjudicator: $ADJUDICATOR_CONTRACT_ADDRESS"
    print_status "   Test Token: $TEST_TOKEN_CONTRACT_ADDRESS"
    print_status "   RPC URL: $ANVIL_RPC_URL"
    
    cd - >/dev/null
    print_success "Contract deployment completed"
}

# Function to run SDK tests
run_sdk_tests() {
    print_status "Running SDK tests..."
    
    cd "$(dirname "$0")/../sdk"
    
    # Install dependencies
    print_status "Installing SDK dependencies..."
    npm ci
    
    # Run unit tests
    print_status "Running unit tests..."
    npm run test
    
    # Run integration tests
    print_status "Running integration tests..."
    ANVIL_RPC_URL="http://$ANVIL_HOST:$ANVIL_PORT" \
    CUSTODY_CONTRACT_ADDRESS="$CUSTODY_CONTRACT_ADDRESS" \
    ADJUDICATOR_CONTRACT_ADDRESS="$ADJUDICATOR_CONTRACT_ADDRESS" \
    TEST_TOKEN_CONTRACT_ADDRESS="$TEST_TOKEN_CONTRACT_ADDRESS" \
    npm run test:integration
    
    cd - >/dev/null
    print_success "SDK tests completed"
}

# Function to run Go tests
run_go_tests() {
    print_status "Running Go tests..."
    
    cd "$(dirname "$0")/../clearnode"
    
    # Check if go is installed
    if ! command -v go &> /dev/null; then
        print_warning "Go not found. Skipping Go tests."
        return 0
    fi
    
    # Download dependencies
    print_status "Downloading Go dependencies..."
    go mod download
    
    # Run tests
    print_status "Running Go integration tests..."
    ETH_RPC_URL="http://$ANVIL_HOST:$ANVIL_PORT" \
    CUSTODY_CONTRACT_ADDRESS="$CUSTODY_CONTRACT_ADDRESS" \
    ADJUDICATOR_CONTRACT_ADDRESS="$ADJUDICATOR_CONTRACT_ADDRESS" \
    TEST_TOKEN_CONTRACT_ADDRESS="$TEST_TOKEN_CONTRACT_ADDRESS" \
    go test -v -race ./pkg/testing/...
    
    cd - >/dev/null
    print_success "Go tests completed"
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up..."
    
    # Kill Anvil if it's running
    if [ -f anvil.pid ]; then
        ANVIL_PID=$(cat anvil.pid)
        if kill -0 $ANVIL_PID 2>/dev/null; then
            print_status "Stopping Anvil (PID: $ANVIL_PID)..."
            kill $ANVIL_PID 2>/dev/null || true
        fi
        rm -f anvil.pid
    fi
    
    # Remove log files
    rm -f anvil.log
    
    print_success "Cleanup completed"
}

# Main execution
main() {
    # Trap to ensure cleanup on exit
    trap cleanup EXIT

    print_status "Starting Nitrolite test suite..."
    
    # Check prerequisites
    if ! command -v anvil &> /dev/null; then
        print_error "Anvil not found. Please install Foundry."
        exit 1
    fi
    
    if ! command -v npm &> /dev/null; then
        print_error "NPM not found. Please install NPM."
        exit 1
    fi
    
    # Start test environment
    start_anvil || exit 1
    
    # Deploy contracts
    deploy_contracts
    
    # Run tests
    run_sdk_tests || exit 1
    run_go_tests
    
    cleanup
    print_success "All tests completed successfully! 🎉"
}

# Function to display contract addresses for FE/BE integration
show_addresses() {
    if [ -f "contract_addresses.test.env" ]; then
        print_success "Contract addresses for FE/BE integration tests:"
        echo ""
        echo "🔗 Add these to your .env file or export them:"
        echo ""
        cat contract_addresses.test.env
        echo ""
        echo "📋 Or source the file directly:"
        echo "   source contract_addresses.test.env"
        echo ""
        echo "🌐 Frontend usage example:"
        echo "   NEXT_PUBLIC_ANVIL_RPC_URL=$ANVIL_RPC_URL"
        echo "   NEXT_PUBLIC_CUSTODY_ADDRESS=$CUSTODY_CONTRACT_ADDRESS"
        echo ""
        echo "🔧 Backend usage example:"
        echo "   export ETH_RPC_URL=$ETH_RPC_URL"
        echo "   export CUSTODY_CONTRACT_ADDRESS=$CUSTODY_CONTRACT_ADDRESS"
    else
        print_error "No contract addresses found. Run deployment first:"
        echo "   $0 deploy"
    fi
}

# Parse command line arguments
case "${1:-}" in
    "start-anvil")
        start_anvil
        print_success "Anvil started with PID $(cat anvil.pid). Use 'cleanup' command to stop."
        ;;
    "deploy")
        deploy_contracts
        ;;
    "sdk-tests")
        run_sdk_tests
        ;;
    "go-tests")
        run_go_tests
        ;;
    "addresses"|"addr"|"contracts")
        show_addresses
        ;;
    "cleanup")
        cleanup
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  start-anvil  Start Anvil test node"
        echo "  deploy       Deploy contracts"
        echo "  sdk-tests    Run SDK tests"
        echo "  go-tests     Run Go tests"
        echo "  addresses    Show deployed contract addresses"
        echo "  cleanup      Stop Anvil and cleanup"
        echo "  help         Show this help"
        echo ""
        echo "Run without arguments to execute full test suite."
        ;;
    "")
        main
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Use '$0 help' for usage information."
        exit 1
        ;;
esac 