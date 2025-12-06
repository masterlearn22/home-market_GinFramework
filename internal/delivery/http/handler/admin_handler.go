package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	service "home-market/internal/service/postgresql"
	entity "home-market/internal/domain" 
)

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// FR-ADMIN-01: List Users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	role := c.MustGet("role_name").(string)
	users, err := h.adminService.ListUsers(role)

	if err != nil {
		if err.Error() == "unauthorized: admin access required" { 
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"data": users}) 
}

// FR-ADMIN-03: Block/Unblock User
func (h *AdminHandler) BlockUser(c *gin.Context) {
	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr) 
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id format"})
		return
	}
	
	var input entity.UpdateUserStatusInput 
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input body", "detail": err.Error()})
		return
	}
	
	adminRole := c.MustGet("role_name").(string)
	err = h.adminService.BlockUser(adminRole, targetID, input.IsActive) 
	
	if err != nil {
		if err.Error() == "unauthorized: admin access required" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "user status updated to", "is_active": input.IsActive})
}

// FR-ADMIN-02: Moderate Item
func (h *AdminHandler) ModerateItem(c *gin.Context) {
	itemIDStr := c.Param("id")
	itemID, err := uuid.Parse(itemIDStr) 
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id format"})
		return
	}
	
	adminRole := c.MustGet("role_name").(string)
	err = h.adminService.ModerateItem(adminRole, itemID) 
	
	if err != nil {
		if err.Error() == "item not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "unauthorized: admin access required" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "item successfully moderated (set inactive)"})
}