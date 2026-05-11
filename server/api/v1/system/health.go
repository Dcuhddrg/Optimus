package system

import (
	"net/http"

	commonModel "optimus/server/model/common"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func (h *HealthHandler) Health(c *gin.Context) {
	service := h.service
	if service == nil {
		service = &systemDefaultStatusService
	}

	payload := service.Health()
	utils.JSON(c, http.StatusOK, commonModel.HealthResponse{
		Status: payload.Status,
	})
}
