package mapper

import (
	"go/types"
	"strings"

	"github.com/lopolopen/shoot/internal/shoot"
)

type TmplData struct {
	*shoot.TmplDataBase
	SrcFieldList          []*Field
	DestFieldList         []*Field
	DestTypeName          string //Order
	DestPkgName           string //model
	DestPkgAlias          string //domain
	DestPkgPath           string //mappersample/domain/model
	QualifiedDestTypeName string //domain.Order

	SrcPtrTypeMap   map[string]string //Model -> gorm.Model
	SrcPtrPathList  []string
	DestPtrTypeMap  map[string]string
	DestPtrPathList []string

	SrcNeedReadCheckMap  map[string]string
	DestNeedReadCheckMap map[string]string

	SrcNeedWriteCheckMap  map[string]string
	DestNeedWriteCheckMap map[string]string

	// MismatchFuncMap  map[string]string //Amount -> Amount
	// SrcToDestFuncMap map[string]string //Amount -> StringToDecimal
	// DestToSrcFuncMap map[string]string //Amount -> DecimalToString

	// MismatchSubMap     map[string]string //Address -> Address
	// DestMismatchSubMap map[string]string
	// SrcPtrSet  map[string]bool //Address -> true
	// DestPtrSet map[string]bool //Address -> false
	// SrcSubTypeMap  map[string]string //Address -> OrderAddress
	// DestSubTypeMap map[string]string //Address -> domain.OrderAddress

	// MismatchSubListMap     map[string]string //AddrList -> AddrList
	// DestMismatchSubListMap map[string]string

	ReadMethodName  string //fromModel
	ReadParamPrefix string
	WriteMethodName string //toModel
}

func NewTmplData(cmdline, version string) *TmplData {
	return &TmplData{
		TmplDataBase: shoot.NewTmplDataBase(cmdline, version),
	}
}

type Flags struct {
	destDir   string
	destTypes map[string]string
	alias     string
}

type Field struct {
	Name        string //ID
	path        string //Model.ID
	typ         types.Type
	depth       int32
	backingName string
	IsGet       bool
	IsSet       bool
	Target      *Field `json:"-"`
	IsSame      bool
	IsConv      bool
	CanMap      bool
	CanEachMap  bool
	Type        string
	Func        string
	IsPtr       bool
}

func (f Field) IsEmbeded() bool {
	return strings.Contains(f.path, dot)
}

func (f Field) CoveredBy(path string) bool {
	if f.path == path { //Model.ID
		return true
	}
	if strings.HasPrefix(f.path, path+dot) { //Model.
		return true
	}
	xs := strings.Split(path, dot)
	return strings.HasSuffix(f.path, dot+xs[len(xs)-1]) //.ID
}

func (f Field) MatchingName() string {
	if f.backingName != "" {
		return f.backingName
	}
	return f.Name
}

type Func struct {
	name   string
	path   string
	param  types.Type
	result types.Type
}
