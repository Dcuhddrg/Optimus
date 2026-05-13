package cloud

import (
	"net/http"
	"strconv"

	cloudUsecase "optimus/server/service/cloud/usecase"

	"github.com/gin-gonic/gin"
)

// AccountHandler handles HTTP requests for cloud accounts.
type AccountHandler struct {
	accountSvc *cloudUsecase.CloudAccountService
	assetSvc   *cloudUsecase.AssetRegistryService
}

// AssetHandler handles HTTP requests for cloud assets.
type AssetHandler struct {
	service *cloudUsecase.AssetRegistryService
}

// CloudHandlers groups all cloud HTTP handlers.
type CloudHandlers struct {
	Account *AccountHandler
	Asset   *AssetHandler
}

// CloudHandlerDeps carries the services needed to build CloudHandlers.
type CloudHandlerDeps struct {
	Account *cloudUsecase.CloudAccountService
	Asset   *cloudUsecase.AssetRegistryService
}

func NewCloudHandlers(deps CloudHandlerDeps) *CloudHandlers {
	return &CloudHandlers{
		Account: &AccountHandler{accountSvc: deps.Account, assetSvc: deps.Asset},
		Asset:   &AssetHandler{service: deps.Asset},
	}
}

func writeCloudServiceUnavailable(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"message": message})
}

func parseCloudUintParam(c *gin.Context, name string) (uint, bool) {
	n, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || n == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid " + name})
		return 0, false
	}
	return uint(n), true
}
