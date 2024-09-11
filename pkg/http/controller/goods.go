package controller

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"smile.expression/destiny/logger"
	"smile.expression/destiny/pkg/cache"
	"smile.expression/destiny/pkg/database"
	"smile.expression/destiny/pkg/database/model"
	"smile.expression/destiny/pkg/http/api"
	"smile.expression/destiny/pkg/storage"
)

type SingleIdle struct {
	Id      string
	Name    string `gorm:"type:varchar(20);not null"`
	Picture string `gorm:"type:varchar(1024);not null"`
	Goods   []model.Goods
}

type GoodsController struct {
	options        *GoodsControllerOptions
	r              *gin.Engine
	db             *gorm.DB
	cacheClient    *cache.Client
	storageClient  *storage.Client
	authController *AuthController
}

type GoodsControllerOptions struct {
	RecentLimit     int `json:"recentLimit"`
	HomeGoodsLimit  int `json:"homeGoodsLimit"`
	CacheExpiration int `json:"cacheExpiration"`
}

func NewGoodsController(options *GoodsControllerOptions, r *gin.Engine, db *gorm.DB, cacheClient *cache.Client, storageClient *storage.Client, authController *AuthController) *GoodsController {
	return &GoodsController{
		options:        options,
		r:              r,
		db:             db,
		cacheClient:    cacheClient,
		storageClient:  storageClient,
		authController: authController,
	}
}

func (c *GoodsController) Register() {
	rg := c.r.Group("/home")

	rg.GET("new", c.new)
	rg.GET("goods", c.goods)

	rg2 := c.r.Group("/member")

	rg2.POST("/release", c.authController.AuthMiddleware(), c.release)
}

func (c *GoodsController) release(ctx *gin.Context) {
	var (
		ctx0 = ctx.Request.Context()
		log  = logger.SmileLog.WithContext(ctx0)
	)

	user, exists := ctx.Get("user")
	if !exists {
		log.Error("need to login")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "need to login"})
		return
	}
	userInfo := user.(*model.User)

	//绑定body
	var goodInfo GoodInfo
	if err := ctx.BindJSON(&goodInfo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	//生成good
	good := model.Goods{
		CateId:      goodInfo.CateId,
		User:        strconv.Itoa(int(userInfo.ID)), //之前将good表的User字段定义成了string
		Name:        goodInfo.Name,
		Picture:     goodInfo.Picture[0],
		Price:       goodInfo.Price,
		Description: goodInfo.Description,
		IsSold:      false,
	}
	if err := c.db.Create(&good).Error; err != nil {
		log.WithError(err).Error("mysql create goods failed")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "mysql create goods failed"})
		return
	}

	//生成good对应的picture
	goodId := strconv.Itoa(int(good.ID)) //之前将picture表goodId字段定义成了string
	picture := model.Picture{
		GoodId: goodId,
	}

	for i := 0; i < len(goodInfo.Picture); i++ {
		field := fmt.Sprintf("Picture%d", i+1)
		reflect.ValueOf(&picture).Elem().FieldByName(field).SetString(goodInfo.Picture[i])
	}

	if err := c.db.Create(&picture).Error; err != nil {
		log.WithError(err).Error("mysql create picture failed")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "mysql create picture failed"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"result": "ok",
	})
	return
}

func (c *GoodsController) goods(ctx *gin.Context) {
	var (
		ctx0 = ctx.Request.Context()
		log  = logger.SmileLog.WithContext(ctx0)
	)

	var cate []model.Category
	if err := c.db.Find(&cate).Error; err != nil {
		log.WithError(err).Error("mysql query category error")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	result := make([]api.Goods, len(cate))
	key := fmt.Sprintf("home/goods_%d", len(cate))

	data, err := c.cacheClient.Get(ctx0, key)
	if err == nil {
		if err = json.Unmarshal(data, &result); err != nil {
			log.WithError(err).Error("failed to unmarshal home goods data")
		} else {
			ctx.JSON(http.StatusOK, gin.H{"result": result})
			return
		}
	}

	var cacheData []byte
	defer func() {
		cacheData, err = json.Marshal(result)
		if err = c.cacheClient.Set(ctx0, key, cacheData, c.options.CacheExpiration); err != nil {
			log.WithError(err).Error("redis set home goods error")
		}
	}()

	for i, g := range cate {
		if err = c.db.Where("cate_id = ? AND is_sold = ?", g.Id, false).Order("created_at DESC").Limit(c.options.HomeGoodsLimit).Find(&result[i].Goods).Error; err != nil {
			log.WithError(err).Error("mysql query goods error")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		result[i].Id = g.Id
		result[i].Name = g.Name
		result[i].Picture = c.storageClient.SetEndpoint(g.Picture)
	}

	ctx.JSON(200, gin.H{
		"result": result,
	})
	return
}

func (c *GoodsController) new(ctx *gin.Context) {
	var (
		ctx0 = ctx.Request.Context()
		log  = logger.SmileLog.WithContext(ctx0)
	)

	limit := c.options.RecentLimit

	var recentGoods []model.Goods
	if err := c.db.Where("is_sold = ?", false).Order("created_at DESC").Limit(limit).Find(&recentGoods).Error; err != nil {
		log.WithError(err).Error("fail to get new goods")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"result": recentGoods})
	return
}

//暂且不考虑id转换错误

func GetOneGood(c *gin.Context) {
	db := database.GetDB()
	idStr := c.Query("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	var target model.Goods
	if err := db.Table("goods").Where("id = ?", id).First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{
				"err warning": "Record of good no found",
			})
		}
	} else {
		var picTarget model.Picture
		if err = db.Table("pictures").Where("good_id = ?", target.ID).First(&picTarget).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(404, gin.H{
					"err warning": "There is a error in pictures database,please contact the administrator",
				})
			}
		} else {
			//用户头像一般不会出错 简化代码不处理
			var user model.User
			db.Table("users").First(&user, target.User)
			p := [5]string{picTarget.Picture1, picTarget.Picture2, picTarget.Picture3, picTarget.Picture4, picTarget.Picture5}
			c.JSON(200, gin.H{
				"result":   target,
				"pictures": p,
				"user":     user,
			})
		}
	}

}

