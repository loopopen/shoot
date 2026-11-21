//go:build !go1.27

package constructor

import (
	"bytes"

	"golang.org/x/tools/go/packages"
)

func supportsPromotedFieldLiterals(*packages.Package) bool {
	return false
}

func newBodyRec(buf *bytes.Buffer, fields []*Field, pointer int, depth int32, nameMap map[string]string, _ bool) int {
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
			writeEmbeddedFieldStart(buf, field)
			i = newBodyRec(buf, fields, i+1, field.depth, nameMap, false)
			buf.WriteString("},\n")
			continue
		}
		writeNewField(buf, field, nameMap)
		i++
	}
	return i
}
