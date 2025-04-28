// Package controllers API.
// @title Greenbook API
// @version 1.0
// @description Greenbook 服务API文档
// @BasePath /api/v1
package controllers

import (
	"net/http"
	"time"

	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/models"
	"github.com/appabin/greenbook/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LoginRequest 登录请求参数
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求参数
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// WechatLoginRequest 微信小程序登录请求参数
type WechatLoginRequest struct {
	Code      string `json:"code" binding:"required"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
	Gender    uint8  `json:"gender"`
	Country   string `json:"country"`
	Province  string `json:"province"`
	City      string `json:"city"`
	Language  string `json:"language"`
}

// WechatCode2SessionResponse 微信登录凭证校验返回结果
type WechatCode2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// WechatLogin 微信小程序登录
// @Summary 微信小程序登录
// @Description 通过微信小程序登录接口，支持新用户注册和老用户登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param data body WechatLoginRequest true "微信登录参数"
// @Success 200 {object} map[string]interface{} "返回用户信息和token"
// @Success 200 {string} token "JWT令牌"
// @Success 200 {object} models.User "用户信息"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /auth/wechat/login [post]
// WechatLogin 微信小程序登录
func WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 调用微信接口获取 OpenID 和 SessionKey
	wxResp, err := utils.Code2Session(req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "微信登录失败"})
		return
	}

	// 查询用户是否存在
	var user models.User
	if err := global.Db.Where("open_id = ?", wxResp.OpenID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 用户不存在，创建新用户
			user = models.User{
				OpenID:     wxResp.OpenID,
				UnionID:    wxResp.UnionID,
				SessionKey: wxResp.SessionKey,
				Nickname:   req.Nickname,
				Avatar:     req.AvatarURL,
				Gender:     req.Gender,
				Country:    req.Country,
				Province:   req.Province,
				City:       req.City,
				Language:   req.Language,
				Username:   wxResp.OpenID, // 使用 OpenID 作为用户名
			}

			if err := global.Db.Create(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
			return
		}
	}

	// 更新用户信息
	user.Nickname = req.Nickname
	user.Avatar = req.AvatarURL
	user.Gender = req.Gender
	user.Country = req.Country
	user.Province = req.Province
	user.City = req.City
	user.Language = req.Language
	user.LastLoginAt = time.Now()

	if err := global.Db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户信息失败"})
		return
	}

	// 生成JWT token
	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
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
		if err == gorm.ErrRecordNotFound {
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

	// 生成JWT token
	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
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

	// 检查用户名是否已存在
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

	// 创建新用户
	user := models.User{
		Username:    req.Username,
		Password:    hashedPassword,
		Nickname:    req.Nickname,
		Email:       req.Email,
		Phone:       req.Phone,
		LastLoginAt: time.Now(), // 添加这一行，设置注册时间为当前时间
	}

	if dbErr := global.Db.Create(&user).Error; dbErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	// 生成JWT token
	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}
