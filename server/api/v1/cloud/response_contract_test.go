package cloud

import (
	"net/http"
	"net/http/httptest"
	"testing"

	cloudService "optimus/server/service/cloud"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCloudErrorsUseGVAResponse(t *testing.T) {
	for _, send := range []func(*gin.Context){
		func(c *gin.Context) { writeCloudServiceUnavailable(c, "unavailable") },
		func(c *gin.Context) { writeCloudError(c, cloudService.ErrAccountNotFound, "unknown") },
	} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		send(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), `"code":7`)
	}
}
