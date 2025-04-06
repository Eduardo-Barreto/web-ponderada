package routes

import (
	"net/http" // Added missing import
	"github.com/gin-contrib/cors" // Import CORS middleware
	"github.com/gin-gonic/gin"
	"github.com/Eduardo-Barreto/web-ponderada/backend/handlers"
	"github.com/Eduardo-Barreto/web-ponderada/backend/middleware"
	"github.com/Eduardo-Barreto/web-ponderada/backend/repository"
)

// SetupRouter configures the Gin router with all routes
func SetupRouter(
	userRepo repository.UserRepository,
	productRepo repository.ProductRepository,
	fileRepo repository.StorageRepository,
) *gin.Engine {

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(userRepo)
	userHandler := handlers.NewUserHandler(userRepo, fileRepo) // Pass fileRepo
	productHandler := handlers.NewProductHandler(productRepo, fileRepo)
	imageHandler := handlers.NewImageHandler(fileRepo)

	// Gin Router
	// router := gin.Default() // Includes logger and recovery middleware
	router := gin.New() // Use New for more control
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS Middleware Configuration
	// WARNING: For development, allow all origins. Restrict in production!
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // Replace with your frontend URL in production
	// config.AllowOrigins = []string{"http://localhost:3000", "https://your-frontend-domain.com"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	// config.AllowCredentials = true // If you need cookies or sessions
	router.Use(cors.New(config))

	// Public Routes (Authentication)
	authRoutes := router.Group("/api/v1/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
	}

	// Public route to serve images
    // Note: Captures everything after /images/ as 'filepath' parameter
	router.GET("/api/v1/images/*filepath", imageHandler.ServeImage)

	// API v1 Group
	apiV1 := router.Group("/api/v1")

	// --- User Routes (Protected) ---
	userRoutes := apiV1.Group("/users")
	userRoutes.Use(middleware.AuthMiddleware()) // Apply auth middleware to this group
	{
		userRoutes.GET("", userHandler.GetUsers)              // GET /api/v1/users
		userRoutes.GET("/:id", userHandler.GetUser)           // GET /api/v1/users/:id
		userRoutes.PUT("/:id", userHandler.UpdateUser)        // PUT /api/v1/users/:id (for name/email)
		userRoutes.POST("/:id/profile-pic", userHandler.UploadProfilePic) // POST /api/v1/users/:id/profile-pic
		userRoutes.DELETE("/:id", userHandler.DeleteUser)     // DELETE /api/v1/users/:id
	}

	// --- Product Routes ---
	productRoutes := apiV1.Group("/products")
	{
		// Publicly viewable products 
		productRoutes.GET("", productHandler.GetProducts)      // GET /api/v1/products
		productRoutes.GET("/:id", productHandler.GetProduct)   // GET /api/v1/products/:id

		// Protected actions (Create, Update, Delete)
		protectedProductRoutes := productRoutes.Group("")
		protectedProductRoutes.Use(middleware.AuthMiddleware())
		{
			protectedProductRoutes.POST("", productHandler.CreateProduct) // POST /api/v1/products
			protectedProductRoutes.PUT("/:id", productHandler.UpdateProduct) // PUT /api/v1/products/:id (Note: PUT/POST for multipart forms)
			protectedProductRoutes.DELETE("/:id", productHandler.DeleteProduct) // DELETE /api/v1/products/:id
		}
	}

    // Health Check Route
    router.GET("/health", func(c *gin.Context) {
        // TODO: Add DB ping check here for more thorough health check
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})


	return router
}
