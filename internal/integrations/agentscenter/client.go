package agentscenter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const JobTypeEcommerceProductCollect = "ecommerce.product.collect"

type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

func NewClient(baseURL, token string) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" || strings.TrimSpace(token) == "" {
		return nil
	}
	return &Client{
		BaseURL: baseURL,
		Token:   strings.TrimSpace(token),
		HTTP:    &http.Client{Timeout: 15 * time.Second},
	}
}

type Agent struct {
	ID            uint64  `json:"id"`
	Name          string  `json:"name"`
	Status        string  `json:"status"`
	SkillsJSON    string  `json:"skillsJson"`
	LastHeartbeat *string `json:"lastHeartbeat"`
}

type CreatedJob struct {
	ID uint64 `json:"id"`
}

type JobStatus struct {
	ID           uint64 `json:"id"`
	Status       string `json:"status"`
	ErrorMessage string `json:"errorMessage"`
	ResultJSON   string `json:"resultJson"`
}

func (c *Client) ListAgents(tenantID uint64, onlineOnly bool, skill string) ([]Agent, error) {
	if c == nil {
		return nil, fmt.Errorf("AgentsCenter 未配置")
	}
	q := url.Values{}
	q.Set("tenantId", fmt.Sprintf("%d", tenantID))
	q.Set("page", "1")
	q.Set("pageSize", "200")
	if onlineOnly {
		q.Set("onlineOnly", "1")
	}
	if strings.TrimSpace(skill) != "" {
		q.Set("skill", strings.TrimSpace(skill))
	}
	var envelope struct {
		Data struct {
			List []Agent `json:"list"`
		} `json:"data"`
	}
	if err := c.get("/api/v1/internal/agents?"+q.Encode(), &envelope); err != nil {
		return nil, err
	}
	return envelope.Data.List, nil
}

func (c *Client) GetJobs(tenantID uint64, ids []uint64) ([]JobStatus, error) {
	if c == nil {
		return nil, fmt.Errorf("AgentsCenter 未配置")
	}
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			parts = append(parts, fmt.Sprintf("%d", id))
		}
	}
	if len(parts) == 0 {
		return nil, nil
	}
	q := url.Values{}
	q.Set("tenantId", fmt.Sprintf("%d", tenantID))
	q.Set("ids", strings.Join(parts, ","))
	var envelope struct {
		Data struct {
			List []JobStatus `json:"list"`
		} `json:"data"`
	}
	if err := c.get("/api/v1/internal/jobs?"+q.Encode(), &envelope); err != nil {
		return nil, err
	}
	return envelope.Data.List, nil
}

func (c *Client) CreateProductCollectJob(tenantID, targetAgentID uint64, platform, paramsJSON string) (*CreatedJob, error) {
	if c == nil {
		return nil, fmt.Errorf("AgentsCenter 未配置")
	}
	if targetAgentID == 0 {
		return nil, fmt.Errorf("targetAgentId 必填")
	}
	if strings.TrimSpace(platform) == "" {
		platform = "taobao"
	}
	aid := targetAgentID
	body := map[string]any{
		"tenantId":         tenantID,
		"jobType":          JobTypeEcommerceProductCollect,
		"platform":         platform,
		"platformShopId":   "product-collect",
		"platformShopName": "电商商品采集",
		"paramsJson":       paramsJSON,
		"source":           "productcore",
		"priority":         80,
		"targetAgentId":    aid,
	}
	raw, _ := json.Marshal(body)
	var envelope struct {
		Data CreatedJob `json:"data"`
	}
	if err := c.post("/api/v1/internal/jobs", raw, &envelope); err != nil {
		return nil, err
	}
	if envelope.Data.ID == 0 {
		return nil, fmt.Errorf("AgentsCenter 未返回任务编号")
	}
	return &envelope.Data, nil
}

func (c *Client) get(path string, out any) error {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *Client) post(path string, raw []byte, out any) error {
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out any) error {
	req.Header.Set("X-Internal-Token", c.Token)
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return fmt.Errorf("AgentsCenter HTTP %d: %s", res.StatusCode, strings.TrimSpace(string(b)))
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(b, out); err != nil {
		return err
	}
	return nil
}
