package cloud

import (
	"errors"
	"net/http"
	"strconv"

	cloudService "optimus/server/service/cloud"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func writeCloudServiceUnavailable(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"message": message})
}

func writeCloudError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, cloudService.ErrAccountNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, cloudService.ErrProviderNotRegistered):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
	case errors.Is(err, cloudService.ErrInvalidTagKey),
		errors.Is(err, cloudService.ErrInvalidTagValue),
		errors.Is(err, cloudService.ErrTooManyTags),
		errors.Is(err, cloudService.ErrInvalidTagFilter):
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": fallback})
	}
}

func parseTagFilters(c *gin.Context) ([]cloudService.TagFilter, error) {
	return parseTagFilterTokens(c.QueryArray("tag"))
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
