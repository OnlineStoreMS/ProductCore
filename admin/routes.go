package admin

import "github.com/gin-gonic/gin"

// RegisterRoutes 管理后台 API（鉴权、CRUD）
func RegisterRoutes(
	g *gin.RouterGroup,
	productH *ProductHandler,
	brandH *BrandHandler,
	categoryH *CategoryHandler,
	groupH *GroupHandler,
	keywordH *KeywordHandler,
	uploadH *UploadHandler,
	importH *ProductImportHandler,
	platformTypeH *PlatformTypeHandler,
	platformShopH *PlatformShopHandler,
) {
	g.GET("/products", productH.List)
	g.GET("/super-search", productH.SuperSearch)
	g.GET("/products/drafts", productH.ListDrafts)
	g.DELETE("/products/:id/edit-draft", productH.DiscardEditDraft)
	g.GET("/products/trash", productH.ListTrash)
	g.POST("/products/batch/delete", productH.BatchDelete)
	g.POST("/products/batch/restore", productH.BatchRestore)
	g.POST("/products/batch/force-delete", productH.BatchForceDelete)
	g.GET("/products/:id", productH.Get)
	g.POST("/products", productH.Create)
	g.POST("/products/draft", productH.CreateDraft)
	g.POST("/products/import", importH.Import)
	g.GET("/products/:id/export", importH.Export)
	g.POST("/products/:id/restore", productH.Restore)
	g.PUT("/products/:id", productH.Update)
	g.GET("/products/:id/skus", productH.GetSkus)
	g.PUT("/products/:id/skus", productH.UpdateSkus)
	g.GET("/products/:id/listings", productH.GetListings)
	g.PUT("/products/:id/listings", productH.SetListings)
	g.DELETE("/products/:id", productH.Delete)
	g.DELETE("/products/:id/force", productH.ForceDelete)
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
	g.GET("/groups/tree", groupH.Tree)
	g.GET("/groups/:id/products", groupH.Products)
	g.POST("/groups", groupH.Create)
	g.PUT("/groups/:id", groupH.Update)
	g.DELETE("/groups/:id", groupH.Delete)

	g.GET("/keywords", keywordH.List)
	g.GET("/keywords/:id/products", keywordH.Products)
	g.PUT("/keywords/:id/products", keywordH.SetProducts)
	g.POST("/keywords/:id/products", keywordH.AddProducts)
	g.DELETE("/keywords/:id/products/:productId", keywordH.RemoveProduct)
	g.POST("/keywords", keywordH.Create)
	g.PUT("/keywords/:id", keywordH.Update)
	g.DELETE("/keywords/:id", keywordH.Delete)

	g.POST("/upload", uploadH.Upload)
	g.POST("/upload/batch", uploadH.UploadBatch)
	g.POST("/upload/video", uploadH.UploadVideo)
	g.POST("/upload/from-url", uploadH.UploadFromURL)

	g.GET("/platform-types", platformTypeH.List)
	g.GET("/platform-types/enabled", platformTypeH.ListEnabled)
	g.POST("/platform-types", platformTypeH.Create)
	g.PUT("/platform-types/:id", platformTypeH.Update)
	g.DELETE("/platform-types/:id", platformTypeH.Delete)

	g.GET("/platform-shops", platformShopH.List)
	g.GET("/platform-shops/:id/products", platformShopH.ListProducts)
	g.POST("/platform-shops", platformShopH.Create)
	g.PUT("/platform-shops/:id", platformShopH.Update)
	g.DELETE("/platform-shops/:id", platformShopH.Delete)
}
