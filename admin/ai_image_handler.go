package admin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"encoding/json"
	"os"

	"productcore/internal/dto"
	"productcore/internal/integrations/aimodel"
	"productcore/internal/integrations/imageerase"
	"productcore/internal/pkg/authcontext"
	"productcore/internal/pkg/imagesize"
	"productcore/internal/pkg/response"
	"productcore/internal/storage"

	"github.com/gin-gonic/gin"
)

type AIImageHandler struct {
	client *aimodel.Client
	erase  *imageerase.Client
	store  storage.Storage
}

func NewAIImageHandler(client *aimodel.Client, erase *imageerase.Client, store storage.Storage) *AIImageHandler {
	return &AIImageHandler{client: client, erase: erase, store: store}
}

// Retouch godoc
//
//	@Summary		AI 修图
//	@Description	用提示词修改商品图，结果写入存储后返回可保存的地址
//	@Tags			admin-上传
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.AIRetouchRequest	true	"原图、提示词与上传上下文"
//	@Success		200		{object}	response.UploadURLResp
//	@Failure		400		{object}	response.Body
//	@Failure		502		{object}	response.Body
//	@Failure		503		{object}	response.Body
//	@Router			/admin/ai/images/retouch [post]
func (h *AIImageHandler) Retouch(c *gin.Context) {
	if h.client == nil {
		response.Fail(c, http.StatusServiceUnavailable, "未配置 AI 修图")
		return
	}
	var body dto.AIRetouchRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数不正确")
		return
	}
	prompt := strings.TrimSpace(body.Prompt)
	if prompt == "" || utf8.RuneCountInString(prompt) > 2000 {
		response.Fail(c, http.StatusBadRequest, "提示词需在 1 到 2000 字")
		return
	}
	imageURL := strings.TrimSpace(body.ImageURL)
	if h.store != nil {
		imageURL = h.store.ResolvePublicURL(imageURL)
	}
	if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		response.Fail(c, http.StatusBadRequest, "原图地址无效")
		return
	}
	opts := storage.UploadOptionsFromForm(map[string]string{
		"scope":     body.Scope,
		"resource":  body.Resource,
		"productId": fmt.Sprintf("%d", body.ProductID),
		"skuId":     fmt.Sprintf("%d", body.SkuID),
	}, false)
	if err := opts.Validate(); err != nil {
		if errors.Is(err, storage.ErrProductIDRequired) {
			response.Fail(c, http.StatusBadRequest, "请先保存商品后再修图")
			return
		}
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 170*time.Second)
	defer cancel()
	size, err := retouchOutputSize(ctx, h.store, imageURL)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	remote, err := h.client.Retouch(ctx, aimodel.RetouchInput{
		TenantID: authcontext.TenantID(c),
		Prompt:   prompt,
		ImageURL: imageURL,
		Size:     size,
	})
	if err != nil {
		response.Fail(c, http.StatusBadGateway, err.Error())
		return
	}
	stored, err := storage.UploadFromURL(h.store, remote, opts)
	if err != nil {
		response.Fail(c, http.StatusBadGateway, "结果图保存失败")
		return
	}
	response.OK(c, gin.H{"url": stored})
}

func (h *AIImageHandler) Snap(c *gin.Context) {
	if h.erase == nil {
		response.Fail(c, http.StatusServiceUnavailable, "未配置图片擦除")
		return
	}
	var body dto.ImageEraseRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数不正确")
		return
	}
	raw, err := h.readSourceImage(c.Request.Context(), body.ImageURL)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	out, err := h.erase.Snap(ctx, raw, jsonText(body.Boxes, "[]"), jsonText(body.Polygons, "[]"))
	if err != nil {
		response.Fail(c, http.StatusBadGateway, err.Error())
		return
	}
	var regions map[string]any
	if err := json.Unmarshal(out, &regions); err != nil {
		response.Fail(c, http.StatusBadGateway, "擦除服务返回异常")
		return
	}
	response.OK(c, regions)
}

func (h *AIImageHandler) Erase(c *gin.Context) {
	if h.erase == nil {
		response.Fail(c, http.StatusServiceUnavailable, "未配置图片擦除")
		return
	}
	var body dto.ImageEraseRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数不正确")
		return
	}
	opts := storage.UploadOptionsFromForm(map[string]string{
		"scope":     body.Scope,
		"resource":  body.Resource,
		"productId": fmt.Sprintf("%d", body.ProductID),
		"skuId":     fmt.Sprintf("%d", body.SkuID),
	}, false)
	if err := opts.Validate(); err != nil {
		if errors.Is(err, storage.ErrProductIDRequired) {
			response.Fail(c, http.StatusBadRequest, "请先保存商品后再擦除")
			return
		}
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	raw, err := h.readSourceImage(c.Request.Context(), body.ImageURL)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	png, err := h.erase.Erase(ctx, raw, jsonText(body.Boxes, "[]"), jsonText(body.Polygons, "[]"), jsonText(body.Shapes, "[]"))
	if err != nil {
		response.Fail(c, http.StatusBadGateway, err.Error())
		return
	}
	tmp, err := os.CreateTemp("", "image-erase-*.png")
	if err != nil {
		response.Fail(c, http.StatusBadGateway, "结果图保存失败")
		return
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(png); err != nil {
		tmp.Close()
		response.Fail(c, http.StatusBadGateway, "结果图保存失败")
		return
	}
	if err := tmp.Close(); err != nil {
		response.Fail(c, http.StatusBadGateway, "结果图保存失败")
		return
	}
	stored, err := h.store.UploadPath(tmpPath, "erased.png", opts)
	if err != nil {
		response.Fail(c, http.StatusBadGateway, "结果图保存失败")
		return
	}
	response.OK(c, gin.H{"url": stored})
}

func (h *AIImageHandler) readSourceImage(ctx context.Context, imageURL string) ([]byte, error) {
	imageURL = strings.TrimSpace(imageURL)
	if h.store != nil {
		imageURL = h.store.ResolvePublicURL(imageURL)
	}
	if imageURL == "" {
		return nil, fmt.Errorf("原图地址无效")
	}
	rc, err := openSourceImage(ctx, h.store, imageURL)
	if err != nil {
		return nil, fmt.Errorf("无法读取原图")
	}
	defer rc.Close()
	raw, err := io.ReadAll(io.LimitReader(rc, 15<<20+1))
	if err != nil {
		return nil, fmt.Errorf("无法读取原图")
	}
	if len(raw) == 0 || len(raw) > 15<<20 {
		return nil, fmt.Errorf("原图为空或超过 15MB")
	}
	return raw, nil
}

func jsonText(raw json.RawMessage, fallback string) string {
	if len(raw) == 0 || string(raw) == "null" {
		return fallback
	}
	return string(raw)
}

func retouchOutputSize(ctx context.Context, store storage.Storage, imageURL string) (string, error) {
	rc, err := openSourceImage(ctx, store, imageURL)
	if err != nil {
		return "", fmt.Errorf("无法读取原图尺寸")
	}
	defer rc.Close()
	w, h, err := imagesize.FromReader(io.LimitReader(rc, 32<<20))
	if err != nil || w <= 0 || h <= 0 {
		return "", fmt.Errorf("无法读取原图尺寸")
	}
	return imagesize.QwenSize(w, h), nil
}

func openSourceImage(ctx context.Context, store storage.Storage, imageURL string) (io.ReadCloser, error) {
	if store != nil {
		if rc, err := store.Open(imageURL); err == nil {
			return rc, nil
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return resp.Body, nil
}
