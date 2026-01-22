package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	db "github.com/erc7824/nitrolite/clearnode/store/database"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	container "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB chooses SQLite or Postgres based on TEST_DB_DRIVER environment variable.
// This is exported so it can be used by tests in other packages.
func SetupTestDB(t testing.TB) (*gorm.DB, func()) {
	t.Helper()

	ctx := context.Background()
	var database *gorm.DB
	var cleanup func()

	switch os.Getenv("TEST_DB_DRIVER") {
	case "postgres":
		log.Println("Using PostgreSQL for testing")
		var container testcontainers.Container
		database, container = setupTestPostgres(ctx, t)
		cleanup = func() {
			if container != nil {
				if err := container.Terminate(ctx); err != nil {
					log.Printf("Failed to terminate PostgreSQL container: %v", err)
				}
			}
		}
	default:
		log.Println("Using SQLite for testing (default)")
		database = setupTestSqlite(t)
		cleanup = func() {}
	}

	return database, cleanup
}

// setupTestSqlite creates an in-memory SQLite DB for testing
func setupTestSqlite(t testing.TB) *gorm.DB {
	t.Helper()

	uniqueDSN := fmt.Sprintf("file::memory:test%s?mode=memory&cache=shared", uuid.NewString())
	database, err := gorm.Open(sqlite.Open(uniqueDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}

	err = database.AutoMigrate(&db.AppLedgerEntryV1{}, &db.Channel{}, &db.AppSessionV1{}, &db.RPCRecord{}, &db.ContractEvent{}, &db.Transaction{}, &db.BlockchainAction{}, &db.SessionKey{})
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
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
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}
	log.Println("Started container:", postgresContainer.GetContainerID())

	url, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}
	log.Println("PostgreSQL URL:", url)

	database, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open PostgreSQL database: %v", err)
	}

	err = database.AutoMigrate(&db.AppLedgerEntryV1{}, &db.Channel{}, &db.AppSessionV1{}, &db.RPCRecord{}, &db.ContractEvent{}, &db.Transaction{}, &db.BlockchainAction{}, &db.SessionKey{})
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database, postgresContainer
}
