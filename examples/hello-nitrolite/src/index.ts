/**tutorial-meta
title: "Hello Nitrolite: State Channels Made Simple"
description: "Learn state channels by building your first Nitrolite application - clearer than any quick start guide"
difficulty: "beginner"
estimatedTime: "15 minutes"
technologies: ["TypeScript", "Nitrolite", "Ethereum"]
concepts: ["State Channels", "ClearNode", "Off-chain Updates", "Channel Lifecycle"]
prerequisites: ["Basic TypeScript", "Wallet basics", "Node.js"]
*/

/**tutorial:architecture
# Why State Channels Matter

Traditional blockchain apps are slow and expensive. Every action requires:
- 15+ second confirmation times
- $5-50+ gas fees per transaction
- Limited to ~15 transactions per second

**State channels solve this:**

```
Traditional Blockchain          vs          State Channels
┌─────────────────┐                      ┌─────────────────┐
│  Every action   │                      │  Only open/close│
│  = Gas fees     │                      │  = Gas fees     │
│  = Wait time    │                      │  = Wait time    │
└─────────────────┘                      └─────────────────┘
        💰💰💰                                   💰  ...  💰
        ⏰⏰⏰                                   ⏰        ⏰
                                               ⚡⚡⚡⚡⚡⚡⚡
                                            Instant & Free!
```

**Result**: Instant transactions with zero gas fees between opening and closing.
*/

import { NitroliteClient } from "@erc7824/nitrolite";
import { ethers } from "ethers";

/**tutorial:step Quick Setup - Get Running in 60 Seconds

Let's get you up and running with Nitrolite. Here's everything you need:

## 1. Install Dependencies

First, create a new project and install the required packages:

```bash
mkdir hello-nitrolite
cd hello-nitrolite
npm init -y
npm install @erc7824/nitrolite ethers
npm install -D typescript @types/node ts-node
```

## 2. Setup TypeScript

Create a `tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  }
}
```

## 3. Get Test Funds

You'll need Mumbai MATIC for gas fees:
- Go to https://faucet.polygon.technology/
- Switch to Mumbai network
- Request test MATIC for your wallet

## 4. Run This Example

```bash
npx ts-node src/index.ts
```

That's it! Now let's understand what this code does.
*/

// Simple configuration - just what you need to get started
const CONFIG = {
  clearNodeUrl: "wss://testnet-clearnode.nitrolite.org",
  network: "polygon-mumbai",
  rpcUrl: "https://rpc-mumbai.polygon.technology",
};

/**tutorial:step Creating Your First State Channel App

Now we'll build a simple counter app that demonstrates state channels. This shows the core pattern you'll use in any Nitrolite application.

Here's what our HelloNitrolite class will do:
1. **Initialize**: Set up wallet and Nitrolite client
2. **Connect**: Authenticate with ClearNode
3. **Create Channel**: Open a state channel for our counter
4. **Update State**: Increment counter with instant updates
5. **Close Channel**: Safely finalize everything on-chain

```typescript
class HelloNitrolite {
  private client: NitroliteClient;
  private wallet: ethers.Wallet;
  
  // Methods will be implemented below...
}
```

This is the basic structure every Nitrolite app follows.
*/

class HelloNitrolite {
  private client: NitroliteClient;
  private wallet: ethers.Wallet;

  /**tutorial:concept What Happens Under the Hood

  When you create a NitroliteClient, it sets up several important components:

  1. **Wallet Integration**: Manages your private key securely for signing transactions and state updates
  2. **Network Connection**: Connects to Ethereum (Mumbai testnet) for on-chain operations like opening/closing channels
  3. **ClearNode Link**: Establishes a WebSocket connection for off-chain coordination and state management
  4. **State Management**: Tracks channel states, validates updates, and handles synchronization

  The key insight: most operations happen off-chain through the ClearNode, with blockchain used only for opening, closing, and dispute resolution.
  */

  /**tutorial:step Initialize Your Wallet and Client

  First, we need to set up our wallet and initialize the Nitrolite client. In production, you'd connect to MetaMask, but for this demo we'll create a wallet programmatically.

  ```typescript
  async initialize(): Promise<void> {
    console.log("Starting Hello Nitrolite...");

    // Create wallet (in production, connect to MetaMask instead)
    this.wallet = ethers.Wallet.createRandom();
    console.log(`Wallet: ${this.wallet.address}`);

    // Initialize Nitrolite client
    this.client = new NitroliteClient({
      privateKey: this.wallet.privateKey,
      network: CONFIG.network,
      rpcUrl: CONFIG.rpcUrl,
      clearNodeUrl: CONFIG.clearNodeUrl,
    });

    console.log("Client initialized");
  }
  ```

  **Security Note**: Never hardcode private keys in production! Use MetaMask or secure key management.
  */

