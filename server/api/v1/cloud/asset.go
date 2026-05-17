package cloud

import (
	"net/http"
	"strconv"

	cloudUsecase "optimus/server/service/cloud/usecase"

	"github.com/gin-gonic/gin"
)

func (h *AssetHandler) ListAssets(c *gin.Context) {
	if h.service == nil {
		writeCloudServiceUnavailable(c, "asset registry service is not available")
		return
	}
	scope := projectScope(c)
	f, err := parseAssetFilter(c)
	if err != nil {
		writeCloudError(c, err, "invalid asset filter")
		return
	}
	f.AccessibleProjectIDs = scope.AccessibleProjectIDs
	page, err := h.service.QueryAssets(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to query assets"})
		return
	}
	c.JSON(http.StatusOK, page)
}

func parseAssetFilter(c *gin.Context) (cloudUsecase.AssetFilter, error) {
	f := cloudUsecase.AssetFilter{Page: 1, PageSize: 20}

	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			f.Page = n
		}
	}
	if v := c.Query("pageSize"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 200 {
			f.PageSize = n
		}
	}
	if v := c.Query("accountId"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			id := uint(n)
			f.AccountID = &id
		}
	}
	if v := c.Query("provider"); v != "" {
		f.Provider = &v
	}
	if v := c.Query("region"); v != "" {
		f.Region = &v
	}
	if v := c.Query("resourceType"); v != "" {
		f.ResourceType = &v
	}
	if v := c.Query("status"); v != "" {
		f.Status = &v
	}
	if v := c.Query("keyword"); v != "" {
		f.Keyword = &v
	}
	tagFilters, err := parseTagFilters(c)
	if err != nil {
		return cloudUsecase.AssetFilter{}, err
	}
	f.Tags = tagFilters
	return f, nil
}
