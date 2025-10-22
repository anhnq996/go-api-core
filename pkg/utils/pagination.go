package utils

import (
	"net/http"
)

// Pagination thông tin phân trang
type Pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	Offset     int   `json:"-"`
	Limit      int   `json:"-"`
}

// NewPagination tạo pagination mới
func NewPagination(page, perPage int, total int64) *Pagination {
	if page < 1 {
		page = 1
	}

	if perPage < 1 {
		perPage = 10
	}

	if perPage > 100 {
		perPage = 100
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	offset := (page - 1) * perPage

	return &Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		Offset:     offset,
		Limit:      perPage,
	}
}

// PaginationFromRequest tạo pagination từ HTTP request
func PaginationFromRequest(r *http.Request, total int64) *Pagination {
	page := GetQueryParamInt(r, "page", 1)
	perPage := GetQueryParamInt(r, "per_page", 10)

	return NewPagination(page, perPage, total)
}

// HasNextPage kiểm tra có trang tiếp theo không
func (p *Pagination) HasNextPage() bool {
	return p.Page < p.TotalPages
}

// HasPrevPage kiểm tra có trang trước không
func (p *Pagination) HasPrevPage() bool {
	return p.Page > 1
}

// NextPage trả về số trang tiếp theo
func (p *Pagination) NextPage() int {
	if p.HasNextPage() {
		return p.Page + 1
	}
	return p.Page
}

// PrevPage trả về số trang trước
func (p *Pagination) PrevPage() int {
	if p.HasPrevPage() {
		return p.Page - 1
	}
	return 1
}
