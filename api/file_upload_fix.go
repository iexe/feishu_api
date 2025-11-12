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
	"strings"

	"github.com/gin-gonic/gin"
)

// 修复版的文件上传函数
func UploadFileFixed(feishuService *service.FeishuService, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求信息
		fmt.Printf("文件上传请求: %s %s\n", c.Request.Method, c.Request.URL.Path)
		fmt.Printf("Content-Type: %s\n", c.GetHeader("Content-Type"))

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			fmt.Printf("文件上传失败 - FormFile error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": "文件上传失败: " + err.Error(),
			})
			return
		}
		defer file.Close()

		fmt.Printf("接收到文件: %s, 大小: %d bytes\n", header.Filename, header.Size)

		// 检查文件大小
		if header.Size > 10*1024*1024 {
			fmt.Printf("文件过大: %d bytes，超过10MB限制\n", header.Size)
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": "文件过大，请选择小于10MB的文件",
			})
			return
		}

		// 读取文件内容
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			fmt.Printf("读取文件失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": "读取文件失败: " + err.Error(),
			})
			return
		}

		// 计算文件哈希
		md5Hash := md5.Sum(fileBytes)
		sha256Hash := sha256.Sum256(fileBytes)

		// 重新创建文件读取器
		fileReader := io.NopCloser(bytes.NewReader(fileBytes))

		// 根据文件扩展名确定文件类型
		fileExt := strings.ToLower(filepath.Ext(header.Filename))
		fileType := "stream" // 默认类型为二进制流
		
		switch fileExt {
		case ".opus":
			fileType = "opus"  // OPUS 音频文件
		case ".mp4", ".m4v", ".m4a":
			fileType = "mp4"  // MP4 格式视频文件
		case ".pdf":
			fileType = "pdf"  // PDF 格式文件
		case ".doc", ".docx":
			fileType = "doc"  // DOC 格式文件
		case ".xls", ".xlsx":
			fileType = "xls"  // XLS 格式文件
		case ".ppt", ".pptx":
			fileType = "ppt"  // PPT 格式文件
		case ".mp3", ".wav", ".flac", ".aac", ".ogg":
			// 非OPUS格式的音频文件，提示用户转换
			fmt.Printf("错误: 音频文件必须是OPUS格式，当前格式为 %s\n", fileExt)
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": "音频文件必须为OPUS格式，请使用以下命令转换: ffmpeg -i " + header.Filename + " -acodec libopus -ac 1 -ar 16000 output.opus",
			})
			return
		case ".avi", ".mov", ".wmv", ".flv", ".mkv":
			// 非MP4格式的视频文件，提示用户转换
			fmt.Printf("错误: 视频文件必须是MP4格式，当前格式为 %s\n", fileExt)
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": "视频文件必须为MP4格式，请使用转换工具将文件转为MP4格式",
			})
			return
		case ".txt", ".log", ".conf", ".ini", ".zip", ".rar", ".7z", ".jpg", ".jpeg", ".png", ".gif", ".webp":
			fileType = "stream"  // 其他文件使用stream类型
		default:
			fileType = "stream"  // 其他文件使用stream类型
		}

		fmt.Printf("文件类型: %s (根据扩展名 %s)\n", fileType, fileExt)

		// 上传到飞书
		result, err := feishuService.UploadFile(
			fileType,
			header.Filename,
			fileReader,
			0, // duration
		)
		if err != nil {
			fmt.Printf("上传文件到飞书失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": "上传文件失败: " + err.Error(),
			})
			return
		}

		if result == nil || result.FileKey == nil || *result.FileKey == "" {
			fmt.Printf("上传文件失败: 返回结果为空或file_key为空\n")
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": "上传文件失败: 服务器返回无效结果",
			})
			return
		}

		fileKey := *result.FileKey
		fmt.Printf("文件上传成功, file_key: %s\n", fileKey)

		// 保存到数据库
		apiResponse, _ := json.Marshal(result)
		_, err = db.Exec(`
			INSERT INTO file_metadata 
			(resource_type, resource_key, original_name, file_size, md5_hash, sha256_hash, upload_user_id, feishu_api_response)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, "file", fileKey, header.Filename, header.Size, 
			hex.EncodeToString(md5Hash[:]), hex.EncodeToString(sha256Hash[:]), 
			"", string(apiResponse))

		if err != nil {
			fmt.Printf("Failed to save file metadata: %v\n", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"file_key": fileKey,
				"md5":      hex.EncodeToString(md5Hash[:]),
				"sha256":   hex.EncodeToString(sha256Hash[:]),
			},
		})
	}
}