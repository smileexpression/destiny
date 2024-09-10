package middleware

import (
	"net/http"
	"smile.expression/destiny/pkg/database"
	"smile.expression/destiny/pkg/database/model"
	"strings"

	"github.com/gin-gonic/gin"

	"smile.expression/destiny/pkg/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//获取authorization header
		tokenString := ctx.GetHeader("Authorization")
		//验证token格式
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "权限不足",
			})
			ctx.Abort()
			return
		}

		tokenString = tokenString[7:]

		token, claims, err := utils.ParseToken(tokenString)
		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "权限不足"})
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
