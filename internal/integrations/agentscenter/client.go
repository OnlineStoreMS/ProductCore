package agentscenter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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

type JobStatus struct {
	ID           uint64 `json:"id"`
	Status       string `json:"status"`
	ErrorMessage string `json:"errorMessage"`
	ResultJSON   string `json:"resultJson"`
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

type ShopItem struct {
	ID               uint64 `json:"id"`
	AgentID          uint64 `json:"agentId"`
	AgentName        string `json:"agentName"`
	Platform         string `json:"platform"`
	PlatformShopID   string `json:"platformShopId"`
	PlatformShopName string `json:"platformShopName"`
	CapabilitiesJSON string `json:"capabilitiesJson"`
	Status           string `json:"status"`
	AgentOnline      bool   `json:"agentOnline"`
}

type CreateJobRequest struct {
	TenantID         uint64 `json:"tenantId"`
	JobType          string `json:"jobType"`
	Platform         string `json:"platform"`
	PlatformShopID   string `json:"platformShopId"`
	PlatformShopName string `json:"platformShopName"`
	ParamsJSON       string `json:"paramsJson"`
	Source           string `json:"source"`
	Priority         int    `json:"priority"`
}

type CreatedJob struct {
	ID     uint64 `json:"id"`
	Status string `json:"status"`
}

func (c *Client) ListShops(tenantID uint64) ([]ShopItem, error) {
	if c == nil {
		return nil, fmt.Errorf("AgentsCenter 未配置")
	}
	q := url.Values{}
	q.Set("tenantId", fmt.Sprintf("%d", tenantID))
	q.Set("page", "1")
	q.Set("pageSize", "100")
	var envelope struct {
		Data struct {
			List []ShopItem `json:"list"`
		} `json:"data"`
	}
	if err := c.get("/api/v1/internal/shops?"+q.Encode(), &envelope); err != nil {
		return nil, err
	}
	return envelope.Data.List, nil
}

func (c *Client) CreateJob(in CreateJobRequest) (*CreatedJob, error) {
	if c == nil {
		return nil, fmt.Errorf("AgentsCenter 未配置")
	}
	var envelope struct {
		Data CreatedJob `json:"data"`
	}
	if err := c.post("/api/v1/internal/jobs", in, &envelope); err != nil {
		return nil, err
	}
	if envelope.Data.ID == 0 {
		return nil, fmt.Errorf("AgentsCenter 未返回任务 ID")
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

func (c *Client) post(path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, strings.NewReader(string(b)))
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
