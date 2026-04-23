package system

import (
	"net/http"

	commonModel "optimus/server/model/common"
	"optimus/server/service"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func (a *Api) Ping(c *gin.Context) {
	payload := service.ServiceGroupApp.System.Ping()
	utils.JSON(c, http.StatusOK, commonModel.PingResponse{
		Message: payload.Message,
	})
}
