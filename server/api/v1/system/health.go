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
	utils.JSON(c, http.StatusOK, commonResponse.HealthResponse{
		Status: payload.Status,
	})
}
