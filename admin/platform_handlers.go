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

type PlatformTypeHandler struct {
	svc *service.PlatformTypeService
}

func NewPlatformTypeHandler(svc *service.PlatformTypeService) *PlatformTypeHandler {
	return &PlatformTypeHandler{svc: svc}
}

// List godoc
//
//	@Summary		平台类型列表
//	@Tags			admin-平台类型
//	@Produce		json
//	@Security		BearerAuth
//	@Param			keyword	query	string	false	"关键字"
//	@Success		200		{object}	response.PlatformTypeListResp
//	@Router			/admin/platform-types [get]
func (h *PlatformTypeHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Query("keyword"))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}

// ListEnabled godoc
//
//	@Summary		已启用平台类型
//	@Tags			admin-平台类型
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.PlatformTypeListResp
//	@Router			/admin/platform-types/enabled [get]
func (h *PlatformTypeHandler) ListEnabled(c *gin.Context) {
	list, err := h.svc.ListEnabled()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}

// Create godoc
//
//	@Summary		创建平台类型
//	@Tags			admin-平台类型
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.PlatformShopTypeDTO	true	"平台类型"
//	@Success		201		{object}	response.PlatformTypeResp
//	@Router			/admin/platform-types [post]
func (h *PlatformTypeHandler) Create(c *gin.Context) {
	var in dto.PlatformShopTypeDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.svc.Create(&in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

// Update godoc
//
//	@Summary		更新平台类型
//	@Tags			admin-平台类型
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"平台类型 ID"
//	@Param			body	body		dto.PlatformShopTypeDTO		true	"平台类型"
//	@Success		200		{object}	response.PlatformTypeResp
//	@Router			/admin/platform-types/{id} [put]
func (h *PlatformTypeHandler) Update(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.PlatformShopTypeDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.svc.Update(id, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

// Delete godoc
//
//	@Summary		删除平台类型
//	@Tags			admin-平台类型
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"平台类型 ID"
//	@Success		200	{object}	response.EmptyResp
//	@Router			/admin/platform-types/{id} [delete]
func (h *PlatformTypeHandler) Delete(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, nil)
}

type PlatformShopHandler struct {
	svc        *service.PlatformShopService
	listingSvc *service.PlatformListingService
}

func NewPlatformShopHandler(svc *service.PlatformShopService, listingSvc *service.PlatformListingService) *PlatformShopHandler {
	return &PlatformShopHandler{svc: svc, listingSvc: listingSvc}
}

func (h *PlatformShopHandler) shops(c *gin.Context) *service.PlatformShopService {
	return h.svc.ForTenant(authcontext.TenantID(c))
}

func (h *PlatformShopHandler) listings(c *gin.Context) *service.PlatformListingService {
	return h.listingSvc.ForTenant(authcontext.TenantID(c))
}

// List godoc
//
//	@Summary		平台店铺列表
//	@Tags			admin-平台店铺
//	@Produce		json
//	@Security		BearerAuth
//	@Param			keyword			query	string	false	"关键字"
//	@Param			platformTypeId	query	int		false	"平台类型 ID"
//	@Param			status			query	int		false	"状态"
//	@Param			page			query	int		false	"页码"
//	@Param			pageSize		query	int		false	"每页条数"
//	@Success		200				{object}	response.PlatformShopPageResp
//	@Router			/admin/platform-shops [get]
func (h *PlatformShopHandler) List(c *gin.Context) {
	var q dto.PlatformShopQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	list, total, err := h.shops(c).List(q)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, q.Page, q.PageSize))
}

// Create godoc
//
//	@Summary		创建平台店铺
//	@Tags			admin-平台店铺
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.PlatformShopDTO	true	"店铺"
//	@Success		201		{object}	response.PlatformShopResp
//	@Router			/admin/platform-shops [post]
func (h *PlatformShopHandler) Create(c *gin.Context) {
	var in dto.PlatformShopDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.shops(c).Create(&in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

// Update godoc
//
//	@Summary		更新平台店铺
//	@Tags			admin-平台店铺
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int					true	"店铺 ID"
//	@Param			body	body		dto.PlatformShopDTO	true	"店铺"
//	@Success		200		{object}	response.PlatformShopResp
//	@Router			/admin/platform-shops/{id} [put]
func (h *PlatformShopHandler) Update(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.PlatformShopDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.shops(c).Update(id, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

// Delete godoc
//
//	@Summary		删除平台店铺
//	@Tags			admin-平台店铺
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"店铺 ID"
//	@Success		200	{object}	response.EmptyResp
//	@Router			/admin/platform-shops/{id} [delete]
func (h *PlatformShopHandler) Delete(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.shops(c).Delete(id); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, nil)
}

// ListProducts godoc
//
//	@Summary		店铺已铺货商品
//	@Tags			admin-平台店铺
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path	int		true	"店铺 ID"
//	@Param			keyword		query	string	false	"关键字"
//	@Param			page		query	int		false	"页码"
//	@Param			pageSize	query	int		false	"每页条数"
//	@Success		200			{object}	response.ProductPageResp
//	@Router			/admin/platform-shops/{id}/products [get]
func (h *PlatformShopHandler) ListProducts(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var q dto.ProductQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	list, total, err := h.listings(c).ListProductsByShop(id, q.Keyword, q.Page, q.PageSize)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, response.PageResult(list, total, q.Page, q.PageSize))
}
