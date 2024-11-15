package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"smile.expression/destiny/pkg/database"
	"smile.expression/destiny/pkg/database/model"
	"smile.expression/destiny/pkg/http/api"
	"smile.expression/destiny/pkg/logger"
	"smile.expression/destiny/pkg/utils"
)

type UserController struct {
	r  *gin.Engine
	db *gorm.DB
}

func NewUserController(r *gin.Engine, db *gorm.DB) *UserController {
	return &UserController{
		r:  r,
		db: db,
	}
}

func (c *UserController) Register() {
	rg := c.r.Group("")

	rg.POST("/login", c.login)
	rg.POST("/register", c.register)
}

// login 登录接口函数
func (c *UserController) login(ctx *gin.Context) {
	var (
		ctx0 = ctx.Request.Context()
		log  = logger.SmileLog.WithContext(ctx0)
	)

	//获取参数
	var receiveUser model.User
	if err := ctx.BindJSON(&receiveUser); err != nil {
		log.WithError(err).Error("login bind json error")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "invalid json"})
		return
	}

	//数据验证
	if len(receiveUser.Telephone) != 11 {
		log.Errorf("invalid telephone number: %s", receiveUser.Telephone)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "invalid telephone number"})
		return
	}

	if len(receiveUser.Password) < 6 || len(receiveUser.Password) > 14 {
		log.Errorf("invalid password: %s", receiveUser.Password)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "invalid password"})
		return
	}

	//验证手机号对应的用户是否存在
	var user model.User
	if err := c.db.Where("telephone = ?", receiveUser.Telephone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.WithError(err).Errorf("mysql not found user: %s", receiveUser.Telephone)
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"msg": "user not found"})
		} else {
			log.WithError(err).Errorf("mysql query error: %s", receiveUser.Telephone)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "mysql query error"})
		}
		return
	}

	//验证密码是否正确
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(receiveUser.Password)); err != nil {
		log.WithError(err).Errorf("password error: %s", receiveUser.Password)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "password error"})
		return
	}

	//发放token
	token, err := utils.ReleaseToken(user)
	if err != nil {
		log.WithError(err).Error("release token error")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "release token error"})
		return
	}

	//获取user对应的所有地址
	var addressArray []model.UserAddress
	if err = c.db.Where("user_id = ?", user.ID).Find(&addressArray).Error; err != nil {
		log.WithError(err).Errorf("mysql query user address error: %d", user.ID)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "mysql query user address error"})
		return
	}

	userResp := api.UserResponse{
		ID:       user.ID,
		Account:  user.Telephone,
		Token:    token,
		Avatar:   user.Avatar,
		Nickname: user.Name,
		Gender:   user.Gender,
	}

	if len(addressArray) > 0 {
		//创建响应数组
		var apiAddresses []api.Address
		for _, address := range addressArray {
			resAddress := api.Address{
				AddressID: strconv.Itoa(int(address.ID)),
				Receiver:  address.Receiver,
				Contact:   address.Contact,
				Address:   address.Address,
			}
			apiAddresses = append(apiAddresses, resAddress)
		}
		userResp.UserAddress = apiAddresses
	}

	ctx.JSON(http.StatusOK, gin.H{
		"result": userResp,
	})
	return
}

