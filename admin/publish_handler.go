package admin

import (
	"errors"
	"net/http"
	"strconv"

	"productcore/internal/dto"
	"productcore/internal/pkg/httputil"
	"productcore/internal/pkg/response"
	"productcore/internal/service"

	"github.com/gin-gonic/gin"
)

// ListPublishShops godoc
//
//	@Summary		可铺货的电商店铺
//	@Description	读取 AgentsCenter 上 WindowsAgent 已登录上报的在线店铺
//	@Tags			admin-商品
//	@Produce		json
//	@Security		BearerAuth
//	@Param			productId	query		int	false	"商品 ID，用于带上该商品在各店的铺货状态"
//	@Success		200			{object}	response.Body
//	@Router			/admin/publish-shops [get]
func (h *ProductHandler) ListPublishShops(c *gin.Context) {
	var productID uint64
	if raw := c.Query("productId"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid productId")
			return
		}
		productID = id
	}
	list, err := h.pub(c).ListShops(productID)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, list)
}

// StartPublish godoc
//
//	@Summary		开始铺货
//	@Description	按商品生成铺货任务。若带了电商标题或主图，选中的店铺使用这些内容
//	@Tags			admin-商品
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"商品 ID"
//	@Param			body	body		dto.StartPublishRequest	true	"店铺"
//	@Success		200		{object}	response.Body
//	@Failure		400		{object}	response.Body
//	@Router			/admin/products/{id}/publish [post]
func (h *ProductHandler) StartPublish(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.StartPublishRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, "请选择店铺")
		return
	}
	item, err := h.pub(c).Start(id, in.Platform, in.PlatformShopID, in.Title, in.MainImages)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Fail(c, http.StatusNotFound, "商品不存在")
			return
		}
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, item)
}

// StartPublishUpdate godoc
//
//	@Summary		更新铺货
//	@Description	按店铺中已有的商品 ID 覆盖该商品。全部信息时若带了电商标题，用该标题覆盖店铺商品
//	@Tags			admin-商品
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int								true	"商品 ID"
//	@Param			body	body		dto.StartPublishUpdateRequest	true	"店铺和平台商品 ID"
//	@Success		200		{object}	response.Body
//	@Failure		400		{object}	response.Body
//	@Router			/admin/products/{id}/publish-update [post]
func (h *ProductHandler) StartPublishUpdate(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.StartPublishUpdateRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, "请选择店铺并填写商品 ID")
		return
	}
	item, err := h.pub(c).StartUpdate(id, in.Platform, in.PlatformShopID, in.PlatformProductID, in.Scope, in.Title, in.MainImages)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Fail(c, http.StatusNotFound, "商品不存在")
			return
		}
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, item)
}

// SuggestListingTitle godoc
//
//	@Summary		生成抖店电商标题
//	@Description	把商品资料交给 DeepSeek，返回一条可用于更新铺货的标题
//	@Tags			admin-商品
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"商品 ID"
//	@Success		200	{object}	response.Body
//	@Failure		400	{object}	response.Body
//	@Router			/admin/products/{id}/listing-title [post]
func (h *ProductHandler) SuggestListingTitle(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	title, err := h.pub(c).SuggestListingTitle(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Fail(c, http.StatusNotFound, "商品不存在")
			return
		}
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, dto.ListingTitleResult{Title: title})
}
