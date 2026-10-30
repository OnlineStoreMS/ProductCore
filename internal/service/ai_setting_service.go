package service

import (
	"context"
	"errors"
	"strings"

	"productcore/internal/dto"
	"productcore/internal/integrations/aimodel"
	"productcore/internal/model"
	"productcore/internal/repo"
)

const defaultRetouchModel = "qwen-image-3.0"

type AISettingService struct {
	repo     *repo.AISettingRepo
	ai       *aimodel.Client
	tenantID uint64
}

func NewAISettingService(repos *repo.Repos, ai *aimodel.Client) *AISettingService {
	return &AISettingService{
		repo:     repos.AISetting,
		ai:       ai,
		tenantID: 1,
	}
}

func (s *AISettingService) ForTenant(tenantID uint64) *AISettingService {
	cp := *s
	cp.tenantID = repo.NormalizeTenantID(tenantID)
	cp.repo = s.repo.WithTenant(tenantID)
	return &cp
}

func (s *AISettingService) Get(ctx context.Context) (*dto.AISettingDTO, error) {
	row, err := s.repo.Get()
	if err != nil {
		return nil, err
	}
	out := &dto.AISettingDTO{
		DefaultImageModel:        strings.TrimSpace(row.DefaultImageModel),
		DefaultImageConnectionID: row.DefaultImageConnectionID,
	}
	if out.DefaultImageModel == "" {
		out.DefaultImageModel = defaultRetouchModel
	}
	if models, err := s.ListImageModels(ctx); err == nil {
		for _, m := range models {
			if m.ModelID == out.DefaultImageModel &&
				(out.DefaultImageConnectionID == 0 || m.ConnectionID == out.DefaultImageConnectionID) {
				out.DefaultImageLabel = modelLabel(m)
				break
			}
		}
	}
	return out, nil
}

func (s *AISettingService) Update(ctx context.Context, in dto.UpdateAISettingRequest) (*dto.AISettingDTO, error) {
	modelID := strings.TrimSpace(in.DefaultImageModel)
	if modelID == "" {
		modelID = defaultRetouchModel
	}
	if models, err := s.ListImageModels(ctx); err == nil && len(models) > 0 {
		ok := false
		for _, m := range models {
			if m.ModelID != modelID {
				continue
			}
			if in.DefaultImageConnectionID > 0 && m.ConnectionID != in.DefaultImageConnectionID {
				continue
			}
			ok = true
			if in.DefaultImageConnectionID == 0 {
				in.DefaultImageConnectionID = m.ConnectionID
			}
			break
		}
		if !ok {
			return nil, errors.New("请选择 AI 大模型管理中心里已启用的图片模型")
		}
	}
	if err := s.repo.Save(&model.AISetting{
		DefaultImageModel:        modelID,
		DefaultImageConnectionID: in.DefaultImageConnectionID,
	}); err != nil {
		return nil, err
	}
	return s.Get(ctx)
}

func (s *AISettingService) ListImageModels(ctx context.Context) ([]dto.AIImageModelDTO, error) {
	if s.ai == nil {
		return nil, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	list, err := s.ai.ListImageModels(ctx, s.tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.AIImageModelDTO, 0, len(list))
	for _, item := range list {
		out = append(out, dto.AIImageModelDTO{
			ID:             item.ID,
			ConnectionID:   item.ConnectionID,
			ConnectionName: item.ConnectionName,
			ProviderCode:   item.ProviderCode,
			ModelID:        item.ModelID,
			DisplayName:    item.DisplayName,
			IsDefault:      item.IsDefault,
		})
	}
	return out, nil
}

func (s *AISettingService) ResolveRetouchModel(ctx context.Context, requestedModel string, requestedConn uint64) (string, uint64, error) {
	modelID := strings.TrimSpace(requestedModel)
	connID := requestedConn
	if modelID == "" {
		cfg, err := s.Get(ctx)
		if err != nil {
			return "", 0, err
		}
		modelID = cfg.DefaultImageModel
		if connID == 0 {
			connID = cfg.DefaultImageConnectionID
		}
	}
	if modelID == "" {
		modelID = defaultRetouchModel
	}
	return modelID, connID, nil
}

func modelLabel(m dto.AIImageModelDTO) string {
	name := strings.TrimSpace(m.DisplayName)
	if name == "" {
		name = m.ModelID
	}
	if m.ConnectionName != "" {
		return m.ConnectionName + " · " + name
	}
	return name
}
