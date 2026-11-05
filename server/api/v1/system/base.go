package system

import (
	"net/http"

	commonResponse "optimus/server/model/common/response"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

type BaseApi struct{}

// @Tags Base
// @Summary 存活探针
// @Produce application/json
// @Success 200 {object} commonResponse.PingResponse
// @Router /ping [get]
// @Router /base/ping [get]
func (h *BaseApi) Ping(c *gin.Context) {
	service := statusService

	payload := service.Ping()
	utils.JSON(c, http.StatusOK, commonResponse.PingResponse{
		Message: payload.Message,
	})
}
