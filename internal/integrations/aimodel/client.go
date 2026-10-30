package aimodel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		HTTP:    &http.Client{Timeout: 180 * time.Second},
	}
}

type RetouchInput struct {
	TenantID     uint64
	Model        string
	ConnectionID uint64
	Prompt       string
	ImageURL     string
	Size         string
}

type ImageModel struct {
	ID             uint64 `json:"id"`
	ConnectionID   uint64 `json:"connectionId"`
	ConnectionName string `json:"connectionName"`
	ProviderCode   string `json:"providerCode"`
	ModelID        string `json:"modelId"`
	DisplayName    string `json:"displayName"`
	IsDefault      int    `json:"isDefault"`
}

type retouchResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		URL     string `json:"url"`
		B64JSON string `json:"b64Json"`
	} `json:"data"`
}

func (c *Client) ListImageModels(ctx context.Context, tenantID uint64) ([]ImageModel, error) {
	if c == nil {
		return nil, fmt.Errorf("未配置 AI 大模型")
	}
	url := fmt.Sprintf("%s/api/v1/internal/models?tenantId=%d&capability=image", c.BaseURL, tenantID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Internal-Token", c.Token)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("读取图片模型失败")
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	var body struct {
		Code    int          `json:"code"`
		Message string       `json:"message"`
		Data    []ImageModel `json:"data"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("图片模型返回异常")
	}
	if resp.StatusCode >= 300 || body.Code >= 300 {
		msg := strings.TrimSpace(body.Message)
		if msg == "" {
			msg = "读取图片模型失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}
	if body.Data == nil {
		return []ImageModel{}, nil
	}
	return body.Data, nil
}

func (c *Client) Retouch(ctx context.Context, in RetouchInput) (string, error) {
	if c == nil {
		return "", fmt.Errorf("AI 修图未配置")
	}
	model := strings.TrimSpace(in.Model)
	if model == "" {
		model = "qwen-image-3.0"
	}
	payload := map[string]any{
		"tenantId":  in.TenantID,
		"callerApp": "productcore",
		"model":     model,
		"prompt":    in.Prompt,
		"imageUrl":  in.ImageURL,
	}
	if in.ConnectionID > 0 {
		payload["connectionId"] = in.ConnectionID
	}
	if size := strings.TrimSpace(in.Size); size != "" {
		payload["size"] = size
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/v1/internal/images", bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", c.Token)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("调用修图服务失败")
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	var body retouchResponse
	if err := json.Unmarshal(raw, &body); err != nil {
		return "", fmt.Errorf("修图服务返回异常")
	}
	if resp.StatusCode >= 300 || body.Code >= 300 {
		msg := strings.TrimSpace(body.Message)
		if msg == "" {
			msg = "修图失败"
		}
		return "", fmt.Errorf("%s", msg)
	}
	if url := strings.TrimSpace(body.Data.URL); url != "" {
		return url, nil
	}
	return "", fmt.Errorf("模型未返回图片")
}

type ChatInput struct {
	TenantID  uint64
	Model     string
	System    string
	User      string
	MaxTokens int
}

func (c *Client) Chat(ctx context.Context, in ChatInput) (string, error) {
	if c == nil {
		return "", fmt.Errorf("未配置 AI 大模型")
	}
	thinking := false
	temp := 0.3
	sys, err := json.Marshal(in.System)
	if err != nil {
		return "", err
	}
	user, err := json.Marshal(in.User)
	if err != nil {
		return "", err
	}
	payload := map[string]any{
		"tenantId":    in.TenantID,
		"callerApp":   "productcore",
		"capability":  "chat",
		"model":       in.Model,
		"temperature": temp,
		"maxTokens":   in.MaxTokens,
		"thinking":    thinking,
		"messages": []map[string]any{
			{"role": "system", "content": json.RawMessage(sys)},
			{"role": "user", "content": json.RawMessage(user)},
		},
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/v1/internal/chat", bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", c.Token)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("调用 DeepSeek 失败")
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Content string `json:"content"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return "", fmt.Errorf("DeepSeek 返回异常")
	}
	if resp.StatusCode >= 300 || body.Code >= 300 {
		msg := strings.TrimSpace(body.Message)
		if msg == "" {
			msg = "标题优化失败"
		}
		return "", fmt.Errorf("%s", msg)
	}
	text := strings.TrimSpace(body.Data.Content)
	if text == "" {
		return "", fmt.Errorf("DeepSeek 没有返回标题")
	}
	return text, nil
}
