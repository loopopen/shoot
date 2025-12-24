package constructor

import (
	"fmt"

	"github.com/lopolopen/shoot/internal/shoot"
)

type TagCase string

func (v *TagCase) Set(value string) error {
	switch value {
	case "pascal", "camel", "lower", "upper":
		*v = TagCase(value)
		return nil
	default:
		return fmt.Errorf("invalid tag case: %s", value)
	}
}

func (v *TagCase) String() string {
	return string(*v)
}

type TmplData struct {
	*shoot.TmplDataBase
	// GoFile  string
	Imports string
	//All = Exported + Unexported
	AllList     []string
	NewList     []string
	GetSet      bool
	GetterList  []string
	SetterList  []string
	Option      bool
	DefaultList []string
	JSON        bool
	//Marshal: Getteer + Exported
	//Unmarshal: Setter + Exported
	ExportedList []string
	EmbedList    []string
	Self         bool
	Short        bool
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
	tagcase string
	opt     bool
	exp     bool
	short   bool
}
