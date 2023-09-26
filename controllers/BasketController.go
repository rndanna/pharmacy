package controllers

import (
	"net/http"
	"pharmacy/models"
	"pharmacy/services/database"
	"strconv"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type BasketController struct {
	Database *database.DatabaseService
}

type BasketControllerRequestAttributes struct {
	Status uint `form:"status" binding:"required"`
}

func (b *BasketController) Create(c *gin.Context) {
	var basket models.Basket
	basket.Status = models.ACTIVE

	claims := jwt.ExtractClaims(c)
	UserID := claims[IdentityJWTKey].(float64)
	basket.UserID = uint(UserID)

	if Err := b.Database.DB.Where("user_id = ? AND status = ?", UserID, models.ACTIVE).
		Find(&basket); Err.Error != nil {
		if createErr := b.Database.DB.Create(&basket); createErr.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"create basket error": createErr.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error create": "basket already created"})
	return
}

func (b *BasketController) Read(c *gin.Context) {
	var contents []models.BasketContent
	var basket models.Basket

	claims := jwt.ExtractClaims(c)
	UserID := claims[IdentityJWTKey].(float64)

	BasketID := c.Param("id")
	if BasketID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no BasketID id provided"})
		return
	}

	if findErr := b.Database.DB.Debug().Where("user_id = ? AND id = ?", UserID, BasketID).Find(&basket); findErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error find basket": findErr.Error.Error()})
		return
	}

	if findErr := b.Database.DB.Debug().Model(&models.BasketContent{}).Where("basket_id = ?", basket.ID).Find(&contents); findErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error find basket": findErr.Error.Error()})
		return
	}
	basket.BasketContents = contents

	c.JSON(http.StatusOK, &basket)
}

func (b *BasketController) Delete(c *gin.Context) {
	var basket models.Basket

	basketID := c.Param("id")
	if basketID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No basket id provided"})
		return
	}

	claims := jwt.ExtractClaims(c)
	UserID := claims[IdentityJWTKey].(float64)

	_, err := strconv.Atoi(basketID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error - basket id must be uint": err.Error()})
		return
	}

	result := b.Database.DB.Where("id = ? AND user_id = ?", basketID, UserID).Find(&basket)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error basket not found": result.Error.Error()})
		return
	} else {
		if resErr := b.Database.DB.Delete(&basket); resErr.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error delete basket": resErr.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func (b *BasketController) ReadActive(c *gin.Context) {
	var basket models.Basket

	claims := jwt.ExtractClaims(c)
	UserID := claims[IdentityJWTKey].(float64)
	basket.UserID = uint(UserID)
	basket.Status = models.ACTIVE

	if findErr := b.Database.DB.Preload("BasketContents.Product").
		Where("user_id = ? AND status = 1", UserID).Find(&basket); findErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error find basket": findErr.Error.Error()})
		return
	}

	var price float64
	for _, basket_content := range basket.BasketContents {
		price += float64(basket_content.Count) * basket_content.Product.Price
	}
	c.JSON(http.StatusOK, gin.H{
		"basket":      &basket,
		"total_price": price,
	})
}
