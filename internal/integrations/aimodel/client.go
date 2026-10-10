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
	TenantID uint64
	Prompt   string
	ImageURL string
	Size     string
}

type retouchResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		URL     string `json:"url"`
		B64JSON string `json:"b64Json"`
	} `json:"data"`
}

func (c *Client) Retouch(ctx context.Context, in RetouchInput) (string, error) {
	if c == nil {
		return "", fmt.Errorf("AI 修图未配置")
	}
	payload := map[string]any{
		"tenantId":  in.TenantID,
		"callerApp": "productcore",
		"model":     "qwen-image-3.0",
		"prompt":    in.Prompt,
		"imageUrl":  in.ImageURL,
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
