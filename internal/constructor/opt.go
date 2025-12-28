package constructor

import (
	"go/ast"
	"regexp"
	"strings"

	"github.com/lopolopen/shoot/internal/shoot"
	"github.com/lopolopen/shoot/internal/tools/logx"
	"github.com/lopolopen/shoot/internal/transfer"
)

func (g *Generator) makeOpt(typeName string) {
	g.RegisterTransfer("defaultof", transfer.ID)

	var defList []string
	defMap := make(map[string]string)

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
					if ast.IsExported(name.Name) {
						continue
					}

					if field.Doc != nil {
						v, ok := parseDefault(field.Doc.Text())
						if ok {
							if !g.flags.opt {
								logx.Fatalf("should not use instruction def(ault) when -opt(ion) disabled")
							}

							if len(g.data.NewList) != len(g.data.AllList)-len(g.data.ExportedList) {
								if shoot.Contains(g.data.NewList, name.Name) {
									logx.Fatalf("should not apply both instruction new and def(ault) to field %s of type %s", name.Name, typeName)
								}
							}

							defList = append(defList, name.Name)
							defMap[name.Name] = strings.TrimSpace(v)
						}
					}
				}
			}
			return false
		})
	}

	if !g.flags.opt {
		return
	}

	g.RegisterTransfer("defaultof", func(key string) string {
		return defMap[key]
	})
	g.data.DefaultList = defList
	g.data.Option = true
	g.data.Short = g.flags.short
}

func parseDefault(doc string) (string, bool) {
	regDef := regexp.MustCompile(`(?im)^shoot:.*?\Wdef(ault)?=([^;\n]+)(;.*|\s*)$`)
	ms := regDef.FindStringSubmatch(doc)
	for idx, m := range ms {
		if (m == "" || m == "ault") && idx+1 < len(ms) {
			return ms[idx+1], true
		}
	}
	return "", false
}