  async initialize(): Promise<void> {
    console.log("Starting Hello Nitrolite...");

    // Create wallet (in production, connect to MetaMask instead)
    this.wallet = ethers.Wallet.createRandom();
    console.log(`Wallet: ${this.wallet.address}`);

    // Initialize Nitrolite client
    this.client = new NitroliteClient({
      privateKey: this.wallet.privateKey,
      network: CONFIG.network,
      rpcUrl: CONFIG.rpcUrl,
      clearNodeUrl: CONFIG.clearNodeUrl,
    });

    console.log("Client initialized");
  }

  /**tutorial:step Connect to ClearNode - Your Off-chain Coordinator

  A ClearNode is like a referee in a game - it doesn't hold your money but helps everyone agree on the current state. Here's what it does:

  - **Coordinates state updates** between participants
  - **Validates signatures** to prevent cheating
  - **Provides conflict resolution** mechanisms
  - **Enables instant finality** for off-chain transactions

  The connection process involves authentication using cryptographic signatures:

  ```typescript
  async connect(): Promise<void> {
    console.log("Connecting to ClearNode...");

    // Connect and authenticate in one step
    await this.client.connect();
    await this.client.authenticate();

    console.log("Connected and authenticated");
  }
  ```

  This establishes a secure WebSocket connection and proves your identity using your wallet signature.
  */

  async connect(): Promise<void> {
    console.log("Connecting to ClearNode...");

    // Connect and authenticate in one step
    await this.client.connect();
    await this.client.authenticate();

    console.log("Connected and authenticated");
  }

  /**tutorial:step Open Your Channel - Like Opening a Tab

  Creating a channel is like opening a tab at a bar - you put some money down upfront, then you can order drinks instantly without paying each time. You settle the final bill when you close the tab.

  Here's how to create a channel:

  ```typescript
  async createCounter(): Promise<string> {
    console.log("Creating counter channel...");

    const channelConfig = {
      participants: [this.wallet.address],        // Who can use this channel
      challengeDuration: 3600,                    // 1 hour dispute window
      asset: "0x0000000000000000000000000000000000000000", // ETH address
      initialDeposit: "0",                        // Start with zero for simplicity
    };

    const channelId = await this.client.createChannel(channelConfig);
    console.log(`Counter channel created: ${channelId.slice(0, 8)}...`);

    return channelId;
  }
  ```

  **What happens**: This creates an on-chain smart contract that locks funds and establishes the channel rules. Once created, you can update state instantly off-chain.
  */

  async createCounter(): Promise<string> {
    console.log("Creating counter channel...");

    const channelConfig = {
      participants: [this.wallet.address],
      challengeDuration: 3600, // 1 hour dispute window
      asset: "0x0000000000000000000000000000000000000000", // ETH
      initialDeposit: "0", // Start with zero for simplicity
    };

    const channelId = await this.client.createChannel(channelConfig);
    console.log(`Counter channel created: ${channelId.slice(0, 8)}...`);

    return channelId;
  }

  /**tutorial:step The Magic - Instant State Updates

  This is where state channels shine! Each update happens instantly with zero gas fees. Watch how we update our counter 3 times in under 2 seconds:

  ```typescript
  async demonstrateInstantUpdates(channelId: string): Promise<void> {
    console.log("\nDemonstrating instant updates...");

    let counter = 0;

    // Update counter 3 times - watch how fast this is!
    for (let i = 1; i <= 3; i++) {
      counter++;

      const newState = {
        counter,
        timestamp: Date.now(),
        nonce: i,  // Ensures state progression
      };

      console.log(`Updating counter to ${counter}...`);
      await this.client.updateChannelState(channelId, newState);
      console.log(`Counter = ${counter} (instant!)`);

      // Small delay so you can see the progression
      await new Promise((resolve) => setTimeout(resolve, 500));
    }

    console.log("\nThree instant updates completed!");
    console.log("Gas cost: $0 (only paid gas when opening the channel)");
    console.log("Time: ~1.5 seconds (vs ~45 seconds on mainnet)");
  }
  ```

  **The secret**: We're just updating a local state that's cryptographically signed. No blockchain interaction needed!
  */

  async demonstrateInstantUpdates(channelId: string): Promise<void> {
    console.log("\nDemonstrating instant updates...");

    let counter = 0;

    // Update counter 3 times - watch how fast this is!
    for (let i = 1; i <= 3; i++) {
      counter++;

      const newState = {
        counter,
        timestamp: Date.now(),
        nonce: i,
      };

      console.log(`Updating counter to ${counter}...`);
      await this.client.updateChannelState(channelId, newState);
      console.log(`Counter = ${counter} (instant!)`);

      // Small delay so you can see the progression
      await new Promise((resolve) => setTimeout(resolve, 500));
    }

    console.log("\nThree instant updates completed!");
    console.log("Gas cost: $0 (only paid gas when opening the channel)");
    console.log("Time: ~1.5 seconds (vs ~45 seconds on mainnet)");
  }

  /**tutorial:step Check Your Channel State

  You can query your channel anytime to see the current state. This is useful for building UIs that show real-time data:

  ```typescript
  async showChannelStatus(channelId: string): Promise<void> {
    console.log("\nCurrent channel status:");

    const info = await this.client.getChannelInfo(channelId);
    console.log(`   Status: ${info.status}`);
    console.log(`   Participants: ${info.participants.length}`);
    console.log(`   Current State:`, info.currentState);
  }
  ```

  This returns the latest state that all participants have agreed on.
  */

