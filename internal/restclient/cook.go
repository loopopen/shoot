package restclient

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/lopolopen/shoot/internal/tools/logx"
)

func (g *Generator) cookClient(typeName string) {
	g.data.SigMap = make(map[string]string)
	g.data.HTTPMethodMap = make(map[string]string)
	g.data.PathMap = make(map[string]string)
	g.data.AliasMap = make(map[string]map[string]string)
	g.data.PathParamsMap = make(map[string][]string)
	g.data.QueryParamsMap = make(map[string][]string)
	g.data.IsParamPtrMap = make(map[string]map[string]bool)
	g.data.BodyParamMap = make(map[string]string)
	g.data.QueryDictMap = make(map[string]string)
	g.data.ReturnResultMap = make(map[string]struct {
		Type  string
		IsPtr bool
	})
	g.data.ErrReturnMap = make(map[string]string)
	g.data.CtxParamMap = make(map[string]string)
	g.data.DefaultHeaders = map[string]map[string]string{
		http.MethodGet: {
			"Accept": "application/json",
		},
		http.MethodPost: {
			"Accept":       "application/json",
			"Content-Type": "application/json",
		},
		http.MethodPut: {
			"Accept":       "application/json",
			"Content-Type": "application/json",
		},
		http.MethodPatch: {
			"Accept":       "application/json",
			"Content-Type": "application/json",
		},
		http.MethodDelete: {},
	}
	g.data.BodyHTTPMethods = []string{http.MethodPost, http.MethodPut, http.MethodPatch}

	for _, f := range g.Pkg().Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if !g.testNode(typeName, n) {
				return true
			}

			ts, _ := n.(*ast.TypeSpec)
			iface, _ := ts.Type.(*ast.InterfaceType)
			for _, field := range iface.Methods.List {
				//shoot.RestClient[Client]
				//todo:
				if len(field.Names) == 0 {
					if field.Doc != nil {
						headers := parseHeaders(field.Doc.Text())
						for k, v := range headers {
							for _, headers := range g.data.DefaultHeaders {
								headers[k] = v
							}
						}
					}
				} else {
					ftype, ok := field.Type.(*ast.FuncType)
					if !ok {
						continue
					}

					doc := field.Doc.Text()
					methodName := field.Names[0].Name
					g.data.SigMap[methodName] = methodSignature(field) //full signature

					if field.Doc == nil {
						logx.Warnf("method %s without comments will be ignored", methodName)
						continue
					}

					httpMethod, path, pathParams, ok := parsePath(doc) //pathParams ~ [id]
					if !ok {
						logx.Warnf("method %s with bad comments will be ignored", methodName)
						continue
					}

					g.data.MethodList = append(g.data.MethodList, methodName)

					g.data.HTTPMethodMap[methodName] = httpMethod //http mehod
					g.data.PathMap[methodName] = path             //http path

					asMap := parseAlias(doc)             //userID -> id
					reversMap := make(map[string]string) //id -> userID
					for k, v := range asMap {
						reversMap[v] = k
					}
					g.data.AliasMap[methodName] = asMap

					var realPathParams []string
					for _, name := range pathParams {
						if real, ok := reversMap[name]; ok {
							realPathParams = append(realPathParams, real) //fix path param
						} else {
							realPathParams = append(realPathParams, name)
						}
					}
					g.data.PathParamsMap[methodName] = realPathParams

					//------------Params---------------
					if g.data.IsParamPtrMap[methodName] == nil {
						g.data.IsParamPtrMap[methodName] = make(map[string]bool)
					}
					if ftype.Params != nil {
						for _, param := range ftype.Params.List {
							for _, name := range param.Names {
								g.handleExpr(param.Type, name, f, methodName, httpMethod)
								if _, ok := param.Type.(*ast.StarExpr); ok {
									g.data.IsParamPtrMap[methodName][name.Name] = true
								}
							}
						}
					}

					//------------Results---------------
					n := 0
					if ftype.Results != nil {
						n = len(ftype.Results.List)
					}
					if n < 2 {
						logx.Fatalf("method %s should at least return response and error", methodName)
					}
					if n > 3 {
						logx.Fatalf("method %s must not return more than three values", methodName)
					}

					second2last := exprToString(ftype.Results.List[n-2].Type)
					if second2last != "*http.Response" {
						logx.Fatalf("the second to last return value of method %s must be a http response pointer", methodName)
					}
					last := exprToString(ftype.Results.List[n-1].Type)
					if last != "error" {
						logx.Fatalf("the last return value of method %s must be an error", methodName)
					}

					if n == 3 {
						r := ftype.Results.List[0]
						if len(r.Names) > 0 {
							logx.Fatalf("method %s with named return list is not supported", methodName)
						}

						name, isPtr := getReturnTypeName(r.Type)
						g.data.ReturnResultMap[methodName] = struct {
							Type  string
							IsPtr bool
						}{
							Type:  name,
							IsPtr: isPtr,
						}
					}
					g.data.ErrReturnMap[methodName] = strings.Repeat("nil, ", n-1) + "err"

					//todo: check alias={a:alias}, a exists?
				}
			}
			return false
		})
	}
}

