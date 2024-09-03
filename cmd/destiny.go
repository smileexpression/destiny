package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	_ "smile.expression/destiny/log"
	"smile.expression/destiny/pkg/app"
	"smile.expression/destiny/pkg/routes"
)

func main() {
	application := &app.App{}
	application.Init()

	r := gin.Default()
	r = routes.CollectRoute(r)

	port := viper.GetString("server.port")
	if port != "" {
		panic(r.Run(":" + port))
	}
	panic(r.Run())
}
