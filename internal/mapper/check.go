package mapper

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lopolopen/shoot/internal/tools/logx"
	"github.com/lopolopen/shoot/internal/transfer"
)

//tips: write src means read rest, and vice versa

func (g *Generator) makeReadWriteCheck() {
	g.makeReadCond()
	g.nilCheckRead()
	g.nilCheckWrite()
	g.neverWriteCheck()
}

func (g *Generator) nilCheckRead() {
	g.data.DestNeedReadCheckMap = make(map[string]string)
	g.data.SrcNeedReadCheckMap = make(map[string]string)

	for _, f := range g.exportedFields {
		s := f.Name

		d, ok := g.readSrcMap[s]
		if ok {
			if _, ok := g.srcPathsMap[s]; ok {
				g.data.SrcNeedReadCheckMap[s] = d
			}
		}

		d, ok = g.readDestMap()[s]
		if ok {
			if _, ok := g.destPathsMap[d]; ok {
				g.data.DestNeedReadCheckMap[f.Name] = d
			}
		}
	}

}

func (g *Generator) neverWriteCheck() {
	var neverWriteSrc []*Field
	var neverWriteDest []*Field

	for _, f := range g.exportedFields {
		s := f.Name

		if g.writeSrcSet[s] || f.IsGet {
			continue
		}
		c := false
		for path := range g.writeSrcSet { //Model covers Model.ID
			if f.CoveredBy(path) {
				c = true
				break
			}
		}
		if c {
			continue
		}
		neverWriteSrc = append(neverWriteSrc, f)
	}

	for _, f := range g.destExportedFields {
		d := f.Name

		if g.writeDestSet[d] || f.IsGet {
			continue
		}
		c := false
		for path := range g.writeDestSet {
			if f.CoveredBy(path) {
				c = true
				break
			}
		}
		if c {
			continue
		}

		neverWriteDest = append(neverWriteDest, f)
	}

	reportWarn(g.data.PackageName+dot+g.data.TypeName, neverWriteSrc)
	reportWarn(g.data.QualifiedDestTypeName, neverWriteDest)
}

func reportWarn(typename string, fields []*Field) {
	var fnames, mnames []string
	for _, f := range fields {
		if f.IsSet {
			mnames = append(mnames, f.path)
		} else {
			fnames = append(fnames, f.path)
		}
	}
	if len(fnames) > 0 {
		logx.Warnf("%s: these fields are never assigned:\n\t🌱 %s", typename, strings.Join(fnames, ", "))
	}
	if len(mnames) > 0 {
		logx.Warnf("%s: these setter methods are never called:\n\t🔧 %s", typename, strings.Join(mnames, ", "))
	}
}

func (g *Generator) makeReadCond() {
	g.RegisterTransfer("condofread", transfer.ID)

	g.srcPathsMap = make(map[string][]string)
	g.destPathsMap = make(map[string][]string)
	prepareReadPaths(g.exportedFields, g.srcPtrTypeMap, g.srcPathsMap)
	prepareReadPaths(g.destExportedFields, g.destPtrTypeMap, g.destPathsMap)

	g.RegisterTransfer("condofread", func(v, name string, isSrc bool) string {
		var pathsMap map[string][]string
		if isSrc {
			pathsMap = g.srcPathsMap
		} else {
			pathsMap = g.destPathsMap
		}
		paths := pathsMap[name]
		sort.Strings(paths)
		ps := make([]string, len(paths))
		for i, p := range paths {
			ps[i] = fmt.Sprintf(" %s.%s != nil ", v, p)
		}
		return strings.Join(ps, "&&")
	})
}

func prepareReadPaths(fields []*Field, ptrTypeMap map[string]string, readPathsMap map[string][]string) {
	for _, f := range fields {
		if !f.IsEmbeded() {
			continue
		}

		var readPaths []string
		ps := strings.Split(f.path, dot)
		for i := 1; i < len(ps); i++ {
			path := strings.Join(ps[:i], dot)
			_, ok := ptrTypeMap[path]
			if !ok {
				continue
			}
			readPaths = append(readPaths, path)
		}
		if len(readPaths) > 0 {
			readPathsMap[f.Name] = readPaths
		}
	}
}

func (g *Generator) nilCheckWrite() {
	g.data.SrcPtrTypeMap = make(map[string]string)
	g.data.DestPtrTypeMap = make(map[string]string)

	var srcPtrPathList []string
	var destPtrPathList []string
	for _, f := range g.exportedFields {
		s := f.Name

		_, ok := g.writeSrcMap[s]
		if ok && f.IsEmbeded() {
			for p, t := range g.srcPtrTypeMap {
				if _, ok := g.data.SrcPtrTypeMap[p]; ok {
					continue
				}

				if f.CoveredBy(p) {
					srcPtrPathList = append(srcPtrPathList, p)
					g.data.SrcPtrTypeMap[p] = t
				}
			}
		}
	}

	for _, f := range g.destExportedFields {
		for _, d := range g.writeDestMap() {
			if f.Name != d {
				continue
			}

			if !f.IsEmbeded() {
				continue
			}

			for p, t := range g.destPtrTypeMap {
				if _, ok := g.data.DestPtrTypeMap[p]; ok {
					continue
				}

				if f.CoveredBy(p) {
					destPtrPathList = append(destPtrPathList, p)
					g.data.DestPtrTypeMap[p] = t
				}
			}
		}
	}

	sort.Strings(srcPtrPathList)
	g.data.SrcPtrPathList = srcPtrPathList
	sort.Strings(destPtrPathList)
	g.data.DestPtrPathList = destPtrPathList
}
