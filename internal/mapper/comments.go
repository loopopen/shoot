package mapper

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"sort"
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
}

type commentSpan struct {
	start int
	end   int
}

// collectCommentUpdates copies field documentation from destination fields
// (-path) onto the corresponding source fields. -comment selects the mode:
// on copies only onto fields that have no comment, off copies nothing, and
// force replaces a source field's existing comment, and clears it when the
// matching destination field has none.
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
		if !force && hasComment(srcField.node) {
			continue
		}

		for _, destField := range g.destExportedFields {
			if destField.node == nil || !canNameMatch(srcField, destField, g.srcTagMap, g.flags.ignoreCase) {
				continue
			}

			comment := commentText(destField.node)
			if comment == "" && (!force || !hasComment(srcField.node)) {
				if force {
					break
				}
				continue
			}

			position := g.Pkg().Fset.Position(srcField.node.Pos())
			if position.Filename == "" || position.Offset < 0 {
				continue
			}
			edit := commentEdit{text: comment}
			if force {
				edit.remove, edit.replace, edit.from, edit.to = commentReplacement(g.Pkg().Fset, srcField.node)
			}
			if g.commentUpdates[position.Filename] == nil {
				g.commentUpdates[position.Filename] = make(map[int]commentEdit)
			}
			g.commentUpdates[position.Filename][position.Offset] = edit
			break
		}
	}
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
		return edits[i].end > edits[j].end
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
