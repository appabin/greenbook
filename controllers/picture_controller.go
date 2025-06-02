package controllers

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/models"
	"github.com/gin-gonic/gin"
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

// UploadPicture 上传图片
func UploadPicture(c *gin.Context) {
	var req UploadPictureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")

	// 生成文件名
	filename := fmt.Sprintf("%d_%d.jpg", userID, time.Now().Unix())
	filePath := fmt.Sprintf("static/images/%s", filename)
	url := fmt.Sprintf("/static/images/%s", filename)

	// 解码并保存图片
	if err := savePictureToFile(req.ImageData, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片失败: " + err.Error()})
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

// UploadPictureMultipart 支持multipart/form-data的图片上传
func UploadPictureMultipart(c *gin.Context) {
	userID := c.GetUint("userID")

	// 获取上传的文件
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败: " + err.Error()})
		return
	}

	// 生成文件名
	filename := fmt.Sprintf("%d_%d_%s", userID, time.Now().Unix(), file.Filename)
	filePath := fmt.Sprintf("static/images/%s", filename)
	url := fmt.Sprintf("/static/images/%s", filename)

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目录失败"})
		return
	}

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败: " + err.Error()})
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

// savePictureToFile 保存图片数据到文件
func savePictureToFile(base64Data, filePath string) error {
	// 解码base64数据
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 写入文件
	return ioutil.WriteFile(filePath, data, 0644)
}
