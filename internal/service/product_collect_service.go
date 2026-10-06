package service

import (
	"encoding/json"
	"strings"
	"time"

	"productcore/internal/dto"
	"productcore/internal/integrations/agentscenter"
	"productcore/internal/model"
	"productcore/internal/pkg/taobaoimg"
	"productcore/internal/repo"
)

type ProductCollectService struct {
	repos    *repo.Repos
	agents   *agentscenter.Client
	products *ProductService
}

func NewProductCollectService(repos *repo.Repos, agents *agentscenter.Client, products *ProductService) *ProductCollectService {
	return &ProductCollectService{repos: repos, agents: agents, products: products}
}

func (s *ProductCollectService) IngestFromExtension(tenantID, userID uint64, productURL string, product *dto.ProductDTO, videoURLs []string) (*dto.ProductCollectTaskDTO, error) {
	tenantID = repo.NormalizeTenantID(tenantID)
	if product == nil {
		return nil, ErrCollectEmpty
	}
	name := strings.TrimSpace(product.Name)
	if name == "" || strings.Contains(name, "评价") {
		return nil, ErrCollectEmpty
	}
	taobaoimg.UnwrapProduct(product)
	if strings.TrimSpace(product.Pic) == "" && len(product.AlbumPics) == 0 && len(product.Skus) == 0 {
		return nil, ErrCollectEmpty
	}

	platform, normalized, err := DetectCollectPlatform(productURL)
	if err != nil {
		sn := strings.TrimSpace(product.ProductSn)
		if sn == "" {
			return nil, err
		}
		platform = "taobao"
		if strings.EqualFold(strings.TrimSpace(product.Source), "tmall") {
			platform = "taobao"
		}
		normalized = "https://item.taobao.com/item.htm?id=" + sn
	}

	product.IsDraft = 1
	product.PublishStatus = 0
	if strings.TrimSpace(product.Source) == "" {
		product.Source = platform
	}
	if strings.TrimSpace(product.Unit) == "" {
		product.Unit = "件"
	}

	detailN := 0
	if product.Media != nil {
		detailN = len(product.Media.DetailPics)
	}
	videos := collectRemoteVideoURLs(product, videoURLs)
	stripRemoteVideos(product)
	msg := "扩展采集「" + name + "」主图" + jsonNumber(uint64(len(product.AlbumPics))) +
		" 详情" + jsonNumber(uint64(detailN)) + " SKU" + jsonNumber(uint64(len(product.Skus))) +
		" 视频" + jsonNumber(uint64(len(videos)))
	payload := map[string]any{
		"platform":   platform,
		"productUrl": normalized,
		"source":     "extension",
		"message":    msg,
		"product":    product,
		"videos":     videos,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	task := &model.ProductCollectTask{
		TenantID:   tenantID,
		UserID:     userID,
		ProductURL: normalized,
		Platform:   platform,
		AgentName:  "浏览器扩展",
		Status:     "succeeded",
		ResultJSON: string(raw),
	}
	if err := s.repos.ProductCollect.Create(task); err != nil {
		return nil, err
	}
	s.ingestIfNeeded(tenantID, task)
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
	for i := range list {
		s.ingestIfNeeded(tenantID, &list[i])
	}
	out := make([]dto.ProductCollectTaskDTO, 0, len(list))
	for _, task := range list {
		out = append(out, toCollectDTO(task))
	}
	return out, total, nil
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
		s.ingestIfNeeded(tenantID, &list[i])
	}
}

