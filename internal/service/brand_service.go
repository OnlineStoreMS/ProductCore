package service

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"productcore/internal/dto"
	"productcore/internal/model"
	"productcore/internal/repo"

	"gorm.io/gorm"
)

var logoBgPattern = regexp.MustCompile(`(?i)^(#([0-9a-f]{3}|[0-9a-f]{4}|[0-9a-f]{6}|[0-9a-f]{8})|(rgb|rgba|hsl|hsla)\(\s*[\d.%,\s/]+\))$`)

func normalizeLogoBg(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" || strings.EqualFold(s, "transparent") {
		return ""
	}
	if !logoBgPattern.MatchString(s) {
		return ""
	}
	return s
}

type BrandService struct {
	repo *repo.BrandRepo
	tenantID uint64
}

func NewBrandService(repos *repo.Repos) *BrandService {
	return &BrandService{repo: repos.Brand, tenantID: 1}
}

func (s *BrandService) ForTenant(tenantID uint64) *BrandService {
	cp := *s
	cp.tenantID = repo.NormalizeTenantID(tenantID)
	return &cp
}

func (s *BrandService) List(keyword string) ([]dto.BrandDTO, error) {
	brands, err := s.repo.List(s.tenantID, keyword)
	if err != nil {
		return nil, err
	}
	out := make([]dto.BrandDTO, 0, len(brands))
	for _, b := range brands {
		out = append(out, s.toDTO(&b))
	}
	return out, nil
}

func (s *BrandService) Create(in *dto.BrandDTO) (*dto.BrandDTO, error) {
	b := model.Brand{TenantID: s.tenantID,
		Name: in.Name, Logo: in.Logo, LogoBg: normalizeLogoBg(in.LogoBg),
		Website: strings.TrimSpace(in.Website), Agent: strings.TrimSpace(in.Agent),
		AuthFilesJSON: encodeBrandAuthFiles(in.AuthFiles),
		FirstLetter: in.FirstLetter, Sort: in.Sort, ShowStatus: in.ShowStatus,
	}
	if b.FirstLetter == "" {
		b.FirstLetter = firstLetter(in.Name)
	}
	if err := s.repo.Create(&b); err != nil {
		return nil, err
	}
	item := s.toDTO(&b)
	return &item, nil
}

func (s *BrandService) Update(id uint64, in *dto.BrandDTO) (*dto.BrandDTO, error) {
	b, err := s.repo.GetByID(s.tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	b.Name, b.Logo, b.LogoBg = in.Name, in.Logo, normalizeLogoBg(in.LogoBg)
	b.Website, b.Agent = strings.TrimSpace(in.Website), strings.TrimSpace(in.Agent)
	b.AuthFilesJSON = encodeBrandAuthFiles(in.AuthFiles)
	b.FirstLetter, b.Sort, b.ShowStatus = in.FirstLetter, in.Sort, in.ShowStatus
	if b.FirstLetter == "" {
		b.FirstLetter = firstLetter(b.Name)
	}
	if err := s.repo.Save(b); err != nil {
		return nil, err
	}
	item := s.toDTO(b)
	return &item, nil
}

func (s *BrandService) Reorder(ids []uint64) error {
	return s.repo.Reorder(s.tenantID, ids)
}

func (s *BrandService) Delete(id uint64) error {
	if err := s.repo.Delete(s.tenantID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *BrandService) toDTO(b *model.Brand) dto.BrandDTO {
	count, _ := s.repo.CountProducts(s.tenantID, b.ID)
	return dto.BrandDTO{
		ID: b.ID, Name: b.Name, Logo: b.Logo, LogoBg: b.LogoBg,
		Website: b.Website, Agent: b.Agent, AuthFiles: decodeBrandAuthFiles(b.AuthFilesJSON),
		FirstLetter: b.FirstLetter, Sort: b.Sort, ShowStatus: b.ShowStatus, ProductCount: count,
	}
}

func encodeBrandAuthFiles(files []dto.BrandAuthFileDTO) string {
	out := make([]dto.BrandAuthFileDTO, 0, len(files))
	for _, f := range files {
		url := strings.TrimSpace(f.URL)
		if url == "" {
			continue
		}
		name := strings.TrimSpace(f.Name)
		if name == "" {
			name = url
		}
		out = append(out, dto.BrandAuthFileDTO{Name: name, URL: url})
	}
	if len(out) == 0 {
		return ""
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return ""
	}
	return string(raw)
}

func decodeBrandAuthFiles(raw string) []dto.BrandAuthFileDTO {
	s := strings.TrimSpace(raw)
	if s == "" {
		return []dto.BrandAuthFileDTO{}
	}
	var files []dto.BrandAuthFileDTO
	if err := json.Unmarshal([]byte(s), &files); err != nil {
		return []dto.BrandAuthFileDTO{}
	}
	return files
}

func firstLetter(name string) string {
	if name == "" {
		return "#"
	}
	r := []rune(name)
	return string(r[0])
}
