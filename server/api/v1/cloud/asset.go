package cloud

import (
	cloudRequest "optimus/server/model/cloud/request"
	cloudResponse "optimus/server/model/cloud/response"
	commonRequest "optimus/server/model/common/request"
	"optimus/server/model/common/response"
	"strconv"

	cloudService "optimus/server/service/cloud"

	"github.com/gin-gonic/gin"
)

// AssetApi handles HTTP requests for cloud assets.
type AssetApi struct{}

// @Tags CloudAsset
// @Summary 分页查询云资产
// @Security ApiKeyAuth
// @Produce application/json
// @Param X-Project-Code header string false "当前项目编码；为空时按可访问项目范围"
// @Param page query int false "页码，从 1 开始" default(1)
// @Param pageSize query int false "每页条数，1～200" default(20)
// @Param accountId query int false "云账号 ID"
// @Param provider query string false "云厂商"
// @Param region query string false "地域"
// @Param resourceType query string false "资源类型"
// @Param status query string false "状态"
// @Param keyword query string false "关键字"
// @Param tag query []string false "标签筛选，key=value，可重复" collectionFormat(multi)
// @Success 200 {object} response.Response{data=response.PageResult[cloudResponse.AssetView]} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /cloud/assets [get]
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
		response.FailWithMessage("failed to query assets", c)
		return
	}
	response.OkWithData(response.PageResult[cloudResponse.AssetView]{
		List: page.Items, Total: page.Total, Page: page.Page, PageSize: page.PageSize,
	}, c)
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
