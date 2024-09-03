package database

import (
	"fmt"
	"net/url"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"smile.expression/destiny/pkg/database/model"
)

var DB *gorm.DB

type Options struct {
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Charset  string `json:"charset"`
	Loc      string `json:"loc"`
}

func NewDB(options *Options) *gorm.DB {
	//可以用navicat或datagrip等数据库操作软件，利用下面的信息登录数据库查看数据
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=%s",
		options.Username,
		options.Password,
		options.Host,
		options.Port,
		options.Database,
		options.Charset,
		url.QueryEscape(options.Loc),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Error to DB connection, err: " + err.Error())
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
