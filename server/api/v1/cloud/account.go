package cloud

import (
	"net/http"
	cloudRequest "optimus/server/model/cloud/request"

	"github.com/gin-gonic/gin"
)

// AccountApi handles HTTP requests for cloud accounts.
type AccountApi struct{}

func (h *AccountApi) ListAccounts(c *gin.Context) {
	svc := accountService
	if !svc.Ready() {
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
	c.JSON(http.StatusOK, gin.H{"items": views, "total": len(views)})
}

func (h *AccountApi) GetAccount(c *gin.Context) {
	svc := accountService
	if !svc.Ready() {
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
	c.JSON(http.StatusOK, view)
}

func (h *AccountApi) CreateAccount(c *gin.Context) {
	svc := accountService
	if !svc.Ready() {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	scope := projectScope(c)
	if scope.CurrentProjectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "X-Project-Code header is required to create a cloud account"})
		return
	}
	var req cloudRequest.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}
	view, err := svc.Create(c.Request.Context(), scope.CurrentProjectID, req)
	if err != nil {
		writeCloudError(c, err, "failed to create cloud account")
		return
	}
	c.JSON(http.StatusCreated, view)
}

func (h *AccountApi) UpdateAccount(c *gin.Context) {
	svc := accountService
	if !svc.Ready() {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	var req cloudRequest.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}
	scope := projectScope(c)
	view, err := svc.Update(c.Request.Context(), scope.AccessibleProjectIDs, id, req)
	if err != nil {
		writeCloudError(c, err, "failed to update cloud account")
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *AccountApi) UpdateAccountTags(c *gin.Context) {
	svc := accountService
	if !svc.Ready() {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	var req cloudRequest.UpdateAccountTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid tags payload"})
		return
	}
	scope := projectScope(c)
	view, err := svc.UpdateTags(c.Request.Context(), scope.AccessibleProjectIDs, id, req)
	if err != nil {
		writeCloudError(c, err, "failed to update cloud account tags")
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *AccountApi) DeleteAccount(c *gin.Context) {
	svc := accountService
	if !svc.Ready() {
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
	c.JSON(http.StatusOK, gin.H{"message": "cloud account deleted"})
}

func (h *AccountApi) SyncAccount(c *gin.Context) {
	svc := assetService
	if !svc.Ready() {
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
	c.JSON(http.StatusOK, result)
}
