package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"smile.expression/destiny/logger"
	"smile.expression/destiny/pkg/database"
	"smile.expression/destiny/pkg/database/model"
)

type BannerController struct {
	r  *gin.Engine
	db *gorm.DB
}

func NewBannerController(r *gin.Engine, db *gorm.DB) *BannerController {
	return &BannerController{
		r:  r,
		db: db,
	}
}

func (b *BannerController) Register() {
	rg := b.r.Group("/api/v1/home")

	rg.GET("/banner", b.getBanners)
}

func (b *BannerController) getBanners(c *gin.Context) {
	var (
		ctx0 = c.Request.Context()
		log  = logger.SmileLog.WithContext(ctx0)
	)

	var banners []model.Banner
	if err := b.db.Find(&banners).Error; err != nil {
		log.WithError(err).Errorf("getBanners fail")
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, gin.H{
		"result": banners,
	})
}

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
