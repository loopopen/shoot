package mapper

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"
)

// commentEdit is one source-field comment update. text is the comment to
// write. remove deletes existing documentation before that write. When
// replace is set, text replaces the trailing comment instead of being
// appended to the field line.
type commentEdit struct {
	text    string
	remove  []commentSpan
	replace bool
	from    int
	to      int
	tag     *sourceEdit
}

type commentSpan struct {
	start int
	end   int
}

// collectCommentUpdates copies field documentation from destination fields
// (-path) onto the corresponding source fields. -comment selects the mode:
// on copies only onto fields that have no comment, off copies nothing, and
// force replaces a source field's existing comment, and clears it when the
// matching destination field has none. With -gorm, the same documentation is
// also copied into the comment setting of the source field's gorm tag.
func (g *Generator) collectCommentUpdates() {
	if g.flags.comment.mode() == CommentOff {
		return
	}
	if g.commentUpdates == nil {
		g.commentUpdates = make(map[string]map[int]commentEdit)
	}
	force := g.flags.comment.mode() == CommentForce

	for _, srcField := range g.exportedFields {
		if srcField.node == nil {
			continue
		}

		for _, destField := range g.destExportedFields {
			if destField.node == nil || !canNameMatch(srcField, destField, g.srcTagMap, g.flags.ignoreCase) {
				continue
			}

			comment := commentText(destField.node)
			edit := commentEdit{}
			changed := false
			if force && (comment != "" || hasComment(srcField.node)) {
				edit.text = comment
				edit.remove, edit.replace, edit.from, edit.to = commentReplacement(g.Pkg().Fset, srcField.node)
				changed = true
			} else if !force && comment != "" && !hasComment(srcField.node) {
				edit.text = comment
				changed = true
			}

			if g.flags.gorm {
				gormEdit, ok := gormCommentEdit(
					g.Pkg().Fset,
					srcField.node,
					strings.TrimPrefix(comment, "// "),
					force,
				)
				if ok {
					edit.tag = &gormEdit
					changed = true
				}
			}
			if !changed {
				break
			}

			position := g.Pkg().Fset.Position(srcField.node.Pos())
			if position.Filename == "" || position.Offset < 0 {
				continue
			}
			if g.commentUpdates[position.Filename] == nil {
				g.commentUpdates[position.Filename] = make(map[int]commentEdit)
			}
			g.commentUpdates[position.Filename][position.Offset] = edit
			break
		}
	}
}

// gormCommentEdit returns a source edit that adds, replaces, or removes the
// comment setting in field's gorm tag. Existing comments are left untouched
// unless force is enabled.
func gormCommentEdit(fset *token.FileSet, field *ast.Field, comment string, force bool) (sourceEdit, bool) {
	raw := ""
	if field.Tag != nil {
		raw = field.Tag.Value
	}
	updated, changed := updateGormTag(raw, comment, force)
	if !changed {
		return sourceEdit{}, false
	}

	if field.Tag != nil {
		return sourceEdit{
			start: fset.Position(field.Tag.Pos()).Offset,
			end:   fset.Position(field.Tag.End()).Offset,
			text:  updated,
			order: 1,
		}, true
	}
	offset := fset.Position(field.Type.End()).Offset
	return sourceEdit{start: offset, end: offset, text: " " + updated, order: 1}, true
}

func updateGormTag(raw, comment string, force bool) (string, bool) {
	if raw == "" {
		if comment == "" {
			return "", false
		}
		return quoteStructTag(`gorm:` + strconv.Quote("comment:"+comment)), true
	}

	tag, err := strconv.Unquote(raw)
	if err != nil {
		return raw, false
	}
	start, end, value, found := structTagValue(tag, "gorm")
	if !found {
		if comment == "" {
			return raw, false
		}
		if tag != "" && !strings.HasSuffix(tag, " ") {
			tag += " "
		}
		tag += `gorm:` + strconv.Quote("comment:"+comment)
		return quoteStructTag(tag), true
	}
	if strings.TrimSpace(value) == "-" {
		return raw, false
	}

	updated, changed := updateGormComment(value, comment, force)
	if !changed {
		return raw, false
	}
	quoted := strconv.Quote(updated)
	tag = tag[:start] + quoted[1:len(quoted)-1] + tag[end:]
	return quoteStructTag(tag), true
}

// structTagValue returns the encoded value range and decoded value for key.
// The range excludes the surrounding double quotes.
func structTagValue(tag, key string) (start, end int, value string, found bool) {
	for i := 0; i < len(tag); {
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		nameStart := i
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != '\\' {
			i++
		}
		if nameStart == i || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			return 0, 0, "", false
		}
		name := tag[nameStart:i]
		i += 2
		valueStart := i
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
				if i >= len(tag) {
					return 0, 0, "", false
				}
			}
			i++
		}
		if i >= len(tag) {
			return 0, 0, "", false
		}
		if name == key {
			decoded, err := strconv.Unquote(`"` + tag[valueStart:i] + `"`)
			if err != nil {
				return 0, 0, "", false
			}
			return valueStart, i, decoded, true
		}
		i++
	}
	return 0, 0, "", false
}

