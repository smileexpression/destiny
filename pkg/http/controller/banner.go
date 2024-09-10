package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"smile.expression/destiny/logger"
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
	rg := b.r.Group("/home")

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
