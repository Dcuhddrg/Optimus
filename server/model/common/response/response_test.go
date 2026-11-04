package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGVAHelpersUseBusinessCodeWithHTTP200(t *testing.T) {
	for _, tt := range []struct {
		name string
		call func(*gin.Context)
		want string
	}{
		{"Ok", Ok, `{"code":0,"data":{},"msg":"操作成功"}`},
		{"OkWithMessage", func(c *gin.Context) { OkWithMessage("done", c) }, `{"code":0,"data":{},"msg":"done"}`},
		{"OkWithData falsy", func(c *gin.Context) { OkWithData(false, c) }, `{"code":0,"data":false,"msg":"成功"}`},
		{"OkWithDetailed", func(c *gin.Context) { OkWithDetailed(0, "done", c) }, `{"code":0,"data":0,"msg":"done"}`},
		{"Fail", Fail, `{"code":7,"data":{},"msg":"操作失败"}`},
		{"FailWithMessage", func(c *gin.Context) { FailWithMessage("permission denied", c) }, `{"code":7,"data":{},"msg":"permission denied"}`},
		{"FailWithDetailed", func(c *gin.Context) { FailWithDetailed(map[string]any{"id": 1}, "bad", c) }, `{"code":7,"data":{"id":1},"msg":"bad"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.call(c)
			require.Equal(t, http.StatusOK, w.Code)
			require.JSONEq(t, tt.want, w.Body.String())
		})
	}
}

func TestGVANoAuthReturnsHTTP401(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	NoAuth("expired", c)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.JSONEq(t, `{"code":7,"data":null,"msg":"expired"}`, w.Body.String())
}
