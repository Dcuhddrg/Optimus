package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Api) Users(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []any{}, "total": 0})
}

func (a *Api) UpdateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "user updated"})
}

func (a *Api) UpdateUserStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "user status updated"})
}

func (a *Api) UpdateUserRoles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "user roles updated"})
}
