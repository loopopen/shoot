package constructor

import (
	"go/ast"
	"regexp"
)

func (g *Generator) makeGetSet(typeName string) {
	defGetSet := g.flags.getset

	var getList []string
	var setList []string

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

					get := defGetSet
					set := defGetSet
					if field.Doc != nil {
						g, s := parseGetSet(field.Doc.Text())
						if g {
							get = true
							set = set && s
						}
						if s {
							get = get && g
							set = true
						}
					}
					if get {
						getList = append(getList, name.Name)
					}
					if set {
						setList = append(setList, name.Name)
					}
				}
			}
			return false
		})
	}

	g.data.GetterList = getList
	g.data.SetterList = setList

	g.data.GetSet = len(getList) > 0 || len(setList) > 0
}

func parseGetSet(doc string) (bool, bool) {
	regGet := regexp.MustCompile(`(?im)^shoot:.*?\Wget(;.*|\s*)$`)
	regSet := regexp.MustCompile(`(?im)^shoot:.*?\Wset(;.*|\s*)$`)
	get := regGet.Match([]byte(doc))
	set := regSet.Match([]byte(doc))
	return get, set
}
