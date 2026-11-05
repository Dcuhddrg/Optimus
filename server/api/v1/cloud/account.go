package cloud

import (
	cloudRequest "optimus/server/model/cloud/request"
	cloudResponse "optimus/server/model/cloud/response"
	"optimus/server/model/common/response"

	"github.com/gin-gonic/gin"
)

// AccountApi handles HTTP requests for cloud accounts.
type AccountApi struct{}

// swag 只能解析本文件导入的包别名；保留此引用，供 Swagger 注释使用 cloudResponse。
var _ = cloudResponse.AccountView{}

// @Tags CloudAccount
// @Summary 云账号列表（不分页，支持标签筛选）
// @Security ApiKeyAuth
// @Produce application/json
// @Param X-Project-Code header string false "当前项目编码；为空时按可访问项目范围"
// @Param tag query []string false "标签筛选，key=value 或 __none__:key，可重复" collectionFormat(multi)
// @Success 200 {object} response.Response{data=object{list=[]cloudResponse.AccountView}} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /cloud/accounts [get]
func (h *AccountApi) ListAccounts(c *gin.Context) {
	svc := accountService
	if !cloudReady() {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	scope := projectScope(c)
	tagFilters, err := parseTagFilters(c)
	if err != nil {
		writeCloudError(c, err, "invalid tag filter")
		return
	}
	views, err := svc.List(c.Request.Context(), scope.AccessibleProjectIDs, tagFilters)
	if err != nil {
		writeCloudError(c, err, "failed to list cloud accounts")
		return
	}
	response.OkWithData(gin.H{"list": views}, c)
}

// @Tags CloudAccount
// @Summary 获取云账号
// @Security ApiKeyAuth
// @Produce application/json
// @Param X-Project-Code header string false "当前项目编码；为空时按可访问项目范围"
// @Param id path int true "云账号 ID"
// @Success 200 {object} response.Response{data=cloudResponse.AccountView} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /cloud/accounts/{id} [get]
func (h *AccountApi) GetAccount(c *gin.Context) {
	svc := accountService
	if !cloudReady() {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	scope := projectScope(c)
	view, err := svc.Get(c.Request.Context(), scope.AccessibleProjectIDs, id)
	if err != nil {
		writeCloudError(c, err, "failed to get cloud account")
		return
	}
	response.OkWithData(view, c)
}

// @Tags CloudAccount
// @Summary 在当前项目下创建云账号
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param X-Project-Code header string true "当前项目编码，创建时必填"
// @Param data body cloudRequest.CreateAccountRequest true "请求体"
// @Success 200 {object} response.Response{data=cloudResponse.AccountView} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /cloud/accounts [post]
func (h *AccountApi) CreateAccount(c *gin.Context) {
	svc := accountService
	if !cloudReady() {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	scope := projectScope(c)
	if scope.CurrentProjectID == 0 {
		response.FailWithMessage("X-Project-Code header is required to create a cloud account", c)
		return
	}
	var req cloudRequest.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request body", c)
		return
	}
	view, err := svc.Create(c.Request.Context(), scope.CurrentProjectID, req)
	if err != nil {
		writeCloudError(c, err, "failed to create cloud account")
		return
	}
	response.OkWithData(view, c)
}

// @Tags CloudAccount
// @Summary 更新云账号
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param X-Project-Code header string false "当前项目编码；为空时按可访问项目范围"
// @Param id path int true "云账号 ID"
// @Param data body cloudRequest.UpdateAccountRequest true "请求体"
// @Success 200 {object} response.Response{data=cloudResponse.AccountView} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /cloud/accounts/{id} [put]
func (h *AccountApi) UpdateAccount(c *gin.Context) {
	svc := accountService
	if !cloudReady() {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	var req cloudRequest.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request body", c)
		return
	}
	scope := projectScope(c)
	view, err := svc.Update(c.Request.Context(), scope.AccessibleProjectIDs, id, req)
	if err != nil {
		writeCloudError(c, err, "failed to update cloud account")
		return
	}
	response.OkWithData(view, c)
}

// @Tags CloudAccount
// @Summary 替换云账号标签
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param X-Project-Code header string false "当前项目编码；为空时按可访问项目范围"
// @Param id path int true "云账号 ID"
// @Param data body cloudRequest.UpdateAccountTagsRequest true "请求体"
// @Success 200 {object} response.Response{data=cloudResponse.AccountView} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /cloud/accounts/{id}/tags [put]
func (h *AccountApi) UpdateAccountTags(c *gin.Context) {
	svc := accountService
	if !cloudReady() {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	var req cloudRequest.UpdateAccountTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid tags payload", c)
		return
	}
	scope := projectScope(c)
	view, err := svc.UpdateTags(c.Request.Context(), scope.AccessibleProjectIDs, id, req)
	if err != nil {
		writeCloudError(c, err, "failed to update cloud account tags")
		return
	}
	response.OkWithData(view, c)
}

// @Tags CloudAccount
// @Summary 删除云账号
// @Security ApiKeyAuth
// @Produce application/json
// @Param X-Project-Code header string false "当前项目编码；为空时按可访问项目范围"
// @Param id path int true "云账号 ID"
// @Success 200 {object} response.Response "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /cloud/accounts/{id} [delete]
func (h *AccountApi) DeleteAccount(c *gin.Context) {
	svc := accountService
	if !cloudReady() {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	scope := projectScope(c)
	if err := svc.Delete(c.Request.Context(), scope.AccessibleProjectIDs, id); err != nil {
		writeCloudError(c, err, "failed to delete cloud account")
		return
	}
	response.Ok(c)
}

// @Tags CloudAccount
// @Summary 同步云账号下的资产
// @Security ApiKeyAuth
// @Produce application/json
// @Param X-Project-Code header string false "当前项目编码；为空时按可访问项目范围"
// @Param id path int true "云账号 ID"
// @Success 200 {object} response.Response{data=cloudResponse.SyncResult} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /cloud/accounts/{id}/sync [post]
func (h *AccountApi) SyncAccount(c *gin.Context) {
	svc := assetService
	if !cloudReady() {
		writeCloudServiceUnavailable(c, "asset registry service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	scope := projectScope(c)
	result, err := svc.SyncAccount(c.Request.Context(), scope.AccessibleProjectIDs, id)
	if err != nil {
		writeCloudError(c, err, "sync failed")
		return
	}
	response.OkWithData(result, c)
}
