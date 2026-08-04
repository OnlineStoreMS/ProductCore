package router

import (
	"productcore/admin"
	adminmw "productcore/admin/middleware"
	openapi "productcore/api"
	"productcore/internal/cache"
	"productcore/internal/config"
	"productcore/internal/event"
	jwtmgr "productcore/internal/pkg/jwt"
	"productcore/internal/repo"
	"productcore/internal/service"
	"productcore/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, cfg *config.Config, rdb *redis.Client, store storage.Storage) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), corsMiddleware(cfg))

	if cfg.Storage.Driver == "local" || cfg.Storage.Driver == "" {
		r.Static("/uploads", cfg.Storage.LocalContentDir())
	}

	repos := repo.New(db)
	productCache := cache.NewProductCache(rdb, cfg.Redis.CacheTTL)
	publisher := event.NewPublisher(rdb, &cfg.Redis)
	productSvc := service.NewProductService(repos, productCache, publisher, store)
	brandSvc := service.NewBrandService(repos)
	categorySvc := service.NewCategoryService(repos)
	groupSvc := service.NewProductGroupService(repos, productSvc)
	keywordSvc := service.NewProductKeywordService(repos, productSvc)
	platformTypeSvc := service.NewPlatformTypeService(repos)
	platformShopSvc := service.NewPlatformShopService(repos)
	listingSvc := service.NewPlatformListingService(repos, productSvc)

	productH := admin.NewProductHandler(productSvc, listingSvc)
	importSvc := service.NewProductImportService(productSvc, store)
	exportSvc := service.NewProductExportService(productSvc, store)
	importH := admin.NewProductImportHandler(importSvc, exportSvc)
	brandH := admin.NewBrandHandler(brandSvc)
	categoryH := admin.NewCategoryHandler(categorySvc)
	groupH := admin.NewGroupHandler(groupSvc)
	keywordH := admin.NewKeywordHandler(keywordSvc)
	platformTypeH := admin.NewPlatformTypeHandler(platformTypeSvc)
	platformShopH := admin.NewPlatformShopHandler(platformShopSvc, listingSvc)
	uploadH := admin.NewUploadHandler(store)
	openProductH := openapi.NewProductHandler(productSvc)
	openCategoryH := openapi.NewCategoryHandler(categorySvc)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "productcore"})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")

	adminGroup := v1.Group("/admin")
	jwtMgr := jwtmgr.NewManager(cfg.Auth.JWTSecret)
	adminGroup.Use(adminmw.AdminAuth(&cfg.Auth, jwtMgr))
	admin.RegisterRoutes(adminGroup, productH, brandH, categoryH, groupH, keywordH, uploadH, importH, platformTypeH, platformShopH)

	openapi.RegisterRoutes(v1.Group("/open"), openProductH, openCategoryH)

	legacy := v1.Group("")
	legacy.Use(adminmw.AdminAuth(&cfg.Auth, jwtMgr))
	admin.RegisterRoutes(legacy, productH, brandH, categoryH, groupH, keywordH, uploadH, importH, platformTypeH, platformShopH)

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
