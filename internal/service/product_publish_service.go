package service

import (
	"encoding/json"
	"errors"
	"math"
	"strings"

	"productcore/internal/dto"
	"productcore/internal/integrations/agentscenter"
	"productcore/internal/model"
	"productcore/internal/repo"
)

const publishJobType = "doudian.product.publish"
const updateJobType = "doudian.product.update"

type ProductPublishService struct {
	repo       *repo.ProductPublishRepo
	categories *repo.CategoryRepo
	productSvc *ProductService
	agents     *agentscenter.Client
	tenantID   uint64
}

func NewProductPublishService(repos *repo.Repos, productSvc *ProductService, agents *agentscenter.Client) *ProductPublishService {
	return &ProductPublishService{
		repo:       repos.ProductPublish,
		categories: repos.Category,
		productSvc: productSvc,
		agents:     agents,
		tenantID:   1,
	}
}

func (s *ProductPublishService) ForTenant(tenantID uint64) *ProductPublishService {
	cp := *s
	cp.tenantID = repo.NormalizeTenantID(tenantID)
	cp.repo = s.repo.WithTenant(tenantID)
	cp.productSvc = s.productSvc.ForTenant(tenantID)
	return &cp
}

func (s *ProductPublishService) ListShops(productID uint64) (*dto.PublishShopListDTO, error) {
	out := &dto.PublishShopListDTO{List: []dto.PublishShopDTO{}}
	if s.agents == nil {
		out.Hint = "未配置 AgentsCenter，无法读取电商店铺"
		return out, nil
	}
	shops, err := s.agents.ListShops(s.tenantID)
	if err != nil {
		return nil, err
	}
	var current map[string]model.ProductPublish
	if productID > 0 {
		rows, err := s.repo.ListByProductIDs([]uint64{productID})
		if err != nil {
			return nil, err
		}
		current = make(map[string]model.ProductPublish, len(rows))
		for _, row := range rows {
			current[row.Platform+"|"+row.PlatformShopID] = row
		}
	}
	if len(shops) == 0 {
		out.Hint = "没有在线的电商店铺。请确认 WindowsAgent 已登录店铺并保持在线。"
		return out, nil
	}
	// 同一抖店可能挂在多台在线电脑上。列表只保留一条，优先已勾选铺货的那台。
	chosen := map[string]dto.PublishShopDTO{}
	order := make([]string, 0, len(shops))
	for _, sh := range shops {
		item := dto.PublishShopDTO{
			ID:               sh.ID,
			AgentID:          sh.AgentID,
			AgentName:        sh.AgentName,
			Platform:         sh.Platform,
			PlatformShopID:   sh.PlatformShopID,
			PlatformShopName: sh.PlatformShopName,
			AgentOnline:      sh.AgentOnline,
		}
		if row, ok := current[sh.Platform+"|"+sh.PlatformShopID]; ok {
			item.PublishStatus = row.Status
			item.PlatformProductID = row.PlatformProductID
		}
		ready, reason := shopPublishReady(sh)
		item.PublishReady = ready
		item.Reason = reason
		key := sh.Platform + "|" + sh.PlatformShopID
		prev, ok := chosen[key]
		if !ok {
			chosen[key] = item
			order = append(order, key)
			continue
		}
		if !prev.PublishReady && item.PublishReady {
			chosen[key] = item
		}
	}
	for _, key := range order {
		out.List = append(out.List, chosen[key])
	}
	if !anyReady(out.List) {
		out.Hint = "在线店铺都还不能铺货。请在 WindowsAgent 电商店铺里勾选「商品铺货」，等心跳上报后再试。"
	}
	return out, nil
}

