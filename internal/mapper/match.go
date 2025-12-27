package mapper

import (
	"go/types"
)

func (g *Generator) makeMatch() {
	g.data.SrcEqMatchMap = make(map[string]string)
	g.data.DestEqMatchMap = make(map[string]string)
	g.data.ConvMatchMap = map[string]string{}
	g.data.SrcToDestTypeMap = make(map[string]string)
	g.data.DestToSrcTypeMap = make(map[string]string)

	for _, f1 := range g.exportedFields {
		for _, f2 := range g.destExportedFields {
			if !canNameMatch(f1, f2, g.tagMap) {
				continue
			}

			same, conv := matchType(f1.typ, f2.typ)
			if !g.writeDestSet[f2.name] && !f2.isGet {
				if same {
					g.data.DestEqMatchMap[f1.name] = f2.name //dest= or dest.Set
				} else if conv {
					g.data.ConvMatchMap[f1.name] = f2.name
					//in ToXxx, type converter needs desc type
					g.data.SrcToDestTypeMap[f1.name] = qualifiedTypeName(f2.typ, g.flags.alias)
				}
				if same || conv {
					g.writeDestSet[f2.name] = true
					g.readSrcMap[f1.name] = f2.name
				}
			}

			if !g.writeSrcSet[f1.name] && !f1.isGet {
				if same {
					g.data.SrcEqMatchMap[f1.name] = f2.name //src= or src.Set
				} else if conv {
					g.data.ConvMatchMap[f1.name] = f2.name
					//in FromXxx, type converter needs src type
					g.data.DestToSrcTypeMap[f1.name] = qualifiedTypeName(f1.typ, g.flags.alias)
				}
				if same || conv {
					g.writeSrcSet[f1.name] = true
					g.writeSrcMap[f1.name] = f2.name
				}
			}
		}
	}
}

func qualifiedTypeName(t types.Type, alias string) string {
	qualifier := func(pkg *types.Package) string {
		if alias != "" {
			return alias
		}
		if pkg == nil {
			return ""
		}
		return pkg.Name()
	}
	return types.TypeString(t, qualifier)
}

func canNameMatch(f1, f2 Field, tagMap map[string]string) bool {
	if tagMap == nil {
		tagMap = make(map[string]string)
	}

	name, ok := tagMap[f1.MatchingName()]
	if !ok {
		name = f1.MatchingName()
	}

	return name == f2.MatchingName()
}

func matchType(type1, type2 types.Type) (bool, bool) {
	same := types.Identical(type1, type2)
	conv := types.ConvertibleTo(type1, type2)
	return same, conv
}
