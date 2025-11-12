package api

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"oapi-sdk-go-demo/service"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 辅助函数
func getStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// 用户搜索请求结构
type SearchUserRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

// 发送消息请求结构
type SendMessageRequest struct {
	ReceiveIdType string `json:"receive_id_type" binding:"required"`
	ReceiveId     string `json:"receive_id" binding:"required"`
	Content       string `json:"content" binding:"required"`
}

// 发送图片消息请求结构
type SendImageMessageRequest struct {
	ReceiveIdType string `json:"receive_id_type" binding:"required"`
	ReceiveId     string `json:"receive_id" binding:"required"`
	ImageKey      string `json:"image_key" binding:"required"`
}

// 发送文件消息请求结构
type SendFileMessageRequest struct {
	ReceiveIdType string `json:"receive_id_type" binding:"required"`
	ReceiveId     string `json:"receive_id" binding:"required"`
	FileKey       string `json:"file_key" binding:"required"`
}

// 简单文本消息推送请求结构
type SimpleMessageRequest struct {
	UserID string `json:"userid" binding:"required"`
	Msg    string `json:"msg" binding:"required"`
}

// 搜索用户
func searchUser(feishuService *service.FeishuService, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SearchUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		users, err := feishuService.SearchUserByPhone(req.PhoneNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 搜索历史功能已停用以保护用户隐私
		// 不再记录搜索历史到数据库

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    users,
			"count":   len(users),
		})
	}
}

// 获取用户搜索历史
func getUserSearchHistory(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "10")
		limitInt, _ := strconv.Atoi(limit)

		rows, err := db.Query(`
			SELECT phone_number, search_time, result_count 
			FROM user_search_logs 
			ORDER BY search_time DESC 
			LIMIT ?
		`, limitInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var history []map[string]interface{}
		for rows.Next() {
			var phoneNumber string
			var searchTime time.Time
			var resultCount int

			if err := rows.Scan(&phoneNumber, &searchTime, &resultCount); err != nil {
				continue
			}

			history = append(history, map[string]interface{}{
				"phone_number": phoneNumber,
				"search_time":  searchTime.Format("2006-01-02 15:04:05"),
				"result_count": resultCount,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    history,
		})
	}
}

// 发送文本消息
func sendMessage(feishuService *service.FeishuService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SendMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := feishuService.SendTextMessage(req.ReceiveIdType, req.ReceiveId, req.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	}
}

// 上传图片
func uploadImage(feishuService *service.FeishuService, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("image")
		if err != nil {
			fmt.Printf("图片上传失败 - FormFile error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "图片上传失败: " + err.Error()})
			return
		}
		defer file.Close()

		fmt.Printf("接收到图片文件: %s, 大小: %d bytes\n", header.Filename, header.Size)

		// 计算文件特征码
		fileBytes, _ := io.ReadAll(file)
		md5Hash := md5.Sum(fileBytes)
		sha256Hash := sha256.Sum256(fileBytes)

		// 重新创建reader用于上传
		fileReader := io.NopCloser(bytes.NewReader(fileBytes))

		// 上传到飞书
		result, err := feishuService.UploadImage(fileReader)
		if err != nil {
			fmt.Printf("上传图片到飞书失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("上传图片失败: %v", err)})
			return
		}

		if result == nil || result.ImageKey == nil || *result.ImageKey == "" {
			fmt.Printf("上传图片失败: 返回结果为空或image_key为空\n")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "上传图片失败: 服务器返回无效结果"})
			return
		}

		fmt.Printf("图片上传成功, image_key: %s\n", getStringValue(result.ImageKey))

		// 保存到数据库
		apiResponse, _ := json.Marshal(result)
		_, err = db.Exec(`
			INSERT INTO file_metadata 
			(resource_type, resource_key, original_name, file_size, md5_hash, sha256_hash, upload_user_id, feishu_api_response)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, "image", getStringValue(result.ImageKey), header.Filename, header.Size, 
			hex.EncodeToString(md5Hash[:]), hex.EncodeToString(sha256Hash[:]), 
			"", string(apiResponse))

		if err != nil {
			fmt.Printf("Failed to save image metadata: %v\n", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"image_key": getStringValue(result.ImageKey),
				"md5":       hex.EncodeToString(md5Hash[:]),
				"sha256":    hex.EncodeToString(sha256Hash[:]),
			},
		})
	}
}

// 发送图片消息
func sendImageMessage(feishuService *service.FeishuService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SendImageMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			fmt.Printf("发送图片消息 - 参数错误: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fmt.Printf("发送图片消息 - 接收者类型: %s, 接收者ID: %s, 图片Key: %s\n", 
			req.ReceiveIdType, req.ReceiveId, req.ImageKey)

		result, err := feishuService.SendImageMessage(req.ReceiveIdType, req.ReceiveId, req.ImageKey)
		if err != nil {
			fmt.Printf("发送图片消息失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 构建响应数据
		responseData := map[string]interface{}{
			"message_id": getStringValue(result.MessageId),
		}

		fmt.Printf("发送图片消息成功, message_id: %s\n", getStringValue(result.MessageId))

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    responseData,
		})
	}
}

// 发送文件消息
func sendFileMessage(feishuService *service.FeishuService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SendFileMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := feishuService.SendFileMessage(req.ReceiveIdType, req.ReceiveId, req.FileKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	}
}

// 上传文件
func uploadFile(feishuService *service.FeishuService, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "文件上传失败"})
			return
		}
		defer file.Close()

		// 计算文件特征码
		fileBytes, _ := io.ReadAll(file)
		md5Hash := md5.Sum(fileBytes)
		sha256Hash := sha256.Sum256(fileBytes)

		// 重新创建reader用于上传
		fileReader := io.NopCloser(bytes.NewReader(fileBytes))

		// 根据文件扩展名确定文件类型
		fileExt := filepath.Ext(header.Filename)
		fileType := "doc" // 默认类型
		switch strings.ToLower(fileExt) {
		case ".doc", ".docx":
			fileType = "doc"
		case ".pdf":
			fileType = "pdf"
		case ".xls", ".xlsx":
			fileType = "xls"
		case ".ppt", ".pptx":
			fileType = "ppt"
		case ".txt":
			fileType = "txt"
		case ".zip", ".rar":
			fileType = "zip"
		}

		// 上传到飞书
		result, err := feishuService.UploadFile(
			fileType,
			header.Filename,
			fileReader,
			0, // duration
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 保存到数据库
		apiResponse, _ := json.Marshal(result)
		_, err = db.Exec(`
			INSERT INTO file_metadata 
			(resource_type, resource_key, original_name, file_size, md5_hash, sha256_hash, upload_user_id, feishu_api_response)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, "file", getStringValue(result.FileKey), header.Filename, header.Size, 
			hex.EncodeToString(md5Hash[:]), hex.EncodeToString(sha256Hash[:]), 
			"", string(apiResponse))

		if err != nil {
			fmt.Printf("Failed to save file metadata: %v\n", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"file_key": getStringValue(result.FileKey),
				"md5":      hex.EncodeToString(md5Hash[:]),
				"sha256":   hex.EncodeToString(sha256Hash[:]),
			},
		})
	}
}

