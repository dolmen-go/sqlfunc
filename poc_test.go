package sqlfunc_test

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"maps"
	"slices"
	"strconv"
	"strings"
	"testing"
	"text/template"

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

		gen := &Generator{
			Pkg:     pkg,
			Imports: make(map[string]*types.Package),
		}

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
					if err := gen.genForEach(t, ti.TypeOf(arg).(*types.Signature)); err != nil {
						t.Logf("%s %v", pkg.Fset.Position(c.Pos()), err)
					}
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
					gen.gen(t, s.Sel.Name, sig)
				}
				return
			}, nil)
		}

		if len(gen.Funcs) > 0 {
			var buf bytes.Buffer
			fmt.Fprintf(&buf, "package %s\n\n", pkg.Name)

			if len(gen.Imports) > 0 {
				buf.WriteString("import (\n")
				// TODO sort imports for deterministic output
				for _, imp := range gen.Imports {
					fmt.Fprintf(&buf, "\t%s %q\n", imp.Name(), imp.Path())
				}
				buf.WriteString(")\n")
			}

			buf.WriteString("\nfunc init() {")
			for _, f := range gen.Funcs {
				buf.WriteString(f)
			}
			buf.WriteString("}\n")

			t.Log("\n" + buf.String())
		}
	}
}

type Generator struct {
	Pkg     *packages.Package
	Imports map[string]*types.Package

	Funcs []string
}

// The qualifier function is used to determine how to print package-qualified type names in the generated code.
// It also collects the imports needed for the generated code.
// It is used in calls to [types.TypeString].
func (g *Generator) qualifier(other *types.Package) string {
	if other == g.Pkg.Types {
		return "" // Same package, no prefix needed
	}
	if typPkg, seen := g.Imports[other.Path()]; seen {
		return typPkg.Name() // Already recorded import, return its name
	}
	g.Imports[other.Path()] = other
	return other.Name()
}

func (g *Generator) checkTypeScope(typ types.Type) error {
	if _, ok := typ.(*types.TypeParam); ok {
		return fmt.Errorf("%q is a type parameter from an enclosing context", types.TypeString(typ, g.qualifier))
	}

	if named, ok := typ.(*types.Named); ok {
		obj := named.Obj()
		// If the type is defined in the current package but not at the package level
		if obj.Pkg() == g.Pkg.Types && obj.Parent() != g.Pkg.Types.Scope() {
			return fmt.Errorf("%q is a local type", obj.Name())
		}
	}

	// FIXME recurse

	return nil
}

func (g *Generator) gen(tb testing.TB, f string, sig *types.Signature) {
	tb.Log(g.Pkg.Name, f, sig)
}

func (g *Generator) genForEach(tb testing.TB, sig *types.Signature) error {
	tb.Log(g.Pkg.Name, "ForEach", types.TypeString(sig, g.qualifier))

	if sig.Params().Len() == 0 {
		return errors.New("function must receive at least one parameter")
	}

	// FIXME We should check if all types of args are available at the package scope (not types defined locally in a function)
	// and skip because we will not be able to generate code that reference those types

	var withError, withBool bool
	switch sig.Results().Len() {
	case 0:
	case 1:
		if sig.Results().At(0).Type().String() == "error" {
			withError = true
			break
		}
		if sig.Results().At(0).Type().String() == "bool" {
			withBool = true
			break
		}
		fallthrough
	default:
		return errors.New("only one return value allowed of type error")
	}

	params := sig.Params()
	nParams := params.Len()
	vars := make([]string, nParams)
	args := make([]string, nParams)

	for i := range nParams {
		p := params.At(i)
		typ := p.Type()

		if err := g.checkTypeScope(typ); err != nil {
			return fmt.Errorf("parameter %d (type %q): %w", i, types.TypeString(typ, g.qualifier), err)
		}

		name := "v" + strconv.Itoa(i)
		// TODO collect reference to an import in p.Type
		vars[i] = name + " " + types.TypeString(typ, g.qualifier)
		args[i] = name
	}

	sigString := types.TypeString(sig, g.qualifier)
	tb.Log("imports:", slices.Collect(maps.Keys(g.Imports)))

	data := map[string]any{
		"Type":      sigString,
		"WithError": withError,
		"WithBool":  withBool,
		"Vars":      strings.Join(vars, "\n\t\t\t"),
		"Args":      strings.Join(args, ", "),
		"ArgsPtr":   "&" + strings.Join(args, ", &"),
	}

	const code = `` +
		`
	sqlfunc.Ř.ForEach.Register(({{.Type}})(nil), func(rows *sql.Rows, cb interface{}) (err error) {
		cb := cb.({{.Type}})
		defer func() {
			err2 := rows.Close()
			if err == nil {
				err = err2
			}
		}()
		for rows.Next() {
			var (
				{{.Vars}}
			)
			if err = rows.Scan({{.ArgsPtr}}); err != nil {
				return
			}
{{- if .WithError}}
			if err = cb({{.Args}}); err != nil {
				return
			}
{{- else if .WithBool}}
			if !cb({{.Args}}) {
				return
			}
{{- else}}
			cb({{.Args}})
{{- end}}
		}
		err = rows.Err()
		return
	})
`

	tmpl := template.New("code")
	tmpl, err := tmpl.Parse(code)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		panic(err)
	}

	// tb.Log("\n" + buf.String())

	g.Funcs = append(g.Funcs, buf.String())

	return nil
}
