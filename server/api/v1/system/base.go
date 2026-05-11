package system

import (
	"net/http"

	commonModel "optimus/server/model/common"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func (h *BaseHandler) Ping(c *gin.Context) {
	service := h.service
	if service == nil {
		service = &systemDefaultStatusService
	}

	payload := service.Ping()
	utils.JSON(c, http.StatusOK, commonModel.PingResponse{
		Message: payload.Message,
	})
}
