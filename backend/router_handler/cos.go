package router_handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// RegisterCOSRoutes 注册 COS 附件对象存储处理路由
func RegisterCOSRoutes(r *gin.Engine) {
	h := &COSHandler{}
	api := r.Group("/api/cos")
	{
		api.POST("/upload", h.UploadFile)
		api.GET("/download", h.DownloadFile)
	}
}

type COSHandler struct{}

func getCOSClient() (*cos.Client, error) {
	bucketURL := "https://csun-server-1444192538.cos.ap-chengdu.myqcloud.com"
	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("解析 COS Bucket URL 失败: %w", err)
	}

	secretID := os.Getenv("SECRETID")
	secretKey := os.Getenv("SECRETKEY")
	log.Printf("[COS Debug] 读取到 SECRETID: %s, SECRETKEY 长度: %d", secretID, len(secretKey))

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})
	return client, nil
}

// UploadFile 处理上传本地文件到腾讯云 COS 对象存储
func (h *COSHandler) UploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "未接收到有效文件: " + err.Error()})
		return
	}

	// 校验文件大小限制：不超过 16MB (16 * 1024 * 1024 字节)
	maxSize := int64(16 * 1024 * 1024)
	if fileHeader.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"message": "文件大小超出 16MB 限制，上传失败"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "打开文件失败: " + err.Error()})
		return
	}
	defer file.Close()

	client, err := getCOSClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "初始化 COS 客户端失败: " + err.Error()})
		return
	}

	// 生成 COS 存储相对路径 key
	timestamp := time.Now().Format("20060102_150405")
	key := fmt.Sprintf("quote_attachments/%s_%s", timestamp, filepath.Base(fileHeader.Filename))

	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: fileHeader.Header.Get("Content-Type"),
		},
	}

	_, err = client.Object.Put(c.Request.Context(), key, file, opt)
	if err != nil {
		log.Printf("[COS Error] 上传文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "上传文件至 COS 失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "文件成功上传至腾讯云 COS 对象存储",
		"data": gin.H{
			"key":          key,
			"filename":     fileHeader.Filename,
			"size":         fileHeader.Size,
			"download_url": fmt.Sprintf("/api/cos/download?key=%s", url.QueryEscape(key)),
		},
	})
}

// DownloadFile 从 COS 对象存储读取文件并推送到客户端下载
func (h *COSHandler) DownloadFile(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "缺少必需的 key 参数"})
		return
	}

	client, err := getCOSClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "初始化 COS 客户端失败: " + err.Error()})
		return
	}

	resp, err := client.Object.Get(c.Request.Context(), key, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "从 COS 获取文件失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("COS 返回状态异常: %d", resp.StatusCode)})
		return
	}

	filename := filepath.Base(key)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		c.Header("Content-Type", contentType)
	} else {
		c.Header("Content-Type", "application/octet-stream")
	}

	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		_ = c.Error(err)
	}
}
