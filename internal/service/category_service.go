package service

import (
	"errors"

	"productcore/internal/dto"
	"productcore/internal/model"
	"productcore/internal/repo"

	"gorm.io/gorm"
)

type CategoryService struct {
	repo *repo.CategoryRepo
	tenantID uint64
}

func NewCategoryService(repos *repo.Repos) *CategoryService {
	return &CategoryService{repo: repos.Category, tenantID: 1}
}

func (s *CategoryService) ForTenant(tenantID uint64) *CategoryService {
	cp := *s
	cp.tenantID = repo.NormalizeTenantID(tenantID)
	return &cp
}

func (s *CategoryService) Tree() ([]dto.CategoryDTO, error) {
	cats, err := s.repo.ListAll(s.tenantID)
	if err != nil {
		return nil, err
	}
	flat := make([]dto.CategoryDTO, 0, len(cats))
	for _, c := range cats {
		flat = append(flat, s.toDTO(&c))
	}
	return buildCategoryTree(flat, 0), nil
}

func (s *CategoryService) Create(in *dto.CategoryDTO) (*dto.CategoryDTO, error) {
	level := 0
	if in.ParentID > 0 {
		parent, err := s.repo.GetByID(s.tenantID, in.ParentID)
		if err != nil {
			return nil, ErrNotFound
		}
		level = parent.Level + 1
	}
	c := model.Category{TenantID: s.tenantID, ParentID: in.ParentID, Name: in.Name, Level: level, Sort: in.Sort, ShowStatus: in.ShowStatus}
	if err := s.repo.Create(&c); err != nil {
		return nil, err
	}
	item := s.toDTO(&c)
	return &item, nil
}

func (s *CategoryService) Update(id uint64, in *dto.CategoryDTO) (*dto.CategoryDTO, error) {
	c, err := s.repo.GetByID(s.tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	c.Name, c.Sort, c.ShowStatus = in.Name, in.Sort, in.ShowStatus
	if err := s.repo.Save(c); err != nil {
		return nil, err
	}
	item := s.toDTO(c)
	return &item, nil
}

func (s *CategoryService) Delete(id uint64) error {
	n, _ := s.repo.CountChildren(s.tenantID, id)
	if n > 0 {
		return errors.New("category has children")
	}
	if err := s.repo.Delete(s.tenantID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *CategoryService) toDTO(c *model.Category) dto.CategoryDTO {
	count, _ := s.repo.CountProducts(s.tenantID, c.ID)
	return dto.CategoryDTO{
		ID: c.ID, ParentID: c.ParentID, Name: c.Name, Level: c.Level,
		Sort: c.Sort, ShowStatus: c.ShowStatus, ProductCount: count,
	}
}

func buildCategoryTree(items []dto.CategoryDTO, parentID uint64) []dto.CategoryDTO {
	tree := make([]dto.CategoryDTO, 0)
	for _, item := range items {
		if item.ParentID == parentID {
			item.Children = buildCategoryTree(items, item.ID)
			tree = append(tree, item)
		}
	}
	return tree
}
