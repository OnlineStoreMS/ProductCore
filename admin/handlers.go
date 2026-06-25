package admin

import (
	"net/http"
	"strings"

	"productcore/internal/dto"
	"productcore/internal/pkg/httputil"
	"productcore/internal/pkg/response"
	"productcore/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	svc       *service.ProductService
	listingSvc *service.PlatformListingService
}

func NewProductHandler(svc *service.ProductService, listingSvc *service.PlatformListingService) *ProductHandler {
	return &ProductHandler{svc: svc, listingSvc: listingSvc}
}

// List godoc
//
//	@Summary		商品列表
//	@Description	分页查询正式商品（不含草稿）
//	@Tags			admin-商品
//	@Produce		json
//	@Security		BearerAuth
//	@Param			keyword			query		string	false	"关键字"
//	@Param			brandId			query		int		false	"品牌 ID"
//	@Param			categoryId		query		int		false	"分类 ID"
//	@Param			groupId			query		int		false	"分组 ID"
//	@Param			publishStatus	query		int		false	"上架状态 0/1"
//	@Param			page			query		int		false	"页码"
//	@Param			pageSize		query		int		false	"每页条数"
//	@Success		200				{object}	response.ProductPageResp
//	@Failure		400				{object}	response.Body
//	@Failure		500				{object}	response.Body
//	@Router			/admin/products [get]
func (h *ProductHandler) List(c *gin.Context) {
	var q dto.ProductQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	list, total, err := h.svc.List(q)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.listingSvc.AttachListedShops(list); err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, q.Page, q.PageSize))
}

// SuperSearch godoc
//
//	@Summary		超级搜索（SKU 规格值 / 编码 / 商品信息）
//	@Description	在全库 SKU 中搜索规格值、编码、商品名称、货号、资料编码等，返回 SKU 图片、价格、库存、铺货店铺
//	@Tags			admin-商品
//	@Produce		json
//	@Security		BearerAuth
//	@Param			keyword		query	string	true	"搜索关键字"
//	@Param			page		query	int		false	"页码"
//	@Param			pageSize	query	int		false	"每页条数"
//	@Success		200			{object}	response.Body
//	@Router			/admin/super-search [get]
func (h *ProductHandler) SuperSearch(c *gin.Context) {
	var q dto.SuperSearchQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(q.Keyword) == "" {
		response.Fail(c, http.StatusBadRequest, "请输入搜索关键字")
		return
	}
	list, total, err := h.svc.SuperSearch(q)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, q.Page, q.PageSize))
}

// Get godoc
//
//	@Summary		商品详情
//	@Tags			admin-商品
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"商品 ID"
//	@Success		200	{object}	response.ProductResp
//	@Failure		400	{object}	response.Body
//	@Failure		404	{object}	response.Body
//	@Router			/admin/products/{id} [get]
func (h *ProductHandler) Get(c *gin.Context) {
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
	response.OK(c, item)
}

// Create godoc
//
//	@Summary		创建商品
//	@Tags			admin-商品
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.ProductDTO	true	"商品"
//	@Success		201		{object}	response.ProductResp
//	@Failure		400		{object}	response.Body
//	@Router			/admin/products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var in dto.ProductDTO
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

