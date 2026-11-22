package constructor

import (
	"go/types"

	"github.com/loopopen/shoot/internal/shoot"
	"github.com/loopopen/shoot/internal/transfer"
)

type TagCase = transfer.Case

const (
	TagCasePascal = transfer.CasePascal
	TagCaseCamel  = transfer.CaseCamel
	TagCaseSnake  = transfer.CaseSnake
	TagCaseLower  = transfer.CaseLower
	TagCaseUpper  = transfer.CaseUpper
)

type TmplData struct {
	*shoot.TmplDataBase
	Imports           string
	AllList           []string
	NewMap            map[string]string
	GetSet            bool
	GetterList        []string
	SetterList        []string
	GetterIfaces      []string
	SetterIfaces      []string
	Option            bool
	DefaultList       []string
	DefaultValueMap   map[string]string
	JSON              bool
	JSONList          []string
	JSONTagMap        map[string]string
	JSONGetterList    []string
	JSONSetterList    []string
	ExportedList      []string
	EmbedList         []string
	Self              bool
	Short             bool
	NewParamsList     string
	NewBody           string
	TypeParamList     string
	TypeParamNameList string
	TypeMap           map[string]string
	Mark              bool
}

func NewTmplData(cmdline, version string) *TmplData {
	return &TmplData{
		TmplDataBase: shoot.NewTmplDataBase(cmdline, version),
	}
}

type Flags struct {
	//if true:
	//[ ] = [get;set] => get+set
	//[get] => get-only
	//[set] => set-only
	//if false:
	//[ ] => neither
	//[get] => get-only
	//[set] => set-only
	//[get;set] => get+set
	getset  bool
	json    bool
	tagcase TagCase
	opt     bool
	exp     bool
	short   bool
	mark    bool
	sign    string
}

type Field struct {
	name          string
	qualifiedType string
	typ           types.Type
	depth         int32
	isPtr         bool
	isShadowed    bool
	isEmbeded     bool
	isGet         bool
	isSet         bool
	isNew         bool
	defValue      string
	jsonTag       string
}

func (f *Field) HasJSONTag() bool {
	return f.jsonTag != ""
}

func (f *Field) JSONTag() string {
	if f.jsonTag == "" {
		return f.name
	}
	return f.jsonTag
}
