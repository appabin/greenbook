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

// ToggleLike 点赞文章
func ArticleToggleLike(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("article_id"), 10, 32)
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

func CommentToggleLike(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("comment_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的评论ID"})
		return
	}

	userID := c.GetUint("userID")
	likeKey := fmt.Sprintf("comment:like:%d:%d", id, userID)
	countKey := fmt.Sprintf("comment:like_count:%d", id)

	// 使用Redis的SETNX命令尝试添加点赞
	setResult := global.RedisDB.SetNX(likeKey, "1", 0).Val()
	if setResult {
		// 点赞成功，增加计数
		global.RedisDB.Incr(countKey)

		// 异步保存到数据库
		go func() {
			commentLike := models.CommentLike{
				UserID:    userID,
				CommentID: uint(id),
			}
			global.Db.Create(&commentLike)
			// 更新评论的点赞数
			global.Db.Model(&models.Comment{}).Where("id = ?", id).Update("like_count", gorm.Expr("like_count + ?", 1))
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
			var commentLike models.CommentLike
			global.Db.Where("user_id = ? AND comment_id = ?", userID, id).Delete(&commentLike)
			// 更新评论的点赞数
			global.Db.Model(&models.Comment{}).Where("id = ?", id).Update("like_count", gorm.Expr("like_count - ?", 1))
		}()

		c.JSON(http.StatusOK, gin.H{"message": "已取消点赞"})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
}
