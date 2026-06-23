package router

import (
	"productcore/admin"
	openapi "productcore/api"
	"productcore/internal/config"
	"productcore/internal/repo"
	"productcore/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, cfg *config.Config) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), corsMiddleware(cfg))

	repos := repo.New(db)
	productSvc := service.NewProductService(repos)
	brandSvc := service.NewBrandService(repos)
	categorySvc := service.NewCategoryService(repos)
	groupSvc := service.NewProductGroupService(repos, productSvc)

	productH := admin.NewProductHandler(productSvc)
	brandH := admin.NewBrandHandler(brandSvc)
	categoryH := admin.NewCategoryHandler(categorySvc)
	groupH := admin.NewGroupHandler(groupSvc)
	openProductH := openapi.NewProductHandler(productSvc)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "productcore"})
	})

	v1 := r.Group("/api/v1")

	admin.RegisterRoutes(v1.Group("/admin"), productH, brandH, categoryH, groupH)
	openapi.RegisterRoutes(v1.Group("/open"), openProductH)

	// 兼容旧路径（过渡期）
	admin.RegisterRoutes(v1, productH, brandH, categoryH, groupH)

	return r
}

func corsMiddleware(cfg *config.Config) gin.HandlerFunc {
	origins := cfg.CORS.AllowOrigins
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := origin == ""
		for _, o := range origins {
			if o == origin || o == "*" {
				allowed = true
				break
			}
		}
		if allowed && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
