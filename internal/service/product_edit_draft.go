package service

import (
	"encoding/json"
	"errors"

	"productcore/internal/dto"
	"productcore/internal/model"
	"productcore/internal/repo"

	"gorm.io/gorm"
)

func encodeEditDraftPayload(in *dto.ProductDTO, productID uint64) (string, error) {
	copy := *in
	copy.ID = productID
	copy.Finalize = false
	copy.HasEditDraft = false
	b, err := json.Marshal(&copy)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func decodeEditDraftPayload(raw string) (*dto.ProductDTO, error) {
	var out dto.ProductDTO
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ProductService) enrichProductMeta(out *dto.ProductDTO) {
	if out.BrandID > 0 {
		out.BrandName = s.meta.Brand.GetName(out.BrandID)
	} else {
		out.BrandName = "无品牌"
	}
	if out.CategoryID > 0 {
		out.CategoryName = s.meta.Category.GetName(out.CategoryID)
	} else {
		out.CategoryName = "无分类"
	}
}

func (s *ProductService) saveEditDraft(tx *repo.ProductRepo, productID uint64, in *dto.ProductDTO) error {
	payload, err := encodeEditDraftPayload(in, productID)
	if err != nil {
		return err
	}
	return tx.UpsertEditDraft(productID, payload)
}

func (s *ProductService) loadEditDraftDTO(tx *repo.ProductRepo, productID uint64, p *model.Product) (*dto.ProductDTO, error) {
	d, err := tx.GetEditDraft(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			svc := s.withRepo(tx)
			return svc.toDTO(p, true)
		}
		return nil, err
	}
	out, err := decodeEditDraftPayload(d.PayloadJSON)
	if err != nil {
		return nil, err
	}
	out.ID = productID
	out.HasEditDraft = true
	out.IsDraft = p.IsDraft
	out.SkuCount = len(out.Skus)
	s.enrichProductMeta(out)
	s.applyPublicURLs(out)
	return out, nil
}

func (s *ProductService) applyProductUpdate(tx *repo.ProductRepo, id uint64, in *dto.ProductDTO, isDraft int8) (*dto.ProductDTO, error) {
	svc := s.withRepo(tx)
	updated, err := svc.fromDTO(in)
	if err != nil {
		return nil, err
	}
	if err := tx.UpdateFields(id, map[string]interface{}{
		"name": updated.Name, "sub_title": updated.SubTitle,
		"material_code": updated.MaterialCode, "source": updated.Source, "product_sn": updated.ProductSn,
		"brand_id": nullableUint64(updated.BrandID), "category_id": nullableUint64(updated.CategoryID), "pic": updated.Pic,
		"album_pics": updated.AlbumPics, "pics34_json": updated.Pics34JSON,
		"product_video": updated.ProductVideo,
		"media_json": updated.MediaJSON,
		"price": updated.Price, "original_price": updated.OriginalPrice, "stock": updated.Stock,
		"unit": updated.Unit, "weight": updated.Weight, "publish_status": updated.PublishStatus,
		"verify_status": updated.VerifyStatus, "sort": updated.Sort, "is_draft": isDraft,
		"description": updated.Description, "detail_html": updated.DetailHTML,
		"sku_specs_json": updated.SkuSpecsJSON, "channel_visible": updated.ChannelVisible,
	}); err != nil {
		return nil, err
	}
	if err := tx.DeleteSkusByProduct(id); err != nil {
		return nil, err
	}
	if err := svc.saveSkus(tx, id, in.Skus); err != nil {
		return nil, err
	}
	if err := tx.DeleteGroupRelations(id); err != nil {
		return nil, err
	}
	if err := svc.saveGroups(tx, id, in.GroupIDs); err != nil {
		return nil, err
	}
	if err := svc.syncSummary(tx, id); err != nil {
		return nil, err
	}
	return svc.loadDTO(tx, id)
}

func (s *ProductService) updateDraftBoxAutoSave(tx *repo.ProductRepo, id uint64, in *dto.ProductDTO) (*dto.ProductDTO, error) {
	svc := s.withRepo(tx)
	updated, err := svc.fromDTO(in)
	if err != nil {
		return nil, err
	}
	if err := tx.UpdateFields(id, map[string]interface{}{
		"name": updated.Name, "sub_title": updated.SubTitle,
		"material_code": updated.MaterialCode, "source": updated.Source, "product_sn": updated.ProductSn,
		"brand_id": nullableUint64(updated.BrandID), "category_id": nullableUint64(updated.CategoryID), "pic": updated.Pic,
		"album_pics": updated.AlbumPics, "pics34_json": updated.Pics34JSON,
		"product_video": updated.ProductVideo,
		"media_json": updated.MediaJSON,
		"price": updated.Price, "original_price": updated.OriginalPrice, "stock": updated.Stock,
		"unit": updated.Unit, "weight": updated.Weight, "publish_status": updated.PublishStatus,
		"verify_status": updated.VerifyStatus, "sort": updated.Sort, "is_draft": int8(1),
		"description": updated.Description, "detail_html": updated.DetailHTML,
		"sku_specs_json": updated.SkuSpecsJSON, "channel_visible": updated.ChannelVisible,
	}); err != nil {
		return nil, err
	}
	skus := fillMissingSkuCodes(in.Skus)
	if len(skus) > 0 {
		if err := tx.DeleteSkusByProduct(id); err != nil {
			return nil, err
		}
		if err := svc.saveSkus(tx, id, skus); err != nil {
			return nil, err
		}
		if err := svc.syncSummary(tx, id); err != nil {
			return nil, err
		}
	}
	if err := tx.DeleteGroupRelations(id); err != nil {
		return nil, err
	}
	if err := svc.saveGroups(tx, id, in.GroupIDs); err != nil {
		return nil, err
	}
	return svc.loadDTO(tx, id)
}
