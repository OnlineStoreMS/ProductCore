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

func (s *ProductGroupService) ForTenant(tenantID uint64) *ProductGroupService {
	cp := *s
	cp.tenantID = repo.NormalizeTenantID(tenantID)
	cp.product = s.product.ForTenant(tenantID)
	return &cp
}

// List 返回扁平分组列表（含 parentId/level，供下拉等场景）。
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

// Tree 返回树形分组。
func (s *ProductGroupService) Tree() ([]dto.ProductGroupDTO, error) {
	flat, err := s.List()
	if err != nil {
		return nil, err
	}
	return buildGroupTree(flat, 0), nil
}

func (s *ProductGroupService) Create(in *dto.ProductGroupDTO) (*dto.ProductGroupDTO, error) {
	level := 0
	if in.ParentID > 0 {
		parent, err := s.repos.Group.GetByID(s.tenantID, in.ParentID)
		if err != nil {
			return nil, ErrNotFound
		}
		level = parent.Level + 1
	}
	g := model.ProductGroup{
		TenantID:    s.tenantID,
		ParentID:    in.ParentID,
		Name:        in.Name,
		Description: in.Description,
		Level:       level,
		Sort:        in.Sort,
	}
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
	n, _ := s.repos.Group.CountChildren(s.tenantID, id)
	if n > 0 {
		return errors.New("group has children")
	}
	if err := s.repos.Group.Delete(s.tenantID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *ProductGroupService) ListProducts(groupID uint64) ([]dto.ProductDTO, error) {
	ids, err := s.repos.Group.ListSelfAndDescendantIDs(s.tenantID, groupID)
	if err != nil {
		return nil, err
	}
	seen := map[uint64]struct{}{}
	out := make([]dto.ProductDTO, 0)
	for _, gid := range ids {
		pids, err := s.repos.Group.ListProductIDs(s.tenantID, gid)
		if err != nil {
			return nil, err
		}
		for _, pid := range pids {
			if _, ok := seen[pid]; ok {
				continue
			}
			seen[pid] = struct{}{}
			item, err := s.product.Get(pid)
			if err != nil {
				continue
			}
			out = append(out, *item)
		}
	}
	return out, nil
}

func (s *ProductGroupService) toDTO(g *model.ProductGroup) dto.ProductGroupDTO {
	count, _ := s.repos.Group.CountProducts(s.tenantID, g.ID)
	return dto.ProductGroupDTO{
		ID: g.ID, ParentID: g.ParentID, Name: g.Name, Description: g.Description,
		Level: g.Level, Sort: g.Sort, ProductCount: count, CreateTime: util.FormatTime(g.CreatedAt),
	}
}

func buildGroupTree(items []dto.ProductGroupDTO, parentID uint64) []dto.ProductGroupDTO {
	tree := make([]dto.ProductGroupDTO, 0)
	for _, item := range items {
		if item.ParentID == parentID {
			item.Children = buildGroupTree(items, item.ID)
			tree = append(tree, item)
		}
	}
	return tree
}
