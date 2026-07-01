package service

import (
	"errors"

	"productcore/internal/dto"
	"productcore/internal/model"
	"productcore/internal/pkg/util"
	"productcore/internal/repo"

	"gorm.io/gorm"
)

type ProductGroupService struct {
	repos    *repo.Repos
	product  *ProductService
	tenantID uint64
}

func NewProductGroupService(repos *repo.Repos, productSvc *ProductService) *ProductGroupService {
	return &ProductGroupService{repos: repos, product: productSvc, tenantID: 1}
}

func (s *ProductGroupService) List() ([]dto.ProductGroupDTO, error) {
	groups, err := s.repos.Group.List(s.tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ProductGroupDTO, 0, len(groups))
	for _, g := range groups {
		out = append(out, s.toDTO(&g))
	}
	return out, nil
}

func (s *ProductGroupService) Create(in *dto.ProductGroupDTO) (*dto.ProductGroupDTO, error) {
	g := model.ProductGroup{TenantID: s.tenantID, Name: in.Name, Description: in.Description, Sort: in.Sort}
	if err := s.repos.Group.Create(&g); err != nil {
		return nil, err
	}
	item := s.toDTO(&g)
	return &item, nil
}

func (s *ProductGroupService) Update(id uint64, in *dto.ProductGroupDTO) (*dto.ProductGroupDTO, error) {
	g, err := s.repos.Group.GetByID(s.tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	g.Name, g.Description, g.Sort = in.Name, in.Description, in.Sort
	if err := s.repos.Group.Save(g); err != nil {
		return nil, err
	}
	item := s.toDTO(g)
	return &item, nil
}

func (s *ProductGroupService) Delete(id uint64) error {
	if err := s.repos.Group.Delete(s.tenantID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *ProductGroupService) ListProducts(groupID uint64) ([]dto.ProductDTO, error) {
	ids, err := s.repos.Group.ListProductIDs(s.tenantID, groupID)
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

func (s *ProductGroupService) toDTO(g *model.ProductGroup) dto.ProductGroupDTO {
	count, _ := s.repos.Group.CountProducts(s.tenantID, g.ID)
	return dto.ProductGroupDTO{
		ID: g.ID, Name: g.Name, Description: g.Description, Sort: g.Sort,
		ProductCount: count, CreateTime: util.FormatTime(g.CreatedAt),
	}
}


func (s *ProductGroupService) ForTenant(tenantID uint64) *ProductGroupService {
	cp := *s
	cp.tenantID = repo.NormalizeTenantID(tenantID)
	cp.product = s.product.ForTenant(tenantID)
	return &cp
}
