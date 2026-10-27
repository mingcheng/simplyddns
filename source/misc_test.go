package source

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	ddns "github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAllSupportSourceFunc(t *testing.T) {
	names := []string{"auto", "auto-ipv4", "auto-ipv6", "lo", "static"}
	for _, tc := range httpSourceTests {
		names = append(names, tc.name)
	}
	assert.ElementsMatch(t, names, ddns.GetAllSupportSourceFunc())
}

func TestHTTPSourceFileLayout(t *testing.T) {
	for _, tc := range httpSourceTests {
		t.Run(tc.name, func(t *testing.T) {
			stem := strings.ReplaceAll(tc.name, "-", "_")
			require.FileExists(t, stem+".go")
			require.FileExists(t, stem+"_test.go")
			file, err := parser.ParseFile(token.NewFileSet(), stem+".go", nil, 0)
			require.NoError(t, err)
			registrations := 0
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				fn, ok := call.Fun.(*ast.Ident)
				if ok && (fn.Name == "registerSource" || fn.Name == "registerHTTPSource") {
					registrations++
				}
				return true
			})
			assert.Equal(t, 1, registrations, "each source file must register exactly one source")

			testFile, err := parser.ParseFile(token.NewFileSet(), stem+"_test.go", nil, 0)
			require.NoError(t, err)
			tests := 0
			for _, decl := range testFile.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && strings.HasPrefix(fn.Name.Name, "TestSource") {
					tests++
				}
			}
			assert.Equal(t, 1, tests, "each source must have its own focused test entry")
		})
	}
}
