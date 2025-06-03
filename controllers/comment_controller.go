package controllers

import (
	"net/http"
	"strconv"

	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	Content string `json:"content" binding:"required" example:"评论内容"`
}

// CreateComment 评论文章
func CreateComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("article_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章ID"})
		return
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")
	comment := models.Comment{
		Content:   req.Content,
		UserID:    userID,
		ArticleID: uint(id),
	}

	if err := global.Db.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建评论失败"})
		return
	}

	// 更新文章评论数
	global.Db.Model(&models.Article{}).Where("id = ?", id).Update("comment_count", gorm.Expr("comment_count + ?", 1))

	// 查询用户信息
	var user models.User
	if err := global.Db.Select("nickname, avatar").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 返回用户昵称、头像和评论内容
	response := gin.H{
		"user_id":  comment.UserID,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"content":  comment.Content,
	}

	c.JSON(http.StatusOK, response)
}
