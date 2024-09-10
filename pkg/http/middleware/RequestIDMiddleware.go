package middleware

import (
	"context"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"smile.expression/destiny/pkg/constant"
)

func GenerateRequestID() gin.HandlerFunc {
	return requestid.New(
		requestid.WithGenerator(func() string {
			return uuid.New().String()
		}),
	)
}

func SetRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(constant.XRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(constant.XRequestID, requestID)

		c.Writer.Header().Set(constant.XRequestID, requestID)

		// 将 request ID 写入原生 context.Context
		ctx := context.WithValue(c.Request.Context(), constant.XRequestID, requestID)

		// 将带有 request ID 的 context 替换到 gin.Context 的请求中
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
