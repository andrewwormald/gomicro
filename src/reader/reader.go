package reader

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
)

func ReadAPI(packageName string, path string) ([]FunctionSignature, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", string(b), parser.ParseComments)
	if err != nil {
		return nil, err
	}

	r := &Reader{PackageName: packageName}

	var fs []FunctionSignature
	//var localStructs []string
	// Traverse file tree for interface methods
	ast.Inspect(node, func(n ast.Node) bool {
		switch t := n.(type) {
		// Find variable declarations
		case *ast.TypeSpec:
			// Which are public
			if t.Name.IsExported() {
				switch assertion := t.Type.(type) {
				// Which are interfaces
				case *ast.InterfaceType:
					fs = r.ListInterfaceMethods(assertion)
				}
			}
		}
		return true
	})

	return fs, nil
}

type FunctionSignature struct {
	Name string
	Params []Variable
	Results []Variable
}

type Reader struct {
	PackageName string
}

// ListInterfaceMethods returns the function signatures of all the interface methods
func (r *Reader) ListInterfaceMethods(it *ast.InterfaceType) []FunctionSignature {
	var fsSlice []FunctionSignature

	methods := it.Methods.List
	for _, method := range methods {
		fn, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		sig := r.CheckFunctionSignature(fn)
		sig.Name = method.Names[0].Name
		fsSlice = append(fsSlice, sig)
	}

	return fsSlice
}

func (r *Reader) CheckFunctionSignature(fn *ast.FuncType) FunctionSignature {
	var fs FunctionSignature

	for _, param := range fn.Params.List {
		fs.Params = append(fs.Params, r.listVariables(param)...)
	}

	for _, result := range fn.Results.List {
		fs.Results = append(fs.Results, r.listVariables(result)...)
	}

	return fs
}

type Variable struct {
	Name string
	ImportType string
}


func (r *Reader) listVariables(field *ast.Field) []Variable {
	// Unnamed variables
	if len(field.Names) == 0 {
		return []Variable{
			{
				Name:       "",
				ImportType: r.importTypeFromASTExpr(field.Type),
			},
		}
	}

	var vr []Variable
	for _, variable := range field.Names {
		vr = append(vr, Variable{
			Name:       variable.Name,
			ImportType: r.importTypeFromASTExpr(field.Type),
		})
	}

	return vr
}

// importTypeFromASTExpr returns an empty string if it cannot find the imported type or if the type is not imported
func (r *Reader) importTypeFromASTExpr(expr ast.Expr) string {
	switch s := expr.(type) {
	case *ast.Ident: // internal type handling
		if !isBuiltInType(s.Name) {
			return r.PackageName + "." + importTypeFromASTIdent(s)
		}
		return importTypeFromASTIdent(s)
	case *ast.SelectorExpr: // external package type handling
		// convert both package and type abstractions to *ast.Ident types
		p1, _ := s.X.(*ast.Ident)
		p2 := s.Sel

		// return the combined Go syntax for imported types
		return strings.Join([]string{importTypeFromASTIdent(p1), importTypeFromASTIdent(p2)}, ".")
	default:
		return ""
	}
}

// importTypeFromASTIdent collects the type name and can be used for non-import types in the file tree
func importTypeFromASTIdent(ident *ast.Ident) string {
	return ident.Name
}

func isBuiltInType(typ string) bool {
	switch typ {
	case "bool", "byte", "complex128", "complex64", "error":
	case "float32", "float64":
	case "int", "int16", "int32", "int64", "int8":
	case "rune", "string":
	case "uint", "uint16", "uint32", "uint64", "uint8", "uintptr":
	default:
		return false
	}
	return true
}