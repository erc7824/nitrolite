package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test internal authRequest method
func TestClient_authRequest(t *testing.T) {
	mockDialer := newMockInternalDialer()
	client := NewClient(mockDialer)

	challengeUUID := uuid.New()

	mockDialer.handlers[AuthRequestMethod] = func(req *Request) (*Response, error) {
		// Verify request structure
		assert.Equal(t, string(AuthRequestMethod), req.Req.Method)

		var authReq AuthRequestRequest
		err := req.Req.Params.Translate(&authReq)
		assert.NoError(t, err)
		assert.Equal(t, "0x1234", authReq.Address)
		assert.Equal(t, "0x5678", authReq.SessionKey)
		assert.Equal(t, "TestApp", authReq.AppName)

		// Return auth_challenge response
		params, _ := NewParams(AuthRequestResponse{
			ChallengeMessage: challengeUUID,
		})
		payload := NewPayload(0, string(AuthChallengeMethod), params)
		res := NewResponse(payload)
		return &res, nil
	}

	authReq := AuthRequestRequest{
		Address:    "0x1234",
		SessionKey: "0x5678",
		AppName:    "TestApp",
	}

	resp, sigs, err := client.authRequest(context.Background(), authReq)
	require.NoError(t, err)
	assert.Equal(t, challengeUUID, resp.ChallengeMessage)
	assert.Empty(t, sigs)
}

// Test authRequest with unexpected response method
func TestClient_authRequest_WrongResponseMethod(t *testing.T) {
	mockDialer := newMockInternalDialer()
	client := NewClient(mockDialer)

	mockDialer.handlers[AuthRequestMethod] = func(req *Request) (*Response, error) {
		// Return wrong method in response
		params, _ := NewParams(AuthRequestResponse{})
		payload := NewPayload(0, string(PongMethod), params) // Wrong method
		res := NewResponse(payload)
		return &res, nil
	}

	authReq := AuthRequestRequest{
		Address: "0x1234",
	}

	_, _, err := client.authRequest(context.Background(), authReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected response method")
}

// Test internal authSigVerify method
func TestClient_authSigVerify(t *testing.T) {
	mockDialer := newMockInternalDialer()
	client := NewClient(mockDialer)

	challengeUUID := uuid.New()
	jwtToken := "test.jwt.token"

	mockDialer.handlers[AuthVerifyMethod] = func(req *Request) (*Response, error) {
		// Verify request structure
		assert.Equal(t, string(AuthVerifyMethod), req.Req.Method)
		assert.Len(t, req.Sig, 1) // Should have one signature

		var verifyReq AuthSigVerifyRequest
		err := req.Req.Params.Translate(&verifyReq)
		assert.NoError(t, err)
		assert.Equal(t, challengeUUID, verifyReq.Challenge)

		// Return successful verification
		params, _ := NewParams(AuthSigVerifyResponse{
			Address:    "0x1234",
			SessionKey: "0x5678",
			JwtToken:   jwtToken,
			Success:    true,
		})
		payload := NewPayload(0, string(AuthVerifyMethod), params)
		res := NewResponse(payload)
		return &res, nil
	}

	verifyReq := AuthSigVerifyRequest{
		Challenge: challengeUUID,
	}

	testSig := sign.Signature{}
	resp, sigs, err := client.authSigVerify(context.Background(), verifyReq, testSig)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, jwtToken, resp.JwtToken)
	assert.Equal(t, "0x1234", resp.Address)
	assert.Equal(t, "0x5678", resp.SessionKey)
	assert.Empty(t, sigs)
}

// Test signChallenge helper function
func TestSignChallenge(t *testing.T) {
	mockSigner := &mockSigner{}

	authReq := AuthRequestRequest{
		Address:            "0x1234567890123456789012345678901234567890",
		SessionKey:         "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		AppName:            "TestApp",
		ApplicationAddress: "0x1111111111111111111111111111111111111111",
		Allowances:         []Allowance{},
		Expire:             "3600",
		Scope:              "trade",
	}

	challengeToken := "test-challenge-token"

	sig, err := signChallenge(mockSigner, authReq, challengeToken)
	require.NoError(t, err)

	// The mock signer should have created a signature
	assert.NotNil(t, sig)
	// Verify the signature ends with our test suffix
	assert.Equal(t, sign.Signature("-signed"), sig[len(sig)-7:])
}