func exprToString(expr ast.Expr) string {
	var buf bytes.Buffer
	err := printer.Fprint(&buf, token.NewFileSet(), expr)
	if err != nil {
		logx.Fatalf("print expr: %s", err)
	}
	return buf.String()
}

func methodSignature(method *ast.Field) string {
	funcName := method.Names[0].Name
	funcType, ok := method.Type.(*ast.FuncType)
	if !ok {
		return ""
	}

	params := formatFieldList(funcType.Params)
	results := formatFieldList(funcType.Results)

	if results != "" {
		return fmt.Sprintf("%s(%s) (%s)", funcName, params, results)
	}
	return fmt.Sprintf("%s(%s)", funcName, params)
}

func formatFieldList(fl *ast.FieldList) string {
	if fl == nil || len(fl.List) == 0 {
		return ""
	}
	var parts []string
	for _, f := range fl.List {
		typeStr := exprToString(f.Type)
		if len(f.Names) == 0 {
			parts = append(parts, typeStr)
		} else {
			for _, name := range f.Names {
				parts = append(parts, fmt.Sprintf("%s %s", name.Name, typeStr))
			}
		}
	}
	return strings.Join(parts, ", ")
}

func parseHeaders(doc string) map[string]string {
	headers := make(map[string]string)
	regHeaders := regexp.MustCompile(`shoot:.*?\Wheaders=((?:\s*{[^\n]+},?)+)`)
	ms := regHeaders.FindStringSubmatch(doc)
	if len(ms) > 0 {
		kvMap := parseKV(ms[1])
		for k, v := range kvMap {
			headers[k] = v
		}
	}
	return headers
}

func parseKV(str string) map[string]string {
	regKV := regexp.MustCompile(`{([\w|-]+)\W*:\W*([^}]+)}`)
	kvLst := regKV.FindAllStringSubmatch(str, -1)
	if len(kvLst) == 0 {
		return nil
	}
	kvMap := make(map[string]string)
	for _, kv := range kvLst {
		kvMap[kv[1]] = kv[2]
	}
	return kvMap
}

func parsePath(doc string) (string, string, []string, bool) {
	regReq := regexp.MustCompile(`(?im)^shoot:\W+(get|post|put|patch|delete)\((.*)\)\W*;?\W*$`)
	ms := regReq.FindStringSubmatch(doc)
	if len(ms) == 0 {
		return "", "", nil, false
	}
	method := strings.ToUpper(ms[1])
	path := strings.TrimSpace(ms[2])

	regPath := regexp.MustCompile(`^("[^"]+"|[^"]+)$`)
	if !regPath.MatchString(path) {
		logx.Fatalf("bad path format: %s", path)
	}

	path = strings.Trim(path, `"`)

	regPathParam := regexp.MustCompile(`{(\w+)}`)
	psLst := regPathParam.FindAllStringSubmatch(path, -1)
	var pathParams []string
	for _, ps := range psLst {
		pathParams = append(pathParams, ps[1])
	}
	return method, path, pathParams, true
}

