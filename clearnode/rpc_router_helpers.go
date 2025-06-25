package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

const (
	DefaultPageSize = 10
	MaxPageSize     = 100
)

type PaginationParams struct {
	Offset   uint32 `json:"offset,omitempty"`
	PageSize uint32 `json:"page_size,omitempty"`
}

func paginate(params *PaginationParams) func(db *gorm.DB) *gorm.DB {
	if params == nil {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}

	offset := params.Offset
	pageSize := params.PageSize

	if pageSize == 0 {
		pageSize = DefaultPageSize
	} else if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(int(offset)).Limit(int(pageSize))
	}
}

type SortType string

const (
	Ascending  SortType = "asc"
	Descending SortType = "desc"
)

func (s SortType) ToString() string {
	return strings.ToUpper(string(s))
}

func applySort(db *gorm.DB, sortBy string, defaultSort SortType, sortType *SortType) *gorm.DB {
	if sortType == nil {
		return db.Order(sortBy + " " + defaultSort.ToString())
	}

	return db.Order(sortBy + " " + sortType.ToString())
}

func parseParams(params []any, unmarshalTo any) error {
	if len(params) > 0 {
		paramsJSON, err := json.Marshal(params[0])
		if err != nil {
			return fmt.Errorf("failed to parse parameters: %w", err)
		}
		err = json.Unmarshal(paramsJSON, &unmarshalTo)
		if err != nil {
			return err
		}
	}
	return validate.Struct(unmarshalTo)
}
