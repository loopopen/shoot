package restclient

import (
	"fmt"
	"go/ast"
	"go/types"
	"net/http"

	"github.com/lopolopen/shoot/internal/shoot"
	"github.com/lopolopen/shoot/internal/tools/logx"
	"github.com/lopolopen/shoot/internal/transfer"
)

func (g *Generator) handleExpr(paramType ast.Expr, name *ast.Ident, file *ast.File, methodName, httpMethod string) {
	switch t := paramType.(type) {
	case *ast.SelectorExpr:
		// fmt.Println("::", "SelectorExpr")
		g.handleSelectorExpr(t, name, methodName)
	case *ast.Ident:
		// fmt.Println("::", "Ident")
		g.handleIdent(t, name, file, methodName)
	case *ast.MapType:
		// fmt.Println("::", "MapType")
		g.handleMapType(name, methodName, httpMethod)
	case *ast.StarExpr:
		// fmt.Println("::", "StarExpr")
		g.handleExpr(t.X, name, file, methodName, httpMethod)
	default:
		logx.Fatalf("unsupported param type %T of method %s", t, methodName)
	}
}

func (g *Generator) handleSelectorExpr(paramType *ast.SelectorExpr, name *ast.Ident, methodName string) {
	typ := g.Pkg().TypesInfo.Types[paramType].Type
	named, ok := typ.(*types.Named)
	if !ok {
		return
	}
	obj := named.Obj()
	pkgPath := obj.Pkg().Path()
	if pkgPath == "context" && obj.Name() == "Context" {
		g.data.CtxParamMap[methodName] = name.Name
	} else {
		g.setBodyParamName(methodName, name.Name)
		g.handleStruct(paramType, paramType.Sel.Name, name, methodName)
	}
}

func (g *Generator) handleStruct(paramType ast.Expr, paramTypeName string, name *ast.Ident, methodName string) {
	typ := g.Pkg().TypesInfo.Types[paramType].Type
	named, ok := typ.(*types.Named)
	if !ok {
		return
	}

	if g.data.IsParamPtrMap[methodName] == nil {
		g.data.IsParamPtrMap[methodName] = make(map[string]bool)
	}
	if g.data.AliasMap[methodName] == nil {
		g.data.AliasMap[methodName] = make(map[string]string)
	}

	obj := named.Obj()
	pkgPath := obj.Pkg().Path()
	fullPath, err := getPkgDir(pkgPath)
	if err != nil {
		logx.Fatalf("get pkg dir: %s", err)
	}
	fields, err := extractStructFields(fullPath, paramTypeName)
	if err != nil {
		logx.Fatalf("extract struct fields: %s", err)
	}
	for _, f := range fields {
		var key, value string
		if f.IsExported {
			key = transfer.ToCamelCase(f.Name)
			value = fmt.Sprintf("%s.%s", name.Name, f.Name)
		} else {
			key = f.Name
			value = fmt.Sprintf("%s.%s()", name.Name, transfer.ToPascalCase(f.Name))
		}
		if f.IsPtr {
			g.data.IsParamPtrMap[methodName][value] = true
		}
		g.data.QueryParamsMap[methodName] = append(g.data.QueryParamsMap[methodName], value)

		if f.Alias != "" {
			g.data.AliasMap[methodName][value] = f.Alias
		} else {
			g.data.AliasMap[methodName][value] = key
		}
	}
}

func (g *Generator) handleIdent(paramType *ast.Ident, name *ast.Ident, file *ast.File, methodName string) {
	if isStructType(paramType.Name, file) {
		g.setBodyParamName(methodName, name.Name)
		g.handleStruct(paramType, paramType.Name, name, methodName)
	} else {
		if shoot.Contains(g.data.PathParamsMap[methodName], name.Name) {
			return
		}
		g.data.QueryParamsMap[methodName] = append(g.data.QueryParamsMap[methodName], name.Name) //basic type
	}
}

func (g *Generator) handleMapType(name *ast.Ident, methodName string, httpMethod string) {
	if httpMethod == http.MethodGet || httpMethod == http.MethodDelete {
		g.data.QueryDictMap[methodName] = name.Name
	} else {
		//todo: error
	}
}

func (g *Generator) setBodyParamName(methodName, paramName string) {
	if _, ok := g.data.BodyParamMap[methodName]; ok {
		logx.Fatalf("ambiguous body binding of method %s", methodName)
	}
	g.data.BodyParamMap[methodName] = paramName
}
