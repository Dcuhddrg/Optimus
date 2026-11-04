package system

import (
	"net/http"

	commonResponse "optimus/server/model/common/response"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

type HealthApi struct{}

func (h *HealthApi) Health(c *gin.Context) {
	service := statusService

	payload := service.Health()
	code := http.StatusOK
	if !payload.Healthy {
		code = http.StatusServiceUnavailable
	}
	utils.JSON(c, code, commonResponse.HealthResponse{
		Status: payload.Status,
	})
}
