package cloud

import (
	"net/http"
	cloudModel "optimus/server/model/cloud"

	"github.com/gin-gonic/gin"
)

func (h *AccountHandler) ListAccounts(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	scope := projectScope(c)
	tagFilters, err := parseTagFilters(c)
	if err != nil {
		writeCloudError(c, err, "invalid tag filter")
		return
	}
	views, err := h.accountSvc.List(c.Request.Context(), scope.AccessibleProjectIDs, tagFilters)
	if err != nil {
		writeCloudError(c, err, "failed to list cloud accounts")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": views, "total": len(views)})
}

func (h *AccountHandler) GetAccount(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	scope := projectScope(c)
	view, err := h.accountSvc.Get(c.Request.Context(), scope.AccessibleProjectIDs, id)
	if err != nil {
		writeCloudError(c, err, "failed to get cloud account")
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *AccountHandler) CreateAccount(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	scope := projectScope(c)
	if scope.CurrentProjectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "X-Project-Code header is required to create a cloud account"})
		return
	}
	var req cloudModel.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}
	view, err := h.accountSvc.Create(c.Request.Context(), scope.CurrentProjectID, req)
	if err != nil {
		writeCloudError(c, err, "failed to create cloud account")
		return
	}
	c.JSON(http.StatusCreated, view)
}

func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	var req cloudModel.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}
	scope := projectScope(c)
	view, err := h.accountSvc.Update(c.Request.Context(), scope.AccessibleProjectIDs, id, req)
	if err != nil {
		writeCloudError(c, err, "failed to update cloud account")
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *AccountHandler) UpdateAccountTags(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	var req cloudModel.UpdateAccountTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid tags payload"})
		return
	}
	scope := projectScope(c)
	view, err := h.accountSvc.UpdateTags(c.Request.Context(), scope.AccessibleProjectIDs, id, req)
	if err != nil {
		writeCloudError(c, err, "failed to update cloud account tags")
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	scope := projectScope(c)
	if err := h.accountSvc.Delete(c.Request.Context(), scope.AccessibleProjectIDs, id); err != nil {
		writeCloudError(c, err, "failed to delete cloud account")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cloud account deleted"})
}

func (h *AccountHandler) SyncAccount(c *gin.Context) {
	if h.assetSvc == nil {
		writeCloudServiceUnavailable(c, "asset registry service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	scope := projectScope(c)
	result, err := h.assetSvc.SyncAccount(c.Request.Context(), scope.AccessibleProjectIDs, id)
	if err != nil {
		writeCloudError(c, err, "sync failed")
		return
	}
	c.JSON(http.StatusOK, result)
}
