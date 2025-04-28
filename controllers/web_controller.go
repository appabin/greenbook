package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/models"
	"github.com/appabin/greenbook/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 请求结构体定义（新增到文件顶部）
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
}

// 响应结构体定义（可选，或用gin.H）
type AuthResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录接口
// @Tags 认证
// @Accept json
// @Produce json
// @Param data body LoginRequest true "登录参数"
// @Success 200 {object} map[string]interface{} "返回用户信息和token"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "认证失败"
// @Router /auth/login [post]
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 查询用户
	var user models.User
	if err := global.Db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
		}
		return
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}

	// 生成JWT token（关键修改点：使用UserID）
	token, err := utils.GenerateJWT(fmt.Sprintf("%d", user.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	// 过滤敏感字段后的用户信息
	userSafe := gin.H{
		"id":        user.ID,
		"username":  user.Username,
		"nickname":  user.Nickname,
		"email":     user.Email,
		"phone":     user.Phone,
		"avatar":    user.Avatar,
		"createdAt": user.CreatedAt,
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  userSafe,
	})
}

// Register 用户注册
// @Summary 用户注册
// @Description 用户注册接口
// @Tags 认证
// @Accept json
// @Produce json
// @Param data body RegisterRequest true "注册参数"
// @Success 200 {object} map[string]interface{} "返回用户信息和token"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /auth/register [post]
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 检查用户名唯一性
	var existingUser models.User
	if err := global.Db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		return
	}

	// 加密密码
	hashedPassword, err := utils.HassPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 创建用户模型（仅初始化必要字段）
	user := models.User{
		Username: req.Username,
		Password: hashedPassword,
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		// 以下字段保持默认值或空值
		OpenID: nil,
		Gender: 0,
		Avatar: "",
	}

	if err := global.Db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	// 生成JWT token（关键修改点：使用UserID）
	token, err := utils.GenerateJWT(fmt.Sprintf("%d", user.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	// 过滤敏感字段后的用户信息
	userSafe := gin.H{
		"id":        user.ID,
		"username":  user.Username,
		"nickname":  user.Nickname,
		"email":     user.Email,
		"phone":     user.Phone,
		"createdAt": user.CreatedAt,
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  userSafe,
	})
}
