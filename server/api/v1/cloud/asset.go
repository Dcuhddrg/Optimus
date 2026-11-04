package cloud

import (
	"net/http"
	cloudRequest "optimus/server/model/cloud/request"
	commonRequest "optimus/server/model/common/request"
	"strconv"

	cloudService "optimus/server/service/cloud"

	"github.com/gin-gonic/gin"
)

// AssetApi handles HTTP requests for cloud assets.
type AssetApi struct{}

func (h *AssetApi) ListAssets(c *gin.Context) {
	svc := assetService
	if !cloudReady() {
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
	page, err := svc.QueryAssets(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to query assets"})
		return
	}
	c.JSON(http.StatusOK, page)
}

func parseAssetFilter(c *gin.Context) (cloudService.AssetFilter, error) {
	page, pageErr := strconv.Atoi(c.Query("page"))
	if pageErr != nil {
		page = 0
	}
	pageSize, sizeErr := strconv.Atoi(c.Query("pageSize"))
	if sizeErr != nil {
		pageSize = 0
	}
	pagination := commonRequest.PageInfo{Page: page, PageSize: pageSize}
	pagination.Normalize()
	f := cloudRequest.AssetSearch{PageInfo: pagination, Tags: c.QueryArray("tag")}

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
	tagFilters, err := parseTagFilterTokens(f.Tags)
	if err != nil {
		return cloudService.AssetFilter{}, err
	}
	return cloudService.AssetFilter{
		AccountID: f.AccountID, Provider: f.Provider, Region: f.Region,
		ResourceType: f.ResourceType, Status: f.Status, Keyword: f.Keyword,
		Tags: tagFilters, Page: f.Page, PageSize: f.PageSize,
	}, nil
}
