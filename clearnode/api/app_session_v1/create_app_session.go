package app_session_v1

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/log"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

// CreateAppSession creates a new application session between participants.
// App sessions are created with 0 allocations by default, as per V1 API specification.
// Deposits must be done through the submit_deposit_state endpoint.
func (h *Handler) CreateAppSession(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var reqPayload rpc.AppSessionsV1CreateAppSessionRequest
	if err := c.Request.Payload.Translate(&reqPayload); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	appDef := unmapAppDefinitionV1(reqPayload.Definition)

	logger.Debug("processing app session creation request",
		"application", reqPayload.Definition.Application,
		"participantsNum", len(reqPayload.Definition.Participants),
		"quorum", reqPayload.Definition.Quorum,
		"nonce", reqPayload.Definition.Nonce)

	// Validate nonce
	if reqPayload.Definition.Nonce == 0 {
		c.Fail(nil, "nonce is zero or not provided")
		return
	}

	// Validate quorum against total weights
	var totalWeights uint8
	participantWeights := make(map[string]uint8)
	for _, participant := range reqPayload.Definition.Participants {
		totalWeights += participant.SignatureWeight
		participantWeights[participant.WalletAddress] = participant.SignatureWeight
	}

	if reqPayload.Definition.Quorum > totalWeights {
		c.Fail(rpc.Errorf("target quorum (%d) cannot be greater than total sum of weights (%d)",
			reqPayload.Definition.Quorum, totalWeights), "")
		return
	}

	// Validate signatures and quorum
	if len(reqPayload.Signatures) == 0 {
		c.Fail(nil, "no signatures provided")
		return
	}

	// Pack the request for signature verification
	packedRequest, err := app.PackCreateAppSessionRequest(appDef, reqPayload.SessionData)
	if err != nil {
		c.Fail(rpc.Errorf("failed to pack request: %v", err), "")
		return
	}

	// Verify signatures and calculate quorum
	sigRecoverer := h.sigValidator[EcdsaSigType]
	signedWeights := make(map[string]bool)
	var achievedQuorum uint8

	for _, sigHex := range reqPayload.Signatures {
		sigBytes, err := hexutil.Decode(sigHex)
		if err != nil {
			c.Fail(rpc.Errorf("failed to decode signature: %v", err), "")
			return
		}

		// Recover the signer address from the signature (this also validates the signature)
		signerAddress, err := sigRecoverer.Recover(packedRequest, sigBytes)
		if err != nil {
			c.Fail(rpc.Errorf("failed to recover signer address: %v", err), "")
			return
		}

		// Check if signer is a participant
		weight, isParticipant := participantWeights[signerAddress]
		if !isParticipant {
			c.Fail(rpc.Errorf("signature from non-participant: %s", signerAddress), "")
			return
		}

		// Add weight if not already counted
		if !signedWeights[signerAddress] {
			signedWeights[signerAddress] = true
			achievedQuorum += weight
		}
	}

	// Check if quorum is met
	if achievedQuorum < reqPayload.Definition.Quorum {
		c.Fail(rpc.Errorf("quorum not met: achieved %d, required %d", achievedQuorum, reqPayload.Definition.Quorum), "")
		return
	}

	// Generate app session ID (deterministic)
	appSessionID, err := app.GenerateAppSessionIDV1(appDef)
	if err != nil {
		c.Fail(rpc.Errorf("failed to generate app session ID: %v", err), "")
		return
	}

	// Create app session with 0 allocations
	appSession := app.AppSessionV1{
		SessionID:    appSessionID,
		Application:  appDef.Application,
		Participants: appDef.Participants,
		Quorum:       appDef.Quorum,
		Nonce:        appDef.Nonce,
		IsClosed:     false,
		Version:      1,
		SessionData:  reqPayload.SessionData,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = h.useStoreInTx(func(store AppStoreV1) error {
		if err := store.CreateAppSession(appSession); err != nil {
			return rpc.Errorf("failed to create app session: %v", err)
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to create app session", "error", err)
		c.Fail(err, "failed to create app session")
		return
	}

	resp := rpc.AppSessionsV1CreateAppSessionResponse{
		AppSessionID: appSessionID,
		Version:      fmt.Sprintf("%d", appSession.Version),
		IsClosed:     false,
	}

	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
	logger.Info("successfully created app session", "appSessionID", appSessionID)
}