// register 注册接口函数
func (c *UserController) register(ctx *gin.Context) {
	var (
		ctx0 = ctx.Request.Context()
		log  = logger.SmileLog.WithContext(ctx0)
	)

	//获取数据
	var receiveUser model.User
	if err := ctx.BindJSON(&receiveUser); err != nil {
		log.WithError(err).Error()
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(receiveUser.Name) == 0 {
		log.Errorf("invalid name: %s", receiveUser.Name)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid name"})
		return
	}

	//账号密码基本数据长度验证
	if len(receiveUser.Telephone) != 11 {
		log.Errorf("invalid telephone number: %s", receiveUser.Telephone)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid telephone number"})
		return
	}

	if len(receiveUser.Password) < 6 || len(receiveUser.Password) > 14 {
		log.Errorf("invalid password: %s", receiveUser.Password)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
		return
	}

	//验证手机号是否被注册过
	if err := c.isTelephoneExist(receiveUser.Telephone); err != nil {
		log.Errorf("telephone number exists: %s", receiveUser.Telephone)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "telephone number exists"})
		return
	}

	//对密码进行加密处理
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(receiveUser.Password), bcrypt.DefaultCost)
	if err != nil {
		log.WithError(err).Error()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newUser := model.User{
		Name:      receiveUser.Name,
		Telephone: receiveUser.Telephone,
		Password:  string(hashPassword),
		Gender:    receiveUser.Gender,
		Avatar:    receiveUser.Avatar,
	}
	if err = c.db.Create(&newUser).Error; err != nil {
		log.WithError(err).Error("create user failed")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//发放Token
	token, err := utils.ReleaseToken(newUser)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"result": &api.RegisterResponse{
		Token: token,
	}})
}

// 验证手机号是否已被注册
func (c *UserController) isTelephoneExist(telephone string) error {
	var user model.User
	if err := c.db.Where("telephone = ?", telephone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func Info(ctx *gin.Context) {
	user, _ := ctx.Get("user")
	userinfo := user.(model.User)
	id := userinfo.ID
	fmt.Println(id)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"user": user}})
}

func UpdateAvatar(ctx *gin.Context) {
	DB := database.GetDB()
	pictureID, isSuccess := ctx.GetQuery("pictureID")
	user, isExist := ctx.Get("user")
	if !isExist {
		ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "user not exist"})
		return
	}
	userInfo := user.(model.User)
	if !isSuccess {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "获取头像失败",
		})
		return
	}
	if pictureID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "头像不能为空",
		})
		return
	}
	if userInfo.Avatar != pictureID {
		userInfo.Avatar = pictureID
	}
	DB.Model(&userInfo).Where("id=?", userInfo.ID).Update("avatar", userInfo.Avatar)
	ctx.JSON(200, gin.H{
		"code":     200,
		"msg":      "更换头像成功",
		"avatarID": userInfo.Avatar,
	})
}

type onPassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func ChangePassword(ctx *gin.Context) {
	DB := database.GetDB()
	var receivePassword onPassword
	if err := ctx.BindJSON(&receivePassword); err != nil {
		ctx.JSON(422, gin.H{"code": 422, "msg": "获取失败"})
		return
	}
	user, isExist := ctx.Get("user")
	if !isExist {
		ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "user not exist"})
		return
	}
	userInfo := user.(model.User)

	//验证新旧密码长度
	if len(receivePassword.OldPassword) < 6 || len(receivePassword.OldPassword) > 14 {
		ctx.JSON(422, gin.H{"code": 422, "msg": "旧密码长度为6-14个字符"})
		return
	}

	if len(receivePassword.NewPassword) < 6 || len(receivePassword.NewPassword) > 14 {
		ctx.JSON(422, gin.H{"code": 422, "msg": "新密码长度为6-14个字符"})
		return
	}

	//验证旧密码输入是否正确
	if err := bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(receivePassword.OldPassword)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "旧密码错误",
		})
		return
	}

	//如果旧密码输入正确且新密码长度符合，对新密码进行加密
	HashPassword, err := bcrypt.GenerateFromPassword([]byte(receivePassword.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(500, gin.H{"code": 500, "msg": "加密错误"})
		return
	}

	//更换密码
	//println("old password:", userInfo.Password)
	userInfo.Password = string(HashPassword)
	//println("new password:", userInfo.Password)
	DB.Model(&userInfo).Where("id=?", userInfo.ID).Update("password", userInfo.Password)
	//返回响应
	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "更换密码成功",
	})
}

