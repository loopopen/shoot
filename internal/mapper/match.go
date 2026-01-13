package mapper

import (
	"go/types"
	"strings"

	"github.com/lopolopen/shoot/internal/shoot"
	"github.com/lopolopen/shoot/internal/transfer"
)

func (g *Generator) makeMatch() {
	for _, f1 := range g.exportedFields {
		for _, f2 := range g.destExportedFields {
			if !canNameMatch(f1, f2, g.srcTagMap) {
				continue
			}

			same, conv := matchType(f1.typ, f2.typ)
			_, convback := matchType(f2.typ, f1.typ)
			if !g.writeDestSet[f2.Name] && !f2.IsGet {
				//f2 = f1
				//f2 = (type)f1
				if same || conv {
					f1.Target = f2
					g.writeDestSet[f2.Name] = true
					g.readSrcMap[f1.Name] = f2.Name
				}
				if same {
					f2.CanAssign = true
				} else if conv {
					f2.IsConv = true
					f2.Type = qualifiedTypeName(f2.typ, g.flags.alias)
				}
			}

			if !g.writeSrcSet[f1.Name] && !f1.IsGet {
				if same || convback {
					f2.Target = f1
					g.writeSrcSet[f1.Name] = true
					g.writeSrcMap[f1.Name] = f2.Name
				}
				if same {
					f1.CanAssign = true
				} else if convback {
					f1.IsConv = true
					f1.Type = qualifiedTypeName(f1.typ, g.flags.alias)
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

func canNameMatch(f1, f2 *Field, tagMap map[string]string) bool {
	if f1.IsGet && f2.IsGet {
		return false
	}
	if f1.IsSet && f2.IsSet {
		return false
	}

	if tagMap == nil {
		tagMap = make(map[string]string)
	}

	m1 := f1.MatchingName()
	m2 := f2.MatchingName()
	tag, ok := tagMap[m1]
	if ok {
		m1 = tag
	}

	return smartMatch(m1, m2)
}

func smartMatch(a, b string) bool {
	if a == b {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	i := len(a)
	for i > 0 && transfer.IsUpper(a[i-1]) {
		i--
	}

	if i == len(a) {
		i = len(b)
		for i > 0 && transfer.IsUpper(b[i-1]) {
			i--
		}
	}

	if i == len(b) {
		return false
	}

	prefixA, suffixA := a[:i], a[i:]
	prefixB, suffixB := b[:i], b[i:]

	if prefixA != prefixB {
		return false
	}

	return strings.EqualFold(suffixA, suffixB)
}

func matchType(type1, type2 types.Type) (bool, bool) {
	same := shoot.TypeEquals(type1, type2)
	conv := types.ConvertibleTo(type1, type2)
	if !same && conv {
		if mayMisConv(type1, type2) {
			conv = false
		}
	}
	return same, conv
}

// mayMisConv returns true if ta and tb are convertible between string and fixed-width integers.
// Otherwise returns true.
func mayMisConv(ta, tb types.Type) bool {
	if isString(ta) && isFixedWidthInt(tb) {
		return true
	}
	if isString(tb) && isFixedWidthInt(ta) {
		return true
	}
	return false
}

func isString(t types.Type) bool {
	b, ok := t.Underlying().(*types.Basic)
	if !ok {
		return false
	}
	return b.Kind() == types.String
}

func isFixedWidthInt(t types.Type) bool {
	b, ok := t.Underlying().(*types.Basic)
	if !ok {
		return false
	}

	switch b.Kind() {
	case types.Int8, types.Int16, types.Int32, types.Int64,
		types.Uint8, types.Uint16, types.Uint32, types.Uint64:
		return true
	}
	return false
}
