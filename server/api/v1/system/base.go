package system

import (
	"net/http"

	commonResponse "optimus/server/model/common/response"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

type BaseApi struct{}

func (h *BaseApi) Ping(c *gin.Context) {
	service := *statusService
	if service == nil {
		service = &systemService.Service{}
	}

	payload := service.Ping()
	utils.JSON(c, http.StatusOK, commonResponse.PingResponse{
		Message: payload.Message,
	})
}
