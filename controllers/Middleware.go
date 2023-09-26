package controllers

import (
	"net/http"
	"pharmacy/models"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/sirupsen/logrus"

	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type JwtWrapper struct {
	SecretKey       string
	Issuer          string
	ExpirationHours int64
}

func (r *Router) JwtMiddleware() *jwt.GinJWTMiddleware {
	m, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "diplom",
		Key:         []byte("superoleg"),
		Timeout:     time.Minute * 15,
		MaxRefresh:  time.Hour * 100,
		IdentityKey: IdentityJWTKey,
		RefreshResponse: func(c *gin.Context, code int, token string, t time.Time) {

			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusOK,
				"token":   token,
				"expire":  t.Format(time.RFC3339),
				"message": "refresh successfully",
			})
		},

		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*models.User); ok {
				return jwt.MapClaims{
					IdentityJWTKey: v.ID,
					"Role":         v.Role,
				}
			}
			return jwt.MapClaims{
				"error": true,
			}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			if v, ok := claims[IdentityJWTKey].(float64); ok {
				return &models.User{
					ID: uint(v),
				}
			}
			return &models.User{
				ID: 0,
			}
		},

		Authenticator: func(c *gin.Context) (interface{}, error) {
			var credentials = struct {
				Email    string `form:"email" json:"email" binding:"required"`
				Password string `form:"password" json:"password" binding:"required"`
			}{}

			if err := c.ShouldBind(&credentials); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			var userModel models.User
			r.Database.DB.Where(models.User{Email: credentials.Email}).First(&userModel)
			if userModel.ID == 0 {
				return "", jwt.ErrFailedAuthentication
			}
			err := bcrypt.CompareHashAndPassword([]byte(userModel.Password), []byte(credentials.Password))
			if err != nil {
				return "", jwt.ErrFailedAuthentication
			}
			return &userModel, nil
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if _, ok := data.(*models.User); ok {
				return true
			}
			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenHeadName:     "Bearer ",
		TokenLookup:       "header: Authorization, query: token, cookie: jwt",
		TimeFunc:          time.Now,
		SendAuthorization: true,
	},
	)

	if err != nil {
		logrus.Errorf("Can't wake up JWT Middleware! Error: %s\n", err.Error())
		return nil
	}

	errInit := m.MiddlewareInit()
	if errInit != nil {
		logrus.Errorf("Can't init JWT Middleware! Error: %s\n", errInit.Error())
		return nil
	}

	return m
}

func (r *Router) AdminMiddleware(c *gin.Context) {
	var currentUser models.User
	claims := jwt.ExtractClaims(c)
	UserIdRaw := claims[IdentityJWTKey]
	r.Database.DB.Where("id = ?", UserIdRaw).First(&currentUser)

	if currentUser.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		c.Abort()
		return
	}

	if currentUser.Role != models.ROLE_ADMIN {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient rights"})
		c.Abort()
		return
	}

	c.Next()
}

func (r *Router) CORSMiddleware(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
	c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Disposition, X-Suggested-Filename")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.Next()
}
