package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type SortType string

const (
	SortTypeAscending  SortType = "asc"
	SortTypeDescending SortType = "desc"
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

const (
	DefaultLimit = 10
	MaxLimit     = 100
)

func paginate(rawOffset, rawLimit *uint32) func(db *gorm.DB) *gorm.DB {
	offset := 0
	if rawOffset != nil {
		offset = int(*rawOffset)
	}

	limit := DefaultLimit
	if rawLimit != nil {
		limit = int(*rawLimit)
	}
	if limit == 0 {
		limit = DefaultLimit
	} else if limit > MaxLimit {
		limit = MaxLimit
	}

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(limit)
	}
}

type ListOptions struct {
	Offset uint32    `json:"offset,omitempty"`
	Limit  uint32    `json:"limit,omitempty"`
	Sort   *SortType `json:"sort,omitempty"` // Optional sort type (asc/desc)
}

func applyListOptions(db *gorm.DB, sortBy string, defaultSort SortType, options *ListOptions) *gorm.DB {
	if options == nil {
		return db
	}

	db = paginate(&options.Offset, &options.Limit)(db)
	db = applySort(db, sortBy, defaultSort, options.Sort)

	return db
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
