package dto

import "errors"

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

var ErrLimitExceeded = errors.New("error.pagination.limit_exceeded")

type PaginationQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

func (q *PaginationQuery) Normalize() {
	if q.Page < 1 {
		q.Page = DefaultPage
	}
	if q.Limit < 1 {
		q.Limit = DefaultLimit
	}
}

func (q *PaginationQuery) Validate() error {
	if q.Limit > MaxLimit {
		return ErrLimitExceeded
	}
	return nil
}

func (q PaginationQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}
