package service

import (
	"errors"
	"regexp"
	"strings"

	"productcore/internal/dto"
	"productcore/internal/model"
	"productcore/internal/pkg/util"
	"productcore/internal/repo"

	"gorm.io/gorm"
)

var typeCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)

type PlatformTypeService struct {
	repo *repo.PlatformTypeRepo
}

func NewPlatformTypeService(repos *repo.Repos) *PlatformTypeService {
	return &PlatformTypeService{repo: repos.PlatformType}
}

func (s *PlatformTypeService) List(keyword string) ([]dto.PlatformShopTypeDTO, error) {
	list, err := s.repo.List(keyword, false)
	if err != nil {
		return nil, err
	}
	out := make([]dto.PlatformShopTypeDTO, 0, len(list))
	for _, item := range list {
		out = append(out, s.toDTO(&item))
	}
	return out, nil
}

func (s *PlatformTypeService) ListEnabled() ([]dto.PlatformShopTypeDTO, error) {
	list, err := s.repo.List("", true)
	if err != nil {
		return nil, err
	}
	out := make([]dto.PlatformShopTypeDTO, 0, len(list))
	for _, item := range list {
		out = append(out, s.toDTO(&item))
	}
	return out, nil
}

func (s *PlatformTypeService) Create(in *dto.PlatformShopTypeDTO) (*dto.PlatformShopTypeDTO, error) {
	code := normalizeTypeCode(in.Code)
	if code == "" || !typeCodePattern.MatchString(code) {
		return nil, errors.New("type code must be lowercase letters, numbers or underscore")
	}
	if _, err := s.repo.GetByCode(code); err == nil {
		return nil, ErrDuplicateTypeCode
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	item := &model.PlatformShopType{
		Code: code, Name: strings.TrimSpace(in.Name), Logo: in.Logo,
		Sort: in.Sort, Enabled: in.Enabled, Remark: in.Remark,
	}
	if item.Enabled != 0 && item.Enabled != 1 {
		item.Enabled = 1
	}
	if err := s.repo.Create(item); err != nil {
		return nil, err
	}
	out := s.toDTO(item)
	return &out, nil
}

func (s *PlatformTypeService) Update(id uint64, in *dto.PlatformShopTypeDTO) (*dto.PlatformShopTypeDTO, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if item.IsBuiltin == 1 {
		item.Name = strings.TrimSpace(in.Name)
		item.Logo = in.Logo
		item.Sort = in.Sort
		item.Enabled = in.Enabled
		item.Remark = in.Remark
	} else {
		code := normalizeTypeCode(in.Code)
		if code == "" || !typeCodePattern.MatchString(code) {
			return nil, errors.New("type code must be lowercase letters, numbers or underscore")
		}
		if code != item.Code {
			if existing, err := s.repo.GetByCode(code); err == nil && existing.ID != id {
				return nil, ErrDuplicateTypeCode
			} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
			item.Code = code
		}
		item.Name = strings.TrimSpace(in.Name)
		item.Logo = in.Logo
		item.Sort = in.Sort
		item.Enabled = in.Enabled
		item.Remark = in.Remark
	}
	if err := s.repo.Save(item); err != nil {
		return nil, err
	}
	out := s.toDTO(item)
	return &out, nil
}

func (s *PlatformTypeService) Delete(id uint64) error {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if item.IsBuiltin == 1 {
		return ErrBuiltinPlatformType
	}
	count, err := s.repo.CountShops(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrTypeHasShops
	}
	if err := s.repo.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *PlatformTypeService) Get(id uint64) (*dto.PlatformShopTypeDTO, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	out := s.toDTO(item)
	return &out, nil
}

func (s *PlatformTypeService) toDTO(item *model.PlatformShopType) dto.PlatformShopTypeDTO {
	count, _ := s.repo.CountShops(item.ID)
	return dto.PlatformShopTypeDTO{
		ID: item.ID, Code: item.Code, Name: item.Name, Logo: item.Logo,
		Sort: item.Sort, Enabled: item.Enabled, IsBuiltin: item.IsBuiltin,
		Remark: item.Remark, ShopCount: count,
		CreateTime: util.FormatTime(item.CreatedAt),
	}
}

func normalizeTypeCode(code string) string {
	return strings.ToLower(strings.TrimSpace(code))
}

type PlatformShopService struct {
	typeRepo *repo.PlatformTypeRepo
	shopRepo *repo.PlatformShopRepo
	listing  *repo.PlatformListingRepo
}

func NewPlatformShopService(repos *repo.Repos) *PlatformShopService {
	return &PlatformShopService{
		typeRepo: repos.PlatformType,
		shopRepo: repos.PlatformShop,
		listing:  repos.PlatformListing,
	}
}

func (s *PlatformShopService) List(q dto.PlatformShopQuery) ([]dto.PlatformShopDTO, int64, error) {
	list, total, err := s.shopRepo.List(q)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.PlatformShopDTO, 0, len(list))
	shopIDs := make([]uint64, len(list))
	for i, item := range list {
		shopIDs[i] = item.ID
	}
	counts, _ := s.listing.CountProductsByShopIDs(shopIDs)
	for _, item := range list {
		d := s.toDTO(&item)
		if counts != nil {
			d.ListedProductCount = counts[item.ID]
		}
		out = append(out, d)
	}
	return out, total, nil
}

