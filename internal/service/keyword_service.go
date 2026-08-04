package service

import (
	"errors"
	"fmt"
	"strings"

	"productcore/internal/dto"
	"productcore/internal/model"
	"productcore/internal/pkg/util"
	"productcore/internal/repo"

	"gorm.io/gorm"
)

type ProductKeywordService struct {
	repos    *repo.Repos
	product  *ProductService
	tenantID uint64
}

func NewProductKeywordService(repos *repo.Repos, productSvc *ProductService) *ProductKeywordService {
	return &ProductKeywordService{repos: repos, product: productSvc, tenantID: 1}
}

func (s *ProductKeywordService) ForTenant(tenantID uint64) *ProductKeywordService {
	cp := *s
	cp.tenantID = repo.NormalizeTenantID(tenantID)
	cp.product = s.product.ForTenant(tenantID)
	return &cp
}

func (s *ProductKeywordService) List() ([]dto.ProductKeywordDTO, error) {
	list, err := s.repos.Keyword.List(s.tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ProductKeywordDTO, 0, len(list))
	for _, k := range list {
		out = append(out, s.toDTO(&k))
	}
	return out, nil
}

func (s *ProductKeywordService) Create(in *dto.ProductKeywordDTO) (*dto.ProductKeywordDTO, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: 关键词名称不能为空", ErrInvalidImport)
	}
	if existing, err := s.repos.Keyword.GetByName(s.tenantID, name); err == nil && existing != nil {
		return nil, fmt.Errorf("%w: 关键词已存在", ErrDuplicateSN)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	k := model.ProductKeyword{
		TenantID:    s.tenantID,
		Name:        name,
		Description: strings.TrimSpace(in.Description),
		Sort:        in.Sort,
	}
	if err := s.repos.Keyword.Create(&k); err != nil {
		return nil, err
	}
	item := s.toDTO(&k)
	return &item, nil
}

func (s *ProductKeywordService) Update(id uint64, in *dto.ProductKeywordDTO) (*dto.ProductKeywordDTO, error) {
	k, err := s.repos.Keyword.GetByID(s.tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: 关键词名称不能为空", ErrInvalidImport)
	}
	if existing, err := s.repos.Keyword.GetByName(s.tenantID, name); err == nil && existing != nil && existing.ID != id {
		return nil, fmt.Errorf("%w: 关键词已存在", ErrDuplicateSN)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	k.Name = name
	k.Description = strings.TrimSpace(in.Description)
	k.Sort = in.Sort
	if err := s.repos.Keyword.Save(k); err != nil {
		return nil, err
	}
	item := s.toDTO(k)
	return &item, nil
}

func (s *ProductKeywordService) Delete(id uint64) error {
	if err := s.repos.Keyword.Delete(s.tenantID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *ProductKeywordService) ListProducts(keywordID uint64) ([]dto.ProductDTO, error) {
	if _, err := s.repos.Keyword.GetByID(s.tenantID, keywordID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	ids, err := s.repos.Keyword.ListProductIDs(s.tenantID, keywordID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ProductDTO, 0, len(ids))
	for _, id := range ids {
		item, err := s.product.Get(id)
		if err != nil {
			continue
		}
		out = append(out, *item)
	}
	return out, nil
}

func (s *ProductKeywordService) SetProducts(keywordID uint64, productIDs []uint64) error {
	if _, err := s.repos.Keyword.GetByID(s.tenantID, keywordID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return s.repos.Keyword.ReplaceProducts(keywordID, productIDs)
}

func (s *ProductKeywordService) AddProducts(keywordID uint64, productIDs []uint64) error {
	if _, err := s.repos.Keyword.GetByID(s.tenantID, keywordID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return s.repos.Keyword.AddProducts(keywordID, productIDs)
}

func (s *ProductKeywordService) RemoveProduct(keywordID, productID uint64) error {
	if _, err := s.repos.Keyword.GetByID(s.tenantID, keywordID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return s.repos.Keyword.RemoveProduct(keywordID, productID)
}

func (s *ProductKeywordService) toDTO(k *model.ProductKeyword) dto.ProductKeywordDTO {
	count, _ := s.repos.Keyword.CountProducts(s.tenantID, k.ID)
	return dto.ProductKeywordDTO{
		ID: k.ID, Name: k.Name, Description: k.Description, Sort: k.Sort,
		ProductCount: count, CreateTime: util.FormatTime(k.CreatedAt),
	}
}
