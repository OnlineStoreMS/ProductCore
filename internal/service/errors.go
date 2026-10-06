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
	ErrNoCollectAgent           = errors.New("没有在线且已启用「电商商品采集」的电脑。请在采集电脑托盘的「本机能力」里新建并启用")
	ErrCollectAgentOutdated     = errors.New("在线电脑还没有启用「电商商品采集」。请在采集电脑托盘的「本机能力」里新建，选好浏览器并启用")
	ErrAgentsCenterUnconfigured = errors.New("AgentsCenter 未配置，无法下发采集任务")
	ErrCollectEmpty             = errors.New("采集结果缺少标题或主图，无法入库")
)
