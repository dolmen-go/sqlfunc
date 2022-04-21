package sqlfunc_test

import (
	"go/ast"
	"go/token"
	"go/types"
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
				if _, isSelector := ti.Selections[s]; isSelector {
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
					// - identifier pointing to a func type
					genForEach(t, pkg, ti.TypeOf(arg).(*types.Signature))
					// As the argument might be a func literal, we want to go deeper in the AST
					return true
				default:
					deeper = false // Skip processing the arguments (just for speed)
					//
					fnPtrArg, ok := arg.(*ast.UnaryExpr)
					if !ok || fnPtrArg.Op != token.AND {
						t.Logf("%s %s.%s SKIP (arg %d is not a pointer)",
							pkg.Fset.Position(c.Pos()),
							path,
							s.Sel.Name,
							len(c.Args)-1,
						)
						return
					}
					ident := fnPtrArg.X.(*ast.Ident)
					if ident.Obj.Kind != ast.Var {
						t.Logf("%s %s.%s SKIP (arg %d is not the address (&) of a variable)",
							pkg.Fset.Position(c.Pos()),
							path,
							s.Sel.Name,
							len(c.Args)-1,
						)
						return
					}
					typ := ti.ObjectOf(ident).Type()
					sig, isSignature := typ.(*types.Signature)
					if !isSignature {
						t.Logf("%s %s.%s SKIP (%s is not function variable)",
							pkg.Fset.Position(c.Pos()),
							path,
							s.Sel.Name,
							ident.Name,
						)
					}
					// t.Logf("%#v", typ)
					gen(t, pkg, s.Sel.Name, sig)
				}
				return
			}, nil)
		}
	}
}

func gen(tb testing.TB, pkg *packages.Package, f string, sig *types.Signature) {
	tb.Log(pkg.Name, f, sig)
}

func genForEach(tb testing.TB, pkg *packages.Package, sig *types.Signature) {
	tb.Log(pkg.Name, "ForEach", sig)
}
