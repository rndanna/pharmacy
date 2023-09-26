package models

import (
	"github.com/jinzhu/gorm"
)

type Feedback struct {
	gorm.Model

	UserID  uint   `json:"user_id"`
	Title   string `json:"title" binding:"required"`
	Created string `json:"created" binding:"required"`
}

type FeedbackDTO struct {
	Id      uint
	Title   string
	Login   string
	Created string `json:"created" binding:"required"`
}
