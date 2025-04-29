package main

import (
	"fmt"
	"log"

	"github.com/appabin/greenbook/router"
	"github.com/appabin/greenbook/config"
	
	// 保留swagger相关包的导入
	_ "github.com/appabin/greenbook/docs"
)

// @title Greenbook API
// @version 1.0
// @description Greenbook社交博客平台API文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api
// @schemes http https
func main() {
	config.InitConfig()

	log.Println("=== 配置加载成功 ===")
	fmt.Printf("应用名称: %s\n", config.AppConfig.App.Name)
	fmt.Printf("应用端口: %s\n", config.AppConfig.App.Port)

	r := router.SetupRouter()
	
	// 删除这一行，因为在router.go中已经添加了swagger路由
	// r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run(":" + config.AppConfig.App.Port)
}