// var (
//
//	testCtx     = context.Background()
//	fixedTime   = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
//	testWallet  = "0x1234"
//	testWallet2 = "0x5678"
//	testChainID = uint32(1)
//	testToken   = "0xUSDC"
//	testSymbol  = "USDC"
//
// )
func TestClient_EventHandling(t *testing.T) {
	mockDialer := newMockInternalDialer()
	client := NewClient(mockDialer)

	// Event channels
	balanceReceived := make(chan BalanceUpdateNotification, 1)
	channelReceived := make(chan ChannelUpdateNotification, 1)
	transferReceived := make(chan TransferNotification, 1)

	// Register handlers
	client.HandleBalanceUpdateEvent(func(ctx context.Context, n BalanceUpdateNotification, _ []sign.Signature) {
		balanceReceived <- n
	})
	client.HandleChannelUpdateEvent(func(ctx context.Context, n ChannelUpdateNotification, _ []sign.Signature) {
		channelReceived <- n
	})
	client.HandleTransferEvent(func(ctx context.Context, n TransferNotification, _ []sign.Signature) {
		transferReceived <- n
	})

	// Start listener
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go client.listenEvents(ctx)

	// Publish events
	balanceUpdate := BalanceUpdateNotification{
		BalanceUpdates: []LedgerBalance{{Asset: "usdc", Amount: decimal.NewFromInt(500)}},
	}
	params, _ := NewParams(balanceUpdate)
	mockDialer.publishNotification(BalanceUpdateEvent, params)

	channelUpdate := ChannelUpdateNotification{ChannelID: "ch123"}
	params, _ = NewParams(channelUpdate)
	mockDialer.publishNotification(ChannelUpdateEvent, params)

	transferNotif := TransferNotification{
		Transactions: []LedgerTransaction{{Id: 1, TxType: "incoming"}},
	}
	params, _ = NewParams(transferNotif)
	mockDialer.publishNotification(TransferEvent, params)

	// Verify events
	select {
	case <-balanceReceived:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("balance update timeout")
	}

	select {
	case <-channelReceived:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("channel update timeout")
	}

	select {
	case <-transferReceived:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("transfer timeout")
	}

	// Cleanup
	cancel()
}

// mockInternalDialer is a simple mock dialer for internal tests
type mockInternalDialer struct {
	handlers map[Method]func(*Request) (*Response, error)
	eventCh  chan *Response
}

func newMockInternalDialer() *mockInternalDialer {
	return &mockInternalDialer{
		handlers: make(map[Method]func(*Request) (*Response, error)),
		eventCh:  make(chan *Response, 10),
	}
}

func (m *mockInternalDialer) Dial(ctx context.Context, url string, handleClosure func(err error)) error {
	return nil
}

func (m *mockInternalDialer) IsConnected() bool { return true }
func (m *mockInternalDialer) EventCh() <-chan *Response {
	return m.eventCh
}

func (m *mockInternalDialer) Call(ctx context.Context, req *Request) (*Response, error) {
	handler, ok := m.handlers[Method(req.Req.Method)]
	if !ok {
		res := NewErrorResponse(req.Req.RequestID, "method not found")
		return &res, nil
	}
	return handler(req)
}

func (m *mockInternalDialer) publishNotification(event Event, params Params) {
	payload := NewPayload(0, string(event), params)
	res := NewResponse(payload)
	m.eventCh <- &res
}

// mockSigner implements sign.Signer for testing
// It simply appends "-signed" to the input data as a signature
type mockSigner struct {
	publicKey sign.PublicKey
}

func (m *mockSigner) Sign(data []byte) (sign.Signature, error) {
	// Create a simple signature by appending "-signed" to the data
	sig := sign.Signature(append(data, []byte("-signed")...))
	return sig, nil
}

func (m *mockSigner) PublicKey() sign.PublicKey {
	return m.publicKey
}
