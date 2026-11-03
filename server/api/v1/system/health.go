package system

import (
	"net/http"

	commonResponse "optimus/server/model/common/response"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

type HealthApi struct{}

func (h *HealthApi) Health(c *gin.Context) {
	service := *statusService
	if service == nil {
		service = &systemService.Service{}
	}

	payload := service.Health()
	utils.JSON(c, http.StatusOK, commonResponse.HealthResponse{
		Status: payload.Status,
	})
}
