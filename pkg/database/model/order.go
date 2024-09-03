package model

import "gorm.io/gorm"

type Order struct {
	gorm.Model // ID gen update del
	GoodId     string
	AddressId  string
	UserId     uint
	PayMoney   int
}
