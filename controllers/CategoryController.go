package controllers

import (
	"net/http"
	"pharmacy/models"
	"pharmacy/services/database"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryController struct {
	Database *database.DatabaseService
}

type CategoryControllerListRequest struct {
	CategoryID uint `form:"main_cat_id"`
}

func (ct *CategoryController) Create(c *gin.Context) {
	var category models.Category

	if err := c.ShouldBindJSON(&category); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ShouldBindJSON error": err.Error()})
		return
	}

	if createErr := ct.Database.DB.Create(&category); createErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error create category": createErr.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (ct *CategoryController) Read(c *gin.Context) {
	var category models.Category

	categoryID := c.Param("id")
	if categoryID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no categoryID id provided"})
		return
	}

	_, err := strconv.Atoi(categoryID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error - category id must be uint": err.Error()})
		return
	}

	if findErr := ct.Database.DB.Where("id = ?", categoryID).Find(&category); findErr.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": findErr.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, &category)
}

func (ct *CategoryController) Delete(c *gin.Context) {
	var category models.Category

	categoryID := c.Param("id")
	if categoryID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no categoryID id provided"})
		return
	}

	_, err := strconv.Atoi(categoryID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error - category id must be uint": err.Error()})
		return
	}

	findErr := ct.Database.DB.Where("id = ?", categoryID).Find(&category)

	if findErr.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": findErr.Error.Error()})
		return
	}

	if resErr := ct.Database.DB.Delete(&category); resErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error delete category": resErr.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (ct *CategoryController) ListMainCategories(c *gin.Context) {
	var categories []models.Category

	if findErr := ct.Database.DB.Preload("Categories").Where("category_id = ?", 0).Find(&categories); findErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error find categories": findErr.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, &categories)
}

func (ct *CategoryController) ListSubCategories(c *gin.Context) {
	var categories []models.Category
	var request CategoryControllerListRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error bind query": err.Error()})
		return
	}

	if findErr := ct.Database.DB.Where("category_id <> ?", 0).Where(&models.Category{CategoryID: request.CategoryID}).Find(&categories); findErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error find categories": findErr.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, &categories)
}
