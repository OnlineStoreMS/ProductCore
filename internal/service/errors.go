package service

import "errors"

var (
	ErrNotFound            = errors.New("record not found")
	ErrDuplicateSN         = errors.New("product sn already exists")
	ErrDuplicateMaterialCode = errors.New("material code already exists")
	ErrMaterialCodeRequired  = errors.New("material code is required")
	ErrDuplicateSku        = errors.New("sku code already exists")
	ErrInvalidSkuCode      = errors.New("sku code must be alphanumeric (A-Z, a-z, 0-9)")
	ErrSkuCodeRequired     = errors.New("sku code is required")
	ErrDuplicateSpecValue  = errors.New("duplicate spec value in the same spec")
	ErrDuplicateSpecCombo  = errors.New("duplicate sku spec combination")
	ErrInvalidImport       = errors.New("invalid product import")
	ErrInvalidExport       = errors.New("invalid product export")
	ErrNoEditDraft         = errors.New("no edit draft")
	ErrNotInTrash          = errors.New("product is not in trash")
	ErrDuplicateTypeCode   = errors.New("platform type code already exists")
	ErrBuiltinPlatformType = errors.New("builtin platform type cannot be deleted")
	ErrTypeHasShops        = errors.New("platform type has bound shops")
)