func ChangeInfo(ctx *gin.Context) {
	DB := database.GetDB()
	var receiveInfo model.User
	if err := ctx.BindJSON(&receiveInfo); err != nil {
		ctx.JSON(422, gin.H{"code": 422, "msg": "获取失败"})
		return
	}

	newName := receiveInfo.Name
	newGender := receiveInfo.Gender

	if len(newName) == 0 || len(newGender) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "个人信息不完整，请重新填写",
		})
		return
	}
	user, isExist := ctx.Get("user")
	if !isExist {
		ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "user not exist"})
		return
	}
	userInfo := user.(model.User)

	if userInfo.Name != newName || userInfo.Gender != newGender {
		userInfo.Name = newName
		userInfo.Gender = newGender
		DB.Model(&userInfo).Where("id=?", userInfo.ID).Updates(map[string]interface{}{"name": userInfo.Name, "gender": userInfo.Gender})
	}
	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "操作成功",
		"result": gin.H{
			"nickname": userInfo.Name,
			"gender":   userInfo.Gender,
		},
	})
}

func AddAddress(ctx *gin.Context) {
	DB := database.GetDB()
	var receiveAddress model.UserAddress
	if err := ctx.BindJSON(&receiveAddress); err != nil {
		ctx.JSON(422, gin.H{"code": 422, "msg": "获取失败"})
		return
	}

	if len(receiveAddress.Receiver) == 0 || len(receiveAddress.Contact) == 0 || len(receiveAddress.Address) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "地址信息不完整，请重新填写",
		})
		return
	}

	user, isExist := ctx.Get("user")
	if !isExist {
		ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "user not exist"})
		return
	}
	userInfo := user.(model.User)
	receiveAddress.UserID = userInfo.ID

	DB.Create(&receiveAddress)

	var addressArray []model.UserAddress
	DB.Model(&model.UserAddress{}).Where("user_id=?", userInfo.ID).Find(&addressArray)

	//for _, v := range addressArray {
	//	println("id", v.ID)
	//	println("Receiver", v.Receiver)
	//	println("Contact", v.Contact)
	//	println("Address", v.Address)
	//	println("UserID", v.UserID)
	//
	//}

	var apiAddresses []api.Address
	for _, address := range addressArray {
		resAddress := api.Address{AddressID: strconv.Itoa(int(address.ID)), Receiver: address.Receiver, Contact: address.Contact, Address: address.Address}
		apiAddresses = append(apiAddresses, resAddress)
	}
	ctx.JSON(200, gin.H{
		"code":          200,
		"msg":           "新增收货地址成功",
		"userAddresses": apiAddresses,
	},
	)
}

func DeleteAddress(ctx *gin.Context) {
	DB := database.GetDB()
	addressID, isExist := ctx.GetQuery("id")

	if !isExist {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "没有传入参数id",
		})
		return
	}
	if len(addressID) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "id为空",
		})
		return
	}

	user, isExist := ctx.Get("user")
	if !isExist {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 400, "msg": "user not exist"})
		return
	}
	userInfo := user.(model.User)

	if err := DB.Where("id = ?", addressID).First(&model.UserAddress{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		//记录为空
		ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "该记录不存在"})
		return
	}

	DB.Where("id=?", addressID).Delete(&model.UserAddress{})

	var addressArray []model.UserAddress
	DB.Model(&model.UserAddress{}).Where("user_id=?", userInfo.ID).Find(&addressArray)

	//for _, v := range addressArray {
	//	println("id", v.ID)
	//	println("Receiver", v.Receiver)
	//	println("Contact", v.Contact)
	//	println("Address", v.Address)
	//	println("UserID", v.UserID)
	//
	//}

	var apiAddresses []api.Address
	for _, address := range addressArray {
		resAddress := api.Address{AddressID: strconv.Itoa(int(address.ID)), Receiver: address.Receiver, Contact: address.Contact, Address: address.Address}
		apiAddresses = append(apiAddresses, resAddress)
	}
	ctx.JSON(200, gin.H{
		"code":          200,
		"msg":           "删除收货地址成功",
		"userAddresses": apiAddresses,
	},
	)
}
