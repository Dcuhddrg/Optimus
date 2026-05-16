package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakeProjectScopeResolver struct {
	accessible map[uint][]uint
	codes      map[string]uint
	roles      map[uint]map[uint]string
}

func (f *fakeProjectScopeResolver) ListAccessibleProjectIDs(_ context.Context, userID uint) ([]uint, error) {
	return f.accessible[userID], nil
}

func (f *fakeProjectScopeResolver) ResolveProjectByCode(_ context.Context, code string) (uint, bool, error) {
	id, ok := f.codes[code]
	return id, ok, nil
}

func (f *fakeProjectScopeResolver) GetMemberRole(_ context.Context, userID, projectID uint) (string, bool, error) {
	role, ok := f.roles[userID][projectID]
	return role, ok, nil
}

type erroringProjectScopeResolver struct{}

func (erroringProjectScopeResolver) ListAccessibleProjectIDs(context.Context, uint) ([]uint, error) {
	return nil, errors.New("boom")
}

func (erroringProjectScopeResolver) ResolveProjectByCode(context.Context, string) (uint, bool, error) {
	return 0, false, nil
}

func (erroringProjectScopeResolver) GetMemberRole(context.Context, uint, uint) (string, bool, error) {
	return "", false, nil
}

func setupProjectScopeRouter(resolver ProjectScopeResolver, withClaims bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if withClaims {
			c.Set(utils.ClaimsContextKey, &utils.CustomClaims{BaseClaims: utils.BaseClaims{ID: 42}})
		}
		c.Next()
	})
	r.Use(ProjectScope(resolver))
	r.GET("/", func(c *gin.Context) {
		scope, _ := c.Get(utils.ProjectScopeContextKey)
		c.JSON(http.StatusOK, scope)
	})
	return r
}

func TestProjectScopeNoClaimsPassesThrough(t *testing.T) {
	r := setupProjectScopeRouter(&fakeProjectScopeResolver{}, false)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestProjectScopeNoHeaderOnlySetsAccessibleProjects(t *testing.T) {
	r := setupProjectScopeRouter(&fakeProjectScopeResolver{
		accessible: map[uint][]uint{42: []uint{7, 8}},
	}, true)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "\"AccessibleProjectIDs\":[7,8]")
}

func TestProjectScopeValidHeaderSetsCurrentProject(t *testing.T) {
	resolver := &fakeProjectScopeResolver{
		accessible: map[uint][]uint{42: []uint{7}},
		codes:      map[string]uint{"prod": 7},
		roles:      map[uint]map[uint]string{42: map[uint]string{7: "owner"}},
	}
	r := setupProjectScopeRouter(resolver, true)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Project-Code", "prod")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "\"CurrentProjectID\":7")
	require.Contains(t, rec.Body.String(), "\"CurrentProjectCode\":\"prod\"")
	require.Contains(t, rec.Body.String(), "\"CurrentProjectRole\":\"owner\"")
}

func TestProjectScopeUnknownCodeReturns404(t *testing.T) {
	r := setupProjectScopeRouter(&fakeProjectScopeResolver{
		accessible: map[uint][]uint{42: []uint{7}},
	}, true)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Project-Code", "missing")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestProjectScopeInaccessibleCodeReturns403(t *testing.T) {
	resolver := &fakeProjectScopeResolver{
		accessible: map[uint][]uint{42: []uint{8}},
		codes:      map[string]uint{"prod": 7},
		roles:      map[uint]map[uint]string{42: map[uint]string{7: "owner"}},
	}
	r := setupProjectScopeRouter(resolver, true)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Project-Code", "prod")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestProjectScopeResolverErrorReturns500(t *testing.T) {
	r := setupProjectScopeRouter(erroringProjectScopeResolver{}, true)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
