package event_handlers

import (
	"time"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/log"
)

// EventHandlerService handles on-chain events
type EventHandlerService struct {
	useStoreInTx StoreTxProvider
	logger       log.Logger
}

// NewEventHandlerService creates a new event handler service
func NewEventHandlerService(useStoreInTx StoreTxProvider, logger log.Logger) *EventHandlerService {
	return &EventHandlerService{
		useStoreInTx: useStoreInTx,
		logger:       logger,
	}
}

// HandleHomeChannelCreated handles the HomeChannelCreated event
func (s *EventHandlerService) HandleHomeChannelCreated(event *core.HomeChannelCreatedEvent) error {
	return s.useStoreInTx(func(tx Store) error {
		chanID := event.ChannelID
		channel, err := tx.GetChannelByID(chanID)
		if err != nil {
			return err
		}
		if channel == nil {
			s.logger.Warn("channel not found in DB during HomeChannelCreated event", "channelId", chanID)
			return nil
		}
		if channel.Type != core.ChannelTypeHome {
			s.logger.Warn("channel type mismatch during HomeChannelCreated event", "channelId", chanID, "expectedType", core.ChannelTypeHome, "actualType", channel.Type)
			return nil
		}
		channel.StateVersion = event.StateVersion
		channel.Status = core.ChannelStatusOpen

		err = tx.UpdateChannel(*channel)
		if err != nil {
			return err
		}
		s.logger.Info("handled HomeChannelCreated event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)

		return nil
	})
}

// HandleHomeChannelMigrated handles the HomeChannelMigrated event
func (s *EventHandlerService) HandleHomeChannelMigrated(event *core.HomeChannelMigratedEvent) error {
	// TODO: Implement HomeChannelMigrated handler logic
	s.logger.Info("Unexpected HomeChannelMigrated event", "channelId", event.ChannelID, "stateVersion", event.StateVersion)
	return nil
}

// HandleHomeChannelCheckpointed handles the HomeChannelCheckpointed event
func (s *EventHandlerService) HandleHomeChannelCheckpointed(event *core.HomeChannelCheckpointedEvent) error {
	return s.useStoreInTx(func(tx Store) error {
		chanID := event.ChannelID
		channel, err := tx.GetChannelByID(chanID)
		if err != nil {
			return err
		}
		if channel == nil {
			s.logger.Warn("channel not found in DB during HomeChannelCheckpointed event", "channelId", chanID)
			return nil
		}
		if channel.Type != core.ChannelTypeHome {
			s.logger.Warn("channel type mismatch during HomeChannelCheckpointed event", "channelId", chanID, "expectedType", core.ChannelTypeHome, "actualType", channel.Type)
			return nil
		}
		channel.StateVersion = event.StateVersion

		if channel.Status == core.ChannelStatusChallenged {
			channel.Status = core.ChannelStatusOpen
		}

		err = tx.UpdateChannel(*channel)
		if err != nil {
			return err
		}
		s.logger.Info("handled HomeChannelCheckpointed event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)

		return nil
	})
}

// HandleHomeChannelChallenged handles the HomeChannelChallenged event
func (s *EventHandlerService) HandleHomeChannelChallenged(event *core.HomeChannelChallengedEvent) error {
	return s.useStoreInTx(func(tx Store) error {
		chanID := event.ChannelID
		channel, err := tx.GetChannelByID(chanID)
		if err != nil {
			return err
		}
		if channel == nil {
			s.logger.Warn("channel not found in DB during HomeChannelChallenged event", "channelId", chanID)
			return nil
		}
		if channel.Type != core.ChannelTypeHome {
			s.logger.Warn("channel type mismatch during HomeChannelChallenged event", "channelId", chanID, "expectedType", core.ChannelTypeHome, "actualType", channel.Type)
			return nil
		}

		if event.StateVersion < channel.StateVersion {
			// TODO: SCHEDULE CHECKPOINT ACTION
		}
		if event.StateVersion > channel.StateVersion {
			channel.StateVersion = event.StateVersion
		}

		unixExpiry := event.ChallengeExpiry
		expirationTime := time.Unix(int64(unixExpiry), 0) // TODO: recheck format

		channel.ChallengeExpiresAt = &expirationTime

		channel.Status = core.ChannelStatusChallenged

		err = tx.UpdateChannel(*channel)
		if err != nil {
			return err
		}
		s.logger.Info("handled HomeChannelChallenged event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)

		return nil
	})
}

// HandleHomeChannelClosed handles the HomeChannelClosed event
func (s *EventHandlerService) HandleHomeChannelClosed(event *core.HomeChannelClosedEvent) error {
	return s.useStoreInTx(func(tx Store) error {
		chanID := event.ChannelID
		channel, err := tx.GetChannelByID(chanID)
		if err != nil {
			return err
		}
		if channel == nil {
			s.logger.Warn("channel not found in DB during HomeChannelClosed event", "channelId", chanID)
			return nil
		}
		if channel.Type != core.ChannelTypeHome {
			s.logger.Warn("channel type mismatch during HomeChannelClosed event", "channelId", chanID, "expectedType", core.ChannelTypeHome, "actualType", channel.Type)
			return nil
		}

		channel.StateVersion = event.StateVersion
		channel.Status = core.ChannelStatusClosed

		err = tx.UpdateChannel(*channel)
		if err != nil {
			return err
		}
		s.logger.Info("handled HomeChannelClosed event", "channelId", event.ChannelID, "stateVersion", event.StateVersion, "userWallet", channel.UserWallet)

		return nil
	})
}

// HandleEscrowDepositInitiated handles the EscrowDepositInitiated event
func (s *EventHandlerService) HandleEscrowDepositInitiated(event *core.EscrowDepositInitiatedEvent) error {
	// TODO: Implement EscrowDepositInitiated handler logic
	return nil
}

// HandleEscrowDepositChallenged handles the EscrowDepositChallenged event
func (s *EventHandlerService) HandleEscrowDepositChallenged(event *core.EscrowDepositChallengedEvent) error {
	// TODO: Implement EscrowDepositChallenged handler logic
	return nil
}

// HandleEscrowDepositFinalized handles the EscrowDepositFinalized event
func (s *EventHandlerService) HandleEscrowDepositFinalized(event *core.EscrowDepositFinalizedEvent) error {
	// TODO: Implement EscrowDepositFinalized handler logic
	return nil
}

// HandleEscrowWithdrawalInitiated handles the EscrowWithdrawalInitiated event
func (s *EventHandlerService) HandleEscrowWithdrawalInitiated(event *core.EscrowWithdrawalInitiatedEvent) error {
	// TODO: Implement EscrowWithdrawalInitiated handler logic
	return nil
}

// HandleEscrowWithdrawalChallenged handles the EscrowWithdrawalChallenged event
func (s *EventHandlerService) HandleEscrowWithdrawalChallenged(event *core.EscrowWithdrawalChallengedEvent) error {
	// TODO: Implement EscrowWithdrawalChallenged handler logic
	return nil
}

// HandleEscrowWithdrawalFinalized handles the EscrowWithdrawalFinalized event
func (s *EventHandlerService) HandleEscrowWithdrawalFinalized(event *core.EscrowWithdrawalFinalizedEvent) error {
	// TODO: Implement EscrowWithdrawalFinalized handler logic
	return nil
}
