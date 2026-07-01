package service

import (
	"errors"

	"productcore/internal/dto"
	"productcore/internal/repo"

	"gorm.io/gorm"
)

type PlatformListingService struct {
	listing    *repo.PlatformListingRepo
	shopRepo   *repo.PlatformShopRepo
	product    *repo.ProductRepo
	productSvc *ProductService
	tenantID   uint64
}

func NewPlatformListingService(repos *repo.Repos, productSvc *ProductService) *PlatformListingService {
	return &PlatformListingService{
		listing:    repos.PlatformListing,
		shopRepo:   repos.PlatformShop,
		product:    repos.Product,
		productSvc: productSvc,
		tenantID:   1,
	}
}

func (s *PlatformListingService) ForTenant(tenantID uint64) *PlatformListingService {
	cp := *s
	cp.tenantID = repo.NormalizeTenantID(tenantID)
	cp.product = s.product.WithTenant(tenantID)
	cp.shopRepo = s.shopRepo.WithTenant(tenantID)
	cp.productSvc = s.productSvc.ForTenant(tenantID)
	return &cp
}

func (s *PlatformListingService) ListShopsByProduct(productID uint64) ([]dto.ListedShopDTO, error) {
	if _, err := s.product.GetByID(productID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	m, err := s.listing.ShopsByProductIDs([]uint64{productID})
	if err != nil {
		return nil, err
	}
	shops := m[productID]
	if shops == nil {
		return []dto.ListedShopDTO{}, nil
	}
	return shops, nil
}

func (s *PlatformListingService) SetProductListings(productID uint64, shopIDs []uint64) ([]dto.ListedShopDTO, error) {
	if _, err := s.product.GetByID(productID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	for _, sid := range shopIDs {
		if sid == 0 {
			continue
		}
		if _, err := s.shopRepo.GetByID(sid); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("shop not found")
			}
			return nil, err
		}
	}
	if err := s.listing.ReplaceProductListings(s.tenantID, productID, shopIDs); err != nil {
		return nil, err
	}
	return s.ListShopsByProduct(productID)
}

func (s *PlatformListingService) AttachListedShops(list []dto.ProductDTO) error {
	if len(list) == 0 {
		return nil
	}
	ids := make([]uint64, len(list))
	for i, p := range list {
		ids[i] = p.ID
	}
	m, err := s.listing.ShopsByProductIDs(ids)
	if err != nil {
		return err
	}
	for i := range list {
		shops := m[list[i].ID]
		if shops == nil {
			list[i].ListedShops = []dto.ListedShopDTO{}
			list[i].ListedShopCount = 0
			continue
		}
		list[i].ListedShops = shops
		list[i].ListedShopCount = len(shops)
	}
	return nil
}

func (s *PlatformListingService) ListProductsByShop(shopID uint64, keyword string, page, pageSize int) ([]dto.ProductDTO, int64, error) {
	if _, err := s.shopRepo.GetByID(shopID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, ErrNotFound
		}
		return nil, 0, err
	}
	ids, total, err := s.listing.ListProductIDsByShop(shopID, keyword, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.ProductDTO, 0, len(ids))
	for _, id := range ids {
		item, err := s.productSvc.Get(id)
		if err != nil {
			continue
		}
		out = append(out, *item)
	}
	return out, total, nil
}
