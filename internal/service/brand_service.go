package service

import (
	"errors"

	"productcore/internal/dto"
	"productcore/internal/model"
	"productcore/internal/repo"

	"gorm.io/gorm"
)

type BrandService struct {
	repo *repo.BrandRepo
}

func NewBrandService(repos *repo.Repos) *BrandService {
	return &BrandService{repo: repos.Brand}
}

func (s *BrandService) List(keyword string) ([]dto.BrandDTO, error) {
	brands, err := s.repo.List(keyword)
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
	b := model.Brand{
		Name: in.Name, Logo: in.Logo, FirstLetter: in.FirstLetter,
		Sort: in.Sort, ShowStatus: in.ShowStatus,
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
	b, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	b.Name, b.Logo, b.FirstLetter, b.Sort, b.ShowStatus = in.Name, in.Logo, in.FirstLetter, in.Sort, in.ShowStatus
	if b.FirstLetter == "" {
		b.FirstLetter = firstLetter(b.Name)
	}
	if err := s.repo.Save(b); err != nil {
		return nil, err
	}
	item := s.toDTO(b)
	return &item, nil
}

func (s *BrandService) Delete(id uint64) error {
	if err := s.repo.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *BrandService) toDTO(b *model.Brand) dto.BrandDTO {
	count, _ := s.repo.CountProducts(b.ID)
	return dto.BrandDTO{
		ID: b.ID, Name: b.Name, Logo: b.Logo, FirstLetter: b.FirstLetter,
		Sort: b.Sort, ShowStatus: b.ShowStatus, ProductCount: count,
	}
}

func firstLetter(name string) string {
	if name == "" {
		return "#"
	}
	r := []rune(name)
	return string(r[0])
}
