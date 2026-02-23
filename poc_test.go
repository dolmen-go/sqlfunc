package sqlfunc_test

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
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
		t.Log("Package", pkg.Name)
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
					if err := gen.add("ForEach", ti.TypeOf(arg).(*types.Signature), (*Generator).genForEach); err != nil {
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
				printFuncCode(&buf, f)
			}
			buf.WriteString("}\n")

			t.Log("\n" + buf.String())
		}
	}
}

type funcCode interface {
	Registry() string
	// Template returns a text/template that will be executed to generate the code for this function.
	Template() string
}

func printFuncCode(w io.Writer, f interface{ Template() string }) error {
	tmpl := template.New("code")
	tmpl, err := tmpl.Parse(f.Template())
	if err != nil {
		return fmt.Errorf("Parse template: %w", err)
	}

	if err = tmpl.Execute(w, f); err != nil {
		return fmt.Errorf("Execute template: %w", err)
	}

	return nil
}

type Generator struct {
	Pkg     *packages.Package
	Imports map[string]*types.Package

	Funcs map[string]funcCode
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

func (g *Generator) add(registry string, sig *types.Signature, build func(g *Generator, sig *types.Signature) (funcCode, error)) error {
	// FIXME We are leaking imports
	key := registry + " " + types.TypeString(sig, g.qualifier)

	// Skip if we already have a function for this signature
	if _, exists := g.Funcs[key]; exists {
		return nil
	}
	f, err := build(g, sig)
	if err != nil {
		return err
	}
	if f == nil {
		return nil
	}
	if g.Funcs == nil {
		g.Funcs = make(map[string]funcCode)
	}
	g.Funcs[key] = f
	return nil
}

func (g *Generator) genForEach(sig *types.Signature) (funcCode, error) {
	// FIXME We are leaking imports if we skip due to an error
	sigString := types.TypeString(sig, g.qualifier)

	if sig.Params().Len() == 0 {
		return nil, errors.New("function must receive at least one parameter")
	}

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
		return nil, errors.New("only one return value allowed of type error or bool")
	}

	params := sig.Params()
	nParams := params.Len()
	vars := make([]string, nParams)
	args := make([]string, nParams)

	for i := range nParams {
		p := params.At(i)
		typ := p.Type()

		// FIXME TypeScope check failures should not prevent generating code for other signatures
		if err := g.checkTypeScope(typ); err != nil {
			return nil, fmt.Errorf("parameter %d (type %q): %w", i, types.TypeString(typ, g.qualifier), err)
		}

		name := "v" + strconv.Itoa(i)
		vars[i] = name + " " + types.TypeString(typ, g.qualifier)
		args[i] = name
	}

	code := funcCodeForEach{
		Signature: sigString,
		WithError: withError,
		WithBool:  withBool,
		Vars:      strings.Join(vars, "\n\t\t\t"),
		Args:      strings.Join(args, ", "),
		ArgsPtr:   "&" + strings.Join(args, ", &"),
	}

	return &code, nil
}

type funcCodeForEach struct {
	Signature string
	WithError bool
	WithBool  bool
	Vars      string
	Args      string
	ArgsPtr   string
}

func (funcCodeForEach) Registry() string {
	return "ForEach"
}

func (f funcCodeForEach) Key() string {
	return f.Registry() + " " + f.Signature
}

func (funcCodeForEach) Template() string {
	return `
	sqlfunc.Ř.ForEach.Register(({{.Signature}})(nil), func(rows *sql.Rows, cb any) (err error) {
		cb := cb.({{.Signature}})
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
}
