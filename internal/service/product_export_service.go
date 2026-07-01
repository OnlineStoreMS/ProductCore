package service

import (
	"fmt"

	"productcore/internal/pkg/productimport"
	"productcore/internal/storage"
)

type ProductExportService struct {
	products *ProductService
	store    storage.Storage
}

func NewProductExportService(products *ProductService, store storage.Storage) *ProductExportService {
	return &ProductExportService{products: products, store: store}
}

func (s *ProductExportService) ForTenant(tenantID uint64) *ProductExportService {
	cp := *s
	cp.products = s.products.ForTenant(tenantID)
	return &cp
}

// ExportToZip 导出商品为 ProductCore 标准 zip 包
func (s *ProductExportService) ExportToZip(productID uint64) (zipPath, downloadName string, cleanup func(), err error) {
	cleanup = func() {}
	product, err := s.products.Get(productID)
	if err != nil {
		return "", "", cleanup, err
	}
	zipPath, cleanup, err = productimport.ExportProduct(product, s.store)
	if err != nil {
		return "", "", cleanup, fmt.Errorf("%w: %s", ErrInvalidExport, err.Error())
	}
	downloadName = productimport.BuildExportFolderName(product.ID) + ".zip"
	return zipPath, downloadName, cleanup, nil
}
