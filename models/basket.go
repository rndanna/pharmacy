package models

import (
	"github.com/jinzhu/gorm"
)

const (
	ACTIVE     = 1
	NOT_ACTIVE = 2
)

type Basket struct {
	gorm.Model

	UserID         uint            `json:"user_id" binding:"required"`
	Status         int             `json:"status" binding:"required"`
	BasketContents []BasketContent `json:"basket_contents"`
}

type BasketContent struct {
	gorm.Model

	BasketID  uint    `json:"basket_id" binding:"required"`
	ProductID uint    `json:"product_id" binding:"required"`
	Count     uint    `json:"count"`
	Product   Product `json:"product"`
}

type BasketContentDTO struct {
	Name  string
	Count uint
}
