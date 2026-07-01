// ProductCore 商品库 API
//
//	@title			ProductCore API
//	@version		1.0
//	@description	多平台电商商品底库（PIM）REST API
//	@BasePath		/api/v1
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				管理端鉴权，格式: Bearer {UserCore JWT}
//
//	@tag.name		admin-商品
//	@tag.description	商品 CRUD、草稿箱、回收站、铺货
//	@tag.name		admin-品牌
//	@tag.description	品牌管理
//	@tag.name		admin-分类
//	@tag.description	商品分类
//	@tag.name		admin-分组
//	@tag.description	商品分组
//	@tag.name		admin-上传
//	@tag.description	图片/视频上传
//	@tag.name		admin-导入
//	@tag.description	Zip 包导入商品
//	@tag.name		admin-平台类型
//	@tag.description	电商平台类型
//	@tag.name		admin-平台店铺
//	@tag.description	平台店铺与铺货
//	@tag.name		open-商品
//	@tag.description	对外 Open API（只读，已上架商品）
package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	_ "productcore/docs/swagger"
	"productcore/internal/cache"
	"productcore/internal/config"
	"productcore/internal/database"
	"productcore/internal/router"
	"productcore/internal/seed"
	"productcore/internal/storage"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "config file path")
	forceSeed := flag.Bool("seed", false, "force re-seed demo data (truncate and reload)")
	flag.Parse()

	absConfig, err := filepath.Abs(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.Load(absConfig)
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}
	log.Printf("database connected: driver=%s", cfg.Database.Driver)

	rdb, err := cache.NewRedis(&cfg.Redis)
	if err != nil {
		log.Printf("redis disabled: %v", err)
	}

	store, err := storage.New(&cfg.Storage)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("storage: driver=%s", cfg.Storage.Driver)
	if cfg.Storage.Driver == "minio" {
		log.Printf("  minio endpoint=%s bucket=%s prefix=%s public_read=%v public_base=%s",
			cfg.Storage.MinIO.Endpoint,
			cfg.Storage.MinIO.Bucket,
			cfg.Storage.EffectivePrefix(),
			cfg.Storage.MinIO.PublicRead,
			cfg.Storage.PublicBaseURL,
		)
	} else {
		log.Printf("  local dir=%s prefix=%s public_base=%s",
			cfg.Storage.LocalContentDir(),
			cfg.Storage.EffectivePrefix(),
			cfg.Storage.PublicBaseURL,
		)
	}

	if *forceSeed {
		if err := seed.Force(db); err != nil {
			log.Fatal(err)
		}
		log.Println("seed completed")
	} else {
		seed.Run(db)
	}
	seed.SeedPlatformTypes(db)

	engine := router.Setup(db, cfg, rdb, store)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("ProductCore API listening on http://localhost%s", addr)
	log.Printf("  admin API: /api/v1/admin/*")
	log.Printf("  open  API: /api/v1/open/*")
	log.Printf("  swagger:   http://localhost%s/swagger/index.html", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatal(err)
	}
}
