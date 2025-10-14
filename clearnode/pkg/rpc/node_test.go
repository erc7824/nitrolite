package rpc_test

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erc7824/nitrolite/clearnode/pkg/log"
	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
)

func TestNewWebsocketNode(t *testing.T) {
	t.Parallel()

	cfg := rpc.WebsocketNodeConfig{}
	_, err := rpc.NewWebsocketNode(cfg)
	require.Equal(t, "signer cannot be nil", err.Error())

	cfg.Signer = sign.NewMockSigner("signer1")
	_, err = rpc.NewWebsocketNode(cfg)
	require.Equal(t, "logger cannot be nil", err.Error())
	cfg.Logger = log.NewNoopLogger()

	node, err := rpc.NewWebsocketNode(cfg)
	require.NoError(t, err)
	require.NotNil(t, node)
}

func TestWebsocketNode_FullFlow(t *testing.T) {
	t.Parallel()

	signer := sign.NewMockSigner("signer1")
	logger := log.NewNoopLogger()

	const (
		methodWelcome        = "welcome"
		methodRegister       = "register"
		methodReauthenticate = "reauthenticate"
		methodUserID         = "user_id"

		userAlice = "alice"
		userBob   = "bob"
	)

	type paramsUserID struct {
		UserID string `json:"user_id"`
	}

	eventMu := &sync.RWMutex{}
	onConnectCount := 0
	onDisconnectCount := 0
	onMessageSentCount := 0
	onAuthenticatedCount := 0
	currentAuthUserID := ""
	disconnectedUserID := ""
	nodeConnCallCount := 0
	var dialErr error

	noceCfg := rpc.WebsocketNodeConfig{
		Signer: signer,
		Logger: logger,
		OnConnectHandler: func(sendRes rpc.SendResponseFunc) {
			eventMu.Lock()
			defer eventMu.Unlock()

			onConnectCount++
			sendRes(methodWelcome, nil)
		},
		OnDisconnectHandler: func(userID string) {
			eventMu.Lock()
			defer eventMu.Unlock()

			disconnectedUserID = userID
			currentAuthUserID = ""
			onDisconnectCount++
		},
		OnMessageSentHandler: func(_ []byte) {
			eventMu.Lock()
			defer eventMu.Unlock()

			onMessageSentCount++
		},
		OnAuthenticatedHandler: func(userID string, sendRes rpc.SendResponseFunc) {
			eventMu.Lock()
			defer eventMu.Unlock()

			currentAuthUserID = userID
			onAuthenticatedCount++

			params, err := rpc.NewParams(paramsUserID{UserID: userID})
			if err != nil {
				sendRes(rpc.ErrorMethod.String(), rpc.NewErrorParams("internal error"))
				return
			}
			sendRes(methodWelcome, params)
		},
	}

	node, err := rpc.NewWebsocketNode(noceCfg)
	require.NoError(t, err)
	require.NotNil(t, node)

	userStore := rpc.NewSafeStorage()
	node.Use(func(ctx *rpc.Context) {
		eventMu.Lock()
		nodeConnCallCount++
		eventMu.Unlock()

		ctx.Next()
	})
	node.Handle(methodRegister, func(ctx *rpc.Context) {
		var params paramsUserID
		if err := ctx.Request.Req.Params.Translate(&params); err != nil {
			ctx.Fail(err, "invalid params")
			return
		}

		userStore.Set(params.UserID, struct{}{})
		ctx.UserID = params.UserID

		ctx.Succeed(methodRegister, nil)
	})
	node.Handle(methodReauthenticate, func(ctx *rpc.Context) {
		var params paramsUserID
		if err := ctx.Request.Req.Params.Translate(&params); err != nil {
			ctx.Fail(err, "invalid params")
			return
		}

		_, ok := userStore.Get(params.UserID)
		if !ok {
			ctx.Fail(nil, "user not found")
			return
		}
		ctx.UserID = params.UserID

		ctx.Succeed(methodReauthenticate, nil)
	})

	privGroup := node.NewGroup("private")
	privGroup.Use(func(ctx *rpc.Context) {
		if ctx.UserID == "" {
			ctx.Fail(nil, "not authenticated")
			return
		}
		ctx.Next()
	})
	privGroup.Handle(methodUserID, func(ctx *rpc.Context) {
		params, err := rpc.NewParams(paramsUserID{UserID: ctx.UserID})
		if err != nil {
			ctx.Fail(err, "internal error")
			return
		}
		ctx.Succeed(methodUserID, params)
	})

	server := httptest.NewServer(node)
	defer server.Close()

	dialerCfg := rpc.DefaultWebsocketDialerConfig
	dialer := rpc.NewWebsocketDialer(dialerCfg)

	ctx, cancel := context.WithCancel(context.Background())
	dialer.Dial(ctx, "ws://"+server.Listener.Addr().String(), func(err error) {
		eventMu.Lock()
		defer eventMu.Unlock()

		dialErr = err
	})
	require.True(t, dialer.IsConnected())

	// Wait for welcome event (OnConnect)
	select {
	case event := <-dialer.EventCh():
		require.Equal(t, methodWelcome, event.Res.Method)
		assert.Equal(t, uint64(0), event.Res.RequestID)

		time.Sleep(50 * time.Millisecond)
		eventMu.RLock()
		assert.Equal(t, 1, onConnectCount)
		assert.Equal(t, 1, onMessageSentCount)
		eventMu.RUnlock()
	case <-time.After(100 * time.Millisecond):
		t.Fatal("OnConnect event timeout")
	}

	// CASE 1: Call private method before authentication
	req := rpc.NewRequest(rpc.NewPayload(1, methodUserID, nil))
	res, err := dialer.Call(ctx, &req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(1), res.Res.RequestID)
	assert.Equal(t, "not authenticated", res.Error().Error())

	time.Sleep(50 * time.Millisecond)
	eventMu.RLock()
	assert.Equal(t, 2, onMessageSentCount)
	assert.Equal(t, 1, nodeConnCallCount)
	eventMu.RUnlock()

	// CASE 2: Register new user
	registerParams, err := rpc.NewParams(paramsUserID{UserID: userAlice})
	require.NoError(t, err)

	req = rpc.NewRequest(rpc.NewPayload(2, methodRegister, registerParams))
	res, err = dialer.Call(ctx, &req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(2), res.Res.RequestID)
	assert.Equal(t, methodRegister, res.Res.Method)
	assert.Nil(t, res.Error())

	eventMu.RLock()
	assert.Equal(t, 2, nodeConnCallCount)
	eventMu.RUnlock()

	// Wait for welcome event (OnAuthenticated)
	select {
	case event := <-dialer.EventCh():
		require.Equal(t, methodWelcome, event.Res.Method)
		assert.Equal(t, uint64(0), event.Res.RequestID)

		time.Sleep(50 * time.Millisecond)
		eventMu.RLock()
		assert.Equal(t, 1, onAuthenticatedCount)
		assert.Equal(t, 1, onConnectCount)
		assert.Equal(t, 4, onMessageSentCount)

		var resParams paramsUserID
		err = event.Res.Params.Translate(&resParams)
		require.NoError(t, err)
		assert.Equal(t, userAlice, resParams.UserID)
		assert.Equal(t, userAlice, currentAuthUserID)
		eventMu.RUnlock()
	case <-time.After(100 * time.Millisecond):
		t.Fatal("OnConnect event timeout")
	}

	// CASE 3: Call private method after authentication
	req = rpc.NewRequest(rpc.NewPayload(3, methodUserID, nil))
	res, err = dialer.Call(ctx, &req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(3), res.Res.RequestID)
	assert.Equal(t, methodUserID, res.Res.Method)
	assert.Nil(t, res.Error())

	time.Sleep(50 * time.Millisecond)
	eventMu.RLock()
	assert.Equal(t, 5, onMessageSentCount)
	assert.Equal(t, 3, nodeConnCallCount)
	eventMu.RUnlock()

	{
		var resParams paramsUserID
		err = res.Res.Params.Translate(&resParams)
		require.NoError(t, err)
		assert.Equal(t, userAlice, resParams.UserID)
	}

	// CASE 4: Reauthenticate as different user (not registered)
	reauthParams, err := rpc.NewParams(paramsUserID{UserID: userBob})
	require.NoError(t, err)

	req = rpc.NewRequest(rpc.NewPayload(4, methodReauthenticate, reauthParams))
	res, err = dialer.Call(ctx, &req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(4), res.Res.RequestID)
	assert.Equal(t, "user not found", res.Error().Error())

	time.Sleep(50 * time.Millisecond)
	eventMu.RLock()
	assert.Equal(t, 6, onMessageSentCount)
	assert.Equal(t, 4, nodeConnCallCount)
	assert.Equal(t, userAlice, currentAuthUserID) // Still Alice
	eventMu.RUnlock()

	// CASE 5: Register second user
	registerParams, err = rpc.NewParams(paramsUserID{UserID: userBob})
	require.NoError(t, err)

	req = rpc.NewRequest(rpc.NewPayload(5, methodRegister, registerParams))
	res, err = dialer.Call(ctx, &req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(5), res.Res.RequestID)
	assert.Equal(t, methodRegister, res.Res.Method)
	assert.Nil(t, res.Error())

	eventMu.RLock()
	assert.Equal(t, 5, nodeConnCallCount)
	eventMu.RUnlock()

	// Wait for welcome event (OnAuthenticated)
	select {
	case event := <-dialer.EventCh():
		require.Equal(t, methodWelcome, event.Res.Method)
		assert.Equal(t, uint64(0), event.Res.RequestID)

		time.Sleep(50 * time.Millisecond)
		eventMu.RLock()
		assert.Equal(t, 2, onAuthenticatedCount)
		assert.Equal(t, 1, onConnectCount)
		assert.Equal(t, 8, onMessageSentCount)

		var resParams paramsUserID
		err = event.Res.Params.Translate(&resParams)
		require.NoError(t, err)
		assert.Equal(t, userBob, resParams.UserID)
		assert.Equal(t, userBob, currentAuthUserID)
		eventMu.RUnlock()
	case <-time.After(100 * time.Millisecond):
		t.Fatal("OnConnect event timeout")
	}

	// CASE 6: Call private method after registering second user
	req = rpc.NewRequest(rpc.NewPayload(6, methodUserID, nil))
	res, err = dialer.Call(ctx, &req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(6), res.Res.RequestID)
	assert.Equal(t, methodUserID, res.Res.Method)
	assert.Nil(t, res.Error())

	time.Sleep(50 * time.Millisecond)
	eventMu.RLock()
	assert.Equal(t, 9, onMessageSentCount)
	assert.Equal(t, 6, nodeConnCallCount)
	eventMu.RUnlock()

	{
		var resParams paramsUserID
		err = res.Res.Params.Translate(&resParams)
		require.NoError(t, err)
		assert.Equal(t, userBob, resParams.UserID)
	}

	// CASE 7: Reauthenticate as first user
	reauthParams, err = rpc.NewParams(paramsUserID{UserID: userAlice})
	require.NoError(t, err)

	req = rpc.NewRequest(rpc.NewPayload(7, methodReauthenticate, reauthParams))
	res, err = dialer.Call(ctx, &req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(7), res.Res.RequestID)
	assert.Equal(t, methodReauthenticate, res.Res.Method)
	assert.Nil(t, res.Error())

	eventMu.RLock()
	assert.Equal(t, 7, nodeConnCallCount)
	eventMu.RUnlock()

	// Wait for welcome event (OnAuthenticated)
	select {
	case event := <-dialer.EventCh():
		require.Equal(t, methodWelcome, event.Res.Method)
		assert.Equal(t, uint64(0), event.Res.RequestID)

		time.Sleep(50 * time.Millisecond)
		eventMu.RLock()
		assert.Equal(t, 3, onAuthenticatedCount)
		assert.Equal(t, 1, onConnectCount)
		assert.Equal(t, 11, onMessageSentCount)

		var resParams paramsUserID
		err = event.Res.Params.Translate(&resParams)
		require.NoError(t, err)
		assert.Equal(t, userAlice, resParams.UserID)
		assert.Equal(t, userAlice, currentAuthUserID)
		eventMu.RUnlock()
	case <-time.After(100 * time.Millisecond):
		t.Fatal("OnConnect event timeout")
	}

	// CASE 8: Send manual welcome event
	welcomeParams, err := rpc.NewParams(paramsUserID{UserID: userAlice})
	require.NoError(t, err)
	node.Notify(userAlice, methodWelcome, welcomeParams)

	// Wait for welcome event (manual)
	select {
	case event := <-dialer.EventCh():
		require.Equal(t, methodWelcome, event.Res.Method)
		assert.Equal(t, uint64(0), event.Res.RequestID)

		var resParams paramsUserID
		err = event.Res.Params.Translate(&resParams)
		require.NoError(t, err)
		assert.Equal(t, userAlice, resParams.UserID)

	case <-time.After(100 * time.Millisecond):
		t.Fatal("manual welcome event timeout")
	}

	// Disconnect
	cancel()
	time.Sleep(100 * time.Millisecond) // Allow some time for the disconnect to be processed

	assert.False(t, dialer.IsConnected())

	eventMu.RLock()
	assert.NoError(t, dialErr)
	assert.Equal(t, 3, onAuthenticatedCount)
	assert.Equal(t, 1, onConnectCount)
	assert.Equal(t, 12, onMessageSentCount)
	assert.Equal(t, 7, nodeConnCallCount)
	assert.Equal(t, 1, onDisconnectCount)
	assert.Equal(t, userAlice, disconnectedUserID)
	assert.Equal(t, "", currentAuthUserID)
	eventMu.RUnlock()
}
