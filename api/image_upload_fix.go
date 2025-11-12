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

	"github.com/gin-gonic/gin"
)

// 上传图片的修复版本
func UploadImageFixed(feishuService *service.FeishuService, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "图片上传失败: " + err.Error()})
			return
		}
		defer file.Close()

		fmt.Printf("接收到图片文件: %s, 大小: %d bytes\n", header.Filename, header.Size)
		
		// 检查文件大小
		if header.Size > 10*1024*1024 {
			fmt.Printf("图片文件过大: %d bytes，超过10MB限制\n", header.Size)
			c.JSON(http.StatusBadRequest, gin.H{"error": "图片文件过大，请选择小于10MB的图片"})
			return
		}

		// 计算文件特征码
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			fmt.Printf("读取图片文件失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取图片文件失败"})
			return
		}
		
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

// 发送图片消息的修复版本
func SendImageMessageFixed(feishuService *service.FeishuService) gin.HandlerFunc {
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