// CreateDraft godoc
//
//	@Summary		创建草稿商品
//	@Description	分配商品 ID，进入草稿箱
//	@Tags			admin-商品
//	@Produce		json
//	@Security		BearerAuth
//	@Success		201	{object}	response.ProductResp
//	@Failure		500	{object}	response.Body
//	@Router			/admin/products/draft [post]
func (h *ProductHandler) CreateDraft(c *gin.Context) {
	item, err := h.svc.CreateDraft()
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

// Update godoc
//
//	@Summary		更新商品
//	@Description	finalize=false 自动保存至草稿箱；finalize=true 完成保存并移出草稿箱
//	@Tags			admin-商品
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int				true	"商品 ID"
//	@Param			body	body		dto.ProductDTO	true	"商品"
//	@Success		200		{object}	response.ProductResp
//	@Failure		400		{object}	response.Body
//	@Router			/admin/products/{id} [put]
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.ProductDTO
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
//	@Summary		删除商品（移入回收站）
//	@Tags			admin-商品
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"商品 ID"
//	@Success		200	{object}	response.EmptyResp
//	@Failure		400	{object}	response.Body
//	@Router			/admin/products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
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

// BatchDelete godoc
//
//	@Summary		批量删除商品
//	@Description	软删除，移入回收站
//	@Tags			admin-商品
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.BatchIDsRequest	true	"商品 ID 列表"
//	@Success		200		{object}	response.BatchResultResp
//	@Failure		400		{object}	response.Body
//	@Router			/admin/products/batch/delete [post]
func (h *ProductHandler) BatchDelete(c *gin.Context) {
	var body dto.BatchIDsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	success, failed := h.svc.BatchDelete(body.IDs)
	response.OK(c, dto.BatchResult{Success: success, Failed: failed})
}

// BatchRestore godoc
//
//	@Summary		批量恢复商品
//	@Tags			admin-商品
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.BatchIDsRequest	true	"商品 ID 列表"
//	@Success		200		{object}	response.BatchResultResp
//	@Failure		400		{object}	response.Body
//	@Router			/admin/products/batch/restore [post]
func (h *ProductHandler) BatchRestore(c *gin.Context) {
	var body dto.BatchIDsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	success, failed := h.svc.BatchRestore(body.IDs)
	response.OK(c, dto.BatchResult{Success: success, Failed: failed})
}

// BatchForceDelete godoc
//
//	@Summary		批量彻底删除商品
//	@Tags			admin-商品
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.BatchIDsRequest	true	"商品 ID 列表"
//	@Success		200		{object}	response.BatchResultResp
//	@Failure		400		{object}	response.Body
//	@Router			/admin/products/batch/force-delete [post]
func (h *ProductHandler) BatchForceDelete(c *gin.Context) {
	var body dto.BatchIDsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	success, failed := h.svc.BatchForceDelete(body.IDs)
	response.OK(c, dto.BatchResult{Success: success, Failed: failed})
}

// ListTrash godoc
//
//	@Summary		回收站商品列表
//	@Tags			admin-商品
//	@Produce		json
//	@Security		BearerAuth
//	@Param			keyword		query	string	false	"关键字"
//	@Param			brandId		query	int		false	"品牌 ID"
//	@Param			categoryId	query	int		false	"分类 ID"
//	@Param			page		query	int		false	"页码"
//	@Param			pageSize	query	int		false	"每页条数"
//	@Success		200			{object}	response.ProductPageResp
//	@Router			/admin/products/trash [get]
func (h *ProductHandler) ListTrash(c *gin.Context) {
	var q dto.ProductQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	list, total, err := h.svc.ListTrash(q)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, q.Page, q.PageSize))
}

// ListDrafts godoc
//
//	@Description	含新建草稿（is_draft=1）与已发布商品的未合并编辑草稿
//	@Tags			admin-商品
//	@Produce		json
//	@Security		BearerAuth
//	@Param			keyword		query	string	false	"关键字"
//	@Param			page		query	int		false	"页码"
//	@Param			pageSize	query	int		false	"每页条数"
//	@Success		200			{object}	response.ProductPageResp
//	@Router			/admin/products/drafts [get]
func (h *ProductHandler) ListDrafts(c *gin.Context) {
	var q dto.ProductQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	list, total, err := h.svc.ListDrafts(q)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, q.Page, q.PageSize))
}

// Restore godoc
//
//	@Summary		恢复商品
//	@Tags			admin-商品
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"商品 ID"
//	@Success		200	{object}	response.EmptyResp
//	@Router			/admin/products/{id}/restore [post]
func (h *ProductHandler) Restore(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Restore(id); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, nil)
}

// ForceDelete godoc
//
//	@Summary		彻底删除商品
//	@Tags			admin-商品
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"商品 ID"
//	@Success		200	{object}	response.EmptyResp
//	@Router			/admin/products/{id}/force [delete]
func (h *ProductHandler) ForceDelete(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.ForceDelete(id); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, nil)
}

// GetSkus godoc
//
//	@Summary		获取商品 SKU（轻量）
//	@Description	仅返回 SKU 管理所需字段，不含详情页、媒体等大字段
//	@Tags			admin-商品
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"商品 ID"
//	@Success		200	{object}	response.ProductSkusResp
//	@Failure		404	{object}	response.Body
//	@Router			/admin/products/{id}/skus [get]
func (h *ProductHandler) GetSkus(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.svc.GetSkus(id)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

// UpdateSkus godoc
//
//	@Summary		更新商品 SKU
//	@Tags			admin-商品
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"商品 ID"
//	@Param			body	body		dto.UpdateSkusRequest	true	"SKU 列表"
//	@Success		200		{object}	response.ProductSkusResp
//	@Router			/admin/products/{id}/skus [put]
func (h *ProductHandler) UpdateSkus(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var body dto.UpdateSkusRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.svc.UpdateSkus(id, body.Skus, body.SkuSpecs)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

// UpdatePublishStatus godoc
//
//	@Summary		更新上架状态
//	@Tags			admin-商品
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int								true	"商品 ID"
//	@Param			body	body		dto.UpdatePublishStatusRequest	true	"上架状态"
//	@Success		200		{object}	response.EmptyResp
//	@Router			/admin/products/{id}/publish-status [patch]
func (h *ProductHandler) UpdatePublishStatus(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var body dto.UpdatePublishStatusRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.UpdatePublishStatus(id, body.PublishStatus); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, nil)
}

// DiscardEditDraft godoc
//
//	@Summary		放弃未发布的编辑草稿
//	@Tags			admin-商品
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"商品 ID"
//	@Success		200	{object}	response.EmptyResp
//	@Failure		404	{object}	response.Body
//	@Router			/admin/products/{id}/edit-draft [delete]
func (h *ProductHandler) DiscardEditDraft(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.DiscardEditDraft(id); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, nil)
}

