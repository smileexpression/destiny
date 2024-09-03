package main

import (
	"os"
	"smile.expression/destiny/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"smile.expression/destiny/pkg/routes"
)

func main() {
	InitConfig()
	database.InitDB()
	r := gin.Default()
	r = routes.CollectRoute(r)

	port := viper.GetString("server.port")
	if port != "" {
		panic(r.Run(":" + port))
	}
	panic(r.Run())
}

func InitConfig() {
	workDir, _ := os.Getwd()
	viper.SetConfigName("application")
	viper.SetConfigType("yml")
	viper.AddConfigPath(workDir + "/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
