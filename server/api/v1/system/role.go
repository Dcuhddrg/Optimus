package system

import (
	"net/http"

	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
)

func (a *Api) Roles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []any{}, "total": 0})
}

func (a *Api) CreateRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "role created"})
}

func (a *Api) UpdateRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

func (a *Api) UpdateRoleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "role status updated"})
}

func (a *Api) UpdateRolePermissions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "role permissions updated"})
}

func (a *Api) PermissionTree(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": systemService.DefaultSystemPermissions()})
}