func (s *ProductCollectService) ingestIfNeeded(tenantID uint64, task *model.ProductCollectTask) {
	if s.products == nil || task == nil || task.Status != "succeeded" || strings.TrimSpace(task.ResultJSON) == "" {
		return
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(task.ResultJSON), &payload); err != nil {
		return
	}
	if _, ok := payload["productId"]; ok {
		return
	}
	raw, ok := payload["product"]
	if !ok || len(raw) == 0 || string(raw) == "null" {
		return
	}
	var in dto.ProductDTO
	if err := json.Unmarshal(raw, &in); err != nil {
		s.patchCollectJSON(task, payload, 0, err.Error())
		return
	}
	taobaoimg.UnwrapProduct(&in)
	in.IsDraft = 1
	in.PublishStatus = 0
	if strings.TrimSpace(in.Unit) == "" {
		in.Unit = "件"
	}
	products := s.products.ForTenant(tenantID)
	if code := strings.TrimSpace(in.MaterialCode); code != "" {
		if id, exists := products.IDByMaterialCode(code); exists {
			s.patchCollectJSON(task, payload, id, "")
			return
		}
	}
	var extra []string
	if rawVideos, ok := payload["videos"]; ok && len(rawVideos) > 0 {
		_ = json.Unmarshal(rawVideos, &extra)
	}
	urls := collectRemoteVideoURLs(&in, extra)
	stripRemoteVideos(&in)
	created, err := products.Create(&in)
	if err != nil {
		s.patchCollectJSON(task, payload, 0, err.Error())
		return
	}
	if created == nil {
		return
	}
	var msg string
	if rawMsg, ok := payload["message"]; ok {
		_ = json.Unmarshal(rawMsg, &msg)
	}
	if imgOK, imgFailed, imgErr := products.ingestRemoteImages(created.ID, &in); imgOK > 0 || imgFailed > 0 || imgErr != nil {
		if imgOK > 0 {
			msg += "，已转存图片" + jsonNumber(uint64(imgOK))
		}
		if imgFailed > 0 {
			msg += "，图片失败" + jsonNumber(uint64(imgFailed))
		}
		if imgErr != nil {
			msg += "：" + imgErr.Error()
		}
	}
	ingestErr := ""
	if n, skipped, vErr := products.ingestRemoteVideos(created.ID, urls); vErr != nil {
		ingestErr = "视频上传失败：" + vErr.Error()
	} else if n > 0 || skipped > 0 {
		if n > 0 {
			msg += "，已上传视频" + jsonNumber(uint64(n))
		}
		if skipped > 0 {
			msg += "（跳过" + jsonNumber(uint64(skipped)) + "个比例不符或无法下载）"
		}
	}
	if b, err := json.Marshal(msg); err == nil {
		payload["message"] = b
	}
	s.patchCollectJSON(task, payload, created.ID, ingestErr)
}

func (s *ProductCollectService) patchCollectJSON(task *model.ProductCollectTask, payload map[string]json.RawMessage, productID uint64, ingestErr string) {
	if productID > 0 {
		b, _ := json.Marshal(productID)
		payload["productId"] = b
		delete(payload, "ingestError")
	}
	if ingestErr != "" {
		b, _ := json.Marshal(ingestErr)
		payload["ingestError"] = b
	}
	out, err := json.Marshal(payload)
	if err != nil {
		return
	}
	task.ResultJSON = string(out)
	_ = s.repos.ProductCollect.Save(task)
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
		ProductID:    collectProductID(task.ResultJSON),
		Status:       task.Status,
		Message:      collectMessage(task),
		CreatedAt:    task.CreatedAt.Format(time.RFC3339),
	}
}

func collectProductID(resultJSON string) uint64 {
	if strings.TrimSpace(resultJSON) == "" {
		return 0
	}
	var payload struct {
		ProductID uint64 `json:"productId"`
	}
	if err := json.Unmarshal([]byte(resultJSON), &payload); err != nil {
		return 0
	}
	return payload.ProductID
}

func collectMessage(task model.ProductCollectTask) string {
	if msg := strings.TrimSpace(task.ErrorMessage); msg != "" {
		return msg
	}
	if strings.TrimSpace(task.ResultJSON) == "" {
		return ""
	}
	var payload struct {
		Message     string `json:"message"`
		ProductID   uint64 `json:"productId"`
		IngestError string `json:"ingestError"`
	}
	if err := json.Unmarshal([]byte(task.ResultJSON), &payload); err != nil {
		return ""
	}
	msg := strings.TrimSpace(payload.Message)
	if payload.ProductID > 0 {
		if msg == "" {
			return "已写入草稿商品 #" + jsonNumber(payload.ProductID)
		}
		return msg + "（草稿商品 #" + jsonNumber(payload.ProductID) + "）"
	}
	if errMsg := strings.TrimSpace(payload.IngestError); errMsg != "" {
		if msg == "" {
			return "采集成功但入库失败：" + errMsg
		}
		return msg + "；入库失败：" + errMsg
	}
	return msg
}

func jsonNumber(v uint64) string {
	b, _ := json.Marshal(v)
	return string(b)
}
