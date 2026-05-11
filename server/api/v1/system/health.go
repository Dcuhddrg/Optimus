package system

import (
	"net/http"

	commonModel "optimus/server/model/common"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func (h *HealthHandler) Health(c *gin.Context) {
	service := h.service
	if service == nil {
		service = &systemService.Service{}
	}

	payload := service.Health()
	utils.JSON(c, http.StatusOK, commonModel.HealthResponse{
		Status: payload.Status,
	})
}
