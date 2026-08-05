package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// Params holds parsed pagination query parameters.
type Params struct {
	Page     int
	PageSize int
}

// Offset returns the SQL OFFSET for these pagination params.
func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// FromQuery reads "page" and "page_size" query params, applying sane defaults and limits.
func FromQuery(c *gin.Context) Params {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		page = defaultPage
	}

	pageSize, err := strconv.Atoi(c.Query("page_size"))
	if err != nil || pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	return Params{Page: page, PageSize: pageSize}
}

// Meta describes pagination info to include in list responses.
type Meta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// NewMeta	 builds pagination Meta from the params and a total item count.
func NewMeta(p Params, totalItems int) Meta {
	totalPages := totalItems / p.PageSize
	if totalItems%p.PageSize != 0 {
		totalPages++
	}
	return Meta{
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
