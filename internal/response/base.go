package response

import "math"

type Response[T any] struct {
	Success bool            `json:"success"`
	Data    *T              `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
	Meta    *PaginationMeta `json:"meta,omitempty"`
}

func Ok[T any](data T) Response[T] {
	return Response[T]{
		Success: true,
		Data:    &data,
	}
}

func Done() Response[struct{}] {
	return Response[struct{}]{Success: true}
}

func Err(msg string) Response[struct{}] {
	return Response[struct{}]{
		Success: false,
		Error:   msg,
	}
}

func Paginated[T any](data []T, total int64, page, limit int) Response[[]T] {
	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}
	return Response[[]T]{
		Success: true,
		Data:    &data,
		Meta: &PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}
}
