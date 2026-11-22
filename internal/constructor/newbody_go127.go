//go:build go1.27

package constructor

import (
	"bytes"
	"go/version"

	"golang.org/x/tools/go/packages"
)

func supportsPromotedFieldLiterals(pkg *packages.Package) bool {
	if pkg == nil || pkg.Module == nil {
		return true
	}
	if pkg.Module.GoVersion == "" {
		return false
	}
	return version.Compare("go"+pkg.Module.GoVersion, "go1.27") >= 0
}

func newBodyRec(buf *bytes.Buffer, fields []*Field, pointer int, depth int32, nameMap map[string]string, directPromotedFields bool) int {
	if len(fields) == 0 {
		return 0
	}
	i := pointer
	for i < len(fields) {
		field := fields[i]
		if field.depth <= depth {
			break
		}
		if field.isEmbeded {
			if field.isPtr || !directPromotedFields {
				writeEmbeddedFieldStart(buf, field)
				i = newBodyRec(buf, fields, i+1, field.depth, nameMap, directPromotedFields)
				buf.WriteString("},\n")
			} else {
				i = newBodyRec(buf, fields, i+1, field.depth, nameMap, directPromotedFields)
			}
			continue
		}
		writeNewField(buf, field, nameMap)
		i++
	}
	return i
}
