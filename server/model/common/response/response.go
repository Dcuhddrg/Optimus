// Package response is GVA's common response protocol: every business result is
// HTTP 200 with {code,data,msg}; code 0 succeeds and code 7 fails. Only an
// unauthenticated request (NoAuth) uses HTTP 401.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int    `json:"code"`
	Data any    `json:"data"`
	Msg  string `json:"msg"`
}

const (
	ERROR   = 7
	SUCCESS = 0
)

func Result(code int, data any, msg string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{Code: code, Data: data, Msg: msg})
}

func Ok(c *gin.Context) {
	Result(SUCCESS, map[string]any{}, "操作成功", c)
}

func OkWithMessage(message string, c *gin.Context) {
	Result(SUCCESS, map[string]any{}, message, c)
}

func OkWithData(data any, c *gin.Context) {
	Result(SUCCESS, data, "成功", c)
}

func OkWithDetailed(data any, message string, c *gin.Context) {
	Result(SUCCESS, data, message, c)
}

func Fail(c *gin.Context) {
	Result(ERROR, map[string]any{}, "操作失败", c)
}

func FailWithMessage(message string, c *gin.Context) {
	Result(ERROR, map[string]any{}, message, c)
}

func NoAuth(message string, c *gin.Context) {
	c.JSON(http.StatusUnauthorized, Response{Code: ERROR, Data: nil, Msg: message})
}

func FailWithDetailed(data any, message string, c *gin.Context) {
	Result(ERROR, data, message, c)
}
