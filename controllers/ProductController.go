package controllers

import (
	"fmt"
	"net/http"
	"pharmacy/models"
	"pharmacy/services/database"
	"strconv"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type ProductController struct {
	Database *database.DatabaseService
}

type ProductControllerListRequest struct {
	CategoryID uint   `form:"category_id"`
	Country    string `form:"country"`
	Substance  string `form:"substance"`
	Search     string `form:"search"`
	PageID     uint   `form:"page_id" bindign:"required,min=1"`
	PageSize   uint   `form:"page_size" binding:"required,min=5"`
}

func (p *ProductController) Create(c *gin.Context) {
	var product models.Product

	if err := c.ShouldBindJSON(&product); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if createErr := p.Database.DB.Create(&product); createErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error create product": createErr.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (p *ProductController) Read(c *gin.Context) {
	var product models.Product

	productID := c.Param("id")
	if productID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no productID id provided"})
		fmt.Println(productID)
		return
	}

	_, err := strconv.Atoi(productID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "productID id must be uint"})
		return
	}

	if result := p.Database.DB.Where("id = ?", productID).Find(&product); result.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "no product"})
		return
	}

	c.JSON(http.StatusOK, &product)
}

func (p *ProductController) Delete(c *gin.Context) {
	var product models.Product

	ProductID := c.Param("id")
	if ProductID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No product id provided"})
		return
	}

	_, err := strconv.Atoi(ProductID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error - product id must be uint": err.Error()})
		return
	}

	result := p.Database.DB.Where("id = ?", ProductID).Find(&product)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error product not found": result.Error.Error()})
		return
	} else {
		if resErr := p.Database.DB.Delete(&product); resErr.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error delete product": resErr.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func (p *ProductController) Update(c *gin.Context) {
	var product models.Product

	if shouldErr := c.ShouldBindJSON(&product); shouldErr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ShouldBindJSON error": shouldErr.Error()})
		return
	}

	productID := c.Param("id")
	if productID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no product id provided"})
		return
	}

	if _, err := strconv.Atoi(productID); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "product id must be uint"})
		return
	}

	resultFind := p.Database.DB.Where("id = ?", productID).Find(&models.Product{})
	if resultFind.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"product found error": resultFind.Error.Error()})
		return
	}

	if resultUpdate := resultFind.Updates(&product); resultUpdate.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error update product": resultUpdate.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (p *ProductController) List(c *gin.Context) {
	var products []models.Product
	var request ProductControllerListRequest
	var category models.Category
	var ids []uint

	if err := c.ShouldBindQuery(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error bind query": err.Error()})
		return
	}

	if request.CategoryID == 0 {
		p.Database.DB.Model(&models.Category{}).Pluck("id", &ids)
	} else {
		if findErr := p.Database.DB.Where("id = ?", request.CategoryID).Find(&category); findErr.RowsAffected == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"category found error": findErr.Error.Error()})
			return
		}

		if category.CategoryID == 0 {
			p.Database.DB.Model(&models.Category{}).Where("category_id = ?", request.CategoryID).Pluck("id", &ids)
		} else {
			if len(ids) == 0 {
				ids = append(ids, request.CategoryID)
			}
		}
	}

	offset := (request.PageID * request.PageSize) - request.PageSize
	limit := request.PageSize

	if findErr := p.Database.DB.Debug().
		Where(&models.Product{Country: request.Country}).
		Where("category_id IN (?) AND price <> ?", ids, 0).
		Where("active_substance LIKE ? AND name LIKE ?", "%"+request.Substance+"%", "%"+request.Search+"%").
		Limit(limit).Offset(offset).
		Find(&products); findErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error find products": findErr.Error.Error()})
		return
	}

	total_records := p.Database.DB.Debug().
		Where("category_id IN (?) AND price <> ?", ids, 0).
		Where("active_substance LIKE ? AND name LIKE ?", "%"+request.Substance+"%", "%"+request.Search+"%").
		Where(&models.Product{Country: request.Country}).
		Find(&models.Product{}).RowsAffected

	c.JSON(http.StatusOK, gin.H{
		"page_id":       request.PageID,
		"page_size":     request.PageSize,
		"records":       &products,
		"total_records": total_records,
	})
}

func (p *ProductController) ListForUser(c *gin.Context) {
	var products []models.ProductDTO
	var request ProductControllerListRequest
	var category models.Category
	var ids []uint

	claims := jwt.ExtractClaims(c)
	UserID := claims[IdentityJWTKey].(float64)

	if err := c.ShouldBindQuery(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error bind query": err.Error()})
		return
	}

	if request.CategoryID == 0 {
		p.Database.DB.Model(&models.Category{}).Pluck("id", &ids)
	} else {
		if findErr := p.Database.DB.Where("id = ?", request.CategoryID).Find(&category); findErr.RowsAffected == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"category found error": findErr.Error.Error()})
			return
		}

		if category.CategoryID == 0 {
			p.Database.DB.Model(&models.Category{}).Where("category_id = ?", request.CategoryID).Pluck("id", &ids)
		} else {
			if len(ids) == 0 {
				ids = append(ids, request.CategoryID)
			}
		}
	}

	offset := (request.PageID * request.PageSize) - request.PageSize
	limit := request.PageSize

	if findErr := p.Database.DB.Debug().Model(&models.Product{}).
		Where(&models.Product{Country: request.Country}).
		Where("category_id IN (?) AND price <> ?", ids, 0).
		Where("active_substance LIKE ? AND name LIKE ?", "%"+request.Substance+"%", "%"+request.Search+"%").
		Limit(limit).Offset(offset).
		Scan(&products); findErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error find products": findErr.Error.Error()})
		return
	}

	total_records := p.Database.DB.Debug().Model(&models.Product{}).
		Where("category_id IN (?) AND price <> ?", ids, 0).
		Where("active_substance LIKE ? AND name LIKE ?", "%"+request.Substance+"%", "%"+request.Search+"%").
		Where(&models.Product{Country: request.Country}).
		Find(&models.Product{}).RowsAffected

	var basket models.Basket
	p.Database.DB.Preload("BasketContents").Where("user_id = ? and status = ?", UserID, models.ACTIVE).Find(&basket)

	for _, content := range basket.BasketContents {
		for _, product := range products {
			if product.ID == content.ProductID {
				products[product.ID-1].InBasket = true
				products[product.ID-1].Count = content.Count
			}
		}
	}

	var favorites []models.Favorites
	p.Database.DB.Where("user_id = ?", UserID).Find(&favorites)
	for _, fav := range favorites {
		for _, product := range products {
			if product.ID == fav.ProductID {
				products[fav.ProductID-1].InFavorites = true
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"page_id":       request.PageID,
		"page_size":     request.PageSize,
		"records":       &products,
		"total_records": total_records,
	})
}

func (p *ProductController) CountryList(c *gin.Context) {
	var CountryList []struct {
		Country string
	}

	p.Database.DB.Debug().Table("products").Select("distinct(country)").Order("country").Scan(&CountryList)

	c.JSON(http.StatusOK, &CountryList)
}
