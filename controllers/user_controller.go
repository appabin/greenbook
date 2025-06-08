package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetCurrentUserInfo 获取当前登录用户的详细信息
func GetCurrentUserInfo(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 查询用户信息
	var user models.User
	if err := global.Db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 获取关注数和粉丝数
	var followingCount, followerCount int64
	global.Db.Model(&models.UserFollow{}).Where("follower_id = ?", userID).Count(&followingCount)
	global.Db.Model(&models.UserFollow{}).Where("followed_id = ?", userID).Count(&followerCount)

	// 获取文章数
	var articleCount int64
	global.Db.Model(&models.Article{}).Where("author_id = ?", userID).Count(&articleCount)
	// 更新用户表中的统计数据 - 修正字段名称
	updates := map[string]interface{}{
		"following_count": followingCount,
		"followers_count": followerCount, // 修改这里，从follower_count改为followers_count
		"posts_count":     articleCount,  // 修改这里，从article_count改为posts_count
	}

	if err := global.Db.Model(&user).Updates(updates).Error; err != nil {
		// 简单打印错误，不使用日志系统
		fmt.Printf("更新用户统计数据失败: %v\n", err)
	}

	// 更新后重新查询用户信息，确保获取最新数据
	if err := global.Db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 查询当前用户的文章列表
	var userArticles []models.Article
	if err := global.Db.Select("id, title, author_id, like_count, created_at").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname")
		}).
		Where("author_id = ?", userID).
		Order("created_at DESC").
		Find(&userArticles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户文章失败"})
		return
	}
	var favoriteArticles []models.Article
	if err := global.Db.Select("articles.id, articles.title, articles.author_id, articles.like_count, articles.created_at").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname")
		}).
		Joins("JOIN favorites ON favorites.article_id = articles.id").
		Where("favorites.user_id =?", userID).
		Order("favorites.created_at DESC").
		Find(&favoriteArticles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取收藏文章失败"})
		return
	}

	// 查询当前用户点赞过的文章列表
	var likedArticles []models.Article
	if err := global.Db.Select("articles.id, articles.title, articles.author_id, articles.like_count, articles.created_at").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname")
		}).
		Joins("JOIN likes ON likes.article_id = articles.id").
		Where("likes.user_id = ?", userID).
		Order("likes.created_at DESC").
		Find(&likedArticles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取点赞文章失败"})
		return
	}

	// 构建用户文章响应数据
	userArticleList := make([]gin.H, 0)
	for _, article := range userArticles {
		// 获取封面图URL（order=0的图片）
		var coverImageURL string
		var picture models.Picture
		if err := global.Db.Select("pictures.url").
			Joins("JOIN article_pictures ON article_pictures.picture_id = pictures.id").
			Where("article_pictures.article_id = ? AND article_pictures.`order` = 0", article.ID).
			First(&picture).Error; err == nil {
			coverImageURL = picture.URL
		}

		// 检查当前用户是否点赞了这篇文章
		var isLiked bool
		var likeCount int64
		global.Db.Model(&models.Like{}).Where("user_id = ? AND article_id = ?", userID, article.ID).Count(&likeCount)
		isLiked = likeCount > 0

		userArticleList = append(userArticleList, gin.H{
			"id":          article.ID,
			"title":       article.Title,
			"author_name": article.Author.Nickname,
			"cover_url":   coverImageURL,
			"like_count":  article.LikeCount,
			"is_liked":    isLiked,
		})
	}
	// 构建收藏文章响应数据
	favoriteArticleList := make([]gin.H, 0)
	for _, article := range favoriteArticles {
		// 获取封面图URL（order=0的图片）
		var coverImageURL string
		var picture models.Picture
		if err := global.Db.Select("pictures.url").
			Joins("JOIN article_pictures ON article_pictures.picture_id = pictures.id").
			Where("article_pictures.article_id =? AND article_pictures.`order` = 0", article.ID).
			First(&picture).Error; err == nil {
			coverImageURL = picture.URL
		}
		// 检查当前用户是否点赞了这篇文章
		var isLiked bool
		var likeCount int64
		global.Db.Model(&models.Like{}).Where("user_id = ? AND article_id = ?", userID, article.ID).Count(&likeCount)
		isLiked = likeCount > 0
		favoriteArticleList = append(favoriteArticleList, gin.H{
			"id":          article.ID,
			"title":       article.Title,
			"author_name": article.Author.Nickname,
			"cover_url":   coverImageURL,
			"like_count":  article.LikeCount,
			"is_liked":    isLiked,
		})
	}

	// 构建点赞文章响应数据
	likedArticleList := make([]gin.H, 0)
	for _, article := range likedArticles {
		// 获取封面图URL（order=0的图片）
		var coverImageURL string
		var picture models.Picture
		if err := global.Db.Select("pictures.url").
			Joins("JOIN article_pictures ON article_pictures.picture_id = pictures.id").
			Where("article_pictures.article_id = ? AND article_pictures.`order` = 0", article.ID).
			First(&picture).Error; err == nil {
			coverImageURL = picture.URL
		}

		likedArticleList = append(likedArticleList, gin.H{
			"id":          article.ID,
			"title":       article.Title,
			"author_name": article.Author.Nickname,
			"cover_url":   coverImageURL,
			"like_count":  article.LikeCount,
			"is_liked":    true, // 这些都是用户点赞过的文章
		})
	}

	// 返回用户信息和统计数据
	c.JSON(http.StatusOK, gin.H{
		"user":              user,
		"user_articles":     userArticleList,
		"favorite_articles": favoriteArticleList,
		"liked_articles":    likedArticleList,
	})
}

