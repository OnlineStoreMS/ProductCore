package api

import (
	"net/http"

	"productcore/internal/dto"
	"productcore/internal/pkg/httputil"
	"productcore/internal/pkg/response"
	"productcore/internal/service"

	"github.com/gin-gonic/gin"
)

// ProductHandler 对外 Open API（只读，供电商/门店读取商品库）
type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (h *ProductHandler) ListPublished(c *gin.Context) {
	var q dto.ProductQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	list, total, err := h.svc.ListPublished(q)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, q.Page, q.PageSize))
}

func (h *ProductHandler) GetPublished(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.svc.Get(id)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	if item.PublishStatus != 1 {
		response.Fail(c, http.StatusNotFound, "product not published")
		return
	}
	response.OK(c, item)
}
