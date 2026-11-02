package initialize

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// GVA enter.go files aggregate groups and aliases. Responsibility API, service,
// router declarations and support interfaces belong to their implementation files.
func TestEnterFilesOnlyDeclareGroupTypes(t *testing.T) {
	for _, tc := range []struct{ pattern, group string }{
		{"../api/v1/enter.go", "ApiGroup"}, {"../api/v1/*/enter.go", "ApiGroup"},
		{"../service/enter.go", "ServiceGroup"}, {"../service/*/enter.go", "ServiceGroup"},
		{"../router/enter.go", "RouterGroup"}, {"../router/*/enter.go", "RouterGroup"},
	} {
		paths, err := filepath.Glob(tc.pattern)
		require.NoError(t, err)
		require.NotEmpty(t, paths)
		for _, path := range paths {
			t.Run(path, func(t *testing.T) {
				f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
				require.NoError(t, err)
				groups := 0
				for _, decl := range f.Decls {
					gen, ok := decl.(*ast.GenDecl)
					if !ok || gen.Tok != token.TYPE {
						continue
					}
					for _, spec := range gen.Specs {
						typ := spec.(*ast.TypeSpec)
						require.Equal(t, tc.group, typ.Name.Name, "responsibility type must live in its own implementation file")
						require.IsType(t, &ast.StructType{}, typ.Type)
						require.False(t, typ.Assign.IsValid(), "group must be a concrete declaration rather than a compatibility alias")
						groups++
					}
				}
				require.Equal(t, 1, groups, "enter.go must own its group declaration")
			})
		}
	}
}

func TestAPIRequestVariablesUseNamedDTOs(t *testing.T) {
	files, err := filepath.Glob("../api/v1/*/*.go")
	require.NoError(t, err)
	require.NotEmpty(t, files)
	for _, path := range files {
		if filepath.Ext(path) != ".go" {
			continue
		}
		t.Run(path, func(t *testing.T) {
			f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			require.NoError(t, err)
			ast.Inspect(f, func(node ast.Node) bool {
				value, ok := node.(*ast.ValueSpec)
				if !ok {
					return true
				}
				if _, anonymous := value.Type.(*ast.StructType); !anonymous {
					return true
				}
				for _, name := range value.Names {
					require.NotContains(t, []string{"req", "request", "payload"}, name.Name, "HTTP inputs must use named request DTOs")
				}
				return true
			})
		})
	}
}
