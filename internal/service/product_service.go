package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"productcore/internal/dto"
	"productcore/internal/model"
	"productcore/internal/pkg/util"
	"productcore/internal/repo"

	"gorm.io/gorm"
)

type ProductService struct {
	repo *repo.ProductRepo
	meta *repo.Repos
}

func NewProductService(repos *repo.Repos) *ProductService {
	return &ProductService{repo: repos.Product, meta: repos}
}

func (s *ProductService) List(q dto.ProductQuery) ([]dto.ProductDTO, int64, error) {
	products, total, err := s.repo.List(q)
	if err != nil {
		return nil, 0, err
	}
	list := make([]dto.ProductDTO, 0, len(products))
	for _, p := range products {
		item, err := s.toDTO(&p, false)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *item)
	}
	return list, total, nil
}

func (s *ProductService) ListPublished(q dto.ProductQuery) ([]dto.ProductDTO, int64, error) {
	published := int8(1)
	q.PublishStatus = &published
	return s.List(q)
}

func (s *ProductService) Get(id uint64) (*dto.ProductDTO, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.toDTO(p, true)
}

func (s *ProductService) Create(in *dto.ProductDTO) (*dto.ProductDTO, error) {
	if err := s.ensureUniqueSN(in.ProductSn, 0); err != nil {
		return nil, err
	}
	var out *dto.ProductDTO
	err := s.repo.Transaction(func(tx *repo.ProductRepo) error {
		svc := &ProductService{repo: tx, meta: s.meta}
		p, err := svc.fromDTO(in)
		if err != nil {
			return err
		}
		if err := tx.Create(p); err != nil {
			return err
		}
		if err := svc.saveSkus(tx, p.ID, in.Skus); err != nil {
			return err
		}
		if err := svc.saveGroups(tx, p.ID, in.GroupIDs); err != nil {
			return err
		}
		if err := svc.syncSummary(tx, p.ID); err != nil {
			return err
		}
		out, err = svc.loadDTO(tx, p.ID)
		return err
	})
	return out, err
}

