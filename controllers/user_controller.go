package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/models"
	"github.com/gin-gonic/gin"
)

// GetCurrentUserInfo 获取当前登录用户的详细信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} models.User "用户详细信息"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /api/user/info [get]
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
	// var articleCount int64
	// global.Db.Model(&models.Article{}).Where("user_id = ?", userID).Count(&articleCount)

	// 更新用户表中的统计数据 - 修正字段名称
	updates := map[string]interface{}{
		"following_count": followingCount,
		"followers_count": followerCount, // 修改这里，从follower_count改为followers_count
		// "posts_count":     articleCount,  // 修改这里，从article_count改为posts_count
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

	// 返回用户信息和统计数据 - 修正字段名称
	c.JSON(http.StatusOK, gin.H{
		"user": user,
		// "stats": gin.H{
		// 	"following_count": followingCount,
		// 	"followers_count": followerCount, // 修改这里，保持一致性
		// 	// "posts_count":     articleCount,  // 修改这里，保持一致性
		// },
	})
}

// GetUserProfile 获取指定用户的公开信息
// @Summary 获取用户资料
// @Description 获取指定ID用户的公开信息
// @Tags 用户
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} map[string]interface{} "用户公开信息"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Router /api/user/{id} [get]
// GetUserProfile 获取指定用户的公开信息
func GetUserProfile(c *gin.Context) {
	// 获取路径参数中的用户ID
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 查询用户信息 - 添加统计字段
	var user models.User
	if err := global.Db.Select("id, username, nickname, avatar, bio, created_at, followers_count, following_count, posts_count").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 获取关注数和粉丝数
	var followingCount, followerCount int64
	global.Db.Model(&models.UserFollow{}).Where("follower_id = ?", userID).Count(&followingCount)
	global.Db.Model(&models.UserFollow{}).Where("followed_id = ?", userID).Count(&followerCount)

	// 获取文章数
	var articleCount int64
	global.Db.Model(&models.Article{}).Where("user_id = ?", userID).Count(&articleCount)

	// 检查当前用户是否已关注该用户
	var isFollowing bool = false
	if currentUserID, exists := c.Get("userID"); exists {
		var count int64
		global.Db.Model(&models.UserFollow{}).
			Where("follower_id = ? AND followed_id = ?", currentUserID, userID).
			Count(&count)
		isFollowing = count > 0
	}

	// 返回用户信息和统计数据
	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"stats": gin.H{
			"following_count": followingCount,
			"followers_count": followerCount, // 修改这里，保持一致性
			"posts_count":     articleCount,  // 修改这里，保持一致性
		},
		"is_following": isFollowing,
	})
}
