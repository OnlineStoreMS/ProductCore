package admin

import (
	"net/http"

	"productcore/internal/dto"
	"productcore/internal/pkg/authcontext"
	"productcore/internal/pkg/response"
	"productcore/internal/service"

	"github.com/gin-gonic/gin"
)

type AISettingHandler struct {
	svc *service.AISettingService
}

func NewAISettingHandler(svc *service.AISettingService) *AISettingHandler {
	return &AISettingHandler{svc: svc}
}

func (h *AISettingHandler) svcFor(c *gin.Context) *service.AISettingService {
	return h.svc.ForTenant(authcontext.TenantID(c))
}

// GetAISetting godoc
//
//	@Summary		AI 修图默认模型
//	@Tags			admin-AI
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Body
//	@Router			/admin/ai/settings [get]
func (h *AISettingHandler) Get(c *gin.Context) {
	out, err := h.svcFor(c).Get(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, out)
}

// UpdateAISetting godoc
//
//	@Summary		保存 AI 修图默认模型
//	@Tags			admin-AI
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.UpdateAISettingRequest	true	"默认图片模型"
//	@Success		200		{object}	response.Body
//	@Router			/admin/ai/settings [put]
func (h *AISettingHandler) Update(c *gin.Context) {
	var in dto.UpdateAISettingRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数不正确")
		return
	}
	out, err := h.svcFor(c).Update(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, out)
}

// ListImageModels godoc
//
//	@Summary		可用的图片模型
//	@Description	读取 AI 大模型管理中心里已启用的生图/修图模型
//	@Tags			admin-AI
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Body
//	@Router			/admin/ai/image-models [get]
func (h *AISettingHandler) ListImageModels(c *gin.Context) {
	list, err := h.svcFor(c).ListImageModels(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, list)
}
