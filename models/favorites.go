package models

import (
	"github.com/jinzhu/gorm"
)

type Favorites struct {
	gorm.Model

	UserID    uint    `json:"user_id"`
	ProductID uint    `json:"product_id" binding:"required"`
	Product   Product `json:"product"`
}

type FavoritesDTO struct {
	gorm.Model

	UserID    uint    `json:"user_id"`
	ProductID uint    `json:"product_id" binding:"required"`
	Product   Product `json:"product"`
}