// 获取文件列表
func getFileList(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "20")
		limitInt, _ := strconv.Atoi(limit)

		rows, err := db.Query(`
			SELECT resource_type, resource_key, original_name, file_size, md5_hash, upload_time
			FROM file_metadata 
			ORDER BY upload_time DESC 
			LIMIT ?
		`, limitInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var files []map[string]interface{}
		for rows.Next() {
			var resourceType, resourceKey, originalName, md5Hash string
			var fileSize int64
			var uploadTime time.Time

			if err := rows.Scan(&resourceType, &resourceKey, &originalName, &fileSize, &md5Hash, &uploadTime); err != nil {
				continue
			}

			files = append(files, map[string]interface{}{
				"resource_type": resourceType,
				"resource_key":  resourceKey,
				"original_name": originalName,
				"file_size":     fileSize,
				"md5_hash":      md5Hash,
				"upload_time":   uploadTime.Format("2006-01-02 15:04:05"),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    files,
		})
	}
}

// 获取文件信息
func getFileInfo(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceKey := c.Param("resource_key")

		var resourceType, originalName, md5Hash, sha256Hash, apiResponse string
		var fileSize int64
		var uploadTime time.Time

		err := db.QueryRow(`
			SELECT resource_type, original_name, file_size, md5_hash, sha256_hash, upload_time, feishu_api_response
			FROM file_metadata 
			WHERE resource_key = ?
		`, resourceKey).Scan(&resourceType, &originalName, &fileSize, &md5Hash, &sha256Hash, &uploadTime, &apiResponse)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": map[string]interface{}{
				"resource_type": resourceType,
				"resource_key":  resourceKey,
				"original_name": originalName,
				"file_size":     fileSize,
				"md5_hash":      md5Hash,
				"sha256_hash":   sha256Hash,
				"upload_time":   uploadTime.Format("2006-01-02 15:04:05"),
				"api_response":  apiResponse,
			},
		})
	}
}

// 获取图片URL
func getImageURL(feishuService *service.FeishuService) gin.HandlerFunc {
	return func(c *gin.Context) {
		imageKey := c.Param("resource_key")
		
		// 飞书没有提供直接获取图片URL的API
		// 图片通常只能通过发送消息或在飞书客户端中查看
		// 这里我们返回一个包含image_key的信息，以及使用建议
		
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "飞书API不提供直接访问图片的URL，图片只能通过以下方式查看：",
			"methods": []string{
				"1. 通过发送图片消息到聊天中查看",
				"2. 在飞书客户端中查看上传的图片",
			},
			"image_key": imageKey,
		})
	}
}

// 简单文本消息推送 - GET方式
func sendSimpleMessageGET(feishuService *service.FeishuService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userid := c.Query("userid")
		msg := c.Query("msg")
		
		if userid == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userid参数不能为空"})
			return
		}
		
		if msg == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "msg参数不能为空"})
			return
		}
		
		// 使用user_id作为接收者类型
		result, err := feishuService.SendTextMessage("user_id", userid, msg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
			"message": "消息发送成功",
		})
	}
}

// 简单文本消息推送 - POST方式
func sendSimpleMessagePOST(feishuService *service.FeishuService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SimpleMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// 使用user_id作为接收者类型
		result, err := feishuService.SendTextMessage("user_id", req.UserID, req.Msg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
			"message": "消息发送成功",
		})
	}
}