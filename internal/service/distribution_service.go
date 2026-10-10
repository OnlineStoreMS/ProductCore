package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"productcore/internal/dto"
	"productcore/internal/model"
	"productcore/internal/pkg/zzbexcel"
	"productcore/internal/repo"
	"productcore/internal/storage"

	"gorm.io/gorm"
)

type DistributionService struct {
	repos    *repo.Repos
	store    storage.Storage
	tenantID uint64
}

func NewDistributionService(repos *repo.Repos, store storage.Storage) *DistributionService {
	return &DistributionService{repos: repos, store: store, tenantID: 1}
}

func (s *DistributionService) ForTenant(tenantID uint64) *DistributionService {
	cp := *s
	cp.tenantID = repo.NormalizeTenantID(tenantID)
	return &cp
}

func (s *DistributionService) shops() *repo.DistributionRepo {
	return s.repos.Distribution.WithTenant(s.tenantID)
}

func (s *DistributionService) ListShops(q dto.DistributionShopQuery) ([]dto.DistributionShopDTO, int64, error) {
	list, total, err := s.shops().ListShops(q)
	if err != nil {
		return nil, 0, err
	}
	ids := make([]uint64, 0, len(list))
	for _, item := range list {
		ids = append(ids, item.ID)
	}
	counts, err := s.shops().CountItems(ids)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.DistributionShopDTO, 0, len(list))
	for i := range list {
		d := s.shopDTO(&list[i])
		d.ItemCount = counts[list[i].ID]
		out = append(out, d)
	}
	return out, total, nil
}

func (s *DistributionService) GetShop(id uint64) (*dto.DistributionShopDTO, error) {
	item, err := s.shops().GetShop(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	d := s.shopDTO(item)
	counts, err := s.shops().CountItems([]uint64{id})
	if err != nil {
		return nil, err
	}
	d.ItemCount = counts[id]
	return &d, nil
}

func (s *DistributionService) CreateShop(in *dto.DistributionShopDTO) (*dto.DistributionShopDTO, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: 请填写店铺名称", ErrInvalidImport)
	}
	if in.PlatformTypeID == 0 {
		return nil, fmt.Errorf("%w: 请选择电商店铺类型", ErrInvalidImport)
	}
	pt, err := s.repos.PlatformType.GetByID(in.PlatformTypeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 店铺类型不存在", ErrInvalidImport)
		}
		return nil, err
	}
	item := &model.DistributionShop{
		PlatformTypeID: pt.ID,
		Name:           name,
		Remark:         strings.TrimSpace(in.Remark),
	}
	if err := s.shops().CreateShop(item); err != nil {
		return nil, err
	}
	d := s.shopDTO(item)
	d.PlatformTypeName = pt.Name
	d.PlatformTypeLogo = pt.Logo
	return &d, nil
}

func (s *DistributionService) UpdateShop(id uint64, in *dto.DistributionShopDTO) (*dto.DistributionShopDTO, error) {
	item, err := s.shops().GetShop(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: 请填写店铺名称", ErrInvalidImport)
	}
	if in.PlatformTypeID == 0 {
		return nil, fmt.Errorf("%w: 请选择电商店铺类型", ErrInvalidImport)
	}
	if _, err := s.repos.PlatformType.GetByID(in.PlatformTypeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 店铺类型不存在", ErrInvalidImport)
		}
		return nil, err
	}
	item.Name = name
	item.PlatformTypeID = in.PlatformTypeID
	item.Remark = strings.TrimSpace(in.Remark)
	if err := s.shops().SaveShop(item); err != nil {
		return nil, err
	}
	d := s.shopDTO(item)
	return &d, nil
}

func (s *DistributionService) DeleteShop(id uint64) error {
	if err := s.shops().DeleteShop(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *DistributionService) ListItems(shopID uint64, q dto.DistributionItemQuery) ([]dto.DistributionItemDTO, int64, error) {
	if _, err := s.GetShop(shopID); err != nil {
		return nil, 0, err
	}
	list, total, err := s.shops().ListItems(shopID, q)
	if err != nil {
		return nil, 0, err
	}
	codes := make([]string, 0, len(list))
	for i := range list {
		if code := strings.TrimSpace(list[i].ItemID); code != "" {
			codes = append(codes, code)
		}
	}
	collected, err := s.repos.Product.WithTenant(s.tenantID).IDsByMaterialCodes(codes)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.DistributionItemDTO, 0, len(list))
	for i := range list {
		d := itemDTO(&list[i])
		if pid, ok := collected[strings.TrimSpace(list[i].ItemID)]; ok && pid > 0 {
			d.Collected = true
			d.ProductID = pid
		}
		out = append(out, d)
	}
	return out, total, nil
}

const maxDistributionImport = 20 << 20

