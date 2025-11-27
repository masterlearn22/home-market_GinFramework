package handler

import (
	"net/http"

	entity "home-market/internal/domain"
	service "home-market/internal/service/postgresql"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CategoryHandler struct {
	categoryService *service.CategoryService
}

func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	rawID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := rawID.(uuid.UUID)

	rawRole, ok := c.Get("role_name")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "role missing"})
		return
	}
	role := rawRole.(string)

	var input entity.CreateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "detail": err.Error()})
		return
	}

	category, err := h.categoryService.CreateCategory(userID, role, input)
	if err != nil {
		switch err {
		case service.ErrNotSeller:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case service.ErrNoShopOwned:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case service.ErrCategoryExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, category)
}