func (s *ProductPublishService) Start(productID uint64, platform, platformShopID string) (*dto.StartPublishResult, error) {
	platform = strings.TrimSpace(platform)
	platformShopID = strings.TrimSpace(platformShopID)
	if platform == "" || platformShopID == "" {
		return nil, errors.New("请选择店铺")
	}
	if s.agents == nil {
		return nil, errors.New("未配置 AgentsCenter")
	}
	shops, err := s.agents.ListShops(s.tenantID)
	if err != nil {
		return nil, err
	}
	var shop *agentscenter.ShopItem
	reason := ""
	for i := range shops {
		if shops[i].Platform != platform || shops[i].PlatformShopID != platformShopID {
			continue
		}
		ready, why := shopPublishReady(shops[i])
		if ready {
			shop = &shops[i]
			break
		}
		if reason == "" {
			reason = why
		}
	}
	if shop == nil && reason == "" {
		return nil, errors.New("店铺不在线，或还没有从 WindowsAgent 上报")
	}
	if shop == nil {
		return nil, errors.New(reason)
	}

	existing, err := s.repo.Get(productID, platform, platformShopID)
	if err != nil && !repo.IsPublishNotFound(err) {
		return nil, err
	}
	if existing != nil && existing.Status == "running" && existing.AgentJobID > 0 {
		_ = s.syncRows([]model.ProductPublish{*existing})
		existing, err = s.repo.Get(productID, platform, platformShopID)
		if err != nil {
			return nil, err
		}
		if existing.Status == "running" {
			return nil, errors.New("该店铺正在铺货，请等本次结束")
		}
	}

	product, err := s.productSvc.Get(productID)
	if err != nil {
		return nil, err
	}
	params, err := buildPublishParams(product, shop, s.categoryPath(product.CategoryID), false)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	job, err := s.agents.CreateJob(agentscenter.CreateJobRequest{
		TenantID:         s.tenantID,
		JobType:          publishJobType,
		Platform:         shop.Platform,
		PlatformShopID:   shop.PlatformShopID,
		PlatformShopName: shop.PlatformShopName,
		ParamsJSON:       string(raw),
		Source:           "product",
		Priority:         50,
	})
	if err != nil {
		return nil, err
	}

	row := &model.ProductPublish{
		ProductID:        productID,
		Platform:         shop.Platform,
		PlatformShopID:   shop.PlatformShopID,
		PlatformShopName: shop.PlatformShopName,
		AgentID:          shop.AgentID,
		AgentJobID:       job.ID,
		Status:           "running",
		ErrorMessage:     "",
	}
	if existing != nil {
		row.ID = existing.ID
		row.PlatformProductID = existing.PlatformProductID
	}
	if err := s.repo.Save(row); err != nil {
		return nil, err
	}
	return &dto.StartPublishResult{
		ID:               row.ID,
		ShopName:         row.PlatformShopName,
		Platform:         row.Platform,
		PlatformTypeName: platformLabel(row.Platform),
		PublishStatus:    row.Status,
		AgentJobID:       row.AgentJobID,
	}, nil
}

