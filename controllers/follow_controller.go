package controllers

import (
	"net/http"
	"strconv"

	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FollowRequest 关注请求结构
type FollowRequest struct {
	UserID uint `json:"user_id" binding:"required"` // 要关注/取消关注的用户ID
}

// FollowAction 关注/取消关注操作
func FollowAction(c *gin.Context) {
	var req FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 获取当前用户ID
	followerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 不能关注自己
	if uint(followerID.(uint)) == req.UserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能关注自己"})
		return
	}

	// 检查目标用户是否存在
	var targetUser models.User
	if err := global.Db.First(&targetUser, req.UserID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "目标用户不存在"})
		return
	}

	// 检查是否已关注或已取消关注
	var follow models.UserFollow

	if err := global.Db.Where("follower_id =? AND followed_id =?", followerID, req.UserID).First(&follow).Error; err == nil {
		// 已关注，执行取消关注操作
		if err := global.Db.Delete(&follow).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "取消关注失败"})
			return
		}

		// 更新关注者的关注数 -1
		if err := global.Db.Model(&models.User{}).Where("id =?", followerID).UpdateColumn("following_count", gorm.Expr("following_count -?", 1)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新关注数失败"})
			return
		}
		// 更新被关注者的粉丝数 -1
		if err := global.Db.Model(&models.User{}).Where("id =?", req.UserID).UpdateColumn("followers_count", gorm.Expr("followers_count -?", 1)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新粉丝数失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "取消关注成功"})
		return
	}

	// 执行关注操作
	follow = models.UserFollow{
		FollowerID: uint(followerID.(uint)),
		FollowedID: req.UserID,
	}
	if err := global.Db.Create(&follow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "关注失败"})
		return
	}
	// 更新关注者的关注数 +1
	if err := global.Db.Model(&models.User{}).Where("id =?", followerID).UpdateColumn("following_count", gorm.Expr("following_count +?", 1)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新关注数失败"})
		return
	}
	// 更新被关注者的粉丝数 +1
	if err := global.Db.Model(&models.User{}).Where("id =?", req.UserID).UpdateColumn("followers_count", gorm.Expr("followers_count +?", 1)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新粉丝数失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "关注成功"})
}

// GetFollowingList 获取关注列表
func GetFollowingList(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	// 查询关注的用户
	var followings []models.User
	err := global.Db.Table("users").
		Joins("JOIN user_follows ON users.id = user_follows.followed_id").
		Where("user_follows.follower_id = ?", userID).
		Offset(offset).Limit(pageSize).
		Select("users.id, users.username, users.nickname, users.avatar").
		Find(&followings).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取关注列表失败"})
		return
	}

	// 获取总数
	var total int64
	global.Db.Model(&models.UserFollow{}).Where("follower_id = ?", userID).Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"data": followings,
		"meta": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetFollowersList 获取粉丝列表
func GetFollowersList(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	// 查询粉丝用户
	var followers []models.User
	err := global.Db.Table("users").
		Joins("JOIN user_follows ON users.id = user_follows.follower_id").
		Where("user_follows.followed_id = ?", userID).
		Offset(offset).Limit(pageSize).
		Select("users.id, users.username, users.nickname, users.avatar").
		Find(&followers).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取粉丝列表失败"})
		return
	}

	// 获取总数
	var total int64
	global.Db.Model(&models.UserFollow{}).Where("followed_id = ?", userID).Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"data": followers,
		"meta": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
