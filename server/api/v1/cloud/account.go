package cloud

import (
	"errors"
	"net/http"

	cloudUsecase "optimus/server/service/cloud/usecase"

	"github.com/gin-gonic/gin"
)

func (h *AccountHandler) ListAccounts(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	views, err := h.accountSvc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to list cloud accounts"})
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
	view, err := h.accountSvc.Get(c.Request.Context(), id)
	if errors.Is(err, cloudUsecase.ErrAccountNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get cloud account"})
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *AccountHandler) CreateAccount(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	var req cloudUsecase.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}
	view, err := h.accountSvc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create cloud account"})
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
	var req cloudUsecase.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}
	view, err := h.accountSvc.Update(c.Request.Context(), id, req)
	if errors.Is(err, cloudUsecase.ErrAccountNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update cloud account"})
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
	if err := h.accountSvc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, cloudUsecase.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to delete cloud account"})
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
	result, err := h.assetSvc.SyncAccount(c.Request.Context(), id)
	if errors.Is(err, cloudUsecase.ErrAccountNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if errors.Is(err, cloudUsecase.ErrProviderNotRegistered) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "sync failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}
