package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

const (
	AWAITING_PAYMENT   = 1
	PAID               = 2
	COLLECTED_IN_STOCK = 3
	SENT               = 4
)

type Order struct {
	gorm.Model

	BasketID          uint `json:"basket_id" binding:"required"`
	UserID            uint `json:"user_id"  binding:"required"`
	PharmacyAddressID uint `json:"pharmacy_address_id" binding:"required"`
	Status            uint `json:"status" binding:"required,max=4"`

	Name        string `json:"name" binding:"required"`
	Surname     string `json:"surname" binding:"required"`
	Patronymic  string `json:"patronymic" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
}

type OrderDTO struct {
	ID                uint      `json:"id"`
	CreatedAt         time.Time `json:"created"`
	BasketID          uint      `json:"basket_id" binding:"required"`
	Status            uint      `json:"status"`
	PharmacyAddressID uint      `json:"pharmacy_address_id" binding:"required"`
	Name              string    `json:"name"`
	Surname           string    `json:"surname" binding:"required"`
	PhoneNumber       string    `json:"phone_number" binding:"required"`

	Products []struct {
		Product Product `json:"product"`
		Count   uint    `json:"count"`
	} `json:"products"`
	PharmacyAddress PharmacyAddress `json:"pharmacy_address"`
}

// func (u *Order) Convert() (OrderDTO, error) {

// 	jsonModel, err := json.Marshal(u)
// 	if err != nil {
// 		return OrderDTO{}, err
// 	}

// 	var orderDTO OrderDTO
// 	dErr := json.Unmarshal(jsonModel, &orderDTO)
// 	if dErr != nil {
// 		return OrderDTO{}, dErr
// 	}

// 	return orderDTO, nil
// }

func (o *Order) AfterCreate(tx *gorm.DB) {
	basket := Basket{
		UserID: o.UserID,
		Status: ACTIVE,
	}
	if createErr := tx.Create(&basket); createErr.Error != nil {
		fmt.Println("error")
	}

	if updateErr := tx.Model(&Basket{}).Where("id = ?", o.BasketID).Update("status", NOT_ACTIVE); updateErr.Error != nil {
		fmt.Println("error")
	}
}
