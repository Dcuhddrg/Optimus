package system

import (
	"optimus/server/model/common/response"

	"github.com/gin-gonic/gin"
)

func writeServiceUnavailable(c *gin.Context, message string) {
	response.FailWithMessage(message, c)
}
