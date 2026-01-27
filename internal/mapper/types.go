package mapper

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/lopolopen/shoot/internal/shoot"
)

type Way string

const (
	WayToOnly   Way = "toonly"
	WayFromOnly Way = "fromonly"
	WayBoth     Way = "both"
)

func (v *Way) Set(value string) error {
	var err error
	switch Way(value) {
	case WayToOnly, "->":
		*v = WayToOnly
	case WayFromOnly, "<-":
		*v = WayFromOnly
	case WayBoth, "<->":
		*v = WayBoth
	default:
		err = fmt.Errorf("invalid mapping way: %s", value)
	}
	return err
}

func (v *Way) String() string {
	return string(*v)
}

type TmplData struct {
	*shoot.TmplDataBase
	SrcCtorParams         []*Field
	DestCtorParams        []*Field
	SrcFieldList          []*Field
	DestFieldList         []*Field
	DestTypeName          string //Order
	DestPkgName           string //model
	DestPkgAlias          string //domain
	DestPkgPath           string //mapperexample/domain/model
	QualifiedDestTypeName string //domain.Order

	SrcPtrTypeMap   map[string]string //Model -> gorm.Model
	SrcPtrPathList  []string
	DestPtrTypeMap  map[string]string
	DestPtrPathList []string

	SrcNeedReadCheckMap  map[string]string
	DestNeedReadCheckMap map[string]string

	SrcNeedWriteCheckMap  map[string]string
	DestNeedWriteCheckMap map[string]string

	ReadMethodName  string //fromModel
	IsReadParamPtr  bool
	WriteMethodName string //toModel
	IsToOnly        bool
	IsFromOnly      bool
}

func NewTmplData(cmdline, version string) *TmplData {
	return &TmplData{
		TmplDataBase: shoot.NewTmplDataBase(cmdline, version),
	}
}

type Flags struct {
	destDir    string
	destTypes  map[string]string
	alias      string
	way        Way
	ignoreCase bool
}

type Field struct {
	Name        string //ID
	Path        string //Model.ID
	typ         types.Type
	depth       int32
	backingName string
	IsGet       bool
	IsSet       bool
	Target      *Field `json:"-"`
	CanAssign   bool
	IsConv      bool
	CanMap      bool
	CanEachMap  bool
	Type        string
	Func        string
	IsPtr       bool
	Zero        string
	warned      bool
}

func (f Field) IsEmbeded() bool {
	return strings.Contains(f.Path, dot)
}

func (f Field) CoveredBy(path string) bool {
	if f.Path == path { //Model.ID
		return true
	}
	if strings.HasPrefix(f.Path, path+dot) { //Model.
		return true
	}
	xs := strings.Split(path, dot)
	return strings.HasSuffix(f.Path, dot+xs[len(xs)-1]) //.ID
}

func (f Field) MatchingName() string {
	if f.backingName != "" {
		return f.backingName
	}
	return f.Name
}