// GetUserProfile 获取指定用户的公开信息
func GetUserProfile(c *gin.Context) {
	// 解析用户ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 查询用户公开信息，过滤敏感字段
	var user models.User
	if err := global.Db.Select("id, nickname, avatar, gender, created_at").First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 获取关注数和粉丝数
	var followingCount, followerCount int64
	global.Db.Model(&models.UserFollow{}).Where("follower_id = ?", id).Count(&followingCount)
	global.Db.Model(&models.UserFollow{}).Where("followed_id = ?", id).Count(&followerCount)

	// 获取文章数
	var articleCount int64
	global.Db.Model(&models.Article{}).Where("author_id = ?", id).Count(&articleCount)

	// 检查当前用户是否关注了该用户
	var isFollowing bool
	currentUserID, exists := c.Get("userID")
	if exists {
		var followCount int64
		global.Db.Model(&models.UserFollow{}).Where("follower_id = ? AND followed_id = ?", currentUserID, id).Count(&followCount)
		isFollowing = followCount > 0
	}

	// 查询该用户的文章列表
	var userArticles []models.Article
	if err := global.Db.Select("id, title, author_id, like_count, created_at").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname")
		}).
		Where("author_id = ?", id).
		Order("created_at DESC").
		Find(&userArticles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户文章失败"})
		return
	}

	// 构建用户文章响应数据
	userArticleList := make([]gin.H, 0)
	for _, article := range userArticles {
		// 获取封面图URL（order=0的图片）
		var coverImageURL string
		var picture models.Picture
		if err := global.Db.Select("pictures.url").
			Joins("JOIN article_pictures ON article_pictures.picture_id = pictures.id").
			Where("article_pictures.article_id = ? AND article_pictures.`order` = 0", article.ID).
			First(&picture).Error; err == nil {
			coverImageURL = picture.URL
		}

		// 检查当前用户是否点赞了这篇文章
		var isLiked bool
		if currentUserID != nil {
			var likeCount int64
			global.Db.Model(&models.Like{}).Where("user_id = ? AND article_id = ?", currentUserID, article.ID).Count(&likeCount)
			isLiked = likeCount > 0
		}

		userArticleList = append(userArticleList, gin.H{
			"id":         article.ID,
			"title":      article.Title,
			"cover_url":  coverImageURL,
			"like_count": article.LikeCount,
			"is_liked":   isLiked,
		})
	}

	// 返回用户公开信息和统计数据
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":              user.ID,
			"nickname":        user.Nickname,
			"avatar":          user.Avatar,
			"gender":          user.Gender,
			"created_at":      user.CreatedAt,
			"following_count": user.FollowingCount,
			"followers_count": user.FollowersCount,
			"posts_count":     user.PostsCount,
		},
		"is_following":  isFollowing,
		"user_articles": userArticleList,
	})
}

// UpdateUserAvatarRequest 更新用户头像请求结构
type UpdateUserAvatarRequest struct {
	PictureID uint `json:"picture_id" binding:"required"`
}

// UpdateUserAvatar 更新用户头像
func UpdateUserAvatar(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 绑定请求参数
	var req UpdateUserAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 根据图片ID查询图片URL
	var picture models.Picture
	if err := global.Db.Select("url").First(&picture, req.PictureID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "图片不存在"})
		return
	}

	// 更新用户头像
	if err := global.Db.Model(&models.User{}).Where("id = ?", userID).Update("avatar", picture.URL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新头像失败"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message":    "头像更新成功",
		"avatar_url": picture.URL,
	})
}
