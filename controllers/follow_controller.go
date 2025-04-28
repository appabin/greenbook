package controllers

import (
	"net/http"
	"strconv"

	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/models"
	"github.com/gin-gonic/gin"
)

// FollowRequest 关注请求结构
type FollowRequest struct {
	UserID uint `json:"user_id" binding:"required"` // 要关注/取消关注的用户ID
	Action int  `json:"action" binding:"required"`  // 1: 关注, 0: 取消关注
}

// FollowAction 关注/取消关注操作
// @Summary 关注/取消关注用户
// @Description 关注或取消关注指定用户
// @Tags 社交
// @Accept json
// @Produce json
// @Param data body FollowRequest true "关注请求参数"
// @Success 200 {object} map[string]interface{} "操作成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /api/follow [post]
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

	// 根据action执行关注或取消关注
	if req.Action == 1 {
		// 关注操作
		follow := models.UserFollow{
			FollowerID: followerID.(uint),
			FollowedID: req.UserID,
		}

		// 检查是否已经关注
		var existingFollow models.UserFollow
		result := global.Db.Where("follower_id = ? AND followed_id = ?", followerID, req.UserID).First(&existingFollow)
		if result.Error == nil {
			c.JSON(http.StatusOK, gin.H{"message": "已经关注该用户"})
			return
		}

		if err := global.Db.Create(&follow).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "关注失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "关注成功"})
	} else if req.Action == 0 {
		// 取消关注操作
		result := global.Db.Where("follower_id = ? AND followed_id = ?", followerID, req.UserID).Delete(&models.UserFollow{})
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "取消关注失败"})
			return
		}
		if result.RowsAffected == 0 {
			c.JSON(http.StatusOK, gin.H{"message": "未关注该用户"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "取消关注成功"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的操作类型"})
	}
}

// GetFollowingList 获取关注列表
// @Summary 获取关注列表
// @Description 获取当前用户的关注列表
// @Tags 社交
// @Accept json
// @Produce json
// @Success 200 {array} models.User "关注列表"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /api/follow/following [get]
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
// @Summary 获取粉丝列表
// @Description 获取当前用户的粉丝列表
// @Tags 社交
// @Accept json
// @Produce json
// @Success 200 {array} models.User "粉丝列表"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /api/follow/followers [get]
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
