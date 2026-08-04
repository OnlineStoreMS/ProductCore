package httputil

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"productcore/internal/pkg/response"
	"productcore/internal/service"

	"github.com/gin-gonic/gin"
)

func ParseID(c *gin.Context) (uint64, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

// ContentDispositionAttachment 生成支持中文文件名的 Content-Disposition（RFC 5987）
func ContentDispositionAttachment(utf8Filename, asciiFallback string) string {
	utf8Filename = strings.TrimSpace(utf8Filename)
	if utf8Filename == "" {
		utf8Filename = "export.zip"
	}
	asciiFallback = strings.TrimSpace(asciiFallback)
	if asciiFallback == "" {
		asciiFallback = "export.zip"
	}
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, asciiFallback, rfc5987Encode(utf8Filename))
}

func rfc5987Encode(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			c == '!' || c == '#' || c == '$' || c == '&' || c == '+' || c == '-' ||
			c == '.' || c == '^' || c == '_' || c == '`' || c == '|' || c == '~' {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

func HandleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound), errors.Is(err, service.ErrNoEditDraft):
		response.Fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrDuplicateSN), errors.Is(err, service.ErrDuplicateSku),
		errors.Is(err, service.ErrDuplicateMaterialCode), errors.Is(err, service.ErrDuplicateTypeCode):
		response.Fail(c, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrInvalidSkuCode), errors.Is(err, service.ErrSkuCodeRequired),
		errors.Is(err, service.ErrMaterialCodeRequired), errors.Is(err, service.ErrNotInTrash),
		errors.Is(err, service.ErrDuplicateSpecCombo), errors.Is(err, service.ErrDuplicateSpecValue),
		errors.Is(err, service.ErrInvalidImport), errors.Is(err, service.ErrInvalidExport),
		errors.Is(err, service.ErrNoEditDraft),
		errors.Is(err, service.ErrBuiltinPlatformType),
		errors.Is(err, service.ErrTypeHasShops):
		response.Fail(c, http.StatusBadRequest, err.Error())
	default:
		if err.Error() == "category has children" {
			response.Fail(c, http.StatusBadRequest, "请先删除子分类")
			return
		}
		if err.Error() == "group has children" {
			response.Fail(c, http.StatusBadRequest, "请先删除子分组")
			return
		}
		response.Fail(c, http.StatusInternalServerError, err.Error())
	}
}
