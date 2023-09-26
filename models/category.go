package models

import "github.com/jinzhu/gorm"

type Category struct {
	gorm.Model

	Name       string     `json:"name"  binding:"required"`
	CategoryID uint       `json:"categoryID"` // If it's 0 - This is Root category, otherwise it's a child category
	Categories []Category `json:"categories"  binding:"required"`
}

// type CategoryDTO struct {
// 	gorm.Model

// 	Name     string     `json:"name"  binding:"required"`

// }
