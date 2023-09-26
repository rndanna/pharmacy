package controllers

import (
	"net/http"
	"pharmacy/models"
	"pharmacy/services/database"

	"strconv"

	email "pharmacy/services/sendEmail"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type OrderController struct {
	Database *database.DatabaseService
}

type OrderControllerRequestAttributes struct {
	Status uint `form:"status" binding:"required"`
}

type OrderControllerListRequest struct {
	PageID   uint `form:"page_id" bindign:"required,min=1"`
	PageSize uint `form:"page_size" binding:"required,min=5"`
}

func (o *OrderController) Create(c *gin.Context) {
	var order models.Order

	claims := jwt.ExtractClaims(c)
	UserID := claims[IdentityJWTKey].(float64)

	order.UserID = uint(UserID)
	order.Status = models.AWAITING_PAYMENT

	if shouldErr := c.ShouldBindJSON(&order); shouldErr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ShouldBindJSON error": shouldErr.Error()})
		return
	}

	if findErr := o.Database.DB.Where("id = ? AND user_id = ?", order.BasketID, UserID).Find(&models.Basket{}); findErr.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"basket not active": findErr.Error.Error()})
		return
	}

	if createErr := o.Database.DB.Create(&order); createErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error create model": createErr.Error.Error()})
		return
	}

	var emailService = email.EmailService{
		Database: o.Database,
	}
	emailService.CreateOrderBody(order)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (o *OrderController) Read(c *gin.Context) {
	var order models.Order

	claims := jwt.ExtractClaims(c)
	userID := claims[IdentityJWTKey].(float64)

	orderID := c.Param("id")
	if orderID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no orderID id provided"})
		return
	}

	if findErr := o.Database.DB.Where("id = ? and user_id = ?", orderID, userID).Find(&order); findErr.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": findErr.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, &order)
}

func (o *OrderController) Update(c *gin.Context) {
	var request OrderControllerRequestAttributes

	orderID := c.Param("id")
	if orderID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no orderID id provided"})
		return
	}

	_, err := strconv.Atoi(orderID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error - order id must be uint": err.Error()})
		return
	}

	if err := c.ShouldBindQuery(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error bind query": err.Error()})
		return
	}

	if updateErr := o.Database.DB.Model(&models.Order{}).Where("id = ?", orderID).Update("status", request.Status); updateErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": updateErr.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (o *OrderController) Delete(c *gin.Context) {
	var order models.Order

	orderID := c.Param("id")
	if orderID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no orderID id provided"})
		return
	}

	_, err := strconv.Atoi(orderID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error - order id must be uint": err.Error()})
		return
	}

	findErr := o.Database.DB.Where("id = ?", orderID).Find(&order)

	if findErr.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": findErr.Error.Error()})
		return
	}

	if resErr := o.Database.DB.Delete(&order); resErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error delete order": resErr.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (o *OrderController) List(c *gin.Context) {
	var orders []models.OrderDTO

	claims := jwt.ExtractClaims(c)
	userID := claims[IdentityJWTKey].(float64)

	if findErr := o.Database.DB.Model(&models.Order{}).
		Where("user_id = ?", userID).
		Scan(&orders); findErr.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, orders)
		return
	}

	for i, order := range orders {
		var basket models.Basket
		var address models.PharmacyAddress
		o.Database.DB.Preload("BasketContents.Product").Where("id = ?", order.BasketID).Find(&basket)
		o.Database.DB.Where("id = ?", order.PharmacyAddressID).Find(&address)

		var products []struct {
			Product models.Product
			Count   uint
		}

		for _, basket_contents := range basket.BasketContents {
			products = append(products, struct {
				Product models.Product
				Count   uint
			}{basket_contents.Product, basket_contents.Count})
		}

		orders[i].Products = []struct {
			Product models.Product "json:\"product\""
			Count   uint           "json:\"count\""
		}(products)

		orders[i].PharmacyAddress = address
	}

	c.JSON(http.StatusOK, orders)
}

func (o *OrderController) GetAll(c *gin.Context) {
	var orders []models.OrderDTO

	if findErr := o.Database.DB.Model(&models.Order{}).
		Scan(&orders); findErr.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, orders)
		return
	}

	for i, order := range orders {
		var basket models.Basket
		var address models.PharmacyAddress
		o.Database.DB.Preload("BasketContents.Product").Where("id = ?", order.BasketID).Find(&basket)
		o.Database.DB.Where("id = ?", order.PharmacyAddressID).Find(&address)

		var products []struct {
			Product models.Product
			Count   uint
		}

		for _, basket_contents := range basket.BasketContents {
			products = append(products, struct {
				Product models.Product
				Count   uint
			}{basket_contents.Product, basket_contents.Count})
		}

		orders[i].Products = []struct {
			Product models.Product "json:\"product\""
			Count   uint           "json:\"count\""
		}(products)

		orders[i].PharmacyAddress = address
	}

	c.JSON(http.StatusOK, orders)
}
