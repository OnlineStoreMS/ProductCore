package admin

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"productcore/internal/pkg/response"
	"productcore/internal/storage"

	"github.com/gin-gonic/gin"
)

const (
	maxImageSize     = 10 << 20  // 10MB
	maxVideoSize     = 200 << 20 // 200MB
	maxBatchImages   = 50
	maxBatchFormSize = 100 << 20 // 100MB
)

type UploadHandler struct {
	store storage.Storage
}

func NewUploadHandler(store storage.Storage) *UploadHandler {
	return &UploadHandler{store: store}
}

type batchUploadFailed struct {
	Filename string `json:"filename"`
	Message  string `json:"message"`
}

func uploadFormValues(c *gin.Context) map[string]string {
	keys := []string{"scope", "resource", "productId", "skuId"}
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		if v := c.PostForm(k); v != "" {
			out[k] = v
		}
	}
	return out
}

// Upload godoc
//
//	@Summary		上传图片
//	@Description	单文件上传（主图/SKU/详情/素材）
//	@Tags			admin-上传
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file		formData	file	true	"图片文件"
//	@Param			scope		formData	string	false	"product|common"
//	@Param			resource	formData	string	false	"资源类型 main/album/detail/..."
//	@Param			productId	formData	int		false	"商品 ID"
//	@Param			skuId		formData	int		false	"SKU ID"
//	@Success		200			{object}	response.UploadURLResp
//	@Failure		400			{object}	response.Body
//	@Router			/admin/upload [post]
func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "file required")
		return
	}
	if msg := validateImageFile(file); msg != "" {
		response.Fail(c, http.StatusBadRequest, msg)
		return
	}
	opts := storage.UploadOptionsFromForm(uploadFormValues(c), false)
	if err := opts.Validate(); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	url, err := h.store.Upload(file, opts)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"url": url})
}

// UploadBatch godoc
//
//	@Summary		批量上传图片
//	@Tags			admin-上传
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			files		formData	file	true	"图片文件（可多选）"
//	@Param			scope		formData	string	false	"product|common"
//	@Param			resource	formData	string	false	"资源类型"
//	@Param			productId	formData	int		false	"商品 ID"
//	@Param			skuId		formData	int		false	"SKU ID"
//	@Success		200			{object}	response.BatchUploadResp
//	@Router			/admin/upload/batch [post]
func (h *UploadHandler) UploadBatch(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(maxBatchFormSize); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid multipart form")
		return
	}
	form := c.Request.MultipartForm
	if form == nil {
		response.Fail(c, http.StatusBadRequest, "files required")
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		response.Fail(c, http.StatusBadRequest, "files required")
		return
	}
	if len(files) > maxBatchImages {
		response.Fail(c, http.StatusBadRequest, fmt.Sprintf("too many files (max %d)", maxBatchImages))
		return
	}

	opts := storage.UploadOptionsFromForm(uploadFormValues(c), false)
	if err := opts.Validate(); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	urls := make([]string, 0, len(files))
	failed := make([]batchUploadFailed, 0)
	for _, file := range files {
		if msg := validateImageFile(file); msg != "" {
			failed = append(failed, batchUploadFailed{Filename: file.Filename, Message: msg})
			continue
		}
		url, err := h.store.Upload(file, opts)
		if err != nil {
			failed = append(failed, batchUploadFailed{Filename: file.Filename, Message: err.Error()})
			continue
		}
		urls = append(urls, url)
	}
	if len(urls) == 0 {
		msg := "upload failed"
		if len(failed) > 0 {
			msg = failed[0].Message
		}
		response.Fail(c, http.StatusBadRequest, msg)
		return
	}
	response.OK(c, gin.H{"urls": urls, "failed": failed})
}

// UploadVideo godoc
//
//	@Summary		上传视频
//	@Tags			admin-上传
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file		formData	file	true	"视频文件"
//	@Param			scope		formData	string	false	"product|common"
//	@Param			resource	formData	string	false	"video_11/video_34/..."
//	@Param			productId	formData	int		false	"商品 ID"
//	@Success		200			{object}	response.UploadURLResp
//	@Router			/admin/upload/video [post]
func (h *UploadHandler) UploadVideo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "file required")
		return
	}
	if file.Size > maxVideoSize {
		response.Fail(c, http.StatusBadRequest, "file too large (max 200MB)")
		return
	}
	if !isAllowedVideo(file) {
		response.Fail(c, http.StatusBadRequest, "unsupported video type (mp4/webm/mov/avi)")
		return
	}
	opts := storage.UploadOptionsFromForm(uploadFormValues(c), true)
	if err := opts.Validate(); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	url, err := h.store.Upload(file, opts)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"url": url})
}

type uploadFromURLBody struct {
	URL       string `json:"url" binding:"required"`
	Scope     string `json:"scope"`
	Resource  string `json:"resource"`
	ProductID uint64 `json:"productId"`
	SkuID     uint64 `json:"skuId"`
}

// UploadFromURL godoc
//
//	@Summary		从 URL 上传图片
//	@Description	下载远程图片并写入存储（SKU 导入等）
//	@Tags			admin-上传
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		uploadFromURLBody	true	"远程图片 URL 与上传上下文"
//	@Success		200		{object}	response.UploadURLResp
//	@Failure		400		{object}	response.Body
//	@Router			/admin/upload/from-url [post]
func (h *UploadHandler) UploadFromURL(c *gin.Context) {
	var body uploadFromURLBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, http.StatusBadRequest, "url required")
		return
	}
	opts := storage.UploadOptionsFromForm(map[string]string{
		"scope":     body.Scope,
		"resource":  body.Resource,
		"productId": fmt.Sprintf("%d", body.ProductID),
		"skuId":     fmt.Sprintf("%d", body.SkuID),
	}, false)
	if err := opts.Validate(); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	url, err := storage.UploadFromURL(h.store, body.URL, opts)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"url": url})
}

func fileContentType(file *multipart.FileHeader) string {
	if ct := file.Header.Get("Content-Type"); ct != "" && ct != "application/octet-stream" {
		return strings.ToLower(ct)
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	default:
		return ""
	}
}

func isAllowedImage(file *multipart.FileHeader) bool {
	ct := fileContentType(file)
	if strings.HasPrefix(ct, "image/") {
		switch ct {
		case "image/jpeg", "image/png", "image/webp", "image/jpg":
			return true
		}
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	}
	return false
}

func isAllowedVideo(file *multipart.FileHeader) bool {
	ct := fileContentType(file)
	if strings.HasPrefix(ct, "video/") {
		return true
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".mp4", ".webm", ".mov", ".avi", ".mpeg", ".mpg":
		return true
	}
	return false
}

func validateImageFile(file *multipart.FileHeader) string {
	if file.Size > maxImageSize {
		return "file too large (max 10MB)"
	}
	if !isAllowedImage(file) {
		return "unsupported image type (jpg/jpeg/png/webp)"
	}
	return ""
}
