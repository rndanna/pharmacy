package models

import "github.com/jinzhu/gorm"

type PharmacyAddress struct {
	gorm.Model

	Address string `json:"address" binding:"required"`
	Time    string `json:"time" binding:"required"`
}
