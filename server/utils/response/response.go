// Package response provides explicit HTTP responses for migrated routes.
package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const envelopeKey = "optimus.response.envelope"

type Envelope struct {
	Code int    `json:"code"`
	Data any    `json:"data"`
	Msg  string `json:"msg"`
}
type Route struct{ Method, Path string }

// EnvelopeForRoutes marks only the registered method+FullPath pairs, before authentication.
// It never intercepts or wraps a response and does not negotiate from a client header.
func EnvelopeForRoutes(prefix string, routes []Route) gin.HandlerFunc {
	paths := make(map[string]struct{}, len(routes))
	base := strings.TrimRight(prefix, "/")
	for _, r := range routes {
		paths[r.Method+" "+base+r.Path] = struct{}{}
	}
	return func(c *gin.Context) {
		if _, ok := paths[c.Request.Method+" "+c.FullPath()]; ok {
			c.Set(envelopeKey, true)
		}
		c.Next()
	}
}
func Success(c *gin.Context, status int, data any) {
	if data == nil {
		data = struct{}{}
	}
	c.JSON(status, Envelope{Code: 0, Data: data, Msg: "success"})
}
func Failure(c *gin.Context, status int, message string) {
	var data any = struct{}{}
	if status == http.StatusUnauthorized {
		data = nil
	}
	c.JSON(status, Envelope{Code: 7, Data: data, Msg: message})
}
func IsEnvelope(c *gin.Context) bool {
	migrated, _ := c.Get(envelopeKey)
	return migrated == true
}

func AbortFailure(c *gin.Context, status int, message string) {
	c.Abort()
	if IsEnvelope(c) {
		Failure(c, status, message)
	} else {
		c.JSON(status, gin.H{"message": message})
	}
}