func parseAlias(doc string) map[string]string {
	regAlias := regexp.MustCompile(`(?m)^shoot:.*?\Walias=([^;\n]+)(;.*|\s*)$`)
	ms := regAlias.FindStringSubmatch(doc)
	if len(ms) == 0 {
		return nil
	}
	aliasLst := ms[1] //{userID:id},...
	kvMap := parseKV(aliasLst)
	return kvMap
}

// func getUnderlyingTypeName(expr ast.Expr) string {
// 	switch t := expr.(type) {
// 	case *ast.Ident:
// 		return t.Name
// 	case *ast.StarExpr:
// 		// 指针类型，递归获取底层类型
// 		return getUnderlyingTypeName(t.X)
// 	case *ast.SelectorExpr:
// 		// 处理像 pkg.Type 这样的类型
// 		return getUnderlyingTypeName(t.X) + "." + t.Sel.Name
// 	case *ast.ArrayType:
// 		// 处理数组或切片类型
// 		return "[]" + getUnderlyingTypeName(t.Elt)
// 	default:
// 		return fmt.Sprintf("%T", expr)
// 	}
// }

func getReturnTypeName(expr ast.Expr) (string, bool) {
	isPtr := false
	var typeName string
	switch t := expr.(type) {
	case *ast.StarExpr:
		typeName = exprToString(t.X)
		isPtr = true
	case *ast.ArrayType:
		typeName = exprToString(t)
	case *ast.MapType:
		typeName = exprToString(t)
	default:
		logx.Fatalf("unsupported return type: %T", expr)
	}
	return typeName, isPtr
}

func isStructType(name string, file *ast.File) bool {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != name {
				continue
			}
			_, isStruct := typeSpec.Type.(*ast.StructType)
			return isStruct
		}
	}
	return false
}

type fieldInfo struct {
	Name       string
	Type       string
	Tag        string
	Alias      string
	IsExported bool
	IsPtr      bool
}

func extractStructFields(pkgPath, typeName string) ([]fieldInfo, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkgPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package: %w", err)
	}

	var fields []fieldInfo
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				if genDecl.Tok != token.TYPE {
					continue
				}
				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if typeSpec.Name.Name != typeName {
						continue
					}
					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}
					for _, field := range structType.Fields.List {
						var alias string
						var rawTag string
						if field.Tag != nil {
							rawTag = field.Tag.Value
							alias = parseFieldAlias(rawTag)
						}
						_, ok := field.Type.(*ast.StarExpr)
						for _, name := range field.Names {
							fields = append(fields, fieldInfo{
								Name:       name.Name,
								Type:       exprToString(field.Type),
								Tag:        rawTag,
								Alias:      alias,
								IsExported: name.IsExported(),
								IsPtr:      ok,
							})
						}
						// Handle anonymous fields (embedded structs)
						if len(field.Names) == 0 {
							fields = append(fields, fieldInfo{
								Name:       exprToString(field.Type),
								Type:       exprToString(field.Type),
								Tag:        rawTag,
								Alias:      alias,
								IsExported: true,
							})
						}
					}
				}
			}
		}
	}
	return fields, nil
}

func getPkgDir(importPath string) (string, error) {
	pkg, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		return "", err
	}
	return pkg.Dir, nil
}

func parseFieldAlias(tag string) string {
	t := reflect.StructTag(strings.Trim(tag, "`"))
	aliasReg := regexp.MustCompile(`alias=(\w+)`)
	ms := aliasReg.FindStringSubmatch(t.Get("shoot"))
	if len(ms) == 0 {
		return ""
	}
	return ms[1]
}
