package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"productcore/internal/cache"
	"productcore/internal/dto"
	"productcore/internal/event"
	"productcore/internal/model"
	"productcore/internal/pkg/util"
	"productcore/internal/repo"
	"productcore/internal/storage"

	"gorm.io/gorm"
)

type ProductService struct {
	repo     *repo.ProductRepo
	meta     *repo.Repos
	cache    *cache.ProductCache
	events   *event.Publisher
	store    storage.Storage
	tenantID uint64
}

func NewProductService(repos *repo.Repos, pc *cache.ProductCache, pub *event.Publisher, store storage.Storage) *ProductService {
	return &ProductService{repo: repos.Product, meta: repos, cache: pc, events: pub, store: store, tenantID: 1}
}

func (s *ProductService) ForTenant(tenantID uint64) *ProductService {
	cp := *s
	cp.tenantID = repo.NormalizeTenantID(tenantID)
	cp.repo = s.repo.WithTenant(tenantID)
	return &cp
}

func (s *ProductService) withRepo(tx *repo.ProductRepo) *ProductService {
	return &ProductService{
		repo: tx, meta: s.meta, cache: s.cache, events: s.events,
		store: s.store, tenantID: s.tenantID,
	}
}

func (s *ProductService) List(q dto.ProductQuery) ([]dto.ProductDTO, int64, error) {
	if q.CategoryID > 0 && len(q.CategoryIDs) == 0 {
		ids, err := s.meta.Category.ListSelfAndDescendantIDs(s.tenantID, q.CategoryID)
		if err != nil {
			return nil, 0, err
		}
		if len(ids) > 0 {
			q.CategoryIDs = ids
			q.CategoryID = 0
		}
	}
	if q.GroupID > 0 && len(q.GroupIDs) == 0 {
		ids, err := s.meta.Group.ListSelfAndDescendantIDs(s.tenantID, q.GroupID)
		if err != nil {
			return nil, 0, err
		}
		if len(ids) > 0 {
			q.GroupIDs = ids
			q.GroupID = 0
		}
	}
	products, total, err := s.repo.List(q)
	if err != nil {
		return nil, 0, err
	}
	list := make([]dto.ProductDTO, 0, len(products))
	ids := make([]uint64, 0, len(products))
	for _, p := range products {
		ids = append(ids, p.ID)
	}
	flags, _ := s.repo.EditDraftFlags(ids)
	for _, p := range products {
		item, err := s.toDTO(&p, false)
		if err != nil {
			return nil, 0, err
		}
		if flags[p.ID] {
			item.HasEditDraft = true
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
	if p.IsDraft == 0 {
		if item, err := s.loadEditDraftDTO(s.repo, id, p); err == nil && item.HasEditDraft {
			return item, nil
		}
	}
	ctx := context.Background()
	if s.cache != nil {
		if item, err := s.cache.Get(ctx, id); err == nil {
			s.applyPublicURLs(item)
			if p.IsDraft == 0 {
				if flags, _ := s.repo.EditDraftFlags([]uint64{id}); flags[id] {
					item.HasEditDraft = true
				}
			}
			return item, nil
		}
	}
	item, err := s.toDTO(p, true)
	if err != nil {
		return nil, err
	}
	if p.IsDraft == 0 {
		if flags, _ := s.repo.EditDraftFlags([]uint64{id}); flags[id] {
			item.HasEditDraft = true
		}
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, item)
	}
	return item, nil
}

func (s *ProductService) Create(in *dto.ProductDTO) (*dto.ProductDTO, error) {
	if err := s.ensureUniqueMaterialCode(in.MaterialCode, 0); err != nil {
		return nil, err
	}
	if err := s.ensureUniqueSN(in.ProductSn, 0); err != nil {
		return nil, err
	}
	if err := s.validateSkuSpecs(in); err != nil {
		return nil, err
	}
	var out *dto.ProductDTO
	err := s.repo.Transaction(func(tx *repo.ProductRepo) error {
		svc := s.withRepo(tx)
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
		if err := svc.saveKeywords(tx, p.ID, in.KeywordIDs); err != nil {
			return err
		}
		if err := svc.syncSummary(tx, p.ID); err != nil {
			return err
		}
		out, err = svc.loadDTO(tx, p.ID)
		return err
	})
	if err == nil && out != nil {
		s.afterProductChange("created", out.ID, out.PublishStatus)
	}
	return out, err
}

func (s *ProductService) Update(id uint64, in *dto.ProductDTO) (*dto.ProductDTO, error) {
	if err := s.ensureUniqueMaterialCode(in.MaterialCode, id); err != nil {
		return nil, err
	}
	if err := s.ensureUniqueSN(in.ProductSn, id); err != nil {
		return nil, err
	}
	var out *dto.ProductDTO
	var notifyChange bool
	err := s.repo.Transaction(func(tx *repo.ProductRepo) error {
		svc := s.withRepo(tx)
		p, err := tx.GetByID(id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		if !in.Finalize {
			if p.IsDraft == 1 {
				notifyChange = true
				out, err = svc.updateDraftBoxAutoSave(tx, id, in)
				return err
			}
			if err := svc.saveEditDraft(tx, id, in); err != nil {
				return err
			}
			out, err = svc.loadEditDraftDTO(tx, id, p)
			return err
		}

		if err := s.validateSkuSpecs(in); err != nil {
			return err
		}
		isDraft := int8(0)
		if p.IsDraft == 1 {
			isDraft = 0
		}
		notifyChange = true
		out, err = svc.applyProductUpdate(tx, id, in, isDraft)
		if err != nil {
			return err
		}
		_ = tx.DeleteEditDraft(id)
		return nil
	})
	if err == nil && out != nil && notifyChange {
		s.afterProductChange("updated", out.ID, out.PublishStatus)
	}
	return out, err
}

func (s *ProductService) Delete(id uint64) error {
	if _, err := s.repo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	_ = s.repo.DeleteEditDraft(id)
	s.afterProductChange("deleted", id, 0)
	return nil
}

func (s *ProductService) ListTrash(q dto.ProductQuery) ([]dto.ProductDTO, int64, error) {
	products, total, err := s.repo.ListTrashed(q)
	if err != nil {
		return nil, 0, err
	}
	list := make([]dto.ProductDTO, 0, len(products))
	for _, p := range products {
		item, err := s.toDTO(&p, false)
		if err != nil {
			return nil, 0, err
		}
		if p.DeletedAt.Valid {
			item.DeleteTime = util.FormatTime(p.DeletedAt.Time)
		}
		list = append(list, *item)
	}
	return list, total, nil
}

func (s *ProductService) ListDrafts(q dto.ProductQuery) ([]dto.ProductDTO, int64, error) {
	entries, total, err := s.repo.ListDraftEntries(q)
	if err != nil {
		return nil, 0, err
	}
	list := make([]dto.ProductDTO, 0, len(entries))
	for _, e := range entries {
		var item *dto.ProductDTO
		var err error
		if e.IsEditOnly {
			item, err = s.loadEditDraftDTO(s.repo, e.Product.ID, &e.Product)
		} else {
			item, err = s.toDTO(&e.Product, false)
		}
		if err != nil {
			return nil, 0, err
		}
		item.DraftSavedAt = util.FormatTime(e.DraftSavedAt)
		if e.IsEditOnly {
			item.HasEditDraft = true
			item.IsDraft = 0
		}
		list = append(list, *item)
	}
	return list, total, nil
}

func (s *ProductService) DiscardEditDraft(id uint64) error {
	p, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if p.IsDraft == 1 {
		return fmt.Errorf("%w: 请使用删除草稿", ErrNoEditDraft)
	}
	if _, err := s.repo.GetEditDraft(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNoEditDraft
		}
		return err
	}
	return s.repo.DeleteEditDraft(id)
}

func (s *ProductService) Restore(id uint64) error {
	p, err := s.repo.GetByIDUnscoped(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if !p.DeletedAt.Valid {
		return ErrNotInTrash
	}
	if err := s.ensureUniqueMaterialCode(p.MaterialCode, p.ID); err != nil {
		return err
	}
	if err := s.ensureUniqueSN(p.ProductSn, p.ID); err != nil {
		return err
	}
	if err := s.repo.RestoreProduct(id); err != nil {
		return err
	}
	s.afterProductChange("restored", id, p.PublishStatus)
	return nil
}

func (s *ProductService) ForceDelete(id uint64) error {
	p, err := s.repo.GetByIDUnscoped(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if !p.DeletedAt.Valid {
		return ErrNotInTrash
	}
	err = s.repo.Transaction(func(tx *repo.ProductRepo) error {
		if err := tx.DeleteSkusByProduct(id); err != nil {
			return err
		}
		if err := tx.DeleteGroupRelations(id); err != nil {
			return err
		}
		if err := tx.DeleteKeywordRelations(id); err != nil {
			return err
		}
		if err := s.meta.PlatformListing.DeleteByProduct(id); err != nil {
			return err
		}
		if err := tx.DeleteEditDraft(id); err != nil {
			return err
		}
		return tx.ForceDeleteProduct(id)
	})
	if err == nil {
		s.afterProductChange("force_deleted", id, 0)
	}
	return err
}

func (s *ProductService) BatchDelete(ids []uint64) (success, failed int) {
	for _, id := range ids {
		if err := s.Delete(id); err != nil {
			failed++
		} else {
			success++
		}
	}
	return success, failed
}

func (s *ProductService) BatchRestore(ids []uint64) (success, failed int) {
	for _, id := range ids {
		if err := s.Restore(id); err != nil {
			failed++
		} else {
			success++
		}
	}
	return success, failed
}

func (s *ProductService) BatchForceDelete(ids []uint64) (success, failed int) {
	for _, id := range ids {
		if err := s.ForceDelete(id); err != nil {
			failed++
		} else {
			success++
		}
	}
	return success, failed
}

func fillMissingSkuCodes(skus []dto.SkuDTO) []dto.SkuDTO {
	if len(skus) == 0 {
		return skus
	}
	used := make(map[string]struct{}, len(skus))
	out := make([]dto.SkuDTO, len(skus))
	for i, item := range skus {
		code := util.SanitizeSkuCode(item.SkuCode)
		if code != "" {
			used[code] = struct{}{}
		}
		out[i] = item
	}
	for i := range out {
		code := util.SanitizeSkuCode(out[i].SkuCode)
		if code == "" {
			code = util.NextRandomSkuCode(used)
		}
		out[i].SkuCode = code
	}
	return out
}

func (s *ProductService) UpdatePublishStatus(id uint64, status int8) error {
	if err := s.repo.UpdatePublishStatus(id, status); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	s.afterProductChange("publish_status_changed", id, status)
	return nil
}

func (s *ProductService) GetSkus(id uint64) (*dto.ProductSkusDTO, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.buildProductSkusDTO(p)
}

func (s *ProductService) buildProductSkusDTO(p *model.Product) (*dto.ProductSkusDTO, error) {
	skus, err := s.repo.ListSkus(p.ID)
	if err != nil {
		return nil, err
	}
	out := &dto.ProductSkusDTO{
		ID:            p.ID,
		Name:          p.Name,
		MaterialCode:  p.MaterialCode,
		Price:         p.Price,
		OriginalPrice: p.OriginalPrice,
		Stock:         p.Stock,
		SkuSpecs:      dto.NormalizeSkuSpecs(p.SkuSpecsJSON),
		Skus:          make([]dto.SkuDTO, 0, len(skus)),
	}
	skuCount, _ := s.repo.CountSkus(p.ID)
	out.SkuCount = int(skuCount)
	for _, sku := range skus {
		specsMap := map[string]string{}
		_ = json.Unmarshal([]byte(sku.SpecData), &specsMap)
		item := dto.SkuDTO{
			ID: sku.ID, SkuCode: sku.SkuCode, Specs: specsMap,
			Price: sku.Price, CostPrice: sku.CostPrice,
			MarketPrice: sku.MarketPrice, Stock: sku.Stock,
			Weight: sku.Weight, Pic: sku.Pic,
		}
		if s.store != nil && item.Pic != "" {
			item.Pic = s.store.ResolvePublicURL(item.Pic)
		}
		out.Skus = append(out.Skus, item)
	}
	for i := range out.SkuSpecs {
		for j := range out.SkuSpecs[i].Values {
			if out.SkuSpecs[i].Values[j].Pic != "" && s.store != nil {
				out.SkuSpecs[i].Values[j].Pic = s.store.ResolvePublicURL(out.SkuSpecs[i].Values[j].Pic)
			}
		}
	}
	return out, nil
}

func syncSkuSpecPicsFromSkus(specs []dto.SkuSpecDTO, skus []dto.SkuDTO) []dto.SkuSpecDTO {
	if len(specs) == 0 || len(skus) == 0 {
		return specs
	}
	picByValue := map[string]string{}
	primarySpec := strings.TrimSpace(specs[0].Name)
	for _, sku := range skus {
		pic := strings.TrimSpace(sku.Pic)
		if pic == "" {
			continue
		}
		val := ""
		if primarySpec != "" {
			val = strings.TrimSpace(sku.Specs[primarySpec])
		}
		if val == "" {
			for _, v := range sku.Specs {
				if strings.TrimSpace(v) != "" {
					val = strings.TrimSpace(v)
					break
				}
			}
		}
		if val != "" {
			picByValue[val] = pic
		}
	}
	if len(picByValue) == 0 {
		return specs
	}
	out := make([]dto.SkuSpecDTO, len(specs))
	for i, spec := range specs {
		out[i] = spec
		out[i].Values = make([]dto.SkuSpecValueDTO, len(spec.Values))
		for j, val := range spec.Values {
			out[i].Values[j] = val
			if pic, ok := picByValue[strings.TrimSpace(val.Value)]; ok {
				out[i].Values[j].Pic = pic
			}
		}
	}
	return out
}

func (s *ProductService) UpdateSkus(id uint64, skus []dto.SkuDTO, skuSpecsIn []dto.SkuSpecDTO) (*dto.ProductSkusDTO, error) {
	var out *dto.ProductSkusDTO
	var publishStatus int8
	err := s.repo.Transaction(func(tx *repo.ProductRepo) error {
		svc := s.withRepo(tx)
		p, err := tx.GetByID(id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		publishStatus = p.PublishStatus
		if err := tx.DeleteSkusByProduct(id); err != nil {
			return err
		}
		if err := svc.saveSkus(tx, id, skus); err != nil {
			return err
		}
		var specs []dto.SkuSpecDTO
		if len(skuSpecsIn) > 0 {
			specs = dto.CompactSkuSpecs(skuSpecsIn)
		} else {
			specs = syncSkuSpecPicsFromSkus(dto.NormalizeSkuSpecs(p.SkuSpecsJSON), skus)
		}
		if err := tx.UpdateFields(id, map[string]interface{}{
			"sku_specs_json": util.ToJSON(specs),
		}); err != nil {
			return err
		}
		if err := svc.syncSummary(tx, id); err != nil {
			return err
		}
		p, err = tx.GetByID(id)
		if err != nil {
			return err
		}
		out, err = svc.buildProductSkusDTO(p)
		return err
	})
	if err == nil && out != nil {
		s.afterProductChange("updated", out.ID, publishStatus)
	}
	return out, err
}

func (s *ProductService) afterProductChange(action string, id uint64, publishStatus int8) {
	ctx := context.Background()
	if s.cache != nil {
		_ = s.cache.Delete(ctx, id)
	}
	if s.events != nil {
		s.events.ProductChanged(ctx, action, id, map[string]interface{}{
			"publish_status": publishStatus,
		})
	}
}

func (s *ProductService) ensureUniqueSN(sn string, excludeID uint64) error {
	sn = strings.TrimSpace(sn)
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

func (s *ProductService) ensureUniqueMaterialCode(code string, excludeID uint64) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil
	}
	count, err := s.repo.CountByMaterialCode(code, excludeID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrDuplicateMaterialCode
	}
	return nil
}

// IDByMaterialCode 导入时解析已有商品 ID（用于资源路径）
func (s *ProductService) IDByMaterialCode(code string) (uint64, bool) {
	p, err := s.repo.GetByMaterialCodeUnscoped(strings.TrimSpace(code))
	if err != nil {
		return 0, false
	}
	return p.ID, true
}

// CreateDraft 创建草稿商品（分配 ID，供新建页面上传资源，仅出现在草稿箱）
func (s *ProductService) CreateDraft() (*dto.ProductDTO, error) {
	p := &model.Product{
		TenantID:       s.tenantID,
		Name:           "未命名商品",
		Unit:           "件",
		IsDraft:        1,
		PublishStatus:  0,
		VerifyStatus:   1,
		ChannelVisible: "both",
	}
	if err := s.repo.CreateDraft(p); err != nil {
		return nil, err
	}
	return s.toDTO(p, false)
}

// EnsureForImport 导入前确保商品存在并返回 ID（新建或恢复回收站）
func (s *ProductService) EnsureForImport(code, name string, brandID, categoryID uint64) (uint64, error) {
	code = strings.TrimSpace(code)
	existing, err := s.repo.GetByMaterialCodeUnscoped(code)
	if err == nil {
		if existing.DeletedAt.Valid {
			if err := s.repo.RestoreProduct(existing.ID); err != nil {
				return 0, err
			}
		}
		return existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	out, err := s.Create(&dto.ProductDTO{
		Name:          strings.TrimSpace(name),
		MaterialCode:  code,
		BrandID:       brandID,
		CategoryID:    categoryID,
		Unit:          "件",
		PublishStatus: 0,
		VerifyStatus:  1,
	})
	if err != nil {
		return 0, err
	}
	return out.ID, nil
}

// ImportUpsert 按资料编码创建或覆盖商品（导入专用，资料编码必填）
func (s *ProductService) ImportUpsert(in *dto.ProductDTO) (*dto.ProductDTO, error) {
	if strings.TrimSpace(in.MaterialCode) == "" {
		return nil, ErrMaterialCodeRequired
	}
	code := strings.TrimSpace(in.MaterialCode)
	existing, err := s.repo.GetByMaterialCodeUnscoped(code)
	if err == nil {
		if existing.DeletedAt.Valid {
			if err := s.repo.RestoreProduct(existing.ID); err != nil {
				return nil, err
			}
		}
		in.Finalize = true
		return s.Update(existing.ID, in)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.Create(in)
	}
	return nil, err
}

func (s *ProductService) validateSkuSpecs(in *dto.ProductDTO) error {
	if len(in.Skus) == 0 {
		return nil
	}
	if err := dto.ValidateSkuSpecsNoDuplicateValues(in.SkuSpecs); err != nil {
		return fmt.Errorf("%w: %s", ErrDuplicateSpecValue, err)
	}
	return nil
}

func specComboKey(specs map[string]string) string {
	if len(specs) == 0 {
		return ""
	}
	keys := make([]string, 0, len(specs))
	for k, v := range specs {
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k == "" || v == "" {
			continue
		}
		keys = append(keys, k)
	}
	if len(keys) == 0 {
		return ""
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+strings.TrimSpace(specs[k]))
	}
	return strings.Join(parts, "|")
}

func validateSkusBatch(skus []dto.SkuDTO) error {
	seenCodes := make(map[string]struct{}, len(skus))
	seenCombos := make(map[string]struct{}, len(skus))
	for _, item := range skus {
		code := util.SanitizeSkuCode(item.SkuCode)
		if code == "" {
			return ErrSkuCodeRequired
		}
		if !util.IsValidSkuCode(code) {
			return ErrInvalidSkuCode
		}
		if _, dup := seenCodes[code]; dup {
			return fmt.Errorf("%w: %s", ErrDuplicateSku, code)
		}
		seenCodes[code] = struct{}{}
		combo := specComboKey(item.Specs)
		if combo != "" {
			if _, dup := seenCombos[combo]; dup {
				return ErrDuplicateSpecCombo
			}
			seenCombos[combo] = struct{}{}
		}
	}
	return nil
}

func (s *ProductService) saveSkus(tx *repo.ProductRepo, productID uint64, skus []dto.SkuDTO) error {
	if err := validateSkusBatch(skus); err != nil {
		return err
	}
	for i, item := range skus {
		code := util.SanitizeSkuCode(item.SkuCode)

		n, err := tx.CountSkuByCode(code, productID)
		if err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("%w: %s", ErrDuplicateSku, code)
		}
		sku := &model.Sku{
			TenantID:  s.tenantID,
			ProductID: productID,
			SkuCode:   code,
			SpecData:  util.ToJSON(item.Specs),
			SortOrder: i,
			Price:     item.Price,
			CostPrice: item.CostPrice,
			MarketPrice: item.MarketPrice,
			Stock:     item.Stock,
			Weight:    item.Weight,
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

func (s *ProductService) saveKeywords(tx *repo.ProductRepo, productID uint64, keywordIDs []uint64) error {
	for _, kid := range keywordIDs {
		if kid == 0 {
			continue
		}
		if err := tx.CreateKeywordRelation(productID, kid); err != nil {
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
	maxMarket := 0.0
	weightAtMinPrice := 0.0
	for _, sku := range skus {
		totalStock += sku.Stock
		if sku.Price > 0 && (minPrice == 0 || sku.Price < minPrice) {
			minPrice = sku.Price
			weightAtMinPrice = sku.Weight
		} else if sku.Price > 0 && sku.Price == minPrice && weightAtMinPrice == 0 && sku.Weight > 0 {
			weightAtMinPrice = sku.Weight
		}
		if sku.MarketPrice > maxMarket {
			maxMarket = sku.MarketPrice
		}
	}
	updates := map[string]interface{}{"stock": totalStock}
	if minPrice > 0 {
		updates["price"] = minPrice
	}
	if maxMarket > 0 {
		updates["original_price"] = maxMarket
	}
	if weightAtMinPrice > 0 {
		updates["weight"] = weightAtMinPrice
	}
	return tx.UpdateFields(productID, updates)
}

func (s *ProductService) fromDTO(in *dto.ProductDTO) (*model.Product, error) {
	channel := in.ChannelVisible
	if channel == "" {
		channel = "both"
	}
	productVideo := in.ProductVideo
	media := dto.ProductMediaWithoutPics34(in.Media)
	if in.Media != nil && in.Media.Videos != nil && in.Media.Videos.Ratio11 != "" {
		productVideo = in.Media.Videos.Ratio11
	}
	mediaJSON := ""
	if media != nil {
		mediaJSON = util.ToJSON(media)
	}
	pics34 := dto.Pics34FromMedia(in.Media)
	skuSpecsJSON := "[]"
	if len(in.Skus) > 0 {
		skuSpecsJSON = util.ToJSON(dto.CompactSkuSpecs(in.SkuSpecs))
	}
	return &model.Product{
		TenantID: s.tenantID,
		Name: in.Name, SubTitle: in.SubTitle,
		MaterialCode: in.MaterialCode, Source: in.Source, ProductSn: in.ProductSn,
		BrandID: in.BrandID, CategoryID: in.CategoryID, Pic: in.Pic,
		AlbumPics: util.ToJSON(in.AlbumPics), Pics34JSON: util.ToJSON(pics34),
		ProductVideo: productVideo,
		MediaJSON: mediaJSON,
		Price: in.Price, OriginalPrice: in.OriginalPrice, Stock: in.Stock,
		Unit: in.Unit, Weight: in.Weight, PublishStatus: in.PublishStatus, IsDraft: in.IsDraft,
		VerifyStatus: in.VerifyStatus, Sort: in.Sort, Description: in.Description,
		DetailHTML: in.DetailHTML, SkuSpecsJSON: skuSpecsJSON,
		ChannelVisible: channel,
	}, nil
}

func (s *ProductService) loadDTO(tx *repo.ProductRepo, id uint64) (*dto.ProductDTO, error) {
	p, err := tx.GetByID(id)
	if err != nil {
		return nil, err
	}
	svc := s.withRepo(tx)
	return svc.toDTO(p, true)
}

func (s *ProductService) toDTO(p *model.Product, withSkus bool) (*dto.ProductDTO, error) {
	out := &dto.ProductDTO{
		ID: p.ID, Name: p.Name, SubTitle: p.SubTitle,
		MaterialCode: p.MaterialCode, Source: p.Source, ProductSn: p.ProductSn,
		BrandID: p.BrandID, CategoryID: p.CategoryID, Pic: p.Pic,
		AlbumPics: util.ParseStringArray(p.AlbumPics), ProductVideo: p.ProductVideo,
		Price: p.Price, OriginalPrice: p.OriginalPrice, Stock: p.Stock,
		Unit: p.Unit, Weight: p.Weight, PublishStatus: p.PublishStatus, IsDraft: p.IsDraft,
		VerifyStatus: p.VerifyStatus, Sort: p.Sort, Sale: p.Sale,
		Description: p.Description, DetailHTML: p.DetailHTML,
		ChannelVisible: p.ChannelVisible,
		CreateTime: util.FormatTime(p.CreatedAt), UpdateTime: util.FormatTime(p.UpdatedAt),
	}
	pics34 := util.ParseStringArray(p.Pics34JSON)
	media := dto.ParseProductMedia(p.MediaJSON)
	if len(pics34) == 0 && media != nil && len(media.Pics34) > 0 {
		pics34 = media.Pics34
	}
	out.Media = dto.MergePics34IntoMedia(media, pics34)
	out.SkuSpecs = dto.NormalizeSkuSpecs(p.SkuSpecsJSON)
	if p.BrandID > 0 {
		out.BrandName = s.meta.Brand.GetName(p.BrandID)
	} else {
		out.BrandName = "无品牌"
	}
	if p.CategoryID > 0 {
		out.CategoryName = s.meta.Category.GetName(p.CategoryID)
	} else {
		out.CategoryName = "无分类"
	}
	groupIDs, _ := s.repo.ListGroupIDs(p.ID)
	out.GroupIDs = groupIDs
	keywordIDs, _ := s.repo.ListKeywordIDs(p.ID)
	out.KeywordIDs = keywordIDs
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
				Price: sku.Price, CostPrice: sku.CostPrice,
				MarketPrice: sku.MarketPrice, Stock: sku.Stock,
				Weight: sku.Weight, Pic: sku.Pic,
			})
		}
	}
	s.applyPublicURLs(out)
	return out, nil
}

func (s *ProductService) applyPublicURLs(out *dto.ProductDTO) {
	if s.store == nil || out == nil {
		return
	}
	rewriteProductURLs(out, s.store.ResolvePublicURL)
}

func nullableUint64(v uint64) interface{} {
	if v == 0 {
		return nil
	}
	return v
}
