package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/erc7824/nitrolite/pkg/sign"
)

func TestClient_EventHandling(t *testing.T) {
	t.Parallel()

	mockDialer := newMockInternalDialer()
	client := NewClient(mockDialer)

	// Event channels
	balanceReceived := make(chan BalanceUpdateNotification, 1)
	channelReceived := make(chan ChannelUpdateNotification, 1)
	appSessionReceived := make(chan AppSessionUpdateNotification, 1)
	transferReceived := make(chan TransferNotification, 1)

	// Register handlers
	client.HandleBalanceUpdateEvent(func(ctx context.Context, n BalanceUpdateNotification, _ []sign.Signature) {
		balanceReceived <- n
	})
	client.HandleChannelUpdateEvent(func(ctx context.Context, n ChannelUpdateNotification, _ []sign.Signature) {
		channelReceived <- n
	})
	client.HandleAppSessionUpdateEvent(func(ctx context.Context, n AppSessionUpdateNotification, _ []sign.Signature) {
		appSessionReceived <- n
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
	params, _ := NewPayload(balanceUpdate)
	mockDialer.publishNotification(BalanceUpdateEvent, params)

	channelUpdate := ChannelUpdateNotification{
		ChannelID: "ch123",
		Status:    "open",
	}
	params, _ = NewPayload(channelUpdate)
	mockDialer.publishNotification(ChannelUpdateEvent, params)

	appSessionUpdate := AppSessionUpdateNotification{
		AppSession: AppSession{
			AppSessionID: "as123",
			Status:       "active",
		},
	}
	params, _ = NewPayload(appSessionUpdate)
	mockDialer.publishNotification(AppSessionUpdateEvent, params)

	transferNotif := TransferNotification{
		Transactions: []LedgerTransaction{{Id: 1, TxType: "incoming"}},
	}
	params, _ = NewPayload(transferNotif)
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
	case <-appSessionReceived:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("app session update timeout")
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
	handlers map[Method]func(*Message) (*Message, error)
	eventCh  chan *Message
}

func newMockInternalDialer() *mockInternalDialer {
	return &mockInternalDialer{
		handlers: make(map[Method]func(*Message) (*Message, error)),
		eventCh:  make(chan *Message, 10),
	}
}

func (m *mockInternalDialer) Dial(ctx context.Context, url string, handleClosure func(err error)) error {
	return nil
}

func (m *mockInternalDialer) IsConnected() bool { return true }
func (m *mockInternalDialer) EventCh() <-chan *Message {
	return m.eventCh
}

func (m *mockInternalDialer) Call(ctx context.Context, req *Message) (*Message, error) {
	handler, ok := m.handlers[Method(req.Method)]
	if !ok {
		res := NewErrorResponse(req.RequestID, "method not found")
		return &res, nil
	}
	return handler(req)
}

func (m *mockInternalDialer) publishNotification(event Event, params Payload) {
	res := NewResponse(0, string(event), params)
	m.eventCh <- &res
}
