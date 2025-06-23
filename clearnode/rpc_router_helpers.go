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

func (pp *PaginationParams) SetDefaultPageSize(defaultPageSize uint32) {
	if pp == nil {
		return
	}

	if pp.PageSize <= 0 {
		pp.PageSize = int32(defaultPageSize)
	}
}

func paginate(params *PaginationParams) func(db *gorm.DB) *gorm.DB {
	if params == nil {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}

	offset := params.Offset
	pageSize := params.PageSize

	if offset < 0 {
		offset = DefaultOffset
	}

	if pageSize <= 0 {
		pageSize = DefaultPageSize
	} else if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(int(offset)).Limit(int(pageSize))
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
