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

	currentUserID, exists := c.Get("userID")
	
	// 预加载作者信息（包含头像）
	query := global.Db.Model(&models.Article{}).Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, nickname, avatar")
	})

	if exists {
		// 混合推荐算法：关注作者(权重3) + 点赞历史作者(权重2) + 热度(权重1) - 已点赞文章(权重-2)
		query = query.Select(`articles.*, 
			(
				CASE WHEN user_follows.followed_id IS NOT NULL THEN 3 ELSE 0 END +
				CASE WHEN liked_authors.author_id IS NOT NULL THEN 2 ELSE 0 END +
				(articles.like_count / 10) -
				CASE WHEN user_likes.article_id IS NOT NULL THEN 2 ELSE 0 END
			) as recommendation_score`).
			Joins("LEFT JOIN user_follows ON articles.author_id = user_follows.followed_id AND user_follows.follower_id = ? AND user_follows.deleted_at IS NULL", currentUserID).
			Joins(`LEFT JOIN (
				SELECT DISTINCT articles.author_id 
				FROM likes 
				JOIN articles ON likes.article_id = articles.id 
				WHERE likes.user_id = ? AND likes.deleted_at IS NULL
			) as liked_authors ON articles.author_id = liked_authors.author_id`, currentUserID).
			Joins("LEFT JOIN likes as user_likes ON articles.id = user_likes.article_id AND user_likes.user_id = ? AND user_likes.deleted_at IS NULL", currentUserID).
			Order("recommendation_score DESC, articles.created_at DESC")
	} else {
		// 未登录用户按热度排序
		query = query.Order("articles.like_count DESC, articles.created_at DESC")
	}

	query.Count(&total)
	result := query.Offset((page - 1) * size).Limit(size).Find(&articles)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文章列表失败"})
		return
	}

	// 构建文章列表响应数据
	articleList := make([]gin.H, 0)
	for _, article := range articles {
		// 获取封面图片URL
		var coverImageURL string
		var picture models.Picture
		err := global.Db.Joins("JOIN article_pictures ON pictures.id = article_pictures.picture_id").
			Where("article_pictures.article_id = ? AND article_pictures.`order` = 0", article.ID).
			First(&picture).Error
		if err == nil {
			coverImageURL = picture.URL
		}

		// 检查当前用户是否点赞了该文章
		var isLiked bool
		if exists {
			var likeCount int64
			global.Db.Model(&models.Like{}).Where("user_id = ? AND article_id = ? AND deleted_at IS NULL", currentUserID, article.ID).Count(&likeCount)
			isLiked = likeCount > 0
		}

		// 生成推荐理由
		recommendationReason := ""
		if exists {
			var isFollowing bool
			var hasLikedAuthor bool
			
			// 检查是否关注作者
			var followCount int64
			global.Db.Model(&models.UserFollow{}).Where("follower_id = ? AND followed_id = ? AND deleted_at IS NULL", currentUserID, article.AuthorID).Count(&followCount)
			isFollowing = followCount > 0
			
			// 检查是否点赞过该作者的其他文章
			var likeCount int64
			global.Db.Model(&models.Like{}).
				Joins("JOIN articles ON likes.article_id = articles.id").
				Where("likes.user_id = ? AND articles.author_id = ? AND likes.deleted_at IS NULL AND articles.id != ?", currentUserID, article.AuthorID, article.ID).
				Count(&likeCount)
			hasLikedAuthor = likeCount > 0
			
			if isLiked {
				recommendationReason = "你已点赞过的文章"
			} else if isFollowing {
				recommendationReason = "来自你关注的作者"
			} else if hasLikedAuthor {
				recommendationReason = "你可能感兴趣的作者"
			} else if article.LikeCount > 50 {
				recommendationReason = "热门文章"
			}
		}

		articleList = append(articleList, gin.H{
			"id":                    article.ID,
			"title":                 article.Title,
			"author_name":           article.Author.Nickname,
			"author_avatar":         article.Author.Avatar,
			"cover_url":             coverImageURL,
			"like_count":            article.LikeCount,
			"is_liked":              isLiked,
			"recommendation_reason": recommendationReason,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": articleList,
	})
}

