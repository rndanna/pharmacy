package controllers

import (
	"net/http"
	"pharmacy/models"
	"pharmacy/services/database"
	"strconv"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type FeedbackController struct {
	Database *database.DatabaseService
}

type FeedbackControllerListRequest struct {
	PageID   uint `form:"page_id" bindign:"required,min=1"`
	PageSize uint `form:"page_size" binding:"required,min=3"`
}

func (f *FeedbackController) Create(c *gin.Context) {
	var feedback models.Feedback

	if err := c.ShouldBindJSON(&feedback); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims := jwt.ExtractClaims(c)
	UserID := claims[IdentityJWTKey].(float64)
	feedback.UserID = uint(UserID)

	if result := f.Database.DB.Create(&feedback); result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (f *FeedbackController) List(c *gin.Context) {
	var response []models.FeedbackDTO
	var request FeedbackControllerListRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error bind query": err.Error()})
		return
	}

	offset := (request.PageID * request.PageSize) - request.PageSize
	limit := request.PageSize

	if findErr := f.Database.DB.
		Model(&models.Feedback{}).Debug().
		Select("users.login, feedbacks.title, feedbacks.created, feedbacks.id").
		Joins("left join users on feedbacks.user_id = users.id").
		Order("feedbacks.created_at desc").
		Limit(limit).Offset(offset).
		Scan(&response); findErr.Error != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error find feedbacks": findErr.Error.Error()})
		return
	}

	total_records := f.Database.DB.Find(&models.Feedback{}).RowsAffected

	c.JSON(http.StatusOK, gin.H{
		"page_id":       request.PageID,
		"page_size":     request.PageSize,
		"records":       &response,
		"total_records": total_records,
	})
}

func (f *FeedbackController) Delete(c *gin.Context) {
	var feedback models.Feedback

	feedbackID := c.Param("id")
	if feedbackID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No feedbackID provided"})
		return
	}

	_, err := strconv.Atoi(feedbackID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error - feedback id must be uint": err.Error()})
		return
	}

	result := f.Database.DB.Where("id = ?", feedbackID).Find(&feedback)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error feedback not found": result.Error.Error()})
		return
	} else {
		if resErr := f.Database.DB.Delete(&feedback); resErr.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error delete feedback": resErr.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
