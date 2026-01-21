package constructor

import (
	"go/ast"
	"strings"

	"github.com/lopolopen/shoot/internal/shoot"
	"github.com/lopolopen/shoot/internal/transfer"
)

func (g *Generator) makeJson() {
	if !g.flags.json {
		return
	}

	needJSON := false

	trans := transfer.ID
	switch g.flags.tagcase {
	case TagCasePascal:
		trans = transfer.ToPascalCase
	case TagCaseCamel:
		trans = transfer.ToCamelCase
	case TagCaseLower:
		trans = strings.ToLower
	case TagCaseUpper:
		trans = strings.ToUpper
	}

	allGetSet := shoot.MakeSet[string]()
	allSetSet := shoot.MakeSet[string]()
	for _, m := range g.getsetMethods {
		if m.IsGetter() {
			allGetSet.Adds(m.Name)
		}
		if m.IsSetter() {
			allSetSet.Adds(m.Name)
		}
	}

	tagMap := make(map[string]string)
	var jsonList []string
	var getterList []string
	var setterList []string
	var exportedList []string
	for _, f := range g.fields {
		if f.isShadowed || f.isEmbeded {
			continue
		}

		jsonTag := trans(f.JSONTag())
		if ast.IsExported(f.name) {
			if !f.HasJSONTag() && jsonTag != f.name {
				needJSON = true
			}
		} else {
			if f.isGet || allGetSet.Has(transfer.ToPascalCase(f.name)) {
				needJSON = true
			}
			if f.isSet || allSetSet.Has(set+transfer.ToPascalCase(f.name)) {
				needJSON = true
			}
		}
		tagMap[f.name] = jsonTag

		if ast.IsExported(f.name) {
			exportedList = append(exportedList, f.name)
			jsonList = append(jsonList, f.name)
		} else {
			getset := false
			if f.isGet || allGetSet.Has(transfer.ToPascalCase(f.name)) {
				getterList = append(getterList, f.name)
				getset = true
			}
			if f.isSet || allSetSet.Has(set+transfer.ToPascalCase(f.name)) {
				setterList = append(setterList, f.name)
				getset = true
			}
			if getset {
				jsonList = append(jsonList, f.name)
			}
		}
	}
	g.data.JSONTagMap = tagMap
	g.data.JSON = needJSON
	g.data.JSONList = jsonList
	g.data.JSONGetterList = getterList
	g.data.JSONSetterList = setterList
	g.data.ExportedList = exportedList
}