func ChooseCategory(ctx *gin.Context) {

	DB := database.GetDB()
	var result SingleIdle

	CateId := ctx.DefaultQuery("id", "3")

	var category model.Category
	DB.Table("categories").Where("id = ?", CateId).Find(&category)
	result.Id = category.Id
	result.Name = category.Name
	result.Picture = category.Picture

	var goods []model.Goods
	DB.Table("goods").Where("cate_Id = ? AND is_sold=?", CateId, false).Find(&goods)
	result.Goods = append(result.Goods, goods...)

	ctx.JSON(200, gin.H{
		"code":   "1",
		"msg":    "获取分类下属物品成功",
		"result": result,
	})

}

type GoodInfo struct { //用于接收body参数
	Name        string   `json:"name"`
	CateId      string   `json:"cate_id"`
	Description string   `json:"description"`
	Picture     []string `json:"picture"`
	Price       string   `json:"price"`
}

type apiGood struct {
	Id      uint   `json:"ID"`
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Price   string `json:"price"`
	Picture string `json:"picture"`
}

func transApiGood(good model.Goods) apiGood {
	return apiGood{Id: good.ID, Name: good.Name, Desc: good.Name, Price: good.Price, Picture: good.Picture}
}

func RecommendGoods(ctx *gin.Context) {
	DB := database.GetDB()
	strLimit := ctx.DefaultQuery("limit", "4")
	intLimit, err := strconv.Atoi(strLimit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get limit"})
		return
	}

	//未售出商品数量
	var isNotSoldNum int64
	DB.Table("goods").Where("is_sold=?", false).Count(&isNotSoldNum)
	//获得未售出的商品数组，数量为isNotSold_Num
	var notSoldGoodArray []model.Goods
	DB.Table("goods").Where("is_sold=?", false).Limit(int(isNotSoldNum)).Find(&notSoldGoodArray)
	println("isNotSold_Num", isNotSoldNum)

	//打印未售出的商品数组
	//println("notSoldGood_Array:")
	//for i, v := range notSoldGood_Array {
	//	println("k", i, ",v", v.ID)
	//}

	//通过比较获得最终展出商品的数量
	times := minNum(intLimit, int(isNotSoldNum))
	result := make([]apiGood, times)
	//记录已随机抽取商品的id
	var idRecord = make([]uint, times)
	//println("idRecord:")
	//for _, v := range idRecord {
	//	println(v)
	//}
	for i := 0; i < times; i++ {
		for {
			//随机抽取一个index,范围[0,total)
			ranIndex, _ := rand.Int(rand.Reader, big.NewInt(isNotSoldNum))
			//println("ranIndex=", ranIndex.Int64())
			index := uint(ranIndex.Int64())
			id := notSoldGoodArray[index].ID
			if checkRanID(idRecord, id) {
				idRecord[i] = id
				notSoldGoodArray[index].IsSold = true
				result[i] = transApiGood(notSoldGoodArray[index])
				break
			}
		}
	}

	//for i := range idRecord {
	//	println(idRecord[i])
	//}
	ctx.JSON(200, gin.H{
		"code":   200,
		"msg":    "操作成功",
		"result": result,
	})
}

func checkRanID(idRecord []uint, checkID uint) bool {
	//检查该id是否已经被抽取,如果已经被抽取，返回false
	for i := 0; i < len(idRecord); i++ {
		if idRecord[i] == checkID {
			return false
		}
	}
	return true
}

func minNum(limit int, total int) int {
	if limit <= total {
		return limit
	} else {
		return total
	}
}
