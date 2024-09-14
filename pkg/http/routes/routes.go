package routes

import (
	"github.com/gin-gonic/gin"

	"smile.expression/destiny/pkg/http/controller"
)

func CollectRoute(r *gin.Engine) *gin.Engine {
	category := r.Group("")
	{
		category.GET("/category", controller.ChooseCategory)
	}

	//r.POST("/login", controller.login)
	//r.POST("/register", controller.register)
	r.GET("/info", controller.AuthMiddleware(), controller.Info)

	//home := r.Group("home")
	//{
	//	//home.GET("/goods", controller.goods)
	//	//home.GET("/banner", controller.GetBanner)
	//	//home.GET("/new", controller.recent)
	//
	//}
	member := r.Group("member")
	{
		member.POST("/order", controller.AuthMiddleware(), controller.CreateOrder)
		member.GET("/order/:id", controller.AuthMiddleware(), controller.GetOrder)
		//member.POST("/release", middleware.AuthMiddleware(), controller.release)
		member.GET("/order/pre", controller.AuthMiddleware(), controller.GetFromCart)
		member.POST("/update_avatar", controller.AuthMiddleware(), controller.UpdateAvatar)
		member.POST("/change_password", controller.AuthMiddleware(), controller.ChangePassword)
		member.POST("/change_info", controller.AuthMiddleware(), controller.ChangeInfo)
		member.POST("/add_address", controller.AuthMiddleware(), controller.AddAddress)
		member.POST("/del_address", controller.AuthMiddleware(), controller.DeleteAddress)
		member.GET("/sold_order", controller.AuthMiddleware(), controller.SoldList)
		member.GET("/get_order", controller.AuthMiddleware(), controller.BoughtList)
		member.GET("/remain", controller.AuthMiddleware(), controller.SaleList)
	}

	goods := r.Group("")
	{
		goods.GET("/goods", controller.GetOneGood)
		goods.GET("/goods/relevant", controller.AuthMiddleware(), controller.RecommendGoods)
	}

	chatList := r.Group("chat")
	{
		chatList.GET("/get_msg", controller.AuthMiddleware(), controller.GetMsg)
		chatList.POST("/send_msg", controller.AuthMiddleware(), controller.SendMsg)
		chatList.POST("/add_chat", controller.AuthMiddleware(), controller.AddChat)
	}

	//member路由完善后可以将下面这个路由整合
	CartGroup := r.Group("member/cart")
	{
		CartGroup.POST("/add", controller.AuthMiddleware(), controller.CartIn)
		CartGroup.GET("/pull", controller.AuthMiddleware(), controller.CartOut)
		CartGroup.DELETE("/del", controller.AuthMiddleware(), controller.CartDel)
		CartGroup.DELETE("/del2", controller.AuthMiddleware(), controller.CartDelOne)
	}

	//imageRoutes := r.Group("/image")
	//{
	//	imageRoutes.POST("/upload", controller.upload)
	//	imageRoutes.GET("/get", controller.HandleImage)
	//	imageRoutes.POST("/delete", controller.removeObject)
	//}

	return r
}
