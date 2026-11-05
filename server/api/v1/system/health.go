package system

import (
	"net/http"

	commonResponse "optimus/server/model/common/response"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

type HealthApi struct{}

// @Tags Base
// @Summary 健康检查：数据库或启动数据初始化失败时返回 503
// @Produce application/json
// @Success 200 {object} commonResponse.HealthResponse
// @Failure 503 {object} commonResponse.HealthResponse
// @Router /health [get]
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
