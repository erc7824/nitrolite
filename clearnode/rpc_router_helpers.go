package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

const (
	DefaultOffset   = 0
	DefaultPageSize = 10
	MaxPageSize     = 100
)

type PaginationParams struct {
	Offset   int32 `json:"offset"`
	PageSize int32 `json:"page_size"`
}

// Normalize applies validation and defaults to pagination parameters
func (pp *PaginationParams) Normalize(defaultPageSize int32) {
	if pp.Offset < 0 {
		pp.Offset = DefaultOffset
	}

	// NOTE: default value can be supplied, and will be preserved if the value is not set or invalid
	if defaultPageSize != 0 && (pp.PageSize <= 0 || pp.PageSize > MaxPageSize) {
		pp.PageSize = defaultPageSize
		return
	}

	if pp.PageSize <= 0 {
		pp.PageSize = DefaultPageSize
	} else if pp.PageSize > MaxPageSize {
		pp.PageSize = MaxPageSize
	}
}

func paginate(params *PaginationParams) func(db *gorm.DB) *gorm.DB {
	if params == nil {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(int(params.Offset)).Limit(int(params.PageSize))
	}
}

func parseParams(params []any, unmarshalTo any) error {
	if len(params) == 0 {
		return errors.New("missing parameters")
	}
	paramsJSON, err := json.Marshal(params[0])
	if err != nil {
		return fmt.Errorf("failed to parse parameters: %w", err)
	}
	return json.Unmarshal(paramsJSON, &unmarshalTo)
}

func parseOptionalParams(params []any, unmarshalTo any) error {
	if len(params) == 0 {
		return nil
	}
	paramsJSON, err := json.Marshal(params[0])
	if err != nil {
		return fmt.Errorf("failed to parse optional parameters: %w", err)
	}
	return json.Unmarshal(paramsJSON, &unmarshalTo)
}
