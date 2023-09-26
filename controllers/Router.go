package controllers

import (
	"log"
	"pharmacy/services/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Router struct {
	Router   *gin.Engine
	Database *database.DatabaseService

	UserController            UserController
	FeedbackController        FeedbackController
	ProductController         ProductController
	BasketController          BasketController
	BasketContentController   BasketContentController
	OrderController           OrderController
	PharmacyAddressController PharmacyAddressController
	CategoryController        CategoryController
	FavoritesController       FavoritesController
}

const IdentityJWTKey = "id"

func (r *Router) InitENV() {

	if errs := godotenv.Load(); errs != nil {
		if errs != nil {
			log.Fatalln("error: ", errs)
		}
	}
}

func (r *Router) Prepare() {
	r.Router = gin.New()
	r.Router.Use(gin.Logger())

	r.Database = &database.DatabaseService{}
	if !r.Database.Init() {
		log.Fatalln("Can't connect to DB")
	}
}

func (r *Router) ImplementControllers() {
	r.UserController = UserController{Database: r.Database}
	r.FeedbackController = FeedbackController{Database: r.Database}
	r.ProductController = ProductController{Database: r.Database}
	r.BasketController = BasketController{Database: r.Database}
	r.BasketContentController = BasketContentController{Database: r.Database}
	r.OrderController = OrderController{Database: r.Database}
	r.PharmacyAddressController = PharmacyAddressController{Database: r.Database}
	r.CategoryController = CategoryController{Database: r.Database}
	r.FavoritesController = FavoritesController{Database: r.Database}
}

func (r *Router) ImplementRouters() {
	r.Router.Use(r.CORSMiddleware)
	jwtMiddleware := r.JwtMiddleware()

	api := r.Router.Group("/api")
	{
		//*===================================User===================================*//
		user := api.Group("/user")
		{
			user.POST("/signup", r.UserController.SignUp)
			user.POST("/login", jwtMiddleware.LoginHandler)
			user.Use(jwtMiddleware.MiddlewareFunc())
			{
				user.GET("/refresh/token", jwtMiddleware.RefreshHandler)
				user.GET("/read", r.UserController.Read)
				user.PUT("/update", r.UserController.Update)

				AdminGroup := user.Group("/admin")
				AdminGroup.Use(r.AdminMiddleware)
				{
					AdminGroup.DELETE("/delete/:id", r.UserController.Delete)
				}
			}
		}

		//*===================================Feedback==================================*//

		feedback := api.Group("/feedback")
		{
			feedback.GET("/list", r.FeedbackController.List)
			feedback.Use(jwtMiddleware.MiddlewareFunc())
			{
				feedback.POST("/create", r.FeedbackController.Create)

				AdminGroup := feedback.Group("/admin")
				AdminGroup.Use(r.AdminMiddleware)
				{
					AdminGroup.DELETE("/delete/:id", r.FeedbackController.Delete)
				}
			}
		}

		//*===================================Category==================================*//

		category := api.Group("/category")
		{
			category.GET("/read/:id", r.CategoryController.Read)
			category.GET("/mainlist", r.CategoryController.ListMainCategories)
			category.GET("/sublist", r.CategoryController.ListSubCategories)

			category.Use(jwtMiddleware.MiddlewareFunc())
			{

			}
		}

		//*===================================Product==================================*//

		product := api.Group("/product")
		{
			product.POST("/create", r.ProductController.Create)
			product.GET("/read/:id", r.ProductController.Read)
			product.GET("/list", r.ProductController.List)
			product.GET("/country", r.ProductController.CountryList)

			product.Use(jwtMiddleware.MiddlewareFunc())
			{
				product.GET("/listuser", r.ProductController.ListForUser)
			}
		}

		//*===================================Favorites==================================*//

		fav := api.Group("/favorites")
		{
			fav.Use(jwtMiddleware.MiddlewareFunc())
			{
				fav.GET("/create", r.FavoritesController.Create)
				fav.GET("/list", r.FavoritesController.List)
				fav.DELETE("/delete/:id", r.FavoritesController.Delete)
			}
		}

		//*===================================Basket===================================*//

		basket := api.Group("/basket")
		{
			basket.Use(jwtMiddleware.MiddlewareFunc())
			{
				basket.GET("/create", r.BasketController.Create)
				basket.GET("/read/:id", r.BasketController.Read)
				basket.GET("/active", r.BasketController.ReadActive)

				AdminGroup := basket.Group("/admin")
				AdminGroup.Use(jwtMiddleware.MiddlewareFunc())
				AdminGroup.Use(r.AdminMiddleware)
				{
					AdminGroup.DELETE("/delete/:id", r.BasketController.Delete)
				}
			}
		}

		//*===================================Content===================================*//

		content := api.Group("/content")
		{
			content.Use(jwtMiddleware.MiddlewareFunc())
			{
				content.GET("/create", r.BasketContentController.Create)
				content.GET("/read/:id", r.BasketContentController.Read)
				content.GET("/list", r.BasketContentController.List)
				content.PUT("/update/:id", r.BasketContentController.Update)
				content.DELETE("/delete/:id", r.BasketContentController.Delete)
			}
		}

		//*===================================Order===================================*//

		order := api.Group("/order")
		{
			order.Use(jwtMiddleware.MiddlewareFunc())
			{
				order.POST("/create", r.OrderController.Create)
				order.GET("/read/:id", r.OrderController.Read)
				order.GET("/list", r.OrderController.List)

				AdminGroup := order.Group("/admin")
				AdminGroup.Use(jwtMiddleware.MiddlewareFunc())
				AdminGroup.Use(r.AdminMiddleware)
				{
					AdminGroup.PUT("/update/:id", r.OrderController.Update)
					AdminGroup.DELETE("/delete/:id", r.OrderController.Delete)
					AdminGroup.GET("/all", r.OrderController.GetAll)
				}
			}
		}

		//*===================================Addresses===================================*//

		addresses := api.Group("/address")
		{
			addresses.GET("/list", r.PharmacyAddressController.List)
			addresses.Use(jwtMiddleware.MiddlewareFunc())
			{
				AdminGroup := addresses.Group("/admin")
				AdminGroup.Use(jwtMiddleware.MiddlewareFunc())
				AdminGroup.Use(r.AdminMiddleware)
				{
					AdminGroup.POST("/create", r.PharmacyAddressController.Create)
				}
			}
		}

	}
}

func (r *Router) StartApp() {
	r.InitENV()
	r.Prepare()
	r.ImplementControllers()
	r.ImplementRouters()

	err := r.Router.Run(":8082")
	if err != nil {
		log.Fatalln("Can't start app on ", err)
	}
}
