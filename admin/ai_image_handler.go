package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"productcore/internal/dto"
	"productcore/internal/integrations/aimodel"
	"productcore/internal/pkg/authcontext"
	"productcore/internal/pkg/response"
	"productcore/internal/storage"

	"github.com/gin-gonic/gin"
)

type AIImageHandler struct {
	client *aimodel.Client
	store  storage.Storage
}

func NewAIImageHandler(client *aimodel.Client, store storage.Storage) *AIImageHandler {
	return &AIImageHandler{client: client, store: store}
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
	remote, err := h.client.Retouch(ctx, aimodel.RetouchInput{
		TenantID: authcontext.TenantID(c),
		Prompt:   prompt,
		ImageURL: imageURL,
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
