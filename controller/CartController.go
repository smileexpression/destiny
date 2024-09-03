package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"smile.expression/destiny/common"
	"smile.expression/destiny/model"
)

type CartGoods struct {
	Id          string `json:"id"`
	CateId      string
	User        string `json:"user" gorm:"type:varchar(255);not null"`
	Name        string `json:"name" gorm:"type:varchar(50);not null"`
	Picture     string `json:"picture" gorm:"type:varchar(1024);not null"`
	Price       string `json:"price" gorm:"type:float;not null"`
	Description string `json:"desc" gorm:"type:varchar(255);not null"`
	IsSold      bool   `json:"is_sold"`
}

type goodIDs struct {
	GIDs []string `json:"ids"`
}

type goodID struct {
	GID string `json:"id"`
}

func CartIn(c *gin.Context) {
	user, _ := c.Get("user")
	userinfo := user.(model.User)
	uId := userinfo.ID
	fmt.Println("uId: ", uId)
	db := common.GetDB()
	var gID goodID
	if err := c.BindJSON(&gID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// 查询是否存在相同的数据
	var count int64
	db.Table("carts").Where("user_id = ? AND  good_id= ?", uId, gID).Count(&count)
	if count == 0 {
		// 不存在相同的数据，插入新数据
		var mId uint
		db.Table("carts").Select("COALESCE(MAX(id),0)").Scan(&mId)
		var re model.Cart
		re.ID = mId + 1
		re.UserId = strconv.Itoa(int(uId))
		re.GoodId = gID.GID
		tx := db.Begin()

		if err := tx.Table("carts").Create(&re).Error; err != nil {
			// 处理错误
			tx.Rollback()
		}
		tx.Commit()
		fmt.Println("check 3")
		c.JSON(200, gin.H{
			"code":   "1",
			"result": "true",
			"msg":    "操作成功",
		})
	} else {
		// 存在相同的数据，不插入新数据
		c.JSON(200, gin.H{
			"code":   "0",
			"result": false,
			"msg":    "已经加入购物车",
		})
	}

}

func CartDel(c *gin.Context) {
	user, _ := c.Get("user")
	userinfo := user.(model.User)
	uId := userinfo.ID
	db := common.GetDB()
	var gIds goodIDs
	if err := c.BindJSON(&gIds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	for _, id := range gIds.GIDs {
		var count int64
		db.Table("carts").Where("user_id = ? AND  good_id= ?", uId, id).Count(&count)
		if count == 0 {
			c.JSON(200, gin.H{
				"code":   "0",
				"result": false,
				"msg":    "商品不在购物车内",
			})
		} else {
			tx := db.Begin()
			if err := tx.Table("carts").Unscoped().Where("user_id = ? AND  good_id= ?", uId, id).Delete(&model.Cart{}).Error; err != nil {
				// 处理错误
				tx.Rollback()
			}
			tx.Commit()
			//有必要可以先查询验证
			c.JSON(200, gin.H{
				"code":   "1",
				"result": true,
				"msg":    "已删除",
			})
		}
	}
}

// CartDelOne 该函数使用query传递单个参数
func CartDelOne(c *gin.Context) {
	user, _ := c.Get("user")
	userinfo := user.(model.User)
	uId := userinfo.ID
	db := common.GetDB()
	gId := c.Query("id")

	var count int64
	db.Table("carts").Where("user_id = ? AND  good_id= ?", uId, gId).Count(&count)
	if count == 0 {
		c.JSON(200, gin.H{
			"code":   "0",
			"result": false,
			"msg":    "商品不在购物车内",
		})
	} else {
		tx := db.Begin()
		if err := tx.Table("carts").Unscoped().Where("user_id = ? AND  good_id= ?", uId, gId).Delete(&model.Cart{}).Error; err != nil {
			// 处理错误
			tx.Rollback()
		}
		tx.Commit()

		//有必要可以先查询验证
		c.JSON(200, gin.H{
			"code":   "1",
			"result": "true",
			"msg":    "已删除",
		})
	}
}

func CartOut(c *gin.Context) {
	user, _ := c.Get("user")
	userinfo := user.(model.User)
	uId := userinfo.ID
	db := common.GetDB()

	var result []CartGoods
	err := db.Table("carts").Joins("left join goods ON carts.good_id = goods.id").Where("carts.user_id = ?", uId).Scan(&result)
	if err.Error != nil {
		//错误处理
		panic(err)
	}
	if err.RowsAffected == 0 {
		//错误处理
		panic(err)
	}
	// 输出查询结果
	c.JSON(200, gin.H{
		"code":   "1",
		"msg":    "获取购物车商品成功",
		"result": result,
	})
}
