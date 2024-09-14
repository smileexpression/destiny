package api

import "smile.expression/destiny/pkg/database/model"

type PutObjectResponse struct {
	URL  string `json:"url"`
	ETag string `json:"etag"`
	Size int64  `json:"size"`
}

type Address struct {
	AddressID string `json:"id"`
	Receiver  string `json:"receiver"`
	Contact   string `json:"contact"`
	Address   string `json:"address"`
}

type RegisterResponse struct {
	Token string `json:"token"`
}

type Goods struct { // "_2" 区分于commodity controller的AllIdle
	Id      string
	Name    string `gorm:"type:varchar(20);not null"`
	Picture string `gorm:"type:varchar(1024);not null"`
	Goods   []model.Goods
}

type UserResponse struct {
	ID          uint      `json:"id"`
	Account     string    `json:"account"`
	Token       string    `json:"token"`
	Avatar      string    `json:"avatar"`
	Nickname    string    `json:"nickname"`
	Gender      string    `json:"gender"`
	UserAddress []Address `json:"userAddresses"`
}