func (s *DistributionService) ImportItems(shopID uint64, file *multipart.FileHeader) (*dto.DistributionImportResult, error) {
	if _, err := s.GetShop(shopID); err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("%w: 请上传至尊宝导出的 Excel", ErrInvalidImport)
	}
	if file.Size > maxDistributionImport {
		return nil, fmt.Errorf("%w: 文件过大（最大 20MB）", ErrInvalidImport)
	}
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	data, err := io.ReadAll(io.LimitReader(src, maxDistributionImport+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxDistributionImport {
		return nil, fmt.Errorf("%w: 文件过大（最大 20MB）", ErrInvalidImport)
	}
	rows, err := zzbexcel.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidImport, err.Error())
	}
	result := &dto.DistributionImportResult{}
	store := s.shops()
	picCache := map[string]string{}
	keep := make([]string, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		if _, ok := seen[row.ItemID]; ok {
			continue
		}
		seen[row.ItemID] = struct{}{}
		keep = append(keep, row.ItemID)
		existing, findErr := store.FindItem(shopID, row.ItemID)
		if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return nil, findErr
		}
		hosted, picErr := s.hostShopPic(row.PicURL, picCache)
		if picErr != nil {
			result.ImageFailed++
			if existing != nil && strings.TrimSpace(existing.PicURL) != "" && !needsRehost(existing.PicURL) {
				hosted = existing.PicURL
			} else {
				hosted = ""
			}
		}
		row.PicURL = hosted
		if existing == nil {
			item := &model.DistributionShopItem{ShopID: shopID}
			applyRow(item, row)
			if err := store.CreateItem(item); err != nil {
				return nil, err
			}
			result.Created++
			continue
		}
		applyRow(existing, row)
		if err := store.SaveItem(existing); err != nil {
			return nil, err
		}
		result.Updated++
	}
	removed, err := store.DeleteItemsExcept(shopID, keep)
	if err != nil {
		return nil, err
	}
	result.Removed = int(removed)
	result.Total = result.Created + result.Updated
	return result, nil
}

func (s *DistributionService) hostShopPic(raw string, cache map[string]string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || !needsRehost(raw) {
		return raw, nil
	}
	if hosted, ok := cache[raw]; ok {
		if hosted == "" {
			return "", fmt.Errorf("图片下载失败")
		}
		return hosted, nil
	}
	if s.store == nil {
		cache[raw] = ""
		return "", fmt.Errorf("存储未配置")
	}
	hosted, err := storage.UploadFromURL(s.store, raw, storage.UploadOptions{
		Scope:    "common",
		Resource: "distribution",
	})
	if err != nil {
		cache[raw] = ""
		return "", err
	}
	cache[raw] = hosted
	return hosted, nil
}

func applyRow(item *model.DistributionShopItem, row zzbexcel.Row) {
	item.ItemID = row.ItemID
	item.Title = row.Title
	item.ItemURL = row.ItemURL
	item.PicURL = row.PicURL
	item.Price = row.Price
	item.Sales = row.Sales
	item.CommentCount = row.CommentCount
	item.MonthDeals = row.MonthDeals
	item.MonthConsign = row.MonthConsign
	item.ShipTime = row.ShipTime
	item.ListedAt = row.ListedAt
	item.Category = row.Category
	item.Tags = row.Tags
	item.SourceShopName = row.SourceShopName
	item.SourceShopURL = row.SourceShopURL
}

func (s *DistributionService) shopDTO(item *model.DistributionShop) dto.DistributionShopDTO {
	out := dto.DistributionShopDTO{
		ID:             item.ID,
		PlatformTypeID: item.PlatformTypeID,
		Name:           item.Name,
		Remark:         item.Remark,
		CreateTime:     item.CreatedAt.Format(time.RFC3339),
	}
	if pt, err := s.repos.PlatformType.GetByID(item.PlatformTypeID); err == nil && pt != nil {
		out.PlatformTypeName = pt.Name
		out.PlatformTypeLogo = pt.Logo
	}
	return out
}

func itemDTO(item *model.DistributionShopItem) dto.DistributionItemDTO {
	return dto.DistributionItemDTO{
		ID:             item.ID,
		ShopID:         item.ShopID,
		ItemID:         item.ItemID,
		Title:          item.Title,
		ItemURL:        item.ItemURL,
		PicURL:         item.PicURL,
		Price:          item.Price,
		Sales:          item.Sales,
		CommentCount:   item.CommentCount,
		MonthDeals:     item.MonthDeals,
		MonthConsign:   item.MonthConsign,
		ShipTime:       item.ShipTime,
		ListedAt:       item.ListedAt,
		Category:       item.Category,
		Tags:           item.Tags,
		SourceShopName: item.SourceShopName,
		SourceShopURL:  item.SourceShopURL,
	}
}
