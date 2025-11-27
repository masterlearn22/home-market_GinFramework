package route

import (
	"database/sql"
	"log"
	"github.com/google/uuid"
	httpHandler "home-market/internal/delivery/http/handler"
	repo "home-market/internal/repository/postgresql"
	service "home-market/internal/service/postgresql"
	"github.com/gin-gonic/gin"
)

func SetupRoute(app *gin.Engine, db *sql.DB) {
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

	// --- 3. Definisikan group route ---
	api := app.Group("/api")

	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)
	auth.GET("/profile", authHandler.Profile) // nanti bisa kamu pasang middleware JWT di sini
}