package common

import "errors"

const CorrelationID = "correlation_id"

type Pagination struct {
	Page     int
	PageSize int
}

func (p Pagination) GetOffset() int {
	return p.Page * p.PageSize
}

func (p Pagination) GetLimit() int {
	if p.PageSize == 0 {
		return 10
	}
	return p.PageSize
}

func (p Pagination) Validate() error {
	if p.Page < 0 {
		return errors.New("page must be a positive number")
	}
	if p.PageSize < 0 {
		return errors.New("pagesize must be a positive number")
	}

	return nil
}
