package controllers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/models"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

// UploadPictureRequest 上传图片请求
type UploadPictureRequest struct {
	ImageData string `json:"image_data" binding:"required" example:"data:image/jpeg;base64,/9j/4AAQ..."`
}

// UploadPictureResponse 上传图片响应
type UploadPictureResponse struct {
	ID  uint   `json:"id"`
	URL string `json:"url"`
}

// UploadPicture 上传图片到 MinIO
func UploadPicture(c *gin.Context) {
	var req UploadPictureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")

	// 生成文件名
	filename := fmt.Sprintf("%d_%d.jpg", userID, time.Now().Unix())
	// MinIO 对象名称
	objectName := fmt.Sprintf("images/%s", filename)
	// 返回的 URL 格式保持不变
	url := fmt.Sprintf("/static/images/%s", filename)

	// 解码 base64 数据
	data, err := base64.StdEncoding.DecodeString(req.ImageData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的图片数据"})
		return
	}

	// 上传到 MinIO
	_, err = global.MinIOClient.PutObject(
		context.Background(),
		global.MinIOConf.BucketName,
		objectName,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{
			ContentType: "image/jpeg",
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传图片失败: " + err.Error()})
		return
	}

	// 创建图片记录
	picture := models.Picture{
		URL:    url,
		UserID: userID,
	}

	if err := global.Db.Create(&picture).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建图片记录失败"})
		return
	}

	c.JSON(http.StatusOK, UploadPictureResponse{
		ID:  picture.ID,
		URL: picture.URL,
	})
}

// UploadPictureMultipart 支持multipart/form-data的图片上传到 MinIO
func UploadPictureMultipart(c *gin.Context) {
	userID := c.GetUint("userID")

	// 获取上传的文件
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败: " + err.Error()})
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败"})
		return
	}
	defer src.Close()

	// 生成文件名
	filename := fmt.Sprintf("%d_%d_%s", userID, time.Now().Unix(), file.Filename)
	// MinIO 对象名称
	objectName := fmt.Sprintf("images/%s", filename)
	// 返回的 URL 格式保持不变
	url := fmt.Sprintf("/static/images/%s", filename)

	// 上传到 MinIO
	_, err = global.MinIOClient.PutObject(
		context.Background(),
		global.MinIOConf.BucketName,
		objectName,
		src,
		file.Size,
		minio.PutObjectOptions{
			ContentType: file.Header.Get("Content-Type"),
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传文件失败: " + err.Error()})
		return
	}

	// 创建图片记录
	picture := models.Picture{
		URL:    url,
		UserID: userID,
	}

	if err := global.Db.Create(&picture).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建图片记录失败"})
		return
	}

	c.JSON(http.StatusOK, UploadPictureResponse{
		ID:  picture.ID,
		URL: picture.URL,
	})
}

// ServeImageFromMinIO 从 MinIO 获取并返回图片
func ServeImageFromMinIO(c *gin.Context) {
	filename := c.Param("filename")
	objectName := fmt.Sprintf("images/%s", filename)

	// 从 MinIO 获取图片
	object, err := global.MinIOClient.GetObject(
		context.Background(),
		global.MinIOConf.BucketName,
		objectName,
		minio.GetObjectOptions{},
	)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer object.Close()

	// 获取对象信息
	objInfo, err := object.Stat()
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	// 设置响应头
	c.Header("Content-Type", objInfo.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", objInfo.Size))
	c.Header("Cache-Control", "public, max-age=31536000") // 1年缓存
	c.Header("ETag", objInfo.ETag)

	// 流式传输图片
	_, err = io.Copy(c.Writer, object)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
}