func (s *ProductService) Update(id uint64, in *dto.ProductDTO) (*dto.ProductDTO, error) {
	if err := s.ensureUniqueSN(in.ProductSn, id); err != nil {
		return nil, err
	}
	var out *dto.ProductDTO
	err := s.repo.Transaction(func(tx *repo.ProductRepo) error {
		svc := &ProductService{repo: tx, meta: s.meta}
		p, err := tx.GetByID(id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		updated, err := svc.fromDTO(in)
		if err != nil {
			return err
		}
		if err := tx.UpdateFields(id, map[string]interface{}{
			"name": updated.Name, "sub_title": updated.SubTitle, "product_sn": updated.ProductSn,
			"brand_id": updated.BrandID, "category_id": updated.CategoryID, "pic": updated.Pic,
			"album_pics": updated.AlbumPics, "product_video": updated.ProductVideo,
			"price": updated.Price, "original_price": updated.OriginalPrice, "stock": updated.Stock,
			"unit": updated.Unit, "weight": updated.Weight, "publish_status": updated.PublishStatus,
			"verify_status": updated.VerifyStatus, "sort": updated.Sort,
			"description": updated.Description, "detail_html": updated.DetailHTML,
			"sku_specs_json": updated.SkuSpecsJSON, "channel_visible": updated.ChannelVisible,
		}); err != nil {
			return err
		}
		_ = p
		if err := tx.DeleteSkusByProduct(id); err != nil {
			return err
		}
		if err := svc.saveSkus(tx, id, in.Skus); err != nil {
			return err
		}
		if err := tx.DeleteGroupRelations(id); err != nil {
			return err
		}
		if err := svc.saveGroups(tx, id, in.GroupIDs); err != nil {
			return err
		}
		if err := svc.syncSummary(tx, id); err != nil {
			return err
		}
		out, err = svc.loadDTO(tx, id)
		return err
	})
	return out, err
}

func (s *ProductService) Delete(id uint64) error {
	return s.repo.Transaction(func(tx *repo.ProductRepo) error {
		if err := tx.DeleteSkusByProduct(id); err != nil {
			return err
		}
		if err := tx.DeleteGroupRelations(id); err != nil {
			return err
		}
		if err := tx.Delete(id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		return nil
	})
}

func (s *ProductService) UpdatePublishStatus(id uint64, status int8) error {
	if err := s.repo.UpdatePublishStatus(id, status); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *ProductService) ensureUniqueSN(sn string, excludeID uint64) error {
	if sn == "" {
		return nil
	}
	count, err := s.repo.CountBySN(sn, excludeID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrDuplicateSN
	}
	return nil
}

func (s *ProductService) saveSkus(tx *repo.ProductRepo, productID uint64, skus []dto.SkuDTO) error {
	for _, item := range skus {
		if item.SkuCode == "" {
			continue
		}
		n, err := tx.CountSkuByCode(item.SkuCode)
		if err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("%w: %s", ErrDuplicateSku, item.SkuCode)
		}
		sku := &model.Sku{
			ProductID: productID,
			SkuCode:   item.SkuCode,
			SpecData:  util.ToJSON(item.Specs),
			Price:     item.Price,
			CostPrice: item.CostPrice,
			Stock:     item.Stock,
			Pic:       item.Pic,
		}
		if err := tx.CreateSku(sku); err != nil {
			return err
		}
	}
	return nil
}

func (s *ProductService) saveGroups(tx *repo.ProductRepo, productID uint64, groupIDs []uint64) error {
	for _, gid := range groupIDs {
		if gid == 0 {
			continue
		}
		if err := tx.CreateGroupRelation(productID, gid); err != nil {
			return err
		}
	}
	return nil
}

func (s *ProductService) syncSummary(tx *repo.ProductRepo, productID uint64) error {
	skus, err := tx.ListSkus(productID)
	if err != nil {
		return err
	}
	totalStock := 0
	minPrice := 0.0
	for i, sku := range skus {
		totalStock += sku.Stock
		if i == 0 || (sku.Price > 0 && sku.Price < minPrice) {
			minPrice = sku.Price
		}
	}
	updates := map[string]interface{}{"stock": totalStock}
	if minPrice > 0 {
		updates["price"] = minPrice
	}
	return tx.UpdateFields(productID, updates)
}

func (s *ProductService) fromDTO(in *dto.ProductDTO) (*model.Product, error) {
	channel := in.ChannelVisible
	if channel == "" {
		channel = "both"
	}
	return &model.Product{
		Name: in.Name, SubTitle: in.SubTitle, ProductSn: in.ProductSn,
		BrandID: in.BrandID, CategoryID: in.CategoryID, Pic: in.Pic,
		AlbumPics: util.ToJSON(in.AlbumPics), ProductVideo: in.ProductVideo,
		Price: in.Price, OriginalPrice: in.OriginalPrice, Stock: in.Stock,
		Unit: in.Unit, Weight: in.Weight, PublishStatus: in.PublishStatus,
		VerifyStatus: in.VerifyStatus, Sort: in.Sort, Description: in.Description,
		DetailHTML: in.DetailHTML, SkuSpecsJSON: util.ToJSON(in.SkuSpecs),
		ChannelVisible: channel,
	}, nil
}

func (s *ProductService) loadDTO(tx *repo.ProductRepo, id uint64) (*dto.ProductDTO, error) {
	p, err := tx.GetByID(id)
	if err != nil {
		return nil, err
	}
	svc := &ProductService{repo: tx, meta: s.meta}
	return svc.toDTO(p, true)
}

func (s *ProductService) toDTO(p *model.Product, withSkus bool) (*dto.ProductDTO, error) {
	out := &dto.ProductDTO{
		ID: p.ID, Name: p.Name, SubTitle: p.SubTitle, ProductSn: p.ProductSn,
		BrandID: p.BrandID, CategoryID: p.CategoryID, Pic: p.Pic,
		AlbumPics: util.ParseStringArray(p.AlbumPics), ProductVideo: p.ProductVideo,
		Price: p.Price, OriginalPrice: p.OriginalPrice, Stock: p.Stock,
		Unit: p.Unit, Weight: p.Weight, PublishStatus: p.PublishStatus,
		VerifyStatus: p.VerifyStatus, Sort: p.Sort, Sale: p.Sale,
		Description: p.Description, DetailHTML: p.DetailHTML,
		ChannelVisible: p.ChannelVisible,
		CreateTime: util.FormatTime(p.CreatedAt), UpdateTime: util.FormatTime(p.UpdatedAt),
	}
	if p.BrandID > 0 {
		out.BrandName = s.meta.Brand.GetName(p.BrandID)
	}
	if p.CategoryID > 0 {
		out.CategoryName = s.meta.Category.GetName(p.CategoryID)
	}
	var specs []dto.SkuSpecDTO
	_ = json.Unmarshal([]byte(p.SkuSpecsJSON), &specs)
	out.SkuSpecs = specs
	groupIDs, _ := s.repo.ListGroupIDs(p.ID)
	out.GroupIDs = groupIDs
	skuCount, _ := s.repo.CountSkus(p.ID)
	out.SkuCount = int(skuCount)
	if withSkus {
		skus, err := s.repo.ListSkus(p.ID)
		if err != nil {
			return nil, err
		}
		out.Skus = make([]dto.SkuDTO, 0, len(skus))
		for _, sku := range skus {
			specsMap := map[string]string{}
			_ = json.Unmarshal([]byte(sku.SpecData), &specsMap)
			out.Skus = append(out.Skus, dto.SkuDTO{
				ID: sku.ID, SkuCode: sku.SkuCode, Specs: specsMap,
				Price: sku.Price, CostPrice: sku.CostPrice, Stock: sku.Stock, Pic: sku.Pic,
			})
		}
	}
	return out, nil
}
