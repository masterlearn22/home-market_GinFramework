package config

import (
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
)

func SetupGin() *gin.Engine {
	// Gin router
	router := gin.New()

	// Middleware: Logger + Recovery (hindari panic menghentikan server)
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Body size limit (10 MB)
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10*1024*1024)
		c.Next()
	})

	// Custom global error handler
	router.Use(func(c *gin.Context) {
		c.Next()

		// Jika terjadi error
		err := c.Errors.Last()
		if err != nil {
			// Secara default gunakan 500
			code := http.StatusInternalServerError

			// Jika error sudah punya status code (err.Meta)
			if meta, ok := err.Meta.(int); ok {
				code = meta
			}

			c.JSON(code, gin.H{
				"error":   true,
				"message": err.Error(),
			})
			return
		}
	})

	fmt.Println("Gin is running")
	return router
}
