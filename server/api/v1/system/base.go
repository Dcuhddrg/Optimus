package system

import (
	"net/http"

	commonResponse "optimus/server/model/common/response"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

type BaseApi struct{}

func (h *BaseApi) Ping(c *gin.Context) {
	service := statusService

	payload := service.Ping()
	utils.JSON(c, http.StatusOK, commonResponse.PingResponse{
		Message: payload.Message,
	})
}
