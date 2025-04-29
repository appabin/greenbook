package docs

import (
	"time"
)

// 这个文件用于为 Swagger 提供类型定义
// 实际上不会被导入到项目中，只是为了让 Swagger 能够正确解析类型

// DeletedAt 用于替代 gorm.DeletedAt 类型
type DeletedAt struct {
	Time  time.Time
	Valid bool
}