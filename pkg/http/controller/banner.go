package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"smile.expression/destiny/logger"
	"smile.expression/destiny/pkg/cache"
	"smile.expression/destiny/pkg/database/model"
)

type BannerController struct {
	r           *gin.Engine
	db          *gorm.DB
	cacheClient *cache.Client
}

func NewBannerController(r *gin.Engine, db *gorm.DB, cacheClient *cache.Client) *BannerController {
	return &BannerController{
		r:           r,
		db:          db,
		cacheClient: cacheClient,
	}
}

func (c *BannerController) Register() {
	rg := c.r.Group("/home")

	rg.GET("/banner", c.getBanners)
}

func (c *BannerController) getBanners(ctx *gin.Context) {
	var (
		ctx0 = ctx.Request.Context()
		log  = logger.SmileLog.WithContext(ctx0)
	)

	key := "banner"
	var banners []model.Banner

	data, err := c.cacheClient.Get(ctx0, key)
	if err == nil {
		if err = json.Unmarshal(data, &banners); err != nil {
			log.WithError(err).Error("failed to unmarshal banners")
		} else {
			ctx.JSON(http.StatusOK, gin.H{"result": banners})
			return
		}
	}

	var cacheData []byte
	defer func() {
		cacheData, err = json.Marshal(banners)
		if err = c.cacheClient.Set(ctx0, key, cacheData); err != nil {
			log.WithError(err).Error("failed to cache banners")
		}
	}()

	if err = c.db.Find(&banners).Error; err != nil {
		log.WithError(err).Errorf("failed to fetch banners")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"result": banners,
	})
	return
}
