---
sidebar_position: 7
title: FAQ
description: State channels frequently asked questions
keywords: [erc7824, statechannels, nitro, sdk, faq, state channels, ethereum scaling, L2]
tags:
  - erc7824
  - nitro
  - faq
---
# FAQ

### State channels FAQ

<details>
  <summary><strong>What is ERC-7824?</strong></summary>
  <p>
    ERC-7824 is a proposed standard for cross-chain trade execution systems that use state channels. It defines structures and interfaces to enable efficient, secure, and scalable off-chain interactions while leveraging the blockchain for finality and dispute resolution.
  </p>
</details>

<details>
  <summary><strong>What is a state channel?</strong></summary>
  <p>
    A state channel can be thought of as an account with multiple balances (often just two). The owners of that account can update those balances according to predefined rules, which are enforceable on a blockchain. This enables peer-to-peer games, payments, and other few-user applications to safely trade blockchain assets with extremely low latency, low cost, and high throughput without requiring trust in a third party.
  </p>
</details>

<details>
  <summary><strong>How do state channels work?</strong></summary>
  <p>
    1. <strong>Setup:</strong> Participants lock assets into a blockchain-based smart contract.<br/>
    2. <strong>Off-Chain Updates:</strong> Transactions or updates occur off-chain through cryptographically signed messages.<br/>
    3. <strong>Finalization:</strong> The final state is submitted on-chain for settlement, or disputes are resolved if necessary.
  </p>
</details>

<details>
  <summary><strong>What are the benefits of state channels?</strong></summary>
  <p>
    - <strong>High Performance:</strong> Transactions are processed off-chain, providing low latency and high throughput.<br/>
    - <strong>Cost Efficiency:</strong> Minimal blockchain interactions significantly reduce gas fees.<br/>
    - <strong>Privacy:</strong> Off-chain interactions keep intermediate states confidential.<br/>
    - <strong>Flexibility:</strong> Supports a wide range of applications, including multi-chain trading.
  </p>
</details>

<details>
  <summary><strong>What kind of applications use state channels?</strong></summary>
  <p>
    State channels enable the redistribution of assets according to arbitrary logic, making them suitable for:
    <ul>
      <li><strong>Games:</strong> Peer-to-peer poker or other interactive games.</li>
      <li><strong>Payments:</strong> Microtransactions and conditional payments.</li>
      <li><strong>Swaps:</strong> Atomic swaps between assets.</li>
      <li><strong>Decentralized Trading:</strong> Real-time, high-frequency trading applications.</li>
    </ul>
  </p>
</details>

<details>
  <summary><strong>How is Nitro Protocol implemented?</strong></summary>
  <p>
    - <strong>On-Chain Components:</strong> Implemented in Solidity and included in the npm package <code>@statechannels/nitro-protocol</code>.<br/>
    - <strong>Off-Chain Components:</strong> A reference implementation provided through <code>go-nitro</code>, a lightweight client written in Go.
  </p>
</details>

<details>
  <summary><strong>Where is Nitro Protocol being used?</strong></summary>
  <p>
    The maintainers of Nitro Protocol are actively integrating it into the Filecoin Retrieval Market and the Filecoin Virtual Machine, enabling decentralized and efficient content distribution.
  </p>
</details>

<details>
  <summary><strong>What is the structure of a state in state channels?</strong></summary>
  <p>
    A state consists of:
    <ol>
      <li><strong>Fixed Part:</strong> Immutable properties like participants, nonce, app definition, and challenge duration.</li>
      <li><strong>Variable Part:</strong> Changeable properties like outcomes, application data, and turn numbers.</li>
    </ol>
    In Nitro, participants sign a keccak256 hash of both these parts to commit to a particular state. The <code>turnNum</code> determines the version of the state, while <code>isFinal</code> can trigger an instant finalization when fully countersigned.
  </p>
</details>

<details>
  <summary><strong>What is a challenge duration?</strong></summary>
  <p>
    The challenge duration is a time window during which disputes can be raised on-chain. If no disputes are raised, the state channel finalizes according to its latest agreed state. In Nitro, it is set at channel creation and cannot be changed later. During this period, an unresponsive or dishonest participant can be forced to progress the channel state via on-chain transactions.
  </p>
</details>

<details>
  <summary><strong>How do disputes get resolved in state channels?</strong></summary>
  <p>
    Participants can:
    <ol>
      <li>Submit signed updates to the blockchain as evidence.</li>
      <li>Resolve disputes based on turn numbers and application-specific rules stored in an on-chain <code>appDefinition</code>.</li>
      <li>Finalize the channel after the challenge duration if no valid disputes arise.</li>
    </ol>
    Nitro Protocol introduces a <strong>challenge</strong> mechanism, enabling any participant to push the channel state on-chain, forcing the other side to respond. This ensures that unresponsive or malicious actors cannot stall the channel indefinitely.
  </p>
</details>

<details>
  <summary><strong>What is the typical channel lifecycle in Nitro Protocol?</strong></summary>
  <p>
    A direct channel often follows these stages:
    <br/><br/>
    1. <strong>Proposed:</strong> A participant signs the initial (prefund) state with <code>turnNum=0</code>.<br/>
    2. <strong>ReadyToFund:</strong> All participants countersign the prefund state, ensuring it is safe to deposit on-chain.<br/>
    3. <strong>Funded:</strong> Deposits appear on-chain, and participants exchange a postfund state (turnNum=1).<br/>
    4. <strong>Running:</strong> The channel can be updated off-chain by incrementing <code>turnNum</code> and exchanging signatures.<br/>
    5. <strong>Finalized:</strong> A state with <code>isFinal=true</code> is fully signed. No more updates are possible; the channel can pay out according to the final outcome.
  </p>
</details>

<details>
  <summary><strong>How do I finalize and withdraw funds from a direct channel in Nitro?</strong></summary>
  <p>
    - <strong>Finalization (Happy Path):</strong> If a fully signed state with <code>isFinal=true</code> exists off-chain, any participant can call <code>conclude</code> on the Nitro Adjudicator to finalize instantly.<br/>
    - <strong>Finalization (Dispute Path):</strong> If participants are unresponsive or disagree, one party can <code>challenge</code> with the latest supported state. After the challenge window, the channel is finalized if unchallenged or out-of-date states are resolved.<br/>
    - <strong>Withdrawing:</strong> Once finalized, participants use the <code>transfer</code> or <code>concludeAndTransferAllAssets</code> method to claim their allocations on-chain.
  </p>
</details>
