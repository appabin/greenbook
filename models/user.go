package models

import (
	"time"

	"gorm.io/gorm"
)

// 在文件顶部添加以下注释

// User 用户模型
// @Description 用户信息
type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// swagger:ignore
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 使用 swagger:ignore 标记

	Username string `gorm:"size:50;comment:用户名;default:'' " json:"username"`
	Password string `gorm:"size:100;comment:密码;default:'' " json:"password"`

	Nickname string `gorm:"size:50;comment:昵称" json:"nickname"`
	Avatar   string `gorm:"size:500;comment:头像URL" json:"avatar"`
	Gender   uint8  `gorm:"default:0;comment:性别(0:未知,1:男,2:女)" json:"gender"`
	Phone    string `gorm:"size:20;uniqueIndex;comment:手机号" json:"phone"`
	Email    string `gorm:"size:100;uniqueIndex;comment:邮箱" json:"email"`

	// 微信小程序相关字段
	OpenID     *string `gorm:"size:50;uniqueIndex;comment:微信openid;default:NULL" json:"open_id"` // 改为指针类型
	UnionID    string  `gorm:"size:50;comment:微信unionid;default:NULL" json:"union_id"`
	SessionKey string  `gorm:"size:100;comment:微信session_key;default:NULL" json:"session_key"`

	// 论坛社交相关字段
	FollowersCount uint `gorm:"default:0;comment:粉丝数" json:"followers_count"`
	FollowingCount uint `gorm:"default:0;comment:关注数" json:"following_count"`
	PostsCount     uint `gorm:"default:0;comment:发帖数" json:"posts_count"`

	Following []User `gorm:"many2many:user_follows;foreignKey:ID;joinForeignKey:FollowerID;joinReferences:FollowedID" json:"-"`
	Followers []User `gorm:"many2many:user_follows;foreignKey:ID;joinForeignKey:FollowedID;joinReferences:FollowerID" json:"-"`
}

// TableName 设置表名
func (User) TableName() string {
	return "users"
}
