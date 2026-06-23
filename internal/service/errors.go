package service

import "errors"

var (
	ErrNotFound     = errors.New("record not found")
	ErrDuplicateSN  = errors.New("product sn already exists")
	ErrDuplicateSku = errors.New("sku code already exists")
)