  async showChannelStatus(channelId: string): Promise<void> {
    console.log("\nCurrent channel status:");

    const info = await this.client.getChannelInfo(channelId);
    console.log(`   Status: ${info.status}`);
    console.log(`   Participants: ${info.participants.length}`);
    console.log(`   Current State:`, info.currentState);
  }

  /**tutorial:step Close Your Channel - Settle the Tab

  When you're done, close the channel to finalize everything on-chain. This is like paying your tab - all the off-chain activity gets settled:

  ```typescript
  async closeChannel(channelId: string): Promise<void> {
    console.log("\nClosing channel...");

    const info = await this.client.getChannelInfo(channelId);
    await this.client.closeChannel(channelId, info.currentState);

    console.log("Channel closure initiated");
    console.log("Final settlement will complete after challenge period");
  }
  ```

  **What happens**: The final state gets published to the blockchain. After the challenge period (1 hour in our example), funds are distributed according to the final state.
  */

  async closeChannel(channelId: string): Promise<void> {
    console.log("\nClosing channel...");

    const info = await this.client.getChannelInfo(channelId);
    await this.client.closeChannel(channelId, info.currentState);

    console.log("Channel closure initiated");
    console.log("Final settlement will complete after challenge period");
  }

  /**tutorial:step Run the Complete Demo

  Now let's put it all together! This method runs through the complete flow:

  ```typescript
  async runDemo(): Promise<void> {
    try {
      // Setup
      await this.initialize();
      await this.connect();

      // Create and use a channel
      const channelId = await this.createCounter();
      await this.demonstrateInstantUpdates(channelId);
      await this.showChannelStatus(channelId);
      await this.closeChannel(channelId);

      // Success message
      console.log("\nDemo completed successfully!");
      console.log("\nWhat you just did:");
      console.log("✓ Connected to Nitrolite infrastructure");
      console.log("✓ Created your first state channel");
      console.log("✓ Performed instant off-chain updates");
      console.log("✓ Closed the channel safely");
      console.log("\nYou're ready to build with state channels!");
    } catch (error) {
      console.error("\nDemo failed:", error);
      console.log("\nTroubleshooting:");
      console.log("• Make sure you have test ETH on Mumbai");
      console.log("• Check your internet connection");
      console.log("• Try again in a few minutes");
    }
  }
  ```

  To run this demo:
  ```bash
  npx ts-node src/index.ts
  ```
  */

  async runDemo(): Promise<void> {
    try {
      // Setup
      await this.initialize();
      await this.connect();

      // Create and use a channel
      const channelId = await this.createCounter();
      await this.demonstrateInstantUpdates(channelId);
      await this.showChannelStatus(channelId);
      await this.closeChannel(channelId);

      // Success message
      console.log("\nDemo completed successfully!");
      console.log("\nWhat you just did:");
      console.log("✓ Connected to Nitrolite infrastructure");
      console.log("✓ Created your first state channel");
      console.log("✓ Performed instant off-chain updates");
      console.log("✓ Closed the channel safely");
      console.log("\nYou're ready to build with state channels!");
    } catch (error) {
      console.error("\nDemo failed:", error);
      console.log("\nTroubleshooting:");
      console.log("• Make sure you have test ETH on Mumbai");
      console.log("• Check your internet connection");
      console.log("• Try again in a few minutes");
    }
  }
}

/**tutorial:concept Real-World Applications
Now that you understand the basics, here's what you can build:

## Gaming
- **Chess/Checkers**: Instant moves, settle winner at end
- **Poker**: Fast betting rounds, final pot distribution
- **Trading Cards**: Quick card battles with instant results

## Payments  
- **Micropayments**: Tip content creators without fees
- **Subscriptions**: Stream payments per second/minute
- **Marketplaces**: Instant escrow for digital goods

## Social
- **Chat Apps**: Pay-per-message to prevent spam  
- **Content Platforms**: Instant monetization for creators
- **Gaming Guilds**: Share loot and rewards instantly

The key insight: any application with frequent state changes benefits from state channels.
*/

/**tutorial:concept Production Checklist
When you're ready to build real applications:

## Security
- Never hardcode private keys
- Validate all state transitions
- Implement proper dispute resolution
- Use secure random number generation

## UX Design  
- Show clear connection status
- Explain gas fees vs instant updates
- Provide offline support
- Handle network failures gracefully

## Performance
- Batch state updates when possible
- Implement efficient state storage
- Monitor ClearNode connectivity
- Plan for scaling to multiple channels

## Testing
- Test channel opening/closing flows
- Verify state update logic
- Simulate network failures
- Test with multiple participants
*/

// Example usage
async function main() {
  const demo = new HelloNitrolite();
  await demo.runDemo();
}

// Run demo if this file is executed directly
if (require.main === module) {
  main().catch(console.error);
}

export default HelloNitrolite;
