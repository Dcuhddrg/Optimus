package system

import (
	"errors"
	"net/http"
	systemModel "optimus/server/model/system"
	"optimus/server/utils/response"

	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
)

// Roles returns the migrated response contract.
// @Summary Roles
// @Tags System roles
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope "code=0; data contains list (SystemRole array) and total; no pagination"
// @Failure 401 {object} response.Envelope "code=7; data=null; authentication required"
// @Failure 403 {object} response.Envelope "code=7; data={}; permission denied or protected administrator role"
// @Failure 500 {object} response.Envelope "code=7; data={}; service failure"
// @Failure 503 {object} response.Envelope "code=7; data={}; role service unavailable"
// @Router /system/roles [get]
func (h *RoleHandler) Roles(c *gin.Context) {
	if h.service == nil {
		response.Failure(c, http.StatusServiceUnavailable, "role service is not available")
		return
	}
	roles, err := h.service.ListRoles(c.Request.Context())
	if err != nil {
		response.Failure(c, http.StatusInternalServerError, "failed to list roles")
		return
	}
	if roles == nil {
		roles = []systemModel.SystemRole{}
	}
	response.Success(c, http.StatusOK, gin.H{"list": roles, "total": len(roles)})
}

// CreateRole returns the migrated response contract.
// @Summary CreateRole
// @Tags System roles
// @Produce json
// @Security BearerAuth
// @Accept json
// @Param request body systemModel.RoleRequest true "Existing role request"
// @Success 200 {object} response.Envelope "code=0; data is SystemRole; existing HTTP 200 retained"
// @Failure 400 {object} response.Envelope "code=7; data={}; invalid request"
// @Failure 401 {object} response.Envelope "code=7; data=null; authentication required"
// @Failure 403 {object} response.Envelope "code=7; data={}; permission denied or protected administrator role"
// @Failure 500 {object} response.Envelope "code=7; data={}; service failure"
// @Failure 503 {object} response.Envelope "code=7; data={}; role service unavailable"
// @Router /system/roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	if h.service == nil {
		response.Failure(c, http.StatusServiceUnavailable, "role service is not available")
		return
	}
	var req systemModel.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failure(c, http.StatusBadRequest, "invalid role request")
		return
	}
	role, err := h.service.CreateRole(c.Request.Context(), actorID(c), req, auditMeta(c))
	writeRoleServiceResult(c, role, err)
}

// UpdateRole returns the migrated response contract.
// @Summary UpdateRole
// @Tags System roles
// @Produce json
// @Security BearerAuth
// @Param id path int true "Role ID"
// @Accept json
// @Param request body systemModel.RoleRequest true "Existing role request"
// @Success 200 {object} response.Envelope "code=0; data is SystemRole"
// @Failure 400 {object} response.Envelope "code=7; data={}; invalid request"
// @Failure 401 {object} response.Envelope "code=7; data=null; authentication required"
// @Failure 403 {object} response.Envelope "code=7; data={}; permission denied or protected administrator role"
// @Failure 404 {object} response.Envelope "code=7; data={}; role not found"
// @Failure 500 {object} response.Envelope "code=7; data={}; service failure"
// @Failure 503 {object} response.Envelope "code=7; data={}; role service unavailable"
// @Router /system/roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	if h.service == nil {
		response.Failure(c, http.StatusServiceUnavailable, "role service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemModel.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failure(c, http.StatusBadRequest, "invalid role request")
		return
	}
	role, err := h.service.UpdateRole(c.Request.Context(), actorID(c), id, req, auditMeta(c))
	writeRoleServiceResult(c, role, err)
}

// UpdateRoleStatus returns the migrated response contract.
// @Summary UpdateRoleStatus
// @Tags System roles
// @Produce json
// @Security BearerAuth
// @Param id path int true "Role ID"
// @Accept json
// @Param request body object true "Existing role request"
// @Success 200 {object} response.Envelope "code=0; data is SystemRole"
// @Failure 400 {object} response.Envelope "code=7; data={}; invalid request"
// @Failure 401 {object} response.Envelope "code=7; data=null; authentication required"
// @Failure 403 {object} response.Envelope "code=7; data={}; permission denied or protected administrator role"
// @Failure 404 {object} response.Envelope "code=7; data={}; role not found"
// @Failure 500 {object} response.Envelope "code=7; data={}; service failure"
// @Failure 503 {object} response.Envelope "code=7; data={}; role service unavailable"
// @Router /system/roles/{id}/status [patch]
func (h *RoleHandler) UpdateRoleStatus(c *gin.Context) {
	if h.service == nil {
		response.Failure(c, http.StatusServiceUnavailable, "role service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failure(c, http.StatusBadRequest, "invalid status request")
		return
	}
	role, err := h.service.UpdateRoleStatus(c.Request.Context(), actorID(c), id, req.Status, auditMeta(c))
	writeRoleServiceResult(c, role, err)
}

// UpdateRolePermissions returns the migrated response contract.
// @Summary UpdateRolePermissions
// @Tags System roles
// @Produce json
// @Security BearerAuth
// @Param id path int true "Role ID"
// @Accept json
// @Param request body object true "Existing role request"
// @Success 200 {object} response.Envelope "code=0; data is an empty object"
// @Failure 400 {object} response.Envelope "code=7; data={}; invalid request"
// @Failure 401 {object} response.Envelope "code=7; data=null; authentication required"
// @Failure 403 {object} response.Envelope "code=7; data={}; permission denied or protected administrator role"
// @Failure 404 {object} response.Envelope "code=7; data={}; role not found"
// @Failure 500 {object} response.Envelope "code=7; data={}; service failure"
// @Failure 503 {object} response.Envelope "code=7; data={}; role service unavailable"
// @Router /system/roles/{id}/permissions [put]
func (h *RoleHandler) UpdateRolePermissions(c *gin.Context) {
	if h.service == nil {
		response.Failure(c, http.StatusServiceUnavailable, "role service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		PermissionCodes []string `json:"permissionCodes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failure(c, http.StatusBadRequest, "invalid permissions request")
		return
	}
	err := h.service.UpdateRolePermissions(c.Request.Context(), actorID(c), id, req.PermissionCodes, auditMeta(c))
	if errors.Is(err, systemService.ErrRoleNotFound) {
		response.Failure(c, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, systemService.ErrProtectedAdminRole) {
		response.Failure(c, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		response.Failure(c, http.StatusInternalServerError, "failed to update role permissions")
		return
	}
	response.Success(c, http.StatusOK, nil)
}

// PermissionTree returns the migrated response contract.
// @Summary PermissionTree
// @Tags System roles
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope "code=0; data contains list (SystemPermission array); no pagination"
// @Failure 401 {object} response.Envelope "code=7; data=null; authentication required"
// @Failure 403 {object} response.Envelope "code=7; data={}; permission denied or protected administrator role"
// @Router /system/permissions/tree [get]
func (h *RoleHandler) PermissionTree(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{"list": systemService.DefaultSystemPermissions()})
}

func writeRoleServiceResult(c *gin.Context, role any, err error) {
	if errors.Is(err, systemService.ErrRoleNotFound) {
		response.Failure(c, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, systemService.ErrProtectedAdminRole) {
		response.Failure(c, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		response.Failure(c, http.StatusInternalServerError, "failed to update role")
		return
	}
	response.Success(c, http.StatusOK, role)
}
