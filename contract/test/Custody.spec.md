# Custody.sol Test Specification

This document outlines the comprehensive test cases needed to verify the functionality, security, and edge cases for the Custody.sol contract.

Test the following scenarios:

To prepare the test, participant must deposit treasury in their account before opening channels with smaller amounts.
by using the IDeposit intergrate methods.

## 1. Channel Creation and Opening

### 1.1 Channel Creation

- Verify that a channel can be created using valid parameters.
- Ensure that the channel ID is calculated consistently and remains unique.
- Verify that creating a channel with the same participants but a different nonce results in a distinct channel.
- Ensure that a channel is created with a target margin and that the target margin is greater than or equal to 0.

### 1.2 Open Function

- Confirm that depositing assets for a new channel transitions its status from VOID to PARTIAL.
- Verify that a participant cannot deposit an amount exceeding the target margin.
- For a channel in the PARTIAL status, ensure that a new participant must deposit the remaining amount (i.e., target margin minus the amount deposited by the initiator) to fully fund the channel.
- Verify that the channel contains exactly 2 participants.
- Ensure that the user joining the channel is listed in the participants array.
- Make sure adjudicator is called when the second participant joined the channel.
- Confirm that adjudicator transitions the channel state from PARTIAL to ACTIVE once fully funded, with the deposited amounts recorded and assigned to each participant.

### 1.3 Invalid Open Attempts

- Creating a channel with less than or more than 2 participants is rejected.
- Verify that attempting to deposit an amount greater than the target margin is rejected.
- Ensure that an open attempt with a deposit less than the required remaining amount in a PARTIAL status fails.
- Confirm that a user not present in the participant list is not allowed to join the channel.
- Validate that deposit attempts on a channel already in an ACTIVE status are rejected.
- Check that deposits with invalid amounts (e.g., negative values) are properly rejected.
- Ensure that any deposit attempt when the channel is in a status other than VOID or PARTIAL is not accepted.
- Test with invalid adjudicator address (zero address)

## 2. Channel Closing Flows

### 2.1 Close Function (Cooperative)

- Channel can be closed only when adjudicator returns that the next status is FINAL. Any other status must reject.
- Ensure that channel status is changed to FINAL.
- Ensure funds are properly distributed according to adjudicator's response.

### 2.2 Invalid Close Attempts

- Test closing a non-existent channel
- Test closing with invalid status (not FINAL from adjudicator)
- Test closing with insufficient funds for allocations
- Test closing an already finalized channel

## 3. Challenge Mechanism

### 3.1 Challenge Function

- Only ACTIVE channel can be challenged. All other statuses must be rejected.
- Test initiating a challenge with valid status
- Test challenge timer starts correctly
- Test challenging with a status that adjudicator validates as FINAL (immediate close)

### 3.2 Counter-Challenge

- Test responding to a challenge with a newer valid status
- Test challenge timer resets with new challenge
- Test multiple back-and-forth challenges

### 3.3 Invalid Challenge Attempts

- Test challenge with invalid state (rejected by adjudicator)
- Test challenge on a non-existent channel
- Test challenge after challenge period expired (should revert)
- Test challenge on already finalized channel

## 4. Checkpoint Mechanism

### 4.1 Checkpoint Function

- Test recording a valid state without starting challenge
- Test checkpoint with state that adjudicator validates as FINAL (immediate close)
- Test checkpoint updates metadata correctly

### 4.2 Invalid Checkpoint Attempts

- Test checkpoint with invalid state (rejected by adjudicator)
- Test checkpoint during active challenge period
- Test checkpoint on non-existent channel

## 5. Reclaim Function

### 5.1 Successful Reclaims

- Test reclaim after challenge period expires
- Test proper fund distribution after reclaim
- Test channel state updates to FINAL after reclaim

### 5.2 Invalid Reclaim Attempts

- Test reclaim before challenge period expires
- Test reclaim on channel not in INVALID state
- Test reclaim on non-existent channel
- Test reclaim on already finalized channel

## 6. Fund Management

### 6.1 Deposit Tracking

- Test deposit tracking for multiple tokens
- Test deposit tracking across multiple deposits
- Test deposit balance updates correctly after distributions

### 6.2 Fund Distribution

- Test distribution to multiple participants
- Test distribution of multiple token types
- Test distribution when allocation exceeds available funds (should revert)
- Test distribution to contract addresses as destinations

## 7. Integration with Adjudicator

### 7.1 Adjudication Integration

- Test different adjudicator status responses and contract behavior
- Test state validation flow with adjudicator
- Test allocation determination based on adjudicator
- Mock adjudicator to test various responses

## 8. Edge Cases and Security

### 8.1 Re-entrancy Protection

- Test potential re-entrancy vectors during token transfers
- Test callback handling with malicious ERC20 tokens

### 8.2 State Consistency

- Test concurrent operations on same channel
- Test channel state consistency after failed operations

### 8.3 Gas Considerations

- Test gas usage for main operations
- Test with large number of tokens/participants

### 8.4 Signature Verification

- Test signature verification through adjudicator
- Test with invalid signatures
- Test with malformed signature data

## 9. Events and Traceability

### 9.1 Event Emission

- Test event emission for all state-changing functions
- Test event parameter accuracy
- Test event sequence during complex operations
