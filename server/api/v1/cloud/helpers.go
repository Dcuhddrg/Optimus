package cloud

import (
	"errors"
	"strconv"

	"optimus/server/global"
	cloudRequest "optimus/server/model/cloud/request"
	"optimus/server/model/common/response"
	cloudService "optimus/server/service/cloud"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

// cloudReady reports whether cloud APIs can serve: the database is up and the
// credential encryption key is valid (#54 keeps cloud unavailable otherwise).
func cloudReady() bool {
	if global.OPTIMUS_DB == nil {
		return false
	}
	_, err := cloudService.CredentialKey()
	return err == nil
}

func writeCloudServiceUnavailable(c *gin.Context, message string) {
	response.FailWithMessage(message, c)
}

func writeCloudError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, cloudService.ErrAccountNotFound):
		response.FailWithMessage(err.Error(), c)
	case errors.Is(err, cloudService.ErrProviderNotRegistered):
		response.FailWithMessage(err.Error(), c)
	case errors.Is(err, cloudService.ErrInvalidTagKey),
		errors.Is(err, cloudService.ErrInvalidTagValue),
		errors.Is(err, cloudService.ErrTooManyTags),
		errors.Is(err, cloudService.ErrInvalidTagFilter):
		response.FailWithMessage(err.Error(), c)
	default:
		response.FailWithMessage(fallback, c)
	}
}

func parseTagFilters(c *gin.Context) ([]cloudService.TagFilter, error) {
	var query cloudRequest.AccountSearch
	if err := c.ShouldBindQuery(&query); err != nil {
		return nil, err
	}
	return parseTagFilterTokens(query.Tags)
}

func parseTagFilterTokens(tokens []string) ([]cloudService.TagFilter, error) {
	if len(tokens) == 0 {
		return nil, nil
	}
	out := make([]cloudService.TagFilter, 0, len(tokens))
	for _, token := range tokens {
		filter, err := cloudService.ParseTagFilter(token)
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
		response.FailWithMessage("invalid "+name, c)
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
