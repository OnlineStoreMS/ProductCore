package api

import (
	"net/http"

	"productcore/internal/pkg/response"
	"productcore/internal/service"

	"github.com/gin-gonic/gin"
)

// CategoryHandler 对外 Open API（只读分类树）
type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// VisibleTree godoc
//
//	@Summary		可见分类树（只读）
//	@Tags			open-分类
//	@Produce		json
//	@Success		200	{object}	response.CategoryTreeResp
//	@Router			/open/categories/tree [get]
func (h *CategoryHandler) VisibleTree(c *gin.Context) {
	list, err := h.svc.VisibleTree()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}
