package api

import "github.com/gin-gonic/gin"

// RegisterRoutes 对外 Open API（只读，已上架商品）
func RegisterRoutes(g *gin.RouterGroup, productH *ProductHandler) {
	g.GET("/products", productH.ListPublished)
	g.GET("/products/:id", productH.GetPublished)
}
