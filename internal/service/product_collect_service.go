package service

import (
	"encoding/json"
	"strings"
	"time"

	"productcore/internal/dto"
	"productcore/internal/integrations/agentscenter"
	"productcore/internal/model"
	"productcore/internal/repo"
)

type ProductCollectService struct {
	repos  *repo.Repos
	agents *agentscenter.Client
}

func NewProductCollectService(repos *repo.Repos, agents *agentscenter.Client) *ProductCollectService {
	return &ProductCollectService{repos: repos, agents: agents}
}

func (s *ProductCollectService) Create(tenantID, userID uint64, productURL string) (*dto.ProductCollectTaskDTO, error) {
	if s.agents == nil {
		return nil, ErrAgentsCenterUnconfigured
	}
	tenantID = repo.NormalizeTenantID(tenantID)
	platform, normalized, err := DetectCollectPlatform(productURL)
	if err != nil {
		return nil, err
	}

	agent, err := s.pickAgent(tenantID)
	if err != nil {
		return nil, err
	}

	params, _ := json.Marshal(map[string]string{
		"productUrl": normalized,
		"platform":   platform,
		"source":     "productcore",
	})
	job, err := s.agents.CreateProductCollectJob(tenantID, agent.ID, platform, string(params))
	if err != nil {
		return nil, err
	}

	task := &model.ProductCollectTask{
		TenantID:   tenantID,
		UserID:     userID,
		ProductURL: normalized,
		Platform:   platform,
		AgentJobID: job.ID,
		AgentID:    agent.ID,
		AgentName:  agent.Name,
		Status:     "pending",
	}
	if err := s.repos.ProductCollect.Create(task); err != nil {
		return nil, err
	}
	dtoTask := toCollectDTO(*task)
	return &dtoTask, nil
}

func (s *ProductCollectService) List(tenantID uint64, page, pageSize int) ([]dto.ProductCollectTaskDTO, int64, error) {
	tenantID = repo.NormalizeTenantID(tenantID)
	list, total, err := s.repos.ProductCollect.List(tenantID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	s.syncJobs(tenantID, list)
	out := make([]dto.ProductCollectTaskDTO, 0, len(list))
	for _, task := range list {
		out = append(out, toCollectDTO(task))
	}
	return out, total, nil
}

func (s *ProductCollectService) pickAgent(tenantID uint64) (*agentscenter.Agent, error) {
	capable, err := s.agents.ListAgents(tenantID, true, agentscenter.JobTypeEcommerceProductCollect)
	if err != nil {
		return nil, err
	}
	if agent := newestAgent(capable); agent != nil {
		return agent, nil
	}
	anyOnline, err := s.agents.ListAgents(tenantID, true, "")
	if err != nil {
		return nil, err
	}
	if len(anyOnline) > 0 {
		return nil, ErrCollectAgentOutdated
	}
	return nil, ErrNoCollectAgent
}

func (s *ProductCollectService) syncJobs(tenantID uint64, list []model.ProductCollectTask) {
	if s.agents == nil || len(list) == 0 {
		return
	}
	ids := make([]uint64, 0, len(list))
	for _, task := range list {
		if task.AgentJobID > 0 && !collectTerminal(task.Status) {
			ids = append(ids, task.AgentJobID)
		}
	}
	if len(ids) == 0 {
		return
	}
	jobs, err := s.agents.GetJobs(tenantID, ids)
	if err != nil {
		return
	}
	byID := make(map[uint64]agentscenter.JobStatus, len(jobs))
	for _, job := range jobs {
		byID[job.ID] = job
	}
	for i := range list {
		job, ok := byID[list[i].AgentJobID]
		if !ok {
			continue
		}
		status := strings.TrimSpace(job.Status)
		if status == "" || (status == list[i].Status && job.ErrorMessage == list[i].ErrorMessage && job.ResultJSON == list[i].ResultJSON) {
			continue
		}
		list[i].Status = status
		list[i].ErrorMessage = job.ErrorMessage
		list[i].ResultJSON = job.ResultJSON
		_ = s.repos.ProductCollect.Save(&list[i])
	}
}

func newestAgent(list []agentscenter.Agent) *agentscenter.Agent {
	var best *agentscenter.Agent
	var bestAt time.Time
	for i := range list {
		at := time.Time{}
		if list[i].LastHeartbeat != nil {
			if parsed, err := time.Parse(time.RFC3339, *list[i].LastHeartbeat); err == nil {
				at = parsed
			}
		}
		if best == nil || at.After(bestAt) {
			best = &list[i]
			bestAt = at
		}
	}
	return best
}

func collectTerminal(status string) bool {
	switch status {
	case "succeeded", "failed", "cancelled":
		return true
	default:
		return false
	}
}

func toCollectDTO(task model.ProductCollectTask) dto.ProductCollectTaskDTO {
	name := "淘宝/天猫"
	if task.Platform != "taobao" {
		name = task.Platform
	}
	return dto.ProductCollectTaskDTO{
		ID:           task.ID,
		ProductURL:   task.ProductURL,
		Platform:     task.Platform,
		PlatformName: name,
		AgentJobID:   task.AgentJobID,
		AgentID:      task.AgentID,
		AgentName:    task.AgentName,
		Status:       task.Status,
		Message:      collectMessage(task),
		CreatedAt:    task.CreatedAt.Format(time.RFC3339),
	}
}

func collectMessage(task model.ProductCollectTask) string {
	if msg := strings.TrimSpace(task.ErrorMessage); msg != "" {
		return msg
	}
	if strings.TrimSpace(task.ResultJSON) == "" {
		return ""
	}
	var payload struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(task.ResultJSON), &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.Message)
}
