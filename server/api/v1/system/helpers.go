package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func writeServiceUnavailable(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"message": message})
}
