package imageerase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil
	}
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Snap(ctx context.Context, image []byte, boxes, polygons string) ([]byte, error) {
	return c.post(ctx, "/api/snap", image, map[string]string{
		"boxes":    boxes,
		"polygons": polygons,
	})
}

func (c *Client) Erase(ctx context.Context, image []byte, boxes, polygons, shapes string) ([]byte, error) {
	return c.post(ctx, "/api/erase", image, map[string]string{
		"boxes":    boxes,
		"polygons": polygons,
		"shapes":   shapes,
		"method":   "sharp",
	})
}

func (c *Client) post(ctx context.Context, path string, image []byte, fields map[string]string) ([]byte, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "image.png")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(image); err != nil {
		return nil, err
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用擦除服务失败")
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s", eraseError(payload, resp.Status))
	}
	return payload, nil
}

func eraseError(payload []byte, status string) string {
	var body struct {
		Detail any `json:"detail"`
	}
	if err := json.Unmarshal(payload, &body); err == nil && body.Detail != nil {
		switch detail := body.Detail.(type) {
		case string:
			if strings.TrimSpace(detail) != "" {
				return detail
			}
		default:
			raw, err := json.Marshal(detail)
			if err == nil && len(raw) > 0 && string(raw) != "null" {
				return string(raw)
			}
		}
	}
	msg := strings.TrimSpace(string(payload))
	if msg == "" {
		return status
	}
	return msg
}
