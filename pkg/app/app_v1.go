package app

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var AppIDV1Regex = regexp.MustCompile(`^[a-z0-9][-a-z0-9]{0,65}$`)

// AppV1 represents an application registry entry.
type AppV1 struct {
	ID                          string
	OwnerWallet                 string
	Metadata                    string
	Version                     uint64
	CreationApprovalNotRequired bool
}

// AppInfoV1 represents full application info including timestamps.
type AppInfoV1 struct {
	App       AppV1
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GetAppOwnerFunc is a function that returns the owner wallet for a given app ID.
type GetAppOwnerFunc func(appID string) (ownerWallet string, err error)

// AppOwnerValidator validates that a wallet is the owner of an application.
type AppOwnerValidator struct {
	getOwner GetAppOwnerFunc
}

// NewAppOwnerValidator creates a new AppOwnerValidator with the provided lookup function.
func NewAppOwnerValidator(getOwner GetAppOwnerFunc) *AppOwnerValidator {
	return &AppOwnerValidator{getOwner: getOwner}
}

// ValidateOwner checks that the given wallet is the owner of the specified app.
// Returns an error if the app is not found or the wallet does not match.
func (v *AppOwnerValidator) ValidateOwner(appID, wallet string) error {
	owner, err := v.getOwner(appID)
	if err != nil {
		return fmt.Errorf("failed to get app owner: %w", err)
	}

	if !strings.EqualFold(owner, wallet) {
		return fmt.Errorf("wallet %s is not the owner of app %s", wallet, appID)
	}

	return nil
}

// PackAppV1 packs the AppV1 for signing using ABI encoding.
func PackAppV1(app AppV1) ([]byte, error) {
	args := abi.Arguments{
		{Type: abi.Type{T: abi.StringTy}},               // id
		{Type: abi.Type{T: abi.AddressTy}},              // ownerWallet
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // metadata (bytes32)
		{Type: abi.Type{T: abi.UintTy, Size: 64}},       // version
		{Type: abi.Type{T: abi.BoolTy}},                 // creationApprovalNotRequired
	}

	appMetadataHash := common.HexToHash(app.Metadata)

	packed, err := args.Pack(
		app.ID,
		common.HexToAddress(app.OwnerWallet),
		appMetadataHash,
		app.Version,
		app.CreationApprovalNotRequired,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pack app: %w", err)
	}

	return crypto.Keccak256(packed), nil
}
