package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminLogin 管理员登录
func AdminLogin(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 简单的硬编码管理员账号验证
	if req.Username != "admin" || req.Password != "admin123" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   "admin_token_" + strconv.FormatInt(time.Now().Unix(), 10),
	})
}

// AdminGetUserList 获取用户列表（分页）
func AdminGetUserList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// 查询用户总数
	var total int64
	global.Db.Model(&models.User{}).Count(&total)

	// 查询用户列表
	var users []models.User
	global.Db.Select("id, nickname, avatar, gender, phone, email, created_at, following_count, followers_count, posts_count").
		Offset(offset).Limit(limit).Order("created_at DESC").Find(&users)

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"limit": limit,
		"users": users,
	})
}

// AdminDeleteUser 软删除用户
func AdminDeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 软删除用户
	if err := global.Db.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
}

// GetStatistics 获取数据统计
func GetStatistics(c *gin.Context) {
	// 用户统计
	var userCount int64
	global.Db.Model(&models.User{}).Count(&userCount)

	// 文章统计
	var articleCount int64
	global.Db.Model(&models.Article{}).Count(&articleCount)

	// 点赞统计
	var likeCount int64
	global.Db.Model(&models.Like{}).Count(&likeCount)

	// 收藏统计
	var favoriteCount int64
	global.Db.Model(&models.Favorite{}).Count(&favoriteCount)

	// 关注统计
	var followCount int64
	global.Db.Model(&models.UserFollow{}).Count(&followCount)

	// 今日新增用户
	var todayUserCount int64
	today := time.Now().Format("2006-01-02")
	global.Db.Model(&models.User{}).Where("DATE(created_at) = ?", today).Count(&todayUserCount)

	// 今日新增文章
	var todayArticleCount int64
	global.Db.Model(&models.Article{}).Where("DATE(created_at) = ?", today).Count(&todayArticleCount)

	// 最近7天用户注册趋势
	type DailyCount struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	var userTrend []DailyCount
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var count int64
		global.Db.Model(&models.User{}).Where("DATE(created_at) = ?", date).Count(&count)
		userTrend = append(userTrend, DailyCount{
			Date:  date,
			Count: count,
		})
	}

	// 最近7天文章发布趋势
	var articleTrend []DailyCount
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var count int64
		global.Db.Model(&models.Article{}).Where("DATE(created_at) = ?", date).Count(&count)
		articleTrend = append(articleTrend, DailyCount{
			Date:  date,
			Count: count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"overview": gin.H{
			"user_count":         userCount,
			"article_count":      articleCount,
			"like_count":         likeCount,
			"favorite_count":     favoriteCount,
			"follow_count":       followCount,
			"today_user_count":   todayUserCount,
			"today_article_count": todayArticleCount,
		},
		"trends": gin.H{
			"user_trend":    userTrend,
			"article_trend": articleTrend,
		},
	})
}

// AdminGetArticleList 获取文章列表（分页）
func AdminGetArticleList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// 查询文章总数
	var total int64
	global.Db.Model(&models.Article{}).Count(&total)

	// 查询文章列表
	var articles []models.Article
	global.Db.Select("id, title, author_id, like_count, created_at").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname")
		}).
		Offset(offset).Limit(limit).Order("created_at DESC").Find(&articles)

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     page,
		"limit":    limit,
		"articles": articles,
	})
}

// AdminDeleteArticle 软删除文章
func AdminDeleteArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章ID"})
		return
	}

	// 软删除文章
	if err := global.Db.Delete(&models.Article{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除文章失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文章删除成功"})
}