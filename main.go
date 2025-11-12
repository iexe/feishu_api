package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"oapi-sdk-go-demo/api"
	"oapi-sdk-go-demo/config"
	"oapi-sdk-go-demo/database"
	"oapi-sdk-go-demo/service"
)

func main() {
	// 初始化配置
	cfg := config.LoadConfig()

	// 初始化数据库
	db, err := database.InitDB(cfg.DatabasePath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// 初始化服务
	feishuService := service.NewFeishuService(cfg)

	// 设置Gin路由
	router := gin.Default()
	
	// 设置静态文件目录（用于前端页面）
	router.Static("/static", "./static")
	
	// 注册API路由
	api.SetupRoutes(router, feishuService, db)

	// 启动服务器
	log.Printf("Server starting on http://localhost:%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}