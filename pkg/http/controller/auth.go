package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"smile.expression/destiny/pkg/cache"
	"smile.expression/destiny/pkg/constant"
	"smile.expression/destiny/pkg/database"
	"smile.expression/destiny/pkg/database/model"
	"smile.expression/destiny/pkg/logger"
	"smile.expression/destiny/pkg/utils"
)

type AuthController struct {
	options     *AuthControllerOptions
	cacheClient *cache.Client
	db          *gorm.DB
}

type AuthControllerOptions struct {
	CacheExpiration int `json:"cacheExpiration"`
}

func NewAuthController(options *AuthControllerOptions, cacheClient *cache.Client, db *gorm.DB) *AuthController {
	return &AuthController{
		options:     options,
		cacheClient: cacheClient,
		db:          db,
	}
}

func (c *AuthController) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			ctx0 = ctx.Request.Context()
			log  = logger.SmileLog.WithContext(ctx0)
		)

		//获取authorization header
		tokenString := ctx.GetHeader(constant.Authorization)
		//验证token格式
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			log.Error("invalid token")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "Unauthorized"})
			return
		}

		tokenString = tokenString[7:]
		token, claims, err := utils.ParseToken(tokenString)
		if err != nil || !token.Valid {
			log.Error("invalid token")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "Unauthorized"})
			return
		}

		//验证通过后获取claim中的userid
		userid := claims.UserID
		key := fmt.Sprintf("user_%d", userid)

		var user model.User

		data, err := c.cacheClient.Get(ctx0, key)
		if err == nil {
			if err = json.Unmarshal(data, &user); err != nil {
				log.WithError(err).Error("failed to unmarshal user")
			} else {
				ctx.Set("user", &user)
				ctx.Next()
				return
			}
		}

		var cacheData []byte
		defer func() {
			cacheData, err = json.Marshal(user)
			if err = c.cacheClient.Set(ctx0, key, cacheData, c.options.CacheExpiration); err != nil {
				log.WithError(err).Error("redis set user error")
			}
		}()

		if err = c.db.First(&user, userid).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.WithError(err).Error("mysql not found user")
			} else {
				log.WithError(err).Error("mysql query user error")
			}
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "Unauthorized"})
			return
		}

		//用户存在 将user的信息写入上下文
		ctx.Set("user", &user)
		ctx.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//获取authorization header
		tokenString := ctx.GetHeader(constant.Authorization)
		//验证token格式
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "Unauthorized"})
			ctx.Abort()
			return
		}

		tokenString = tokenString[7:]
		token, claims, err := utils.ParseToken(tokenString)
		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "Unauthorized"})
			ctx.Abort()
			return
		}

		//验证通过后获取claim中的userid
		userid := claims.UserID

		DB := database.GetDB()
		var user model.User
		DB.First(&user, userid)

		//用户不存在
		if user.ID == 0 {
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "权限不足"})
			ctx.Abort()
			return
		}

		//用户存在 将user的信息写入上下文
		ctx.Set("user", user)

		ctx.Next()
	}
}