// GetFollowArticleList 获取关注用户的文章列表
func GetFollowArticleList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	// 获取当前用户ID
	currentUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	var articles []models.Article
	var total int64

	// 查询当前用户关注的用户发布的文章
	query := global.Db.Model(&models.Article{}).
		Joins("JOIN user_follows ON articles.author_id = user_follows.followed_id").
		Where("user_follows.follower_id = ?", currentUserID).
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, avatar")
		})
	
	query.Count(&total)
	result := query.Order("articles.created_at DESC").Offset((page - 1) * size).Limit(size).Find(&articles)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取关注文章列表失败"})
		return
	}

	// 构建文章列表响应数据
	articleList := make([]gin.H, 0)
	for _, article := range articles {
		// 获取封面图片URL
		var coverImageURL string
		var picture models.Picture
		err := global.Db.Joins("JOIN article_pictures ON pictures.id = article_pictures.picture_id").
			Where("article_pictures.article_id = ? AND article_pictures.`order` = 0", article.ID).
			First(&picture).Error
		if err == nil {
			coverImageURL = picture.URL
		}

		// 检查当前用户是否点赞了该文章
		var isLiked bool
		var likeCount int64
		global.Db.Model(&models.Like{}).Where("user_id = ? AND article_id = ? AND deleted_at IS NULL", currentUserID, article.ID).Count(&likeCount)
		isLiked = likeCount > 0

		articleList = append(articleList, gin.H{
			"id":            article.ID,
			"title":         article.Title,
			"author_name":   article.Author.Nickname,
			"author_avatar": article.Author.Avatar,
			"cover_url":     coverImageURL,
			"like_count":    article.LikeCount,
			"is_liked":      isLiked,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": articleList,
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

	// 检查当前用户是否点赞了该文章
	var isLiked bool
	// 检查当前用户是否收藏了该文章
	var isFavorited bool
	// 检查当前用户是否关注了作者
	var isFollowed bool
	// 检查当前用户是否为文章作者
	var isAuthor bool
	currentUserID, exists := c.Get("userID")
	if exists {
		var likeCount int64
		global.Db.Model(&models.Like{}).Where("user_id = ? AND article_id = ?", currentUserID, article.ID).Count(&likeCount)
		isLiked = likeCount > 0

		var favoriteCount int64
		global.Db.Model(&models.Favorite{}).Where("user_id = ? AND article_id = ?", currentUserID, article.ID).Count(&favoriteCount)
		isFavorited = favoriteCount > 0

		// 检查是否关注了作者
		var followCount int64
		global.Db.Model(&models.UserFollow{}).Where("follower_id = ? AND followed_id = ?", currentUserID, article.AuthorID).Count(&followCount)
		isFollowed = followCount > 0

		// 检查是否为文章作者
		isAuthor = currentUserID == article.AuthorID
	}
	// 构建评论列表
	var filteredComments []gin.H
	for _, comment := range article.Comments {
		// 检查当前用户是否对该评论点赞
		var commentIsLiked bool
		if exists {
			var commentLike models.CommentLike
			err := global.Db.Where("user_id = ? AND comment_id = ?", currentUserID, comment.ID).First(&commentLike).Error
			commentIsLiked = err == nil
		}

		filteredComments = append(filteredComments, gin.H{
			"id":         comment.ID,
			"content":    comment.Content,
			"created_at": comment.CreatedAt,
			"like_count": comment.LikeCount,
			"is_liked":   commentIsLiked,
			"user": gin.H{
				"id":       comment.User.ID,
				"nickname": comment.User.Nickname,
				"avatar":   comment.User.Avatar,
			},
		})
	}

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
		"comments": filteredComments,
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
		"is_liked":       isLiked,
		"is_favorited":   isFavorited,
		"is_followed":    isFollowed,
		"is_author":      isAuthor,
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

// SearchArticles 搜索文章
func SearchArticles(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	// 获取当前用户ID（可选）
	currentUserID, exists := c.Get("userID")

	var articles []models.Article
	query := global.Db.Model(&models.Article{}).
		Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, avatar")
		}).
		Joins("LEFT JOIN pictures ON pictures.id = (SELECT picture_id FROM article_pictures WHERE article_id = articles.id LIMIT 1)").
		Select("articles.*, pictures.url as cover_url").
		Order("articles.created_at DESC")

	if err := query.Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索文章失败"})
		return
	}

	// 构建响应数据
	articleList := make([]gin.H, 0)
	for _, article := range articles {
		// 获取封面图片URL
		var coverURL string
		var picture models.Picture
		if err := global.Db.Table("pictures").
			Joins("JOIN article_pictures ON pictures.id = article_pictures.picture_id").
			Where("article_pictures.article_id = ?", article.ID).
			First(&picture).Error; err == nil {
			coverURL = picture.URL
		}

		// 检查当前用户是否点赞了该文章
		isLiked := false
		if exists {
			var likeCount int64
			global.Db.Model(&models.Like{}).
				Where("user_id = ? AND article_id = ? AND deleted_at IS NULL", currentUserID, article.ID).
				Count(&likeCount)
			isLiked = likeCount > 0
		}

		articleList = append(articleList, gin.H{
			"id":            article.ID,
			"title":         article.Title,
			"author_name":   article.Author.Nickname,
			"author_avatar": article.Author.Avatar,
			"cover_url":     coverURL,
			"like_count":    article.LikeCount,
			"is_liked":      isLiked,
		})
	}

	c.JSON(http.StatusOK, articleList)
}