// GetListings godoc
//
//	@Summary		查询商品已铺货店铺
//	@Tags			admin-商品
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"商品 ID"
//	@Success		200	{object}	response.ListedShopListResp
//	@Router			/admin/products/{id}/listings [get]
func (h *ProductHandler) GetListings(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	list, err := h.listingSvc.ListShopsByProduct(id)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, list)
}

// SetListings godoc
//
//	@Summary		设置商品铺货店铺
//	@Tags			admin-商品
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int								true	"商品 ID"
//	@Param			body	body		dto.UpdateProductListingsRequest	true	"店铺 ID 列表"
//	@Success		200		{object}	response.ListedShopListResp
//	@Router			/admin/products/{id}/listings [put]
func (h *ProductHandler) SetListings(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.UpdateProductListingsRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	list, err := h.listingSvc.SetProductListings(id, in.ShopIDs)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, list)
}

type BrandHandler struct {
	svc *service.BrandService
}

func NewBrandHandler(svc *service.BrandService) *BrandHandler {
	return &BrandHandler{svc: svc}
}

// List godoc
//
//	@Summary		品牌列表
//	@Tags			admin-品牌
//	@Produce		json
//	@Security		BearerAuth
//	@Param			keyword	query	string	false	"关键字"
//	@Success		200		{object}	response.BrandListResp
//	@Router			/admin/brands [get]
func (h *BrandHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Query("keyword"))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}

// Create godoc
//
//	@Summary		创建品牌
//	@Tags			admin-品牌
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.BrandDTO	true	"品牌"
//	@Success		201		{object}	response.BrandResp
//	@Router			/admin/brands [post]
func (h *BrandHandler) Create(c *gin.Context) {
	var in dto.BrandDTO
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
//	@Summary		更新品牌
//	@Tags			admin-品牌
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int				true	"品牌 ID"
//	@Param			body	body		dto.BrandDTO	true	"品牌"
//	@Success		200		{object}	response.BrandResp
//	@Router			/admin/brands/{id} [put]
func (h *BrandHandler) Update(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.BrandDTO
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
//	@Summary		删除品牌
//	@Tags			admin-品牌
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"品牌 ID"
//	@Success		200	{object}	response.EmptyResp
//	@Router			/admin/brands/{id} [delete]
func (h *BrandHandler) Delete(c *gin.Context) {
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

type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// Tree godoc
//
//	@Summary		分类树
//	@Tags			admin-分类
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.CategoryTreeResp
//	@Router			/admin/categories/tree [get]
func (h *CategoryHandler) Tree(c *gin.Context) {
	tree, err := h.svc.Tree()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, tree)
}

// Create godoc
//
//	@Summary		创建分类
//	@Tags			admin-分类
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.CategoryDTO	true	"分类"
//	@Success		201		{object}	response.CategoryResp
//	@Router			/admin/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var in dto.CategoryDTO
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
//	@Summary		更新分类
//	@Tags			admin-分类
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int					true	"分类 ID"
//	@Param			body	body		dto.CategoryDTO		true	"分类"
//	@Success		200		{object}	response.CategoryResp
//	@Router			/admin/categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.CategoryDTO
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
//	@Summary		删除分类
//	@Tags			admin-分类
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"分类 ID"
//	@Success		200	{object}	response.EmptyResp
//	@Router			/admin/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
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

type GroupHandler struct {
	svc *service.ProductGroupService
}

func NewGroupHandler(svc *service.ProductGroupService) *GroupHandler {
	return &GroupHandler{svc: svc}
}

// List godoc
//
//	@Summary		商品分组列表
//	@Tags			admin-分组
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.ProductGroupListResp
//	@Router			/admin/groups [get]
func (h *GroupHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}

// Create godoc
//
//	@Summary		创建商品分组
//	@Tags			admin-分组
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.ProductGroupDTO	true	"分组"
//	@Success		201		{object}	response.ProductGroupResp
//	@Router			/admin/groups [post]
func (h *GroupHandler) Create(c *gin.Context) {
	var in dto.ProductGroupDTO
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
//	@Summary		更新商品分组
//	@Tags			admin-分组
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"分组 ID"
//	@Param			body	body		dto.ProductGroupDTO		true	"分组"
//	@Success		200		{object}	response.ProductGroupResp
//	@Router			/admin/groups/{id} [put]
func (h *GroupHandler) Update(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.ProductGroupDTO
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
//	@Summary		删除商品分组
//	@Tags			admin-分组
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"分组 ID"
//	@Success		200	{object}	response.EmptyResp
//	@Router			/admin/groups/{id} [delete]
func (h *GroupHandler) Delete(c *gin.Context) {
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

// Products godoc
//
//	@Summary		分组内商品列表
//	@Tags			admin-分组
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"分组 ID"
//	@Success		200	{object}	response.ProductPageResp
//	@Router			/admin/groups/{id}/products [get]
func (h *GroupHandler) Products(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	list, err := h.svc.ListProducts(id)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}
