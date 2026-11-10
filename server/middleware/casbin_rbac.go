package middleware

import (
	"strconv"
	"strings"

	"optimus/server/model/common/response"
	"optimus/server/service"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

// PolicySyncingMessage is returned while the Casbin policy is being reloaded
// after a failure.
const PolicySyncingMessage = "permission system is syncing, please retry later"

var casbinService = &service.ServiceGroupApp.SystemServiceGroup.CasbinService

// CasbinHandler authorizes the request path (without router prefix) and method
// against casbin_rule, as GVA. Optimus users may hold several roles, so any
// role allowing the request suffices. Roles come from JWTAuth's per-request
// load, not from the token. While the policy is stale after a failed reload
// every request is denied instead of being judged by the old policy.
func CasbinHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := c.Get(utils.RoleIDsContextKey)
		roleIDs, _ := raw.([]uint)
		if !ok {
			response.FailWithDetailed(gin.H{}, "permission denied", c)
			c.Abort()
			return
		}
		e, ready := casbinService.Enforcer()
		if !ready {
			response.FailWithDetailed(gin.H{}, PolicySyncingMessage, c)
			c.Abort()
			return
		}
		obj := strings.TrimPrefix(c.Request.URL.Path, strings.TrimRight(utils.RouterPrefix(), "/"))
		act := c.Request.Method
		for _, roleID := range roleIDs {
			if success, _ := e.Enforce(strconv.Itoa(int(roleID)), obj, act); success {
				c.Next()
				return
			}
		}
		response.FailWithDetailed(gin.H{}, "permission denied", c)
		c.Abort()
	}
}
