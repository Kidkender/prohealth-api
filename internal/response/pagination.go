package response

import "math"

type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedResponse[T any] struct {
	Data []T            `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

func NewPaginatedResponse[T any](data []T, total int64, page, limit int) PaginatedResponse[T] {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	return PaginatedResponse[T]{
		Data: data,
		Meta: PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}
}
