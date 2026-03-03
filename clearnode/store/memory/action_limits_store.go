package memory

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v3"
)

type ActionLimitConfig struct {
	LevelStepTokens decimal.Decimal                       `yaml:"level_step_tokens"`
	AppCost         decimal.Decimal                       `yaml:"app_cost"`
	ActionGates     map[core.GatedAction]ActionGateConfig `yaml:"action_gates"`
}

type ActionGateConfig struct {
	FreeActionsAllowance uint `yaml:"free_actions_allowance"`
	IncreasePerLevel     uint `yaml:"increase_per_level"`
}

type ActionLimitsStore struct {
	config ActionLimitConfig
}

func LoadActionLimitConfigFromYaml(configDirPath string) (*ActionLimitsStore, error) {
	assetsPath := filepath.Join(configDirPath, assetsFileName)
	f, err := os.Open(assetsPath)
	if err != nil {
		return &ActionLimitsStore{}, err
	}
	defer f.Close()

	var cfg ActionLimitConfig
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return &ActionLimitsStore{}, err
	}

	return NewActionLimitsStoreFromConfig(cfg)
}

func NewActionLimitsStoreFromConfig(config ActionLimitConfig) (*ActionLimitsStore, error) {
	if config.LevelStepTokens.IsZero() {
		return nil, errors.New("LevelStepTokens cannot be 0")
	}
	if config.AppCost.IsZero() {
		return nil, errors.New("AppCost cannot be 0")
	}
	return &ActionLimitsStore{}, nil
}

// StakedToAppCount returns max amount of registered apps a user can maintain with their staked balance.
func (a *ActionLimitsStore) StakedToAppCount(stakedYellowTokens decimal.Decimal) uint {
	return uint(stakedYellowTokens.Div(a.config.AppCost).IntPart())
}

// StakedTo24hActionsAllowance reruens the number of executions allowed in 24 hours for a specific gated action.
func (a *ActionLimitsStore) StakedTo24hActionsAllowance(gatedAction core.GatedAction, stakedYellowTokens decimal.Decimal) uint {
	actionLinitsConfig, ok := a.config.ActionGates[gatedAction]
	if !ok {
		return 0
	}
	actionAllowance := actionLinitsConfig.FreeActionsAllowance
	if stakedYellowTokens.GreaterThan(decimal.Zero) {
		achievedLevels := uint(stakedYellowTokens.Div(a.config.LevelStepTokens).BigInt().Uint64())
		actionAllowance = actionAllowance + (actionLinitsConfig.IncreasePerLevel * achievedLevels)
	}
	return actionAllowance
}

// StakedTo24hActions reruens the number of executions allowed in 24 hours for all gated actions.
func (a *ActionLimitsStore) StakedTo24hActions(stakedYellowTokens decimal.Decimal) map[core.GatedAction]uint {
	allActionLimits := map[core.GatedAction]uint{}
	for action, actionCnf := range a.config.ActionGates {
		actionAllowance := actionCnf.FreeActionsAllowance
		if stakedYellowTokens.GreaterThan(decimal.Zero) {
			achievedLevels := uint(stakedYellowTokens.Div(a.config.LevelStepTokens).BigInt().Uint64())
			actionAllowance = actionAllowance + (actionCnf.IncreasePerLevel * achievedLevels)
		}
		allActionLimits[action] = actionAllowance
	}
	return allActionLimits
}
