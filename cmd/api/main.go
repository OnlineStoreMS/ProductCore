// ProductCore 商品库 API
//
//	@title			ProductCore API
//	@version		1.0
//	@description	多平台电商商品底库（PIM）REST API
//	@BasePath		/api/v1
//
//	@tag.name		admin
//	@tag.description	管理后台 API（CRUD）
//
//	@tag.name		open
//	@tag.description	对外 Open API（只读，已上架商品）
package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"productcore/internal/config"
	"productcore/internal/database"
	"productcore/internal/router"
	"productcore/internal/seed"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "config file path")
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
	seed.Run(db)

	engine := router.Setup(db, cfg)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("ProductCore API listening on http://localhost%s", addr)
	log.Printf("  admin API: /api/v1/admin/*")
	log.Printf("  open  API: /api/v1/open/*")
	if err := engine.Run(addr); err != nil {
		log.Fatal(err)
	}
}
