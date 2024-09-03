package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"smile.expression/destiny/pkg/common"
	"smile.expression/destiny/pkg/model"
)

type OrderInfo struct {
	Id        string `json:"goodId"` //goods_Id
	AddressId string `json:"addressId"`
}

func CreateOrder(ctx *gin.Context) {

	user, isExist := ctx.Get("user")
	if !isExist {
		ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "user not exist"})
		return
	}
	userInfo := user.(model.User)

	var orderInfo OrderInfo
	//绑定结构体,接收body
	if err := ctx.BindJSON(&orderInfo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	//生成订单
	var good model.Goods
	DB := common.GetDB()
	DB.Where("ID=?", orderInfo.Id).Find(&good)
	if good.IsSold {
		ctx.JSON(200, gin.H{
			"code": "0",
			"msg":  "手速太慢，商品被抢",
		})
		return
	}
	DB.Model(&good).Update("is_sold", true) //要不要考虑多线程？
	pMoney, _ := strconv.Atoi(good.Price)   //商品的价格为string，订单pay为int（字段类型）

	order := model.Order{
		GoodId:    orderInfo.Id,
		AddressId: orderInfo.AddressId,
		UserId:    userInfo.ID,
		PayMoney:  pMoney,
	}
	if err := DB.Create(&order).Error; err != nil {
		fmt.Println("插入失败", err)
		return
	}
	//fmt.Print(good.Is_Sold)

	//返回id
	ctx.JSON(200, gin.H{
		"code": "1",
		"msg":  "操作成功",
		"id":   order.ID, //订单id
	})

}

func GetOrder(ctx *gin.Context) {

	orderId := ctx.Param("id")
	fmt.Print(orderId)
	DB := common.GetDB()

	var Order model.Order
	err := DB.Where("ID=?", orderId).Find(&Order)

	if err != nil { //应该没有
		// ctx.JSON(400, gin.H{
		// 	"code": "0",
		// 	"msg":  "订单不存在",
		// })
		fmt.Print(err)
	}
	ctx.JSON(200, gin.H{
		"code":      "1",
		"msg":       "操作成功",
		"countdown": "1800",
		"payMoney":  Order.PayMoney,
	})
	fmt.Print(Order.PayMoney)
	fmt.Print("???\n")
	fmt.Print(Order.AddressId)
}

type SendAddress struct {
	Id       string `json:"id"`
	Receiver string `json:"receiver"`
	Contact  string `json:"contact"`
	Address  string `json:"address"`
}
type SendGood struct {
	Id          string `json:"id"`
	User        string `json:"user"`
	Name        string `json:"name"`
	Description string `json:"desc"`
	Picture     string `json:"picture"`
	Price       string `json:"price"`
}
type Result struct {
	UserAddresses []SendAddress `json:"userAddresses"`
	Goods         SendGood      `json:"goods"`
	Price         string        `json:"price"`
}

func GetFromCart(ctx *gin.Context) {

	idStr := ctx.Query("goodID")

	user, isExist := ctx.Get("user")
	if !isExist {
		ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "user not exist"})
		return
	}
	userInfo := user.(model.User)

	DB := common.GetDB()

	//获取数据库的相关数据
	var good model.Goods
	e := DB.Table("goods").Where("id = ?", idStr).Find(&good)
	if e.Error != nil {
		fmt.Print(e.Error)
	}
	var address []model.UserAddress
	e2 := DB.Table("user_addresses").Where("user_id = ?", userInfo.ID).Find(&address)
	if e2.Error != nil {
		fmt.Print(e.Error)
	}

	//填充发送的具体数据
	var sendGood SendGood
	sendGood.Id = strconv.Itoa(int(good.ID))
	sendGood.Name = good.Name
	sendGood.User = good.User
	sendGood.Description = good.Description
	sendGood.Picture = good.Picture
	sendGood.Price = good.Price
	//
	addrNum := len(address)
	sendAddress := make([]SendAddress, addrNum)
	for i := 0; i < addrNum; i++ {
		sendAddress[i].Id = strconv.Itoa(int(address[i].ID))
		sendAddress[i].Receiver = address[i].Receiver
		sendAddress[i].Contact = address[i].Contact
		sendAddress[i].Address = address[i].Address
	}
	//
	var result Result
	result.UserAddresses = sendAddress
	result.Goods = sendGood
	result.Price = good.Price

	ctx.JSON(200, gin.H{
		"code":   "1",
		"msg":    "操作成功",
		"result": result,
	})
}

// 以下是用户查询已发布及买到的商品的接口
type gif struct {
	Id        uint
	Name      string
	Image     string
	AttrsText string
	RealPay   float64
}

type summary struct {
	Id        uint
	CreatTime string
	Skus      gif
	PayMoney  float64
}

