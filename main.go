package main

import (
	"fmt"
	"log"

	"github.com/appabin/greenbook/router"

	"github.com/appabin/greenbook/config"
)

func main() {
	config.InitConfig()

	log.Println("=== 配置加载成功 ===")
	fmt.Printf("应用名称: %s\n", config.AppConfig.App.Name)
	fmt.Printf("应用端口: %s\n", config.AppConfig.App.Port)

	r := router.SetupRouter()

	r.Run(":" + config.AppConfig.App.Port)
}
