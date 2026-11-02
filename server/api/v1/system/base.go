package system

import (
	"net/http"

	commonResponse "optimus/server/model/common/response"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func (h *BaseHandler) Ping(c *gin.Context) {
	service := h.service
	if service == nil {
		service = &systemService.Service{}
	}

	payload := service.Ping()
	utils.JSON(c, http.StatusOK, commonResponse.PingResponse{
		Message: payload.Message,
	})
}