// SoldList 临时收录获取订单记录的所有接口
func SoldList(c *gin.Context) {

	user, _ := c.Get("user")
	userinfo := user.(model.User)
	uId := userinfo.ID
	db := common.GetDB()
	p, _ := strconv.Atoi(c.Query("page"))
	ps, _ := strconv.Atoi(c.Query("pageSize"))
	var gList []model.Goods
	if err := db.Table("goods").Where("user = ? AND is_sold = ?", uId, 1).Find(&gList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{
				"err warning": "Record of good no found",
			})
		} else {
			c.JSON(500, gin.H{
				"error warning": "unknown error",
			})
		}
	} else {
		count := len(gList)
		idList := make([]uint, count)
		for i, g := range gList {
			idList[i] = g.ID
		}
		var oList []model.Order
		//如果is sold与order绑定成功这里应该没有错误，为了可读性就不检错了
		db.Table("orders").Where("good_id IN (?)", idList).Find(&oList)
		var result []summary
		b := (p - 1) * ps
		e := b + ps
		if (b + 1) > count {
			c.JSON(200, gin.H{
				"count":  count,
				"result": result,
			})
		} else {
			for i := b; (i < count) && (i < e); i++ {
				var r summary
				r.Id = oList[i].ID
				r.CreatTime = oList[i].CreatedAt.Format("2006-01-02 15:04:05")
				r.Skus.Id = gList[i].ID
				r.Skus.Name = gList[i].Name
				r.Skus.Image = gList[i].Picture
				r.Skus.AttrsText = gList[i].Description
				r.Skus.RealPay, _ = strconv.ParseFloat(gList[i].Price, 64)
				r.PayMoney, _ = strconv.ParseFloat(gList[i].Price, 64)
				result = append(result, r)
			}
			c.JSON(200, gin.H{
				"count":  count,
				"result": result,
			})
		}
	}

}

func BoughtList(c *gin.Context) {

	user, _ := c.Get("user")
	userinfo := user.(model.User)
	uId := userinfo.ID
	db := common.GetDB()
	p, _ := strconv.Atoi(c.Query("page"))
	ps, _ := strconv.Atoi(c.Query("pageSize"))
	var oList []model.Order
	if err := db.Table("orders").Where("user_id", uId).Find(&oList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{
				"err warning": "Record of good no found",
			})
		} else {
			c.JSON(500, gin.H{
				"error warning": "unknown error",
			})
		}
	} else {
		count := len(oList)
		idList := make([]int, count)
		for i, o := range oList {
			idList[i], _ = strconv.Atoi(o.GoodId)
		}
		var gList []model.Goods
		db.Table("goods").Where("id IN (?)", idList).Find(&gList)
		var result []summary
		b := (p - 1) * ps
		println(b)
		e := b + ps
		println(e)
		if (b + 1) > count {
			c.JSON(200, gin.H{
				"count":  count,
				"result": result,
			})
		} else {
			for i := b; (i < count) && (i < e); i++ {
				println("循环开始：", i)
				var r summary
				r.Id = oList[i].ID
				r.CreatTime = oList[i].CreatedAt.Format("2006-01-02 15:04:05")
				r.Skus.Id = gList[i].ID
				r.Skus.Name = gList[i].Name
				r.Skus.Image = gList[i].Picture
				r.Skus.AttrsText = gList[i].Description
				r.Skus.RealPay, _ = strconv.ParseFloat(gList[i].Price, 64)
				r.PayMoney, _ = strconv.ParseFloat(gList[i].Price, 64)
				result = append(result, r)
				println(" 循环结束：", i)

			}

			c.JSON(200, gin.H{
				"count":  count,
				"result": result,
			})
		}
	}

}

type summary2 struct {
	CreatTime string
	Skus      gif
}

func SaleList(c *gin.Context) {

	user, _ := c.Get("user")
	userinfo := user.(model.User)
	uId := userinfo.ID
	db := common.GetDB()
	p, _ := strconv.Atoi(c.Query("page"))
	ps, _ := strconv.Atoi(c.Query("pageSize"))
	var gList []model.Goods
	if err := db.Table("goods").Where("user = ? AND is_sold = ?", uId, 0).Find(&gList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{
				"err warning": "Record of good no found",
			})
		} else {
			c.JSON(500, gin.H{
				"error warning": "unknown error",
			})
		}
	} else {
		count := len(gList)
		var result []summary2
		b := (p - 1) * ps
		e := b + ps
		if (b + 1) > count {
			c.JSON(200, gin.H{
				"count":  count,
				"result": result,
			})
		} else {
			for i := b; (i < count) && (i < e); i++ {
				var r summary2
				r.CreatTime = gList[i].CreatedAt.Format("2006-01-02 15:04:05")
				r.Skus.Id = gList[i].ID
				r.Skus.Name = gList[i].Name
				r.Skus.Image = gList[i].Picture
				r.Skus.AttrsText = gList[i].Description
				r.Skus.RealPay, _ = strconv.ParseFloat(gList[i].Price, 64)
				result = append(result, r)
			}
			c.JSON(200, gin.H{
				"count":  count,
				"result": result,
			})
		}
	}

}
