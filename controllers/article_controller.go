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
	Title   string   `json:"title" binding:"required" example:"文章标题"`
	Content string   `json:"content" binding:"required" example:"文章内容"`
	Tags    []string `json:"tags" example:"[\"标签1\",\"标签2\"]"`
}

// UpdateArticleRequest 更新文章请求
type UpdateArticleRequest struct {
	Title   string   `json:"title" example:"更新后的标题"`
	Content string   `json:"content" example:"更新后的内容"`
	Tags    []string `json:"tags" example:"[\"标签1\",\"标签2\"]"`
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	Content  string `json:"content" binding:"required" example:"评论内容"`
	ParentID *uint  `json:"parent_id" example:"1"`
}

// @Summary 创建文章
// @Description 创建一篇新文章
// @Tags 文章管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param article body CreateArticleRequest true "文章信息"
// @Success 200 {object} models.Article
// @Router /api/articles [post]
func (ac *ArticleController) Create(c *gin.Context) {
	var req CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")
	article := models.Article{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: userID,
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

	c.JSON(http.StatusOK, article)
}

// @Summary 获取文章列表
// @Description 获取文章列表，支持分页
// @Tags 文章管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {array} models.Article
// @Router /api/articles [get]
func (ac *ArticleController) List(c *gin.Context) {
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

// @Summary 获取文章详情
// @Description 获取指定文章的详细信息
// @Tags 文章管理
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} models.Article
// @Router /api/articles/{id} [get]
func (ac *ArticleController) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章ID"})
		return
	}

	var article models.Article
	result := global.Db.Preload("Author").Preload("Tags").Preload("Comments", func(db *gorm.DB) *gorm.DB {
		return db.Preload("User").Order("created_at DESC")
	}).First(&article, id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	c.JSON(http.StatusOK, article)
}

// @Summary 更新文章
// @Description 更新指定文章的信息
// @Tags 文章管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path int true "文章ID"
// @Param article body UpdateArticleRequest true "文章更新信息"
// @Success 200 {object} models.Article
// @Router /api/articles/{id} [put]
func (ac *ArticleController) Update(c *gin.Context) {
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

// @Summary 删除文章
// @Description 删除指定的文章
// @Tags 文章管理
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path int true "文章ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/articles/{id} [delete]
func (ac *ArticleController) Delete(c *gin.Context) {
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

// @Summary 点赞文章
// @Description 为指定文章点赞或取消点赞
// @Tags 文章管理
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path int true "文章ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/articles/{id}/like [post]
func (ac *ArticleController) ToggleLike(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章ID"})
		return
	}

	userID := c.GetUint("userID")
	likeKey := fmt.Sprintf("article:like:%d:%d", id, userID)
	countKey := fmt.Sprintf("article:like_count:%d", id)

	// 使用Redis的SETNX命令尝试添加点赞
	setResult := global.RedisDB.SetNX(likeKey, "1", 0).Val()
	if setResult {
		// 点赞成功，增加计数
		global.RedisDB.Incr(countKey)

		// 异步保存到数据库
		go func() {
			like := models.Like{
				UserID:    userID,
				ArticleID: uint(id),
			}
			global.Db.Create(&like)
			global.Db.Model(&models.Article{}).Where("id = ?", id).Update("like_count", gorm.Expr("like_count + ?", 1))
		}()

		c.JSON(http.StatusOK, gin.H{"message": "点赞成功"})
		return
	}

	// 如果SETNX失败，说明已经点赞，尝试取消点赞
	delResult := global.RedisDB.Del(likeKey).Val()
	if delResult > 0 {
		// 取消点赞成功，减少计数
		global.RedisDB.Decr(countKey)

		// 异步从数据库中删除
		go func() {
			var like models.Like
			global.Db.Where("user_id = ? AND article_id = ?", userID, id).Delete(&like)
			global.Db.Model(&models.Article{}).Where("id = ?", id).Update("like_count", gorm.Expr("like_count - ?", 1))
		}()

		c.JSON(http.StatusOK, gin.H{"message": "已取消点赞"})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
}

// @Summary 评论文章
// @Description 为指定文章添加评论
// @Tags 文章管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path int true "文章ID"
// @Param comment body CreateCommentRequest true "评论信息"
// @Success 200 {object} models.Comment
// @Router /api/articles/{id}/comments [post]
func (ac *ArticleController) CreateComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
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
		ParentID:  req.ParentID,
	}

	if err := global.Db.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建评论失败"})
		return
	}

	// 更新文章评论数
	global.Db.Model(&models.Article{}).Where("id = ?", id).Update("comment_count", gorm.Expr("comment_count + ?", 1))

	c.JSON(http.StatusOK, comment)
}
