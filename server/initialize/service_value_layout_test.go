package initialize

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceGroupsEmbedConcreteValues(t *testing.T) {
	cases := []struct {
		path  string
		names []string
	}{
		{"../service/system/enter.go", []string{"AuthService", "UserService", "RoleService", "LDAPService", "AuditService", "JwtService", "CasbinService", "Service"}},
		{"../service/project/enter.go", []string{"ProjectService"}},
		{"../service/cloud/enter.go", []string{"CloudAccountService", "AssetRegistryService"}},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), tc.path, nil, 0)
			require.NoError(t, err)
			found := map[string]bool{}
			ast.Inspect(file, func(n ast.Node) bool {
				spec, ok := n.(*ast.TypeSpec)
				if !ok || spec.Name.Name != "ServiceGroup" {
					return true
				}
				group, ok := spec.Type.(*ast.StructType)
				require.True(t, ok)
				for _, field := range group.Fields.List {
					typ, ok := field.Type.(*ast.Ident)
					if !ok || len(field.Names) != 0 {
						continue
					}
					found[typ.Name] = true
				}
				return false
			})
			for _, name := range tc.names {
				require.True(t, found[name], "%s must be an embedded concrete value", name)
			}
		})
	}
}

func TestAPIAliasesCopyServiceValues(t *testing.T) {
	for _, path := range []string{"../api/v1/system/enter.go", "../api/v1/project/enter.go", "../api/v1/cloud/enter.go"} {
		t.Run(path, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			require.NoError(t, err)
			found := 0
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.VAR {
					continue
				}
				for _, spec := range gen.Specs {
					values := spec.(*ast.ValueSpec)
					for _, value := range values.Values {
						require.IsType(t, &ast.SelectorExpr{}, value, "API service alias must copy the grouped value")
						found++
					}
				}
			}
			require.Positive(t, found)
		})
	}
}

func TestRouterAliasesCopyAPIValues(t *testing.T) {
	for _, path := range []string{"../router/system/enter.go", "../router/project/enter.go", "../router/cloud/enter.go"} {
		t.Run(path, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			require.NoError(t, err)
			found := 0
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.VAR {
					continue
				}
				for _, spec := range gen.Specs {
					for _, value := range spec.(*ast.ValueSpec).Values {
						require.IsType(t, &ast.SelectorExpr{}, value, "router API alias must copy the grouped value")
						found++
					}
				}
			}
			require.Positive(t, found)
		})
	}
}
