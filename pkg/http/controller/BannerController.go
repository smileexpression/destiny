package controller

import (
	"github.com/gin-gonic/gin"

	"smile.expression/destiny/pkg/database"
	"smile.expression/destiny/pkg/database/model"
)

func GetBanner(ctx *gin.Context) {

	DB := database.GetDB()
	var banners []model.Banner
	DB.Find(&banners)

	ctx.JSON(200, gin.H{
		"code":   "1",
		"msg":    "获取轮播图数据成功",
		"result": banners,
	})

}
