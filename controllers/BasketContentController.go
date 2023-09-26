package controllers

import (
	"net/http"
	"pharmacy/models"
	"pharmacy/services/database"
	"strconv"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type BasketContentController struct {
	Database *database.DatabaseService
}

type BasketContentControllerRequestAttributes struct {
	BasketID uint `form:"basket_id" binding:"required"`
}

type BasketContentControllerCreateRequest struct {
	BasketContentControllerRequestAttributes
	ProductID uint `form:"product_id" binding:"required,min=1"`
	Count     uint `form:"count"`
}

type BasketContentControllerListRequest struct {
	BasketContentControllerRequestAttributes
	PageID   uint `form:"page_id" bindign:"required,min=1"`
	PageSize uint `form:"page_size" binding:"required,min=5"`
}

func (b *BasketContentController) Create(c *gin.Context) {
	var basketContent models.BasketContent
	var request BasketContentControllerCreateRequest

	claims := jwt.ExtractClaims(c)
	UserID := claims[IdentityJWTKey].(float64)

	if err := c.ShouldBindQuery(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ShouldBindQuery error"})
		return
	}

	if findErr := b.Database.DB.Debug().Where("user_id = ? AND status = ? AND id = ?", UserID, models.ACTIVE, request.BasketID).Find(&models.Basket{}); findErr.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error find basket": findErr.Error.Error()})
		return
	}

	basketContent.BasketID = request.BasketID
	basketContent.ProductID = request.ProductID

	if findErr := b.Database.DB.Where("id = ?", basketContent.ProductID).Find(&models.Product{}); findErr.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Product not found": findErr.Error.Error()})
		return
	}

	if findErr := b.Database.DB.Where("product_id = ? and basket_id = ?", basketContent.ProductID, basketContent.BasketID).
		Find(&basketContent); findErr.RowsAffected != 0 {
		b.Database.DB.Model(&models.BasketContent{}).
			Where("product_id = ? and basket_id = ?", basketContent.ProductID, basketContent.BasketID).
			Update("count", basketContent.Count+1)
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}
	if request.Count == 0 {
		basketContent.Count = 1
	} else {
		basketContent.Count = request.Count
	}

	if createErr := b.Database.DB.Create(&basketContent); createErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error create content": createErr.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (b *BasketContentController) Read(c *gin.Context) {
	var content models.BasketContent
	var request BasketContentControllerRequestAttributes

	claims := jwt.ExtractClaims(c)
	UserID := claims[IdentityJWTKey].(float64)

	if err := c.ShouldBindQuery(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "basket id not provided"})
		return
	}

	contentID := c.Param("id")
	if contentID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no contentID id provided"})
		return
	}

	if findErr := b.Database.DB.Where("id = ? AND user_id = ?", request.BasketID, UserID).
		Find(&models.Basket{}); findErr.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": findErr.Error.Error()})
		return
	}

	if findErr := b.Database.DB.Where("id = ?", contentID).Find(&content); findErr.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": findErr.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, &content)
}

func (b *BasketContentController) List(c *gin.Context) {
	var basketContents []models.BasketContent
	var request BasketContentControllerListRequest

	claims := jwt.ExtractClaims(c)
	UserID := claims[IdentityJWTKey].(float64)

	if err := c.ShouldBindQuery(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "basket id not provided"})
		return
	}

	if findErr := b.Database.DB.Model(&models.Basket{}).Where("user_id = ? AND id = ?", UserID, request.BasketID).Find(&models.Basket{}); findErr.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": findErr.Error.Error()})
		return
	}

	offset := (request.PageID * request.PageSize) - request.PageSize
	limit := request.PageSize
	total_records := b.Database.DB.
		Where(models.BasketContent{BasketID: request.BasketID}).
		Find(&models.BasketContent{}).RowsAffected

	if ret := b.Database.DB.
		Where(models.BasketContent{BasketID: request.BasketID}).
		Limit(limit).Offset(offset).
		Find(&basketContents); ret.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, basketContents)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"page_id":       request.PageID,
			"page_size":     request.PageSize,
			"records":       basketContents,
			"total_records": total_records,
		})
	}
}

func (b *BasketContentController) Update(c *gin.Context) {
	var count models.BasketContentDTO
	var basket_content models.BasketContent

	contentID := c.Param("id")
	if contentID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no content id provided"})
		return
	}

	claims := jwt.ExtractClaims(c)
	UserID := claims[IdentityJWTKey].(float64)

	if shouldErr := c.ShouldBindJSON(&count); shouldErr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ShouldBindJSON error": shouldErr.Error()})
		return
	}

	resultFind := b.Database.DB.Where("id = ? ", contentID).Find(&basket_content)
	if resultFind.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"content found error": resultFind.Error.Error()})
		return
	}

	if findErr := b.Database.DB.Model(&models.Basket{}).Where("user_id = ? AND id = ?", UserID, basket_content.BasketID).Find(&models.Basket{}); findErr.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": findErr.Error.Error()})
		return
	}

	basket_content.Count = count.Count
	if resultUpdate := resultFind.Select("count").Updates(&basket_content); resultUpdate.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error update content": resultUpdate.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (b *BasketContentController) Delete(c *gin.Context) {
	var basketContent models.BasketContent
	var request BasketContentControllerRequestAttributes

	if err := c.ShouldBindQuery(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "basket id not provided"})
		return
	}
	IDbasket := request.BasketID

	contentID := c.Param("id")
	if contentID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no content id provided"})
		return
	}

	_, err := strconv.Atoi(contentID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error - content id must be uint": err.Error()})
		return
	}

	result := b.Database.DB.Where("id = ? AND basket_id = ?", contentID, IDbasket).Find(&basketContent)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error content not found": result.Error.Error()})
		return
	} else {
		if resErr := b.Database.DB.Delete(&basketContent); resErr.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error delete content": resErr.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
