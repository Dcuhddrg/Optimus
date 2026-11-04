package project

import (
	"net/http"
	"net/http/httptest"
	"testing"

	projectSvc "optimus/server/service/project"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestProjectErrorsUseGVAResponse(t *testing.T) {
	for _, send := range []func(*gin.Context){
		writeProjectServiceUnavailable,
		func(c *gin.Context) { writeProjectError(c, projectSvc.ErrProjectNotFound) },
	} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		send(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), `"code":7`)
		require.Contains(t, w.Body.String(), `"data":{}`)
	}
}
