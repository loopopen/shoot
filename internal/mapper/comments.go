package mapper

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"os"
	"sort"
	"strings"
)

// collectCommentUpdates copies field documentation from the package named by
// -path to the corresponding field in the package being generated for.  The
// destination field is deliberately never overwritten when it already has a
// comment of its own.
func (g *Generator) collectCommentUpdates() {
	if g.commentUpdates == nil {
		g.commentUpdates = make(map[string]map[int]string)
	}

	for _, target := range g.exportedFields {
		if target.node == nil || hasComment(target.node) {
			continue
		}

		for _, source := range g.destExportedFields {
			if source.node == nil || !canNameMatch(target, source, g.srcTagMap, g.flags.ignoreCase) {
				continue
			}

			comment := commentText(source.node)
			if comment == "" {
				continue
			}

			position := g.Pkg().Fset.Position(target.node.Pos())
			if position.Filename == "" || position.Offset < 0 {
				continue
			}
			if g.commentUpdates[position.Filename] == nil {
				g.commentUpdates[position.Filename] = make(map[int]string)
			}
			g.commentUpdates[position.Filename][position.Offset] = comment
			break
		}
	}
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

func applyCommentUpdates(filename string, updates map[int]string) error {
	if len(updates) == 0 {
		return nil
	}

	src, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read source for comment update: %w", err)
	}

	offsets := make([]int, 0, len(updates))
	for offset := range updates {
		offsets = append(offsets, offset)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(offsets)))

	for _, offset := range offsets {
		if offset < 0 || offset > len(src) {
			return fmt.Errorf("comment update offset %d is outside %s", offset, filename)
		}
		lineStart := bytes.LastIndexByte(src[:offset], '\n') + 1
		indent := string(src[lineStart:offset])
		if strings.Trim(indent, " \t") != "" {
			return fmt.Errorf("comment update offset %d does not start a field in %s", offset, filename)
		}

		lineEnd := bytes.IndexByte(src[offset:], '\n')
		if lineEnd < 0 {
			lineEnd = len(src)
		} else {
			lineEnd += offset
		}
		src = append(src[:lineEnd], append([]byte(" "+updates[offset]), src[lineEnd:]...)...)
	}

	formatted, err := format.Source(src)
	if err != nil {
		return fmt.Errorf("format source after comment update: %w", err)
	}
	return os.WriteFile(filename, formatted, 0o644)
}
