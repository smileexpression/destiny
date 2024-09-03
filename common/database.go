package common

import (
	"fmt"
	"net/url"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"smile.expression/destiny/model"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	//可以用navicat或datagrip等数据库操作软件，利用下面的信息登录数据库查看数据
	host := viper.GetString("datasource.host")
	port := viper.GetString("datasource.port")
	database := viper.GetString("datasource.database")
	username := viper.GetString("datasource.username")
	password := viper.GetString("datasource.password")
	charset := viper.GetString("datasource.charset")
	loc := viper.GetString("datasource.loc")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=%s",
		username,
		password,
		host,
		port,
		database,
		charset,
		url.QueryEscape(loc))

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Error to Db connection, err: " + err.Error())
	}
	_ = db.AutoMigrate(&model.User{}) // 此处创建了model文件夹下的user实体类，仅作参考
	_ = db.AutoMigrate(&model.Goods{})
	_ = db.AutoMigrate(&model.Category{})
	_ = db.AutoMigrate(&model.Banner{})
	_ = db.AutoMigrate(&model.Picture{})
	_ = db.AutoMigrate(&model.Chat{})
	_ = db.AutoMigrate(&model.ChatList{})
	_ = db.AutoMigrate(&model.Cart{})
	_ = db.AutoMigrate(&model.Order{})
	_ = db.AutoMigrate(&model.Image{})
	_ = db.AutoMigrate(&model.UserAddress{})

	DB = db
	return db
}

func GetDB() *gorm.DB {
	return DB
}
