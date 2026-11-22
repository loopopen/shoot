package constructor

import (
	"bytes"
	"fmt"
)

func newBody(fields []*Field, nameMap map[string]string, directPromotedFields bool) string {
	if len(fields) == 0 {
		return ""
	}
	var buf bytes.Buffer
	newBodyRec(&buf, fields, 0, -1, nameMap, directPromotedFields)
	return buf.String()
}

func writeEmbeddedFieldStart(buf *bytes.Buffer, field *Field) {
	ref := ""
	if field.isPtr {
		ref = "&"
	}
	buf.WriteString(fmt.Sprintf("%s: %s%s{\n", field.name, ref, field.qualifiedType))
}

func writeNewField(buf *bytes.Buffer, field *Field, nameMap map[string]string) {
	name, ok := nameMap[field.name]
	if ok && !field.isShadowed {
		buf.WriteString(fmt.Sprintf("%s: %s,\n", field.name, name))
	} else if field.defValue != "" {
		buf.WriteString(fmt.Sprintf("%s: %s,\n", field.name, field.defValue))
	}
}
