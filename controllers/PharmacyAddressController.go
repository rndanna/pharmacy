package controllers

import (
	"net/http"
	"pharmacy/models"
	"pharmacy/services/database"

	"github.com/gin-gonic/gin"
)

type PharmacyAddressController struct {
	Database *database.DatabaseService
}

func (p *PharmacyAddressController) Create(c *gin.Context) {
	var address models.PharmacyAddress

	if err := c.ShouldBindJSON(&address); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if result := p.Database.DB.Create(&address); result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (p *PharmacyAddressController) List(c *gin.Context) {
	var addresses []models.PharmacyAddress

	if result := p.Database.DB.Find(&addresses); result.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": addresses})
		return
	}

	c.JSON(http.StatusOK, &addresses)
}
