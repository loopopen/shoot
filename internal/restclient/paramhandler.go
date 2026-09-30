package restclient

import (
	"fmt"
	"go/ast"
	"go/types"

	"github.com/loopopen/shoot/internal/shoot"
	"github.com/loopopen/shoot/internal/transfer"
)

func (g *Generator) handleExpr(paramType ast.Expr, name *ast.Ident, methodName, httpMethod string) error {
	if g.isContextParam(paramType) {
		g.data.CtxParamMap[methodName] = name.Name
		return nil
	}
	if shoot.Contains(g.data.PathParamsMap[methodName], name.Name) {
		return nil
	}
	if shoot.Contains(g.data.BodyHTTPMethods, httpMethod) {
		return g.setBodyParamName(methodName, name.Name)
	}

	switch t := paramType.(type) {
	case *ast.SelectorExpr:
		return g.handleSelectorExpr(t, name, methodName)
	case *ast.Ident:
		return g.handleIdent(t, name, methodName)
	case *ast.MapType:
		return g.handleMapType(t, name, methodName)
	case *ast.StarExpr:
		return g.handleExpr(t.X, name, methodName, httpMethod)
	default:
		return fmt.Errorf("unsupported parameter %q with type %T in method %s", name.Name, t, methodName)
	}
}

func (g *Generator) handleSelectorExpr(paramType *ast.SelectorExpr, name *ast.Ident, methodName string) error {
	typ := g.Pkg().TypesInfo.Types[paramType].Type
	switch underlying := typ.Underlying().(type) {
	case *types.Struct:
		st := underlying
		g.handleStruct(st, name, methodName)
		return nil
	case *types.Map:
		return g.handleQueryMapType(underlying, name, methodName)
	}
	g.data.QueryParamsMap[methodName] = append(g.data.QueryParamsMap[methodName], name.Name)
	return nil
}

func extractFieldsFromTypes(st *types.Struct) []fieldInfo {
	var fields []fieldInfo

	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		tag := st.Tag(i)

		_, isPtr := f.Type().(*types.Pointer)

		fields = append(fields, fieldInfo{
			Name:       f.Name(),
			Type:       f.Type().String(),
			Tag:        tag,
			Alias:      parseFieldAlias(tag),
			IsExported: f.Exported(),
			IsPtr:      isPtr,
		})
	}
	return fields
}

func (g *Generator) handleStruct(st *types.Struct, name *ast.Ident, methodName string) {
	if g.data.IsParamPtrMap[methodName] == nil {
		g.data.IsParamPtrMap[methodName] = make(map[string]bool)
	}
	if g.data.AliasMap[methodName] == nil {
		g.data.AliasMap[methodName] = make(map[string]string)
	}

	fields := extractFieldsFromTypes(st)

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

func (g *Generator) handleIdent(paramType *ast.Ident, name *ast.Ident, methodName string) error {
	typ := g.Pkg().TypesInfo.Types[paramType].Type
	switch underlying := typ.Underlying().(type) {
	case *types.Struct:
		st := underlying
		g.handleStruct(st, name, methodName)
	case *types.Map:
		return g.handleQueryMapType(underlying, name, methodName)
	default:
		g.data.QueryParamsMap[methodName] = append(g.data.QueryParamsMap[methodName], name.Name) //basic type
	}
	return nil
}

func (g *Generator) handleMapType(paramType *ast.MapType, name *ast.Ident, methodName string) error {
	typ, ok := g.Pkg().TypesInfo.TypeOf(paramType).Underlying().(*types.Map)
	if !ok {
		return fmt.Errorf("parameter %q of method %s is not a map", name.Name, methodName)
	}
	return g.handleQueryMapType(typ, name, methodName)
}

func (g *Generator) handleQueryMapType(typ *types.Map, name *ast.Ident, methodName string) error {
	key, ok := typ.Key().Underlying().(*types.Basic)
	if !ok || key.Kind() != types.String {
		return fmt.Errorf("query map parameter %q of method %s must have string keys", name.Name, methodName)
	}
	if previous, ok := g.data.QueryDictMap[methodName]; ok {
		return fmt.Errorf("method %s has multiple query map parameters %q and %q", methodName, previous, name.Name)
	}
	g.data.QueryDictMap[methodName] = name.Name
	return nil
}

func (g *Generator) setBodyParamName(methodName, paramName string) error {
	if previous, ok := g.data.BodyParamMap[methodName]; ok {
		return fmt.Errorf("method %s has ambiguous body parameters %q and %q; POST, PUT, and PATCH methods accept at most one non-path parameter", methodName, previous, paramName)
	}
	g.data.BodyParamMap[methodName] = paramName
	return nil
}

func (g *Generator) isContextParam(expr ast.Expr) bool {
	typ := g.Pkg().TypesInfo.TypeOf(expr)
	if typ == nil {
		return false
	}
	named, ok := types.Unalias(typ).(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return false
	}
	return named.Obj().Pkg().Path() == "context" && named.Obj().Name() == "Context"
}
