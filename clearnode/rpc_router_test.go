package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	container "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestSqlite creates an in-memory SQLite DB for testing
func setupTestSqlite(t testing.TB) *gorm.DB {
	t.Helper()

	uniqueDSN := fmt.Sprintf("file::memory:test%s?mode=memory&cache=shared", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(uniqueDSN), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&Entry{}, &Channel{}, &AppSession{}, &RPCRecord{}, &Asset{}, &SignerWallet{})
	require.NoError(t, err)

	return db
}

// setupTestPostgres creates a PostgreSQL database using testcontainers
func setupTestPostgres(ctx context.Context, t testing.TB) (*gorm.DB, testcontainers.Container) {
	t.Helper()

	const dbName = "postgres"
	const dbUser = "postgres"
	const dbPassword = "postgres"

	postgresContainer, err := container.Run(ctx,
		"postgres:16-alpine",
		container.WithDatabase(dbName),
		container.WithUsername(dbUser),
		container.WithPassword(dbPassword),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "trust",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForLog("database system is ready to accept connections"),
				wait.ForListeningPort("5432/tcp"),
			)))
	require.NoError(t, err)
	log.Println("Started container:", postgresContainer.GetContainerID())

	url, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	log.Println("PostgreSQL URL:", url)

	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&Entry{}, &Channel{}, &AppSession{}, &RPCRecord{}, &Asset{})
	require.NoError(t, err)

	return db, postgresContainer
}

// setupTestDB chooses SQLite or Postgres based on TEST_DB_DRIVER
func setupTestDB(t testing.TB) (*gorm.DB, func()) {
	t.Helper()

	ctx := context.Background()
	var db *gorm.DB
	var cleanup func()

	switch os.Getenv("TEST_DB_DRIVER") {
	case "postgres":
		log.Println("Using PostgreSQL for testing")
		var container testcontainers.Container
		db, container = setupTestPostgres(ctx, t)
		cleanup = func() {
			if container != nil {
				if err := container.Terminate(ctx); err != nil {
					log.Printf("Failed to terminate PostgreSQL container: %v", err)
				}
			}
		}
	default:
		log.Println("Using SQLite for testing (default)")
		db = setupTestSqlite(t)
		cleanup = func() {}
	}

	return db, cleanup
}

func setupTestRPCRouter(t *testing.T) (*RPCRouter, func()) {
	db, dbCleanup := setupTestDB(t)

	// Use a test private key
	privateKeyHex := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	signer, err := NewSigner(privateKeyHex)
	require.NoError(t, err)

	logger := NewLoggerIPFS("root.test")

	// Create an instance of RPCRouter
	router := &RPCRouter{
		Signer: signer,
		DB:     db,
		lg:     logger.NewSystem("rpc-router"),
	}

	return router, func() {
		dbCleanup()
	}
}

func TestRPCRouterHandlePing(t *testing.T) {
	router, cleanup := setupTestRPCRouter(t)
	defer cleanup()

	c := &RPCContext{
		Context: context.TODO(),
		Message: RPCMessage{
			Req: &RPCData{
				RequestID: 1,
				Method:    "ping",
				Params:    []any{nil},
				Timestamp: uint64(time.Now().Unix()),
			},
			Sig: []string{"dummy-signature"},
		},
	}

	router.HandlePing(c)
	res := c.Message.Res
	require.NotNil(t, res)

	assert.Equal(t, "pong", res.Method)
}
