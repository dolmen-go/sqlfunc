package sqlfunc_test

import (
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

func TestScanSrc(t *testing.T) {
	// Helpful article: https://blog.afoolishmanifesto.com/posts/writing-a-golang-linter/

	cfg := &packages.Config{
		Mode:  packages.NeedDeps | packages.NeedImports | packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Tests: true,
	}

	pkgs, err := packages.Load(cfg, "pattern=.")
	// pkgs, err := packages.Load(cfg, "file=./stmt_test.go")
	if err != nil {
		t.Fatalf("load: %s", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return
	}

	// Lint each package we find.
	for _, pkg := range pkgs {
		t.Log(pkg.Name)
		ti := pkg.TypesInfo

		// Each of these is a parsed file.
		for _, f := range pkg.Syntax {

			// Here's where we walk over the syntax tree.  We can
			// return false to stop walking early.  The code could
			// probably be faster by carefully stopping the walk
			// early, but I decided that probably wasn't worth the
			// effort.
			astutil.Apply(f, func(cur *astutil.Cursor) (deeper bool) {
				deeper = true

				c, ok := cur.Node().(*ast.CallExpr)
				// if it's not a call, bail out.
				if !ok {
					return
				}
				// verify that the function being called is a
				// selector.  A selector in Go looks like
				// `foo.bar`.  Read more here:
				// https://golang.org/ref/spec#Selectors
				s, ok := c.Fun.(*ast.SelectorExpr) // possibly method calls
				if !ok {
					return
				}

				// package functions
				nv, isSelector := ti.Selections[s]
				if isSelector {
					return
				}
				pkgName := ti.Uses[s.X.(*ast.Ident)].(*types.PkgName)
				path := pkgName.Imported().Path()

				if path != "github.com/dolmen-go/sqlfunc" {
					return
				}

				t.Logf("%s %s.%s",
					pkg.Fset.Position(c.Pos()),
					path,
					s.Sel.Name)
				// t.Logf("%+v", c)

				// Look at the last parameter
				arg := c.Args[len(c.Args)-1]

				switch s.Sel.Name {
				case "ForEach":
					// Function expected:
					// - literal
					// - identifier
					return false
				default:
					//
					fnPtrArg, ok := arg.(*ast.UnaryExpr)
					if !ok || fnPtrArg.Op != token.AND {
						t.Logf("%s %s.%s SKIP (arg %d is not a pointer)",
							pkg.Fset.Position(c.Pos()),
							path,
							s.Sel.Name,
							len(c.Args)-1,
						)
						return false // Do not go deeper
					}
					ident := fnPtrArg.X.(*ast.Ident)
					if ident.Obj.Kind != ast.Var {
						t.Logf("%s %s.%s SKIP (arg %d is not the address (&) of a variable)",
							pkg.Fset.Position(c.Pos()),
							path,
							s.Sel.Name,
							len(c.Args)-1,
						)
						return false
					}
					// t.Log(astutil.NodeDescription(ident.Obj))
					t.Logf("%#v", ident)
					t.Logf("%#v", ident.Obj)
					t.Logf("%#v", ident.Obj.Decl.(*ast.ValueSpec))
					identType := ident.Obj.Decl.(*ast.ValueSpec).Type
					funcType, isFuncType := identType.(*ast.FuncType)
					if !isFuncType {
						t.Logf("%s %s.%s SKIP (%s is not function)",
							pkg.Fset.Position(c.Pos()),
							path,
							s.Sel.Name,
							ident.Name,
						)
					}
					t.Logf("%s %#v", ident.Name, funcType)
					var sb strings.Builder
					printer.Fprint(&sb, pkg.Fset, funcType)
					t.Logf("%s %s", ident.Name, sb.String())
				}

				_ = nv
				return false // Do not go deeper
			}, nil)
		}
	}
}
