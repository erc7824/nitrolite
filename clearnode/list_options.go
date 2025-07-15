package main

import (
	"strings"

	"github.com/invopop/jsonschema"
	"gorm.io/gorm"
)

type SortType string

const (
	SortTypeAscending  SortType = "asc"
	SortTypeDescending SortType = "desc"
)

func (s SortType) String() string {
	return strings.ToUpper(string(s))
}

func (SortType) JSONSchema() *jsonschema.Schema {
	schema := &jsonschema.Schema{Type: "enum"}
	values := []SortType{SortTypeAscending, SortTypeDescending}
	for _, enum := range values {
		schema.Enum = append(schema.Enum, enum)
	}
	return schema
}

func applySort(db *gorm.DB, sortBy string, defaultSort SortType, sortType *SortType) *gorm.DB {
	if sortType == nil {
		return db.Order(sortBy + " " + defaultSort.String())
	}

	return db.Order(sortBy + " " + sortType.String())
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
		return applySort(db, sortBy, defaultSort, nil)
	}

	db = applySort(db, sortBy, defaultSort, options.Sort)
	db = paginate(&options.Offset, &options.Limit)(db)

	return db
}