func (s *ProductPublishService) StartUpdate(productID uint64, platform, platformShopID, platformProductID, scope string) (*dto.StartPublishResult, error) {
	platform = strings.TrimSpace(platform)
	platformShopID = strings.TrimSpace(platformShopID)
	platformProductID = strings.TrimSpace(platformProductID)
	scope = strings.TrimSpace(scope)
	if scope == "" {
		scope = "all"
	}
	if scope != "all" && scope != "sku" {
		return nil, errors.New("更新范围不正确")
	}
	if platform == "" || platformShopID == "" {
		return nil, errors.New("请选择店铺")
	}
	if !validPlatformProductID(platformProductID) {
		return nil, errors.New("请输入店铺里的商品 ID")
	}
	if s.agents == nil {
		return nil, errors.New("未配置 AgentsCenter")
	}
	shops, err := s.agents.ListShops(s.tenantID)
	if err != nil {
		return nil, err
	}
	var shop *agentscenter.ShopItem
	reason := ""
	for i := range shops {
		if shops[i].Platform != platform || shops[i].PlatformShopID != platformShopID {
			continue
		}
		ready, why := shopPublishReady(shops[i])
		if ready {
			shop = &shops[i]
			break
		}
		if reason == "" {
			reason = why
		}
	}
	if shop == nil && reason == "" {
		return nil, errors.New("店铺不在线，或还没有从 WindowsAgent 上报")
	}
	if shop == nil {
		return nil, errors.New(reason)
	}

	existing, err := s.repo.Get(productID, platform, platformShopID)
	if err != nil && !repo.IsPublishNotFound(err) {
		return nil, err
	}
	if existing != nil && existing.Status == "running" && existing.AgentJobID > 0 {
		_ = s.syncRows([]model.ProductPublish{*existing})
		existing, err = s.repo.Get(productID, platform, platformShopID)
		if err != nil {
			return nil, err
		}
		if existing.Status == "running" {
			return nil, errors.New("该店铺正在铺货，请等本次结束")
		}
	}

	product, err := s.productSvc.Get(productID)
	if err != nil {
		return nil, err
	}
	params, err := buildPublishParams(product, shop, s.categoryPath(product.CategoryID), scope == "sku")
	if err != nil {
		return nil, err
	}
	params.PlatformProductID = platformProductID
	params.UpdateScope = scope
	raw, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	job, err := s.agents.CreateJob(agentscenter.CreateJobRequest{
		TenantID:         s.tenantID,
		JobType:          updateJobType,
		Platform:         shop.Platform,
		PlatformShopID:   shop.PlatformShopID,
		PlatformShopName: shop.PlatformShopName,
		ParamsJSON:       string(raw),
		Source:           "product",
		Priority:         50,
	})
	if err != nil {
		return nil, err
	}

	row := &model.ProductPublish{
		ProductID:         productID,
		Platform:          shop.Platform,
		PlatformShopID:    shop.PlatformShopID,
		PlatformShopName:  shop.PlatformShopName,
		AgentID:           shop.AgentID,
		AgentJobID:        job.ID,
		Status:            "running",
		PlatformProductID: platformProductID,
		ErrorMessage:      "",
	}
	if existing != nil {
		row.ID = existing.ID
	}
	if err := s.repo.Save(row); err != nil {
		return nil, err
	}
	return &dto.StartPublishResult{
		ID:               row.ID,
		ShopName:         row.PlatformShopName,
		Platform:         row.Platform,
		PlatformTypeName: platformLabel(row.Platform),
		PublishStatus:    row.Status,
		AgentJobID:       row.AgentJobID,
	}, nil
}

