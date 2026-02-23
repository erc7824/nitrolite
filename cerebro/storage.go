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

func (s *Storage) SetSessionKeyPrivateKey(privateKey string) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO config (key, value) VALUES ('session_key_private_key', ?)", privateKey)
	return err
}

func (s *Storage) GetSessionKeyPrivateKey() (string, error) {
	var privateKey string
	err := s.db.QueryRow("SELECT value FROM config WHERE key = 'session_key_private_key'").Scan(&privateKey)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("no session key private key configured")
	}
	return privateKey, err
}

func (s *Storage) SetSessionKey(privateKey, metadataHash, authSig string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, kv := range []struct{ key, value string }{
		{"session_key_private_key", privateKey},
		{"session_key_metadata_hash", metadataHash},
		{"session_key_auth_sig", authSig},
	} {
		if _, err := tx.Exec("INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)", kv.key, kv.value); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Storage) GetSessionKey() (privateKey, metadataHash, authSig string, err error) {
	for _, kv := range []struct {
		key  string
		dest *string
	}{
		{"session_key_private_key", &privateKey},
		{"session_key_metadata_hash", &metadataHash},
		{"session_key_auth_sig", &authSig},
	} {
		err = s.db.QueryRow("SELECT value FROM config WHERE key = ?", kv.key).Scan(kv.dest)
		if err == sql.ErrNoRows {
			return "", "", "", fmt.Errorf("no session key configured")
		}
		if err != nil {
			return "", "", "", err
		}
	}
	return
}

func (s *Storage) ClearSessionKey() error {
	_, err := s.db.Exec("DELETE FROM config WHERE key IN ('session_key_private_key', 'session_key_metadata_hash', 'session_key_auth_sig')")
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}
