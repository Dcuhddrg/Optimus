package project

import (
	"net/http"

	projectUsecase "optimus/server/service/project/usecase"

	"github.com/gin-gonic/gin"
)

func (h *ProjectHandlers) ListMembers(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	members, err := h.svc.ListMembers(c.Request.Context(), projectActorID(c), c.Param("code"))
	if err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": members})
}

func (h *ProjectHandlers) AddMember(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	var req projectUsecase.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid member payload"})
		return
	}
	if req.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "userId is required"})
		return
	}
	if err := h.svc.AddMember(c.Request.Context(), projectActorID(c), c.Param("code"), req.UserID, req.ProjectRole); err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "member added"})
}

func (h *ProjectHandlers) UpdateMemberRole(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	userID, ok := projectParseUintParam(c, "userId")
	if !ok {
		return
	}
	var req projectUsecase.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid member payload"})
		return
	}
	if err := h.svc.UpdateMemberRole(c.Request.Context(), projectActorID(c), c.Param("code"), userID, req.ProjectRole); err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member role updated"})
}

func (h *ProjectHandlers) RemoveMember(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	userID, ok := projectParseUintParam(c, "userId")
	if !ok {
		return
	}
	if err := h.svc.RemoveMember(c.Request.Context(), projectActorID(c), c.Param("code"), userID); err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}
