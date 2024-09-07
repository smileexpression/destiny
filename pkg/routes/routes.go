package routes

import (
	"github.com/gin-gonic/gin"

	"smile.expression/destiny/pkg/controller"
	"smile.expression/destiny/pkg/middleware"
)

func CollectRoute(r *gin.Engine) *gin.Engine {
	r.Use(middleware.CORSMiddleware(), middleware.RecoveryMiddleware())
	category := r.Group("")
	{
		category.GET("/category", controller.ChooseCategory)
	}

	r.POST("/login", controller.Login)
	//r.POST("/register", controller.register)
	r.GET("/info", middleware.AuthMiddleware(), controller.Info)

	home := r.Group("home")
	{
		home.GET("/goods", controller.GetGoods)
		home.GET("/banner", controller.GetBanner)
		home.GET("/new", controller.RecentIdle)

	}
	member := r.Group("member")
	{
		member.POST("/order", middleware.AuthMiddleware(), controller.CreateOrder)
		member.GET("/order/:id", middleware.AuthMiddleware(), controller.GetOrder)
		member.POST("/release", middleware.AuthMiddleware(), controller.Release)
		member.GET("/order/pre", middleware.AuthMiddleware(), controller.GetFromCart)
		member.POST("/update_avatar", middleware.AuthMiddleware(), controller.UpdateAvatar)
		member.POST("/change_password", middleware.AuthMiddleware(), controller.ChangePassword)
		member.POST("/change_info", middleware.AuthMiddleware(), controller.ChangeInfo)
		member.POST("/add_address", middleware.AuthMiddleware(), controller.AddAddress)
		member.POST("/del_address", middleware.AuthMiddleware(), controller.DeleteAddress)
		member.GET("/sold_order", middleware.AuthMiddleware(), controller.SoldList)
		member.GET("/get_order", middleware.AuthMiddleware(), controller.BoughtList)
		member.GET("/remain", middleware.AuthMiddleware(), controller.SaleList)
	}

	goods := r.Group("")
	{
		goods.GET("/goods", controller.GetOneGood)
		goods.GET("/goods/relevant", middleware.AuthMiddleware(), controller.RecommendGoods)
	}

	chatList := r.Group("chat")
	{
		chatList.GET("/get_msg", middleware.AuthMiddleware(), controller.GetMsg)
		chatList.POST("/send_msg", middleware.AuthMiddleware(), controller.SendMsg)
		chatList.POST("/add_chat", middleware.AuthMiddleware(), controller.AddChat)
	}

	//member路由完善后可以将下面这个路由整合
	CartGroup := r.Group("member/cart")
	{
		CartGroup.POST("/add", middleware.AuthMiddleware(), controller.CartIn)
		CartGroup.GET("/pull", middleware.AuthMiddleware(), controller.CartOut)
		CartGroup.DELETE("/del", middleware.AuthMiddleware(), controller.CartDel)
		CartGroup.DELETE("/del2", middleware.AuthMiddleware(), controller.CartDelOne)
	}

	imageRoutes := r.Group("/image")
	{
		imageRoutes.POST("/upload", controller.HandleUpload)
		imageRoutes.GET("/get", controller.HandleImage)
		imageRoutes.POST("/delete", controller.DeleteImage)
	}

	return r
}
