package enumer

import (
	"github.com/lopolopen/shoot/internal/shoot"
)

type TmplData struct {
	*shoot.TmplDataBase
	NameList []string
	Bitwise  bool
	Json     bool
	Text     bool
	Sql      bool
}

func NewTmplData(cmdline, version string) *TmplData {
	return &TmplData{
		TmplDataBase: shoot.NewTmplDataBase(cmdline, version),
	}
}

type Flags struct {
	bitwise bool
	json    bool
	text    bool
	sql     bool
}
