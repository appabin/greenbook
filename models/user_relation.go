package models

import (
	"gorm.io/gorm"
)

// UserRelation 用户关注关系模型
// @Description 用户关注关系
type UserRelation struct {
	gorm.Model
	FollowerID uint `gorm:"index:idx_follower;comment:关注者ID" json:"follower_id"`  // 粉丝ID
	FollowedID uint `gorm:"index:idx_followed;comment:被关注者ID" json:"followed_id"` // 被关注者ID

	// 关联用户模型
	Follower User `gorm:"foreignKey:FollowerID" json:"follower"`
	Followed User `gorm:"foreignKey:FollowedID" json:"followed"`
}

// TableName 设置表名
func (UserRelation) TableName() string {
	return "user_relations"
}
