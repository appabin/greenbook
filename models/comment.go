package models

import (
	"time"

	"gorm.io/gorm"
)

// Comment 评论模型
type Comment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Content   string         `gorm:"type:text;not null" json:"content"`   // 评论内容
	UserID    uint           `gorm:"not null;index" json:"user_id"`       // 评论用户ID
	User      User           `gorm:"foreignKey:UserID" json:"user"`       // 评论用户
	ArticleID uint           `gorm:"not null;index" json:"article_id"`    // 评论文章ID
	Article   Article        `gorm:"foreignKey:ArticleID" json:"article"` // 评论文章

	LikeCount int `gorm:"default:0" json:"like_count"` // 点赞数
}

// CommentLike 评论点赞模型
type CommentLike struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID    uint    `gorm:"not null;index" json:"user_id"`       // 点赞用户ID
	User      User    `gorm:"foreignKey:UserID" json:"user"`       // 点赞用户
	CommentID uint    `gorm:"not null;index" json:"comment_id"`    // 被点赞评论ID
	Comment   Comment `gorm:"foreignKey:CommentID" json:"comment"` // 被点赞评论
}

func (Comment) TableName() string {
	return "comments"
}

func (CommentLike) TableName() string {
	return "comment_likes"
}