// SearchUsers 搜索用户
func SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	// 获取当前用户ID（可选）
	currentUserID, exists := c.Get("userID")

	var users []models.User
	query := global.Db.Model(&models.User{}).
		Where("nickname LIKE ?", "%"+keyword+"%").
		Select("id, nickname, avatar").
		Order("created_at DESC")

	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索用户失败"})
		return
	}

	// 构建响应数据
	userList := make([]gin.H, 0)
	for _, user := range users {
		// 检查当前用户是否关注了该用户
		isFollowing := false
		if exists && currentUserID != user.ID {
			var followCount int64
			global.Db.Model(&models.UserFollow{}).
				Where("follower_id = ? AND followed_id = ?", currentUserID, user.ID).
				Count(&followCount)
			isFollowing = followCount > 0
		}

		userList = append(userList, gin.H{
			"id":          user.ID,
			"name":        user.Nickname,
			"avatar":      user.Avatar,
			"description": user.Nickname, // 使用昵称作为描述
			"isFollowing": isFollowing,
		})
	}

	c.JSON(http.StatusOK, userList)
}

// SearchTags 搜索标签
func SearchTags(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	var tags []models.Tag
	query := global.Db.Model(&models.Tag{}).
		Where("name LIKE ?", "%"+keyword+"%").
		Order("created_at DESC")

	if err := query.Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索标签失败"})
		return
	}

	// 构建响应数据
	tagList := make([]gin.H, 0)
	for _, tag := range tags {
		// 统计该标签下的文章数量
		var articleCount int64
		global.Db.Table("article_tags").
			Where("tag_id = ?", tag.ID).
			Joins("JOIN articles ON articles.id = article_tags.article_id AND articles.deleted_at IS NULL").
			Count(&articleCount)

		tagList = append(tagList, gin.H{
			"id":           tag.ID,
			"name":         tag.Name,
			"articleCount": articleCount,
		})
	}

	c.JSON(http.StatusOK, tagList)
}
