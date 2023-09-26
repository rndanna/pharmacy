package models

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
)

type Product struct {
	gorm.Model

	Name            string  `json:"name"  binding:"required"`
	Price           float64 `json:"price" binding:"required"`
	ActiveSubstance string  `json:"active_substance" binding:"required"` //активное вещество
	Manufacturer    string  `json:"manufacturer" binding:"required"`     //производитель
	Country         string  `json:"country" binding:"required"`          //страна производеля
	ReleaseForm     string  `json:"release_form" binding:"required"`     //форма выпуска
	Conditions      string  `json:"conditions" binding:"required"`       //условия выпуска из аптеки

	Description          string `json:"description" gorm:"size:max"`           // описание
	Indications          string `json:"indications" gorm:"size:max"`           //показания
	Contraindications    string `json:"contraindications" gorm:"size:max"`     //противопоказания
	PharmachologicEffect string `json:"pharmachologic_effect" gorm:"size:max"` //фармакологический эффект
	Dosage               string `json:"dosage" gorm:"size:max"`                //способ применения и дозы

	CategoryID uint   `json:"category_id" binding:"required"`
	ImageURL   string `json:"img_href"`
}

type ProductDTO struct {
	gorm.Model

	Name            string  `json:"name"  binding:"required"`
	Price           float64 `json:"price" binding:"required"`
	ActiveSubstance string  `json:"active_substance" binding:"required"` //активное вещество
	Manufacturer    string  `json:"manufacturer" binding:"required"`     //производитель
	Country         string  `json:"country" binding:"required"`          //страна производеля
	ReleaseForm     string  `json:"release_form" binding:"required"`     //форма выпуска
	Conditions      string  `json:"conditions" binding:"required"`       //условия выпуска из аптеки

	Description          string `json:"description" gorm:"size:max"`           // описание
	Indications          string `json:"indications" gorm:"size:max"`           //показания
	Contraindications    string `json:"contraindications" gorm:"size:max"`     //противопоказания
	PharmachologicEffect string `json:"pharmachologic_effect" gorm:"size:max"` //фармакологический эффект
	Dosage               string `json:"dosage" gorm:"size:max"`                //способ применения и дозы

	CategoryID uint   `json:"category_id" binding:"required"`
	ImageURL   string `json:"img_href"`

	InFavorites bool `json:"in_favorites" default:"0"`
	InBasket    bool `json:"in_basket" default:"false"`
	Count       uint `json:"count" default:"0"`
}

func (u *Product) Convert() (ProductDTO, error) {

	jsonModel, err := json.Marshal(u)
	if err != nil {
		return ProductDTO{}, err
	}

	var productDTO ProductDTO
	dErr := json.Unmarshal(jsonModel, &productDTO)
	if dErr != nil {
		return ProductDTO{}, dErr
	}

	return productDTO, nil
}
