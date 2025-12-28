package constructor

import (
	"go/ast"
	"regexp"
	"strings"

	"github.com/lopolopen/shoot/internal/transfer"
)

func (g *Generator) makeJson(typeName string) {
	g.RegisterTransfer("jsontagof", transfer.ID)

	if !g.flags.json {
		return
	}

	trans := transfer.ID
	switch g.flags.tagcase {
	case "pascal":
		trans = transfer.ToPascalCase
	case "camel":
		trans = transfer.ToCamelCase
	case "lower":
		trans = strings.ToLower
	case "upper":
		trans = strings.ToUpper
	}

	var exportedList []string
	tagMap := make(map[string]string)

	for _, f := range g.Pkg().Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if !g.testNode(typeName, n) {
				return true
			}

			ts, _ := n.(*ast.TypeSpec)
			st, _ := ts.Type.(*ast.StructType)

			for _, field := range st.Fields.List {
				for _, name := range field.Names {
					if name.Obj.Kind != ast.Var {
						continue
					}

					tag := name.Name
					if field.Tag != nil {
						jtag := jsonTag(field.Tag.Value)
						if jtag != "" {
							tag = jtag
						}
					}
					tagMap[name.Name] = trans(tag)

					if ast.IsExported(name.Name) {
						exportedList = append(exportedList, name.Name)
					}
				}
			}
			return false
		})
	}
	g.data.ExportedList = exportedList
	//todo: refactor
	g.RegisterTransfer("jsontagof", func(key string) string {
		return tagMap[key]
	})
	g.data.JSON = true
}

func jsonTag(tag string) string {
	reg := regexp.MustCompile(`json:"([^"]*)"`)
	matches := reg.FindStringSubmatch(tag)
	if len(matches) <= 1 {
		return ""
	}
	return matches[1]
}
