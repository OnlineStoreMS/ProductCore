package admin

import (
	"net/http"

	"productcore/internal/dto"
	"productcore/internal/pkg/authcontext"
	"productcore/internal/pkg/httputil"
	"productcore/internal/pkg/response"
	"productcore/internal/service"

	"github.com/gin-gonic/gin"
)

type DistributionHandler struct {
	svc *service.DistributionService
}

func NewDistributionHandler(svc *service.DistributionService) *DistributionHandler {
	return &DistributionHandler{svc: svc}
}

func (h *DistributionHandler) forTenant(c *gin.Context) *service.DistributionService {
	return h.svc.ForTenant(authcontext.TenantID(c))
}

// ListShops godoc
//
//	@Summary		铺货店铺列表
//	@Tags			admin-铺货
//	@Produce		json
//	@Security		BearerAuth
//	@Param			keyword			query	string	false	"店铺名称"
//	@Param			platformTypeId	query	int		false	"店铺类型"
//	@Param			page			query	int		false	"页码"
//	@Param			pageSize		query	int		false	"每页条数"
//	@Success		200				{object}	response.Body
//	@Router			/admin/distribution-shops [get]
func (h *DistributionHandler) ListShops(c *gin.Context) {
	var q dto.DistributionShopQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	list, total, err := h.forTenant(c).ListShops(q)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	page, pageSize := q.Page, q.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

// CreateShop godoc
//
//	@Summary		新增铺货店铺
//	@Tags			admin-铺货
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body	dto.DistributionShopDTO	true	"店铺"
//	@Success		201		{object}	response.Body
//	@Router			/admin/distribution-shops [post]
func (h *DistributionHandler) CreateShop(c *gin.Context) {
	var in dto.DistributionShopDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.forTenant(c).CreateShop(&in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

// UpdateShop godoc
//
//	@Summary		更新铺货店铺
//	@Tags			admin-铺货
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int						true	"店铺 ID"
//	@Param			body	body	dto.DistributionShopDTO	true	"店铺"
//	@Success		200		{object}	response.Body
//	@Router			/admin/distribution-shops/{id} [put]
func (h *DistributionHandler) UpdateShop(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.DistributionShopDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.forTenant(c).UpdateShop(id, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

// DeleteShop godoc
//
//	@Summary		删除铺货店铺
//	@Tags			admin-铺货
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"店铺 ID"
//	@Success		200	{object}	response.EmptyResp
//	@Router			/admin/distribution-shops/{id} [delete]
func (h *DistributionHandler) DeleteShop(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.forTenant(c).DeleteShop(id); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, nil)
}

// GetShop godoc
//
//	@Summary		铺货店铺详情
//	@Tags			admin-铺货
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"店铺 ID"
//	@Success		200	{object}	response.Body
//	@Router			/admin/distribution-shops/{id} [get]
func (h *DistributionHandler) GetShop(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.forTenant(c).GetShop(id)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

// ListItems godoc
//
//	@Summary		铺货店铺商品
//	@Tags			admin-铺货
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path	int		true	"店铺 ID"
//	@Param			keyword		query	string	false	"标题或商品 ID"
//	@Param			collected	query	string	false	"是否采集：1 是，0 否"
//	@Param			sortBy		query	string	false	"排序字段：sales price monthDeals monthConsign shipTime listedAt"
//	@Param			sortOrder	query	string	false	"asc 或 desc，默认 desc"
//	@Param			page		query	int		false	"页码"
//	@Param			pageSize	query	int		false	"每页条数"
//	@Success		200			{object}	response.Body
//	@Router			/admin/distribution-shops/{id}/items [get]
func (h *DistributionHandler) ListItems(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var q dto.DistributionItemQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	list, total, err := h.forTenant(c).ListItems(id, q)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	page, pageSize := q.Page, q.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

// ImportItems godoc
//
//	@Summary		导入铺货店铺商品
//	@Description	上传至尊宝导出的 Excel（HTML 表格）。以本次表格为准整表替换：表内商品覆盖更新，表外商品删除。
//	@Tags			admin-铺货
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int		true	"店铺 ID"
//	@Param			file	formData	file	true	"至尊宝导出 Excel"
//	@Success		200		{object}	response.Body
//	@Router			/admin/distribution-shops/{id}/items/import [post]
func (h *DistributionHandler) ImportItems(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "请上传文件")
		return
	}
	result, err := h.forTenant(c).ImportItems(id, file)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, result)
}
