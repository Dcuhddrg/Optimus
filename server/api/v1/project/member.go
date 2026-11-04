package project

import (
	"optimus/server/global"
	"optimus/server/model/common/response"
	projectRequest "optimus/server/model/project/request"

	"github.com/gin-gonic/gin"
)

type MemberApi struct{}

func (h *MemberApi) ListMembers(c *gin.Context) {
	svc := projectService
	if global.OPTIMUS_DB == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	members, err := svc.ListMembers(c.Request.Context(), projectActorID(c), c.Param("code"))
	if err != nil {
		writeProjectError(c, err)
		return
	}
	response.OkWithData(gin.H{"list": members}, c)
}

func (h *MemberApi) AddMember(c *gin.Context) {
	svc := projectService
	if global.OPTIMUS_DB == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	var req projectRequest.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid member payload", c)
		return
	}
	if req.UserID == 0 {
		response.FailWithMessage("userId is required", c)
		return
	}
	if err := svc.AddMember(c.Request.Context(), projectActorID(c), c.Param("code"), req.UserID, req.ProjectRole); err != nil {
		writeProjectError(c, err)
		return
	}
	response.Ok(c)
}

func (h *MemberApi) UpdateMemberRole(c *gin.Context) {
	svc := projectService
	if global.OPTIMUS_DB == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	userID, ok := projectParseUintParam(c, "userId")
	if !ok {
		return
	}
	var req projectRequest.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid member payload", c)
		return
	}
	if err := svc.UpdateMemberRole(c.Request.Context(), projectActorID(c), c.Param("code"), userID, req.ProjectRole); err != nil {
		writeProjectError(c, err)
		return
	}
	response.Ok(c)
}

func (h *MemberApi) RemoveMember(c *gin.Context) {
	svc := projectService
	if global.OPTIMUS_DB == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	userID, ok := projectParseUintParam(c, "userId")
	if !ok {
		return
	}
	if err := svc.RemoveMember(c.Request.Context(), projectActorID(c), c.Param("code"), userID); err != nil {
		writeProjectError(c, err)
		return
	}
	response.Ok(c)
}
