package utils

import (
	"net/http"
	"strings"
)

// QueryParams struct chứa các query parameters chung
type QueryParams struct {
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Sort    string `json:"sort"`
	Order   string `json:"order"`
	Search  string `json:"search"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
}

// ParseQueryParams parse query parameters từ HTTP request
func ParseQueryParams(r *http.Request) *QueryParams {
	params := &QueryParams{
		Page:    GetQueryParamInt(r, "page", 1),
		PerPage: GetQueryParamInt(r, "per_page", 10),
		Sort:    strings.TrimSpace(r.URL.Query().Get("sort")),
		Order:   strings.TrimSpace(r.URL.Query().Get("order")),
		Search:  strings.TrimSpace(r.URL.Query().Get("search")),
	}

	// Validate và set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 10
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}
	if params.Order == "" {
		params.Order = "asc"
	}
	if params.Order != "asc" && params.Order != "desc" {
		params.Order = "asc"
	}

	// Calculate offset and limit
	params.Offset = (params.Page - 1) * params.PerPage
	params.Limit = params.PerPage

	return params
}

// GetQueryParamString lấy query parameter dạng string với default value
func GetQueryParamString(r *http.Request, key string, defaultValue string) string {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return defaultValue
	}
	return value
}

// QueryParamsOptions options cho ParseQueryParamsFromOptions
type QueryParamsOptions struct {
	Page    int
	PerPage int
	Sort    string
	Order   string
	Search  string
}

// ParseQueryParamsFromOptions parse query parameters từ options
func ParseQueryParamsFromOptions(options QueryParamsOptions) *QueryParams {
	params := &QueryParams{
		Page:    options.Page,
		PerPage: options.PerPage,
		Sort:    strings.TrimSpace(options.Sort),
		Order:   strings.TrimSpace(options.Order),
		Search:  strings.TrimSpace(options.Search),
	}

	// Validate và set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 10
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}
	if params.Order == "" {
		params.Order = "asc"
	}
	if params.Order != "asc" && params.Order != "desc" {
		params.Order = "asc"
	}

	// Calculate offset and limit
	params.Offset = (params.Page - 1) * params.PerPage
	params.Limit = params.PerPage

	return params
}

// PaginatedResponse tạo response data chung cho pagination
func PaginatedResponse(items interface{}, pagination *Pagination) map[string]interface{} {
	return map[string]interface{}{
		"items":      items,
		"pagination": pagination,
	}
}
