package system

import (
	"net/http"

	commonModel "optimus/server/model/common"
	"optimus/server/service"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

type Api struct{}

func (a *Api) Health(c *gin.Context) {
	payload := service.ServiceGroupApp.System.Health()
	utils.JSON(c, http.StatusOK, commonModel.HealthResponse{
		Status: payload.Status,
	})
}
