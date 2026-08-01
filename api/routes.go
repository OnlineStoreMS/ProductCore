package api

import "github.com/gin-gonic/gin"

// RegisterRoutes 对外 Open API（只读，已上架商品/分类）
func RegisterRoutes(g *gin.RouterGroup, productH *ProductHandler, categoryH *CategoryHandler) {
	g.GET("/products", productH.ListPublished)
	g.GET("/products/:id", productH.GetPublished)
	g.GET("/categories/tree", categoryH.VisibleTree)
}
