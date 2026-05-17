package cloud

import (
	"errors"
	"net/http"
	"strconv"

	cloudUsecase "optimus/server/service/cloud/usecase"
	"optimus/server/utils"

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

func writeCloudError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, cloudUsecase.ErrAccountNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, cloudUsecase.ErrProviderNotRegistered):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
	case errors.Is(err, cloudUsecase.ErrInvalidTagKey),
		errors.Is(err, cloudUsecase.ErrInvalidTagValue),
		errors.Is(err, cloudUsecase.ErrTooManyTags),
		errors.Is(err, cloudUsecase.ErrInvalidTagFilter):
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": fallback})
	}
}

func parseTagFilters(c *gin.Context) ([]cloudUsecase.TagFilter, error) {
	tokens := c.QueryArray("tag")
	if len(tokens) == 0 {
		return nil, nil
	}
	out := make([]cloudUsecase.TagFilter, 0, len(tokens))
	for _, token := range tokens {
		filter, err := cloudUsecase.ParseTagFilter(token)
		if err != nil {
			return nil, err
		}
		out = append(out, filter)
	}
	return out, nil
}

func parseCloudUintParam(c *gin.Context, name string) (uint, bool) {
	n, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || n == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid " + name})
		return 0, false
	}
	return uint(n), true
}

func projectScope(c *gin.Context) utils.ProjectScope {
	raw, exists := c.Get(utils.ProjectScopeContextKey)
	if !exists {
		return utils.ProjectScope{}
	}
	scope, ok := raw.(utils.ProjectScope)
	if !ok {
		return utils.ProjectScope{}
	}
	return scope
}
