package api

import (
	"database/sql"
	"oapi-sdk-go-demo/service"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置API路由
func SetupRoutes(router *gin.Engine, feishuService *service.FeishuService, db *sql.DB) {
	apiGroup := router.Group("/api")
	{
		// 用户相关接口
		userGroup := apiGroup.Group("/users")
		{
			userGroup.POST("/search", searchUser(feishuService, db))
			// 搜索历史功能已停用以保护用户隐私
		}

		// 消息相关接口
		messageGroup := apiGroup.Group("/messages")
		{
			messageGroup.POST("/send", sendMessage(feishuService))
			messageGroup.POST("/send-image", sendImageMessage(feishuService))
			messageGroup.POST("/send-file", sendFileMessage(feishuService))
			// 简单文本消息推送接口
			messageGroup.GET("/send-simple", sendSimpleMessageGET(feishuService))
			messageGroup.POST("/send-simple", sendSimpleMessagePOST(feishuService))
		}

		// 文件上传相关接口
		fileGroup := apiGroup.Group("/files")
		{
			fileGroup.POST("/upload", UploadFileFixed(feishuService, db))
			fileGroup.POST("/upload-image", uploadImage(feishuService, db))
			fileGroup.GET("/list", getFileList(db))
fileGroup.GET("/:resource_key", getFileInfo(db))
		fileGroup.GET("/:resource_key/view", getImageURL(feishuService))
	}

		// 首页
		apiGroup.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "飞书API服务运行中",
				"version": "1.0.0",
			})
		})
	}

	// Web页面路由
	router.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	router.GET("/user-search", func(c *gin.Context) {
		c.File("./static/user-search.html")
	})

	router.GET("/message-send", func(c *gin.Context) {
		c.File("./static/message-send.html")
	})
	
	router.GET("/image-test", func(c *gin.Context) {
		c.File("./static/image-test.html")
	})
	
	router.GET("/debug-image", func(c *gin.Context) {
		c.File("./static/debug_image.html")
	})
	
	router.GET("/file-test", func(c *gin.Context) {
		c.File("./static/file-test.html")
	})
}