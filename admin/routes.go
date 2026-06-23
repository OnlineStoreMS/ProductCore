package admin

import "github.com/gin-gonic/gin"

// RegisterRoutes 管理后台 API（鉴权、CRUD）
func RegisterRoutes(g *gin.RouterGroup, productH *ProductHandler, brandH *BrandHandler, categoryH *CategoryHandler, groupH *GroupHandler) {
	g.GET("/products", productH.List)
	g.GET("/products/:id", productH.Get)
	g.POST("/products", productH.Create)
	g.PUT("/products/:id", productH.Update)
	g.DELETE("/products/:id", productH.Delete)
	g.PATCH("/products/:id/publish-status", productH.UpdatePublishStatus)

	g.GET("/brands", brandH.List)
	g.POST("/brands", brandH.Create)
	g.PUT("/brands/:id", brandH.Update)
	g.DELETE("/brands/:id", brandH.Delete)

	g.GET("/categories/tree", categoryH.Tree)
	g.POST("/categories", categoryH.Create)
	g.PUT("/categories/:id", categoryH.Update)
	g.DELETE("/categories/:id", categoryH.Delete)

	g.GET("/groups", groupH.List)
	g.GET("/groups/:id/products", groupH.Products)
	g.POST("/groups", groupH.Create)
	g.PUT("/groups/:id", groupH.Update)
	g.DELETE("/groups/:id", groupH.Delete)
}