func updateGormComment(value, comment string, force bool) (string, bool) {
	parts := strings.Split(value, ";")
	found := false
	for _, part := range parts {
		key, _, ok := strings.Cut(strings.TrimSpace(part), ":")
		if ok && strings.EqualFold(strings.TrimSpace(key), "comment") {
			found = true
			break
		}
	}
	if found && !force {
		return value, false
	}
	if found {
		kept := parts[:0]
		for _, part := range parts {
			key, _, ok := strings.Cut(strings.TrimSpace(part), ":")
			if ok && strings.EqualFold(strings.TrimSpace(key), "comment") {
				continue
			}
			kept = append(kept, part)
		}
		value = strings.Join(kept, ";")
		if comment == "" {
			return value, true
		}
	}
	if comment == "" {
		return value, false
	}
	if value != "" && !strings.HasSuffix(value, ";") {
		value += ";"
	}
	return value + "comment:" + comment, true
}

func quoteStructTag(tag string) string {
	if !strings.ContainsRune(tag, '`') {
		return "`" + tag + "`"
	}
	return strconv.Quote(tag)
}

// commentReplacement describes how to clear comments already attached to
// field. Doc comments are removed with their lines. A trailing comment is
// replaced in place.
func commentReplacement(fset *token.FileSet, field *ast.Field) (remove []commentSpan, replace bool, from, to int) {
	if field.Doc != nil {
		remove = append(remove, docSpan(fset, field.Doc))
	}
	if field.Comment != nil {
		replace = true
		from = fset.Position(field.Comment.Pos()).Offset
		to = fset.Position(field.Comment.End()).Offset
	}
	return remove, replace, from, to
}

func docSpan(fset *token.FileSet, group *ast.CommentGroup) commentSpan {
	file := fset.File(group.Pos())
	startLine := file.Line(group.Pos())
	endLine := file.Line(group.End())
	if endLine > startLine && file.Offset(group.End()) == file.Offset(file.LineStart(endLine)) {
		endLine--
	}
	start := file.Offset(file.LineStart(startLine))
	end := file.Size()
	if endLine < file.LineCount() {
		end = file.Offset(file.LineStart(endLine + 1))
	}
	return commentSpan{start: start, end: end}
}

func hasComment(field *ast.Field) bool {
	return field.Doc != nil || field.Comment != nil
}

func commentText(field *ast.Field) string {
	var texts []string
	for _, group := range []*ast.CommentGroup{field.Doc, field.Comment} {
		if group == nil {
			continue
		}
		text := strings.TrimSpace(group.Text())
		if text != "" {
			texts = append(texts, text)
		}
	}
	text := strings.Join(texts, " ")
	if text == "" {
		return ""
	}
	return "// " + strings.Join(strings.Fields(text), " ")
}

// ApplySourceChanges writes the pending documentation updates.  It is called
// by the command only after generated mapping files have been written, so the
// package analysis remains based on one consistent source snapshot.
func (g *Generator) ApplySourceChanges() error {
	for filename, updates := range g.commentUpdates {
		if err := applyCommentUpdates(filename, updates); err != nil {
			return err
		}
	}
	return nil
}

type sourceEdit struct {
	start int
	end   int
	text  string
	order int
}

func applyCommentUpdates(filename string, updates map[int]commentEdit) error {
	if len(updates) == 0 {
		return nil
	}

	src, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read source for comment update: %w", err)
	}

	edits := make([]sourceEdit, 0, len(updates))
	for offset, update := range updates {
		if offset < 0 || offset > len(src) {
			return fmt.Errorf("comment update offset %d is outside %s", offset, filename)
		}
		lineStart := bytes.LastIndexByte(src[:offset], '\n') + 1
		indent := string(src[lineStart:offset])
		if strings.Trim(indent, " \t") != "" {
			return fmt.Errorf("comment update offset %d does not start a field in %s", offset, filename)
		}

		for _, span := range update.remove {
			if span.start < 0 || span.end > len(src) || span.start > span.end {
				return fmt.Errorf("comment removal [%d,%d) is outside %s", span.start, span.end, filename)
			}
			edits = append(edits, sourceEdit{start: span.start, end: span.end})
		}
		if update.tag != nil {
			if update.tag.start < 0 || update.tag.end > len(src) || update.tag.start > update.tag.end {
				return fmt.Errorf("gorm tag update [%d,%d) is outside %s", update.tag.start, update.tag.end, filename)
			}
			edits = append(edits, *update.tag)
		}
		if update.replace {
			if update.from < 0 || update.to > len(src) || update.from > update.to {
				return fmt.Errorf("comment replacement [%d,%d) is outside %s", update.from, update.to, filename)
			}
			edits = append(edits, sourceEdit{start: update.from, end: update.to, text: update.text})
			continue
		}
		if update.text == "" {
			continue
		}

		lineEnd := bytes.IndexByte(src[offset:], '\n')
		if lineEnd < 0 {
			lineEnd = len(src)
		} else {
			lineEnd += offset
		}
		edits = append(edits, sourceEdit{start: lineEnd, end: lineEnd, text: " " + update.text})
	}

	sort.Slice(edits, func(i, j int) bool {
		if edits[i].start != edits[j].start {
			return edits[i].start > edits[j].start
		}
		if edits[i].end != edits[j].end {
			return edits[i].end > edits[j].end
		}
		return edits[i].order < edits[j].order
	})
	for _, edit := range edits {
		src = append(src[:edit.start], append([]byte(edit.text), src[edit.end:]...)...)
	}

	formatted, err := format.Source(src)
	if err != nil {
		return fmt.Errorf("format source after comment update: %w", err)
	}
	return os.WriteFile(filename, formatted, 0o644)
}
