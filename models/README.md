# Models 包

本目录包含应用程序的数据模型定义，使用GORM进行数据库映射。

## 模型列表

### User 模型

`User` 模型定义了用户信息，包含基本用户信息和微信小程序相关字段：

#### 基本字段
- ID, CreatedAt, UpdatedAt, DeletedAt (来自gorm.Model)
- Username: 用户名
- Nickname: 昵称
- Avatar: 头像URL
- Gender: 性别(0:未知,1:男,2:女)
- Phone: 手机号
- Email: 邮箱

#### 微信小程序相关字段
- OpenID: 微信openid (唯一索引)
- UnionID: 微信unionid
- SessionKey: 微信session_key
- Country: 国家
- Province: 省份
- City: 城市
- Language: 语言

#### 其他字段
- LastLoginAt: 最后登录时间

## 使用方法

在应用程序中引入模型：

```go
import "github.com/appabin/greenbook/models"
```

创建用户示例：

```go
user := models.User{
    Username: "example_user",
    Nickname: "示例用户",
    OpenID:   "wx_openid_example",
    // 其他字段...
}
```