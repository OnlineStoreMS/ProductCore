package admin

import (
	"net/http"
	"strconv"

	"productcore/internal/dto"
	"productcore/internal/pkg/authcontext"
	"productcore/internal/pkg/httputil"
	"productcore/internal/pkg/response"
	"productcore/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductCollectHandler struct {
	svc *service.ProductCollectService
}

func NewProductCollectHandler(svc *service.ProductCollectService) *ProductCollectHandler {
	return &ProductCollectHandler{svc: svc}
}

// Ingest godoc
//
//	@Summary		扩展回传入库
//	@Description	浏览器扩展在人工点至尊宝工具后，把汇总的商品草稿写入商品管理系统
//	@Tags			admin-采集
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.IngestProductCollectRequest	true	"采集结果"
//	@Success		200		{object}	response.ProductCollectTaskResp
//	@Failure		400		{object}	response.Body
//	@Failure		500		{object}	response.Body
//	@Router			/admin/product-collects/ingest [post]
func (h *ProductCollectHandler) Ingest(c *gin.Context) {
	var in dto.IngestProductCollectRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, "请提交采集到的商品数据")
		return
	}
	task, err := h.svc.IngestFromExtension(authcontext.TenantID(c), authcontext.UserID(c), in.ProductURL, in.Product, in.Videos)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, task)
}

// List godoc
//
//	@Summary		采集任务列表
//	@Description	查看本租户的商品采集任务
//	@Tags			admin-采集
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"页码"
//	@Param			pageSize	query		int	false	"每页条数"
//	@Success		200			{object}	response.ProductCollectPageResp
//	@Failure		500			{object}	response.Body
//	@Router			/admin/product-collects [get]
func (h *ProductCollectHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := h.svc.List(authcontext.TenantID(c), page, pageSize)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

// Clear godoc
//
//	@Summary		清空采集记录
//	@Description	删除本租户全部商品采集记录，不删除已生成的商品
//	@Tags			admin-采集
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.EmptyResp
//	@Failure		500	{object}	response.Body
//	@Router			/admin/product-collects [delete]
func (h *ProductCollectHandler) Clear(c *gin.Context) {
	if err := h.svc.Clear(authcontext.TenantID(c)); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, nil)
}
