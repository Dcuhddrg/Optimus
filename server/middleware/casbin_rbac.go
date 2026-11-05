package middleware

import (
	"strconv"
	"strings"

	"optimus/server/model/common/response"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

// CasbinHandler authorizes the request path (without router prefix) and method
// against casbin_rule, as GVA. Optimus users may hold several roles, so any
// role allowing the request suffices.
func CasbinHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var waitUse *utils.CustomClaims
		if claims, ok := c.Get(utils.ClaimsContextKey); ok {
			waitUse, _ = claims.(*utils.CustomClaims)
		} else {
			waitUse, _ = utils.GetClaims(c)
		}
		if waitUse == nil {
			response.FailWithDetailed(gin.H{}, "permission denied", c)
			c.Abort()
			return
		}
		obj := strings.TrimPrefix(c.Request.URL.Path, strings.TrimRight(utils.RouterPrefix(), "/"))
		act := c.Request.Method
		e := utils.GetCasbin()
		for _, roleID := range waitUse.RoleIDs {
			if success, _ := e.Enforce(strconv.Itoa(int(roleID)), obj, act); success {
				c.Next()
				return
			}
		}
		response.FailWithDetailed(gin.H{}, "permission denied", c)
		c.Abort()
	}
}
