package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS rpcs (
			chain_id INTEGER PRIMARY KEY,
			rpc_url TEXT NOT NULL
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SetPrivateKey(privateKey string) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO config (key, value) VALUES ('private_key', ?)", privateKey)
	return err
}

func (s *Storage) GetPrivateKey() (string, error) {
	var privateKey string
	err := s.db.QueryRow("SELECT value FROM config WHERE key = 'private_key'").Scan(&privateKey)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("no private key configured")
	}
	return privateKey, err
}

func (s *Storage) SetRPC(chainID uint64, rpcURL string) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO rpcs (chain_id, rpc_url) VALUES (?, ?)", chainID, rpcURL)
	return err
}

func (s *Storage) GetRPC(chainID uint64) (string, error) {
	var rpcURL string
	err := s.db.QueryRow("SELECT rpc_url FROM rpcs WHERE chain_id = ?", chainID).Scan(&rpcURL)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("no RPC configured for chain %d", chainID)
	}
	return rpcURL, err
}

func (s *Storage) GetAllRPCs() (map[uint64]string, error) {
	rows, err := s.db.Query("SELECT chain_id, rpc_url FROM rpcs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rpcs := make(map[uint64]string)
	for rows.Next() {
		var chainID uint64
		var rpcURL string
		if err := rows.Scan(&chainID, &rpcURL); err != nil {
			return nil, err
		}
		rpcs[chainID] = rpcURL
	}
	return rpcs, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
