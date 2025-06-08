package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/models"
)

// ArticleController 文章控制器
type ArticleController struct{}

// CreateArticleRequest 创建文章请求
type CreateArticleRequest struct {
	Title      string   `json:"title" binding:"required" example:"文章标题"`
	Content    string   `json:"content" binding:"required" example:"文章内容"`
	Tags       []string `json:"tags" example:"[\"标签1\",\"标签2\"]"`
	PictureIDs []uint   `json:"picture_ids" example:"[1,2,3]"` // 图片ID数组
}

// UpdateArticleRequest 更新文章请求
type UpdateArticleRequest struct {
	Title    string   `json:"title" example:"更新后的标题"`
	Content  string   `json:"content" example:"更新后的内容"`
	Tags     []string `json:"tags" example:"[\"标签1\",\"标签2\"]"`
	Pictures []string `json:"picture" example:"[\"图片1\",\"图片2\"]"`
}

// CreateArticle 创建文章
func CreateArticle(c *gin.Context) {
	var req CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")
	article := models.Article{
		Title:        req.Title,
		Content:      req.Content,
		AuthorID:     userID,
		LikeCount:    0,
		CommentCount: 0,
	}

	// 处理标签
	for _, tagName := range req.Tags {
		var tag models.Tag
		result := global.Db.Where("name = ?", tagName).FirstOrCreate(&tag, models.Tag{Name: tagName})
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建标签失败"})
			return
		}
		article.Tags = append(article.Tags, tag)
	}

	if err := global.Db.Create(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文章失败"})
		return
	}

	// 关联图片到文章
	for i, pictureID := range req.PictureIDs {
		// 验证图片是否属于当前用户
		var picture models.Picture
		if err := global.Db.Where("id = ? AND user_id = ?", pictureID, userID).First(&picture).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("图片ID %d 不存在或不属于当前用户", pictureID)})
			return
		}

		// 创建文章图片关联记录
		articlePicture := models.ArticlePicture{
			ArticleID: article.ID,
			PictureID: pictureID,
			Order:     i, // 0为封面
		}
		if err := global.Db.Create(&articlePicture).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "关联图片失败"})
			return
		}

		article.Pictures = append(article.Pictures, picture)
	}

	// 更新用户的文章数量
	global.Db.Model(&models.User{}).Where("id = ?", userID).UpdateColumn("posts_count", gorm.Expr("posts_count + ?", 1))

	// 构建不包含用户信息的图片数组
	var picturesResponse []gin.H
	// 查询文章图片关联信息以获取顺序
	var articlePictures []models.ArticlePicture

	global.Db.Where("article_id = ?", article.ID).Order("`order`").Preload("Picture").Find(&articlePictures)

	for _, ap := range articlePictures {
		picturesResponse = append(picturesResponse, gin.H{
			"id":         ap.Picture.ID,
			"created_at": ap.Picture.CreatedAt,
			"updated_at": ap.Picture.UpdatedAt,
			"url":        ap.Picture.URL,
			"order":      ap.Order,
		})
	}

	// 返回文章信息
	c.JSON(http.StatusOK, gin.H{
		"id":             article.ID,
		"created_at":     article.CreatedAt,
		"updated_at":     article.UpdatedAt,
		"title":          article.Title,
		"content":        article.Content,
		"author_id":      article.AuthorID,
		"likes":          article.Likes,
		"comments":       article.Comments,
		"tags":           article.Tags,
		"pictures":       picturesResponse,
		"like_count":     article.LikeCount,
		"favorite_count": article.FavoriteCount,
		"comment_count":  article.CommentCount,
	})
}

// List 获取文章列表
func GetArticleList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	var articles []models.Article
	var total int64

	query := global.Db.Model(&models.Article{}).Preload("Author").Preload("Tags")
	query.Count(&total)
	result := query.Offset((page - 1) * size).Limit(size).Find(&articles)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文章列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": articles,
	})
}

// Get 获取文章详情
func GetArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章ID"})
		return
	}

	var article models.Article
	result := global.Db.Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, nickname, avatar")
	}).Preload("Tags").Preload("Comments", func(db *gorm.DB) *gorm.DB {
		return db.Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, avatar")
		}).Order("created_at DESC")
	}).First(&article, id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	// 获取文章图片
	var pictures []models.Picture
	global.Db.Joins("JOIN article_pictures ON pictures.id = article_pictures.picture_id").
		Where("article_pictures.article_id = ?", id).
		Order("article_pictures.`order`").
		Find(&pictures)

	// 构建返回数据

	c.JSON(http.StatusOK, gin.H{
		"id":         article.ID,
		"title":      article.Title,
		"content":    article.Content,
		"created_at": article.CreatedAt,
		"author": gin.H{
			"id":       article.Author.ID,
			"nickname": article.Author.Nickname,
			"avatar":   article.Author.Avatar,
		},
		"tags": func() []string {
			var tagNames []string
			for _, tag := range article.Tags {
				tagNames = append(tagNames, tag.Name)
			}
			return tagNames
		}(),
		"comments": func() []gin.H {
			var filteredComments []gin.H
			for _, comment := range article.Comments {
				filteredComments = append(filteredComments, gin.H{
					"id":         comment.ID,
					"content":    comment.Content,
					"created_at": comment.CreatedAt,
					"like_count": comment.LikeCount,
					"user": gin.H{
						"id":       comment.User.ID,
						"nickname": comment.User.Nickname,
						"avatar":   comment.User.Avatar,
					},
				})
			}
			return filteredComments
		}(),
		"pictures": func() []gin.H {
			var filteredPictures []gin.H
			for _, picture := range pictures {
				filteredPictures = append(filteredPictures, gin.H{
					"id":  picture.ID,
					"url": picture.URL,
				})
			}
			return filteredPictures
		}(),
		"like_count":     article.LikeCount,
		"favorite_count": article.FavoriteCount,
		"comment_count":  article.CommentCount,
	})
}

// Update 更新文章
func UpdateArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章ID"})
		return
	}

	var article models.Article
	if err := global.Db.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	userID := c.GetUint("userID")
	if article.AuthorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权更新此文章"})
		return
	}

	var req UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title != "" {
		article.Title = req.Title
	}
	if req.Content != "" {
		article.Content = req.Content
	}

	// 更新标签
	if len(req.Tags) > 0 {
		var tags []models.Tag
		for _, tagName := range req.Tags {
			var tag models.Tag
			result := global.Db.Where("name = ?", tagName).FirstOrCreate(&tag, models.Tag{Name: tagName})
			if result.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "更新标签失败"})
				return
			}
			tags = append(tags, tag)
		}
		global.Db.Model(&article).Association("Tags").Replace(tags)
	}

	if err := global.Db.Save(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新文章失败"})
		return
	}

	c.JSON(http.StatusOK, article)
}

// Delete 删除文章
func DeleteArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章ID"})
		return
	}

	var article models.Article
	if err := global.Db.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	userID := c.GetUint("userID")
	if article.AuthorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除此文章"})
		return
	}

	if err := global.Db.Delete(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除文章失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文章已删除"})
}
