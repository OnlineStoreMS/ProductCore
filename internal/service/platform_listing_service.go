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
}

func NewPlatformListingService(repos *repo.Repos, productSvc *ProductService) *PlatformListingService {
	return &PlatformListingService{
		listing:    repos.PlatformListing,
		shopRepo:   repos.PlatformShop,
		product:    repos.Product,
		productSvc: productSvc,
	}
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
	if err := s.listing.ReplaceProductListings(productID, shopIDs); err != nil {
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
			shops = []dto.ListedShopDTO{}
		}
		list[i].ListedShops = shops
		list[i].ListedShopCount = len(shops)
	}
	return nil
}

func (s *PlatformListingService) ListProductsByShop(shopID uint64, q dto.ProductQuery) ([]dto.ProductDTO, int64, error) {
	if _, err := s.shopRepo.GetByID(shopID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, ErrNotFound
		}
		return nil, 0, err
	}
	ids, total, err := s.listing.ListProductIDsByShop(shopID, q.Keyword, q.Page, q.PageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(ids) == 0 {
		return []dto.ProductDTO{}, total, nil
	}
	out := make([]dto.ProductDTO, 0, len(ids))
	for _, id := range ids {
		p, err := s.product.GetByID(id)
		if err != nil {
			continue
		}
		item, err := s.productSvc.toDTO(p, false)
		if err != nil {
			continue
		}
		out = append(out, *item)
	}
	_ = s.AttachListedShops(out)
	return out, total, nil
}

func (s *PlatformListingService) ProductCountsByShopIDs(shopIDs []uint64) (map[uint64]int64, error) {
	return s.listing.CountProductsByShopIDs(shopIDs)
}
