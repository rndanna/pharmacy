package controllers

import (
	"net/http"
	"pharmacy/models"
	"pharmacy/services/database"
	"strconv"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type FavoritesController struct {
	Database *database.DatabaseService
}

type FavoritesControllerCreateRequest struct {
	ProductID uint `form:"product_id" binding:"required"`
}

func (f *FavoritesController) Create(c *gin.Context) {
	var request FavoritesControllerCreateRequest
	var fav models.Favorites

	if err := c.ShouldBindQuery(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims := jwt.ExtractClaims(c)
	UserID := claims[IdentityJWTKey].(float64)

	fav.UserID = uint(UserID)
	fav.ProductID = request.ProductID

	if findErr := f.Database.DB.Where("product_id = ? AND user_id = ?", fav.ProductID, UserID).Find(&models.Favorites{}); findErr.RowsAffected != 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Product added to favorites"})
		return
	}

	if createErr := f.Database.DB.Create(&fav); createErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": createErr.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (f *FavoritesController) List(c *gin.Context) {
	var fav []models.Favorites

	claims := jwt.ExtractClaims(c)
	UserID := claims[IdentityJWTKey].(float64)

	f.Database.DB.Debug().Preload("Product").Where("user_id = ?", UserID).Find(&fav)
		
	c.JSON(http.StatusOK, &fav)
}

func (f *FavoritesController) Delete(c *gin.Context) {
	var fav models.Favorites

	productID := c.Param("id")
	if productID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No productID provided"})
		return
	}

	_, err := strconv.Atoi(productID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error - productID id must be uint": err.Error()})
		return
	}

	claims := jwt.ExtractClaims(c)
	UserID := claims[IdentityJWTKey].(float64)

	if findErr := f.Database.DB.Where("user_id = ? and product_id = ?", UserID, productID).Find(&fav); findErr.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "product not found"})
		return
	}

	if resErr := f.Database.DB.Delete(&fav); resErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error delete feedback": resErr.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