func validPlatformProductID(id string) bool {
	if len(id) < 6 || len(id) > 32 {
		return false
	}
	for _, r := range id {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (s *ProductPublishService) Sync() {
	if s.agents == nil {
		return
	}
	rows, err := s.repo.ListOpen()
	if err != nil || len(rows) == 0 {
		return
	}
	_ = s.syncRows(rows)
}

func (s *ProductPublishService) syncRows(rows []model.ProductPublish) error {
	if s.agents == nil || len(rows) == 0 {
		return nil
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		if row.AgentJobID > 0 {
			ids = append(ids, row.AgentJobID)
		}
	}
	jobs, err := s.agents.GetJobs(s.tenantID, ids)
	if err != nil {
		return err
	}
	byID := make(map[uint64]agentscenter.JobStatus, len(jobs))
	for _, job := range jobs {
		byID[job.ID] = job
	}
	for i := range rows {
		job, ok := byID[rows[i].AgentJobID]
		if !ok {
			continue
		}
		next := rows[i]
		switch job.Status {
		case "succeeded":
			next.Status = "succeeded"
			next.ErrorMessage = ""
			if id := platformProductID(job.ResultJSON); id != "" {
				next.PlatformProductID = id
			}
		case "failed", "cancelled":
			next.Status = "failed"
			next.ErrorMessage = job.ErrorMessage
		default:
			continue
		}
		_ = s.repo.Save(&next)
	}
	return nil
}

func (s *ProductPublishService) AttachListedShops(list []dto.ProductDTO) error {
	if len(list) == 0 {
		return nil
	}
	ids := make([]uint64, len(list))
	for i, p := range list {
		ids[i] = p.ID
	}
	rows, err := s.repo.ListByProductIDs(ids)
	if err != nil {
		return err
	}
	grouped := make(map[uint64][]dto.ListedShopDTO)
	for _, row := range rows {
		grouped[row.ProductID] = append(grouped[row.ProductID], toListedShop(row))
	}
	for i := range list {
		shops := grouped[list[i].ID]
		if shops == nil {
			shops = []dto.ListedShopDTO{}
		}
		list[i].ListedShops = shops
		list[i].ListedShopCount = len(shops)
	}
	return nil
}

func (s *ProductPublishService) ShopsByProductIDs(productIDs []uint64) (map[uint64][]dto.ListedShopDTO, error) {
	rows, err := s.repo.ListByProductIDs(productIDs)
	if err != nil {
		return nil, err
	}
	out := make(map[uint64][]dto.ListedShopDTO)
	for _, row := range rows {
		out[row.ProductID] = append(out[row.ProductID], toListedShop(row))
	}
	return out, nil
}

func toListedShop(row model.ProductPublish) dto.ListedShopDTO {
	return dto.ListedShopDTO{
		ShopID:           row.ID,
		ShopName:         row.PlatformShopName,
		PlatformTypeName: platformLabel(row.Platform),
		PublishStatus:    row.Status,
		ErrorMessage:     row.ErrorMessage,
	}
}

func shopPublishReady(sh agentscenter.ShopItem) (bool, string) {
	if sh.Platform != "doudian" {
		return false, "暂只支持抖店铺货"
	}
	if !hasCapability(sh.CapabilitiesJSON, publishJobType) {
		return false, "请在 WindowsAgent 电商店铺里勾选「商品铺货」"
	}
	return true, ""
}

func hasCapability(raw, skill string) bool {
	var caps []string
	if err := json.Unmarshal([]byte(raw), &caps); err != nil {
		return strings.Contains(raw, skill)
	}
	for _, c := range caps {
		if c == skill {
			return true
		}
	}
	return false
}

func anyReady(list []dto.PublishShopDTO) bool {
	for _, item := range list {
		if item.PublishReady {
			return true
		}
	}
	return false
}

func platformLabel(platform string) string {
	switch platform {
	case "doudian":
		return "抖店"
	case "taobao":
		return "淘宝"
	case "pdd":
		return "拼多多"
	default:
		return platform
	}
}

func platformProductID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var body struct {
		PlatformProductID string `json:"platformProductId"`
	}
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		return ""
	}
	return strings.TrimSpace(body.PlatformProductID)
}

type publishParams struct {
	Platform          string       `json:"platform"`
	PlatformShopID    string       `json:"platformShopId"`
	PlatformShopName  string       `json:"platformShopName"`
	ProductID         uint64       `json:"productId"`
	Title             string       `json:"title"`
	CategoryPath      []string     `json:"categoryPath,omitempty"`
	CategoryKeyword   string       `json:"categoryKeyword"`
	CategoryPathHint  string       `json:"categoryPathHint"`
	Brand             string       `json:"brand"`
	CategoryAttrName  string       `json:"categoryAttrName,omitempty"`
	CategoryAttrValue string       `json:"categoryAttrValue,omitempty"`
	SpecName          string       `json:"specName"`
	MainImages        []string     `json:"mainImages"`
	DetailImages      []string     `json:"detailImages"`
	Skus              []publishSku `json:"skus"`
	PlatformProductID string       `json:"platformProductId,omitempty"`
	UpdateScope       string       `json:"updateScope,omitempty"`
}

type publishSku struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
	Stock int    `json:"stock"`
	Pic   string `json:"pic,omitempty"`
}

