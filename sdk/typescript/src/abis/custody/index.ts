export { custodyAbi as CustodyAbi } from '../generated';

// Channel lifecycle events
export const ChannelCreatedEvent = 'ChannelCreated';
export const ChannelDepositedEvent = 'ChannelDeposited';
export const ChannelWithdrawnEvent = 'ChannelWithdrawn';
export const ChannelChallengedEvent = 'ChannelChallenged';
export const ChannelCheckpointedEvent = 'ChannelCheckpointed';
export const ChannelClosedEvent = 'ChannelClosed';

// Escrow deposit events
export const EscrowDepositInitiatedEvent = 'EscrowDepositInitiated';
export const EscrowDepositInitiatedOnHomeEvent = 'EscrowDepositInitiatedOnHome';
export const EscrowDepositChallengedEvent = 'EscrowDepositChallenged';
export const EscrowDepositFinalizedEvent = 'EscrowDepositFinalized';
export const EscrowDepositFinalizedOnHomeEvent = 'EscrowDepositFinalizedOnHome';
export const EscrowDepositsPurgedEvent = 'EscrowDepositsPurged';

// Escrow withdrawal events
export const EscrowWithdrawalInitiatedEvent = 'EscrowWithdrawalInitiated';
export const EscrowWithdrawalInitiatedOnHomeEvent = 'EscrowWithdrawalInitiatedOnHome';
export const EscrowWithdrawalChallengedEvent = 'EscrowWithdrawalChallenged';
export const EscrowWithdrawalFinalizedEvent = 'EscrowWithdrawalFinalized';
export const EscrowWithdrawalFinalizedOnHomeEvent = 'EscrowWithdrawalFinalizedOnHome';

// Migration events
export const MigrationInInitiatedEvent = 'MigrationInInitiated';
export const MigrationInFinalizedEvent = 'MigrationInFinalized';
export const MigrationOutInitiatedEvent = 'MigrationOutInitiated';
export const MigrationOutFinalizedEvent = 'MigrationOutFinalized';

// Vault events
export const DepositedEvent = 'Deposited';
export const WithdrawnEvent = 'Withdrawn';
