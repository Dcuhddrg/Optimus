package project

import (
	"net/http"
	projectRequest "optimus/server/model/project/request"

	"github.com/gin-gonic/gin"
)

func (h *MemberApi) ListMembers(c *gin.Context) {
	if projectServices.ProjectService == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	members, err := projectServices.ProjectService.ListMembers(c.Request.Context(), projectActorID(c), c.Param("code"))
	if err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": members})
}

func (h *MemberApi) AddMember(c *gin.Context) {
	if projectServices.ProjectService == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	var req projectRequest.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid member payload"})
		return
	}
	if req.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "userId is required"})
		return
	}
	if err := projectServices.ProjectService.AddMember(c.Request.Context(), projectActorID(c), c.Param("code"), req.UserID, req.ProjectRole); err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "member added"})
}

func (h *MemberApi) UpdateMemberRole(c *gin.Context) {
	if projectServices.ProjectService == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	userID, ok := projectParseUintParam(c, "userId")
	if !ok {
		return
	}
	var req projectRequest.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid member payload"})
		return
	}
	if err := projectServices.ProjectService.UpdateMemberRole(c.Request.Context(), projectActorID(c), c.Param("code"), userID, req.ProjectRole); err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member role updated"})
}

func (h *MemberApi) RemoveMember(c *gin.Context) {
	if projectServices.ProjectService == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	userID, ok := projectParseUintParam(c, "userId")
	if !ok {
		return
	}
	if err := projectServices.ProjectService.RemoveMember(c.Request.Context(), projectActorID(c), c.Param("code"), userID); err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}