func (s *ProductPublishService) categoryPath(categoryID uint64) []string {
	if s.categories == nil || categoryID == 0 {
		return nil
	}
	var names []string
	seen := map[uint64]struct{}{}
	id := categoryID
	for id > 0 {
		if _, ok := seen[id]; ok {
			break
		}
		seen[id] = struct{}{}
		cat, err := s.categories.GetByID(s.tenantID, id)
		if err != nil || cat == nil {
			break
		}
		name := strings.TrimSpace(cat.Name)
		if name != "" && name != "无分类" {
			names = append([]string{name}, names...)
		}
		id = cat.ParentID
	}
	return names
}

func buildPublishParams(p *dto.ProductDTO, shop *agentscenter.ShopItem, path []string, skuOnly bool) (*publishParams, error) {
	title := trimDouyinTitle(strings.TrimSpace(p.Name), 60)
	if title == "" && !skuOnly {
		return nil, errors.New("商品标题为空")
	}
	mains := uniqueNonEmpty(append([]string{p.Pic}, p.AlbumPics...))
	if len(mains) > 5 {
		mains = mains[:5]
	}
	if len(mains) == 0 && !skuOnly {
		return nil, errors.New("商品没有主图")
	}
	var details []string
	if p.Media != nil {
		details = uniqueNonEmpty(p.Media.DetailPics)
	}
	if len(details) > 49 {
		details = details[:49]
	}
	specName := "商品规格"
	if len(p.SkuSpecs) > 0 && strings.TrimSpace(p.SkuSpecs[0].Name) != "" {
		specName = strings.TrimSpace(p.SkuSpecs[0].Name)
	}
	if len(p.Skus) == 0 {
		return nil, errors.New("商品没有 SKU")
	}
	skus := make([]publishSku, 0, len(p.Skus))
	seen := map[string]struct{}{}
	for _, sku := range p.Skus {
		name := skuLabel(sku, specName)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		skus = append(skus, publishSku{
			Name:  name,
			Price: int(math.Round(sku.Price)),
			Stock: sku.Stock,
			Pic:   strings.TrimSpace(sku.Pic),
		})
	}
	if len(skus) == 0 {
		return nil, errors.New("商品没有可用的规格值")
	}
	category := strings.TrimSpace(p.CategoryName)
	if len(path) > 0 {
		category = path[len(path)-1]
	}
	keyword := category
	if i := strings.IndexAny(category, "/／"); i > 0 {
		keyword = strings.TrimSpace(category[:i])
	}
	hint := category
	if len(path) > 0 {
		hint = strings.Join(path, " > ")
	}
	params := &publishParams{
		Platform:         shop.Platform,
		PlatformShopID:   shop.PlatformShopID,
		PlatformShopName: shop.PlatformShopName,
		ProductID:        p.ID,
		Title:            title,
		CategoryPath:     path,
		CategoryKeyword:  keyword,
		CategoryPathHint: hint,
		Brand:            "无品牌",
		SpecName:         specName,
		MainImages:       mains,
		DetailImages:     details,
		Skus:             skus,
	}
	if strings.Contains(category, "变速器") {
		params.CategoryAttrName = "变速器/保护器分类"
		params.CategoryAttrValue = "后拔/后变速器"
	}
	return params, nil
}

func skuLabel(sku dto.SkuDTO, specName string) string {
	if sku.Specs != nil {
		if v := strings.TrimSpace(sku.Specs[specName]); v != "" {
			return v
		}
		for _, v := range sku.Specs {
			if strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
	}
	return strings.TrimSpace(sku.SkuCode)
}

func uniqueNonEmpty(urls []string) []string {
	out := make([]string, 0, len(urls))
	seen := map[string]struct{}{}
	for _, raw := range urls {
		u := strings.TrimSpace(raw)
		if u == "" {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

func trimDouyinTitle(s string, max int) string {
	n := 0
	var b strings.Builder
	for _, r := range s {
		w := 1
		if r > 127 {
			w = 2
		}
		if n+w > max {
			break
		}
		n += w
		b.WriteRune(r)
	}
	return b.String()
}