func (s *PlatformShopService) Create(in *dto.PlatformShopDTO) (*dto.PlatformShopDTO, error) {
	pt, err := s.typeRepo.GetByID(in.PlatformTypeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("platform type not found")
		}
		return nil, err
	}
	item := &model.PlatformShop{
		PlatformTypeID: pt.ID,
		Name:           strings.TrimSpace(in.Name),
		SourceChannel:  pt.Code,
		ShopCode:       strings.TrimSpace(in.ShopCode),
		ExternalShopID: strings.TrimSpace(in.ExternalShopID),
		Status:         in.Status,
		Remark:         in.Remark,
		Sort:           in.Sort,
	}
	if item.Status != 0 && item.Status != 1 {
		item.Status = 1
	}
	if err := s.shopRepo.Create(item); err != nil {
		return nil, err
	}
	out := s.toDTO(item)
	return &out, nil
}

func (s *PlatformShopService) Update(id uint64, in *dto.PlatformShopDTO) (*dto.PlatformShopDTO, error) {
	item, err := s.shopRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if in.PlatformTypeID > 0 && in.PlatformTypeID != item.PlatformTypeID {
		pt, err := s.typeRepo.GetByID(in.PlatformTypeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("platform type not found")
			}
			return nil, err
		}
		item.PlatformTypeID = pt.ID
		item.SourceChannel = pt.Code
	}
	item.Name = strings.TrimSpace(in.Name)
	item.ShopCode = strings.TrimSpace(in.ShopCode)
	item.ExternalShopID = strings.TrimSpace(in.ExternalShopID)
	item.Status = in.Status
	item.Remark = in.Remark
	item.Sort = in.Sort
	if err := s.shopRepo.Save(item); err != nil {
		return nil, err
	}
	out := s.toDTO(item)
	return &out, nil
}

func (s *PlatformShopService) Delete(id uint64) error {
	if _, err := s.shopRepo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if err := s.listing.DeleteByShop(id); err != nil {
		return err
	}
	if err := s.shopRepo.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *PlatformShopService) toDTO(item *model.PlatformShop) dto.PlatformShopDTO {
	out := dto.PlatformShopDTO{
		ID: item.ID, PlatformTypeID: item.PlatformTypeID,
		Name: item.Name, SourceChannel: item.SourceChannel,
		ShopCode: item.ShopCode, ExternalShopID: item.ExternalShopID,
		Status: item.Status, Remark: item.Remark, Sort: item.Sort,
		CreateTime: util.FormatTime(item.CreatedAt),
		UpdateTime: util.FormatTime(item.UpdatedAt),
	}
	if pt, err := s.typeRepo.GetByID(item.PlatformTypeID); err == nil {
		out.PlatformTypeName = pt.Name
		out.PlatformTypeLogo = pt.Logo
	}
	return out
}
