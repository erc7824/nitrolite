package storage

import (
	"fmt"
	"time"

	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	defaultDBPath = "cerebro.db"
)

type Storage struct {
	db *gorm.DB
}

func NewStorage(path string) (*Storage, error) {
	if path == "" {
		path = defaultDBPath
	}

	dsn := fmt.Sprintf("file:%s?cache=shared", path)

	dial := sqlite.Open(dsn)
	dbConf := &gorm.Config{
		Logger: logger.Default.LogMode(logger.LogLevel(0)), // Disable logging
	}
	db, err := gorm.Open(dial, dbConf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
	}

	if err := db.AutoMigrate(&PrivateKeyDTO{}, &ChainRPCDTO{}); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database schema: %w", err)
	}

	return &Storage{db: db}, nil
}

type PrivateKeyDTO struct {
	Address    string `gorm:"column:address;primaryKey"`
	Name       string `gorm:"column:name;not null;unique"`
	PrivateKey string `gorm:"column:private_key;not null;unique"`
	IsSigner   bool   `gorm:"column:is_signer;not null"`
}

func (s *Storage) AddPrivateKey(name, privateKeyHex string, isSigner bool) (*PrivateKeyDTO, error) {
	signer, err := sign.NewEthereumSigner(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	dto := PrivateKeyDTO{
		Address:    signer.PublicKey().Address().String(),
		Name:       name,
		PrivateKey: privateKeyHex,
		IsSigner:   isSigner,
	}

	if err := s.db.Create(&dto).Error; err != nil {
		return nil, fmt.Errorf("failed to add private key: %w", err)
	}

	return &dto, nil
}

func (s *Storage) GetPrivateKeys(isSigner bool) ([]PrivateKeyDTO, error) {
	var keys []PrivateKeyDTO
	if err := s.db.Where("is_signer = ?", isSigner).Find(&keys).Error; err == gorm.ErrRecordNotFound {
		// No keys found, return an empty slice
		return []PrivateKeyDTO{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to retrieve private keys: %w", err)
	}
	return keys, nil
}

func (s *Storage) GetPrivateKeyByName(name string) (*PrivateKeyDTO, error) {
	var key PrivateKeyDTO
	if err := s.db.Where("name = ?", name).First(&key).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("private key not found for name %s", name)
		}
		return nil, fmt.Errorf("failed to retrieve private key: %w", err)
	}
	return &key, nil
}

func (s *Storage) DeletePrivateKey(address string) error {
	if err := s.db.Where("address = ?", address).Delete(&PrivateKeyDTO{}).Error; err != nil {
		return fmt.Errorf("failed to delete private key: %w", err)
	}
	return nil
}

type ChainRPCDTO struct {
	URL        string    `gorm:"column:url;primaryKey"`
	ChainID    uint32    `gorm:"column:chain_id;not null"`
	LastUsedAt time.Time `gorm:"column:last_used_at;not null"`
}

func (s *Storage) AddChainRPC(url string, chainID uint32) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	dto := ChainRPCDTO{
		URL:        url,
		ChainID:    chainID,
		LastUsedAt: time.Unix(0, 0), // Initialize to zero time
	}

	if err := s.db.Create(&dto).Error; err != nil {
		return fmt.Errorf("failed to add chain RPC: %w", err)
	}

	return nil
}

func (s *Storage) GetChainRPCs(chainID uint32) ([]ChainRPCDTO, error) {
	var rpcs []ChainRPCDTO
	if err := s.db.Where("chain_id = ?", chainID).Find(&rpcs).Order("last_used_at ASC").Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve chain RPCs: %w", err)
	}

	return rpcs, nil
}

func (s *Storage) UpdateChainRPCUsage(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	if err := s.db.Model(&ChainRPCDTO{}).Where("url = ?", url).
		Update("last_used_at", time.Now().UTC()).Error; err != nil {
		return fmt.Errorf("failed to update chain RPC usage: %w", err)
	}

	return nil
}
