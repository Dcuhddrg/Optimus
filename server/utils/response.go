package utils

import "github.com/gin-gonic/gin"

func JSON(c *gin.Context, status int, payload any) {
	c.JSON(status, payload)
}
