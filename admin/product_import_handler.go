package admin

import (
	"fmt"
	"net/http"

	"productcore/internal/pkg/httputil"
	"productcore/internal/pkg/response"
	"productcore/internal/service"

	"github.com/gin-gonic/gin"
)

const maxImportMultipart = 310 << 20 // zip 300MB + form overhead

type ProductImportHandler struct {
	importSvc *service.ProductImportService
	exportSvc *service.ProductExportService
}

func NewProductImportHandler(importSvc *service.ProductImportService, exportSvc *service.ProductExportService) *ProductImportHandler {
	return &ProductImportHandler{importSvc: importSvc, exportSvc: exportSvc}
}

// Import godoc
//
//	@Summary		Zip 导入商品
//	@Tags			admin-导入
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			name		formData	string	true	"商品名称"
//	@Param			subTitle	formData	string	false	"副标题"
//	@Param			brandId		formData	int		true	"品牌 ID"
//	@Param			categoryId	formData	int		true	"分类 ID"
//	@Param			file		formData	file	true	"Zip 包"
//	@Success		201			{object}	response.ProductResp
//	@Failure		400			{object}	response.Body
//	@Router			/admin/products/import [post]
func (h *ProductImportHandler) Import(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(maxImportMultipart); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid multipart form")
		return
	}
	var form struct {
		Name       string `form:"name" binding:"required"`
		SubTitle   string `form:"subTitle"`
		BrandID    uint64 `form:"brandId" binding:"required"`
		CategoryID uint64 `form:"categoryId" binding:"required"`
	}
	if err := c.ShouldBind(&form); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "file required")
		return
	}

	item, err := h.importSvc.ImportFromZip(service.ProductImportInput{
		Name:       form.Name,
		SubTitle:   form.SubTitle,
		BrandID:    form.BrandID,
		CategoryID: form.CategoryID,
	}, file)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

// Export godoc
//
//	@Summary		导出商品 Zip
//	@Description	ProductCore 标准导出包，根目录名为 商品中心_商品ID_{id}；含 sku明细.txt
//	@Tags			admin-导入
//	@Produce		application/zip
//	@Security		BearerAuth
//	@Param			id	path	int	true	"商品 ID"
//	@Success		200	{file}	binary
//	@Failure		400	{object}	response.Body
//	@Failure		404	{object}	response.Body
//	@Router			/admin/products/{id}/export [get]
func (h *ProductImportHandler) Export(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	zipPath, downloadName, cleanup, err := h.exportSvc.ExportToZip(id)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	defer cleanup()
	fallback := fmt.Sprintf("productcore_%d.zip", id)
	c.Header("Content-Disposition", httputil.ContentDispositionAttachment(downloadName, fallback))
	c.File(zipPath)
}
