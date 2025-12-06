package route

import (
	"database/sql"
	"log"
	"github.com/google/uuid"
	httpHandler "home-market/internal/delivery/http/handler"
	repo "home-market/internal/repository/postgresql"
	mongorepo "home-market/internal/repository/mongodb"
	service "home-market/internal/service/postgresql"
	"github.com/gin-gonic/gin"
	"home-market/internal/delivery/http/middleware"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoute(app *gin.Engine, db *sql.DB,mongoclient *mongo.Client) {
    // --- 1. Ambil default role untuk user baru (misal: "buyer") ---
	var defaultRoleID uuid.UUID
	err := db.QueryRow(`SELECT id FROM roles WHERE name = $1`, "buyer").Scan(&defaultRoleID)
	if err != nil {
		log.Printf("warning: gagal mengambil default role 'buyer': %v", err)
	}

	// --- 2. Init repository, service, handler ---
	userRepo := repo.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, defaultRoleID)
	authHandler := httpHandler.NewAuthHandler(authService)
	shopRepo := repo.NewShopRepository(db)
	shopService := service.NewShopService(shopRepo)
	shopHandler := httpHandler.NewShopHandler(shopService)
	categoryRepo := repo.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryHandler := httpHandler.NewCategoryHandler(categoryService)
	itemRepo := repo.NewItemRepository(db)
	logRepo := mongorepo.NewLogRepository(mongoclient)
	itemService := service.NewItemService(itemRepo,logRepo)
	itemHandler := httpHandler.NewItemHandler(itemService)
	adminService := service.NewAdminService(userRepo, itemRepo)
    adminHandler := httpHandler.NewAdminHandler(adminService)

	// --- 3. Definisikan group route ---
	api := app.Group("/api")

	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)
	auth.GET("/profile", middleware.AuthRequired(), authHandler.Profile)

	
	shop := api.Group("/shops")
	shop.POST("/", middleware.AuthRequired(), shopHandler.CreateShop)
	cat := api.Group("/categories")
	cat.POST("/", middleware.AuthRequired(), categoryHandler.CreateCategory)

	items := api.Group("/items", middleware.AuthRequired())
	items.POST("/", itemHandler.CreateItem)
	items.PUT("/:id", itemHandler.UpdateItem)    
    items.DELETE("/:id", itemHandler.DeleteItem) 

	offers := api.Group("/offers", middleware.AuthRequired())
    offers.POST("", itemHandler.CreateOffer)
    offers.GET("/my", itemHandler.GetMyOffers)
    offers.GET("/inbox", middleware.AuthRequired(), itemHandler.GetOffersToSeller)
    offers.POST("/:id/accept", middleware.AuthRequired(), itemHandler.AcceptOffer)
    offers.POST("/:id/reject", middleware.AuthRequired(), itemHandler.RejectOffer)	

	market := api.Group("/market")
    market.GET("/items", itemHandler.GetMarketplaceItems)
    market.GET("/items/:id", itemHandler.GetItemDetail)  
    orders := api.Group("/orders")
    orders.POST("", middleware.AuthRequired(), itemHandler.CreateOrder)
	// Endpoint Seller/Admin (FR-ORDER-02 & FR-ORDER-03)
    orders.PATCH("/:id/status", middleware.AuthRequired(), itemHandler.UpdateOrderStatus)
    orders.POST("/:id/shipping", middleware.AuthRequired(), itemHandler.InputShippingReceipt)
    
    // Endpoint Buyer/Admin (FR-ORDER-04)
    orders.GET("/:id/tracking", middleware.AuthRequired(), itemHandler.GetOrderTracking)


	// --- Admin Group ---
    admin := api.Group("/admin")
    // Pasang AuthRequired dan RoleAllowed global untuk semua endpoint Admin
    admin.Use(middleware.AuthRequired(), middleware.RoleAllowed("admin")) 
    
    // FR-ADMIN-01: List Users
    admin.GET("/users", adminHandler.ListUsers)
    
    // FR-ADMIN-03: Blokir User
    admin.PATCH("/users/:id/status", adminHandler.BlockUser) 
    
    // FR-ADMIN-02: Moderasi Barang
    admin.PATCH("/items/:id/moderate", adminHandler.ModerateItem)
}