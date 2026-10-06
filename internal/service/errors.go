package service

import "errors"

var (
	ErrNotFound                 = errors.New("record not found")
	ErrDuplicateSN              = errors.New("product sn already exists")
	ErrDuplicateMaterialCode    = errors.New("material code already exists")
	ErrMaterialCodeRequired     = errors.New("material code is required")
	ErrDuplicateSku             = errors.New("sku code already exists")
	ErrInvalidSkuCode           = errors.New("sku code must be alphanumeric (A-Z, a-z, 0-9)")
	ErrSkuCodeRequired          = errors.New("sku code is required")
	ErrDuplicateSpecValue       = errors.New("duplicate spec value in the same spec")
	ErrDuplicateSpecCombo       = errors.New("duplicate sku spec combination")
	ErrInvalidImport            = errors.New("invalid product import")
	ErrInvalidExport            = errors.New("invalid product export")
	ErrNoEditDraft              = errors.New("no edit draft")
	ErrNotInTrash               = errors.New("product is not in trash")
	ErrDuplicateTypeCode        = errors.New("platform type code already exists")
	ErrBuiltinPlatformType      = errors.New("builtin platform type cannot be deleted")
	ErrTypeHasShops             = errors.New("platform type has bound shops")
	ErrUnsupportedProductURL    = errors.New("暂只支持淘宝、天猫商品链接")
	ErrNoCollectAgent           = errors.New("没有在线的 WindowsAgent。请确认采集电脑上的代理已连接")
	ErrCollectAgentOutdated     = errors.New("在线的 WindowsAgent 还没有电商商品采集能力，请在采集电脑上更新并重启")
	ErrAgentsCenterUnconfigured = errors.New("AgentsCenter 未配置，无法下发采集任务")
)
