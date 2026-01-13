package shoot

import (
	"go/types"
	"strings"
)

type Func struct {
	Name   string
	Path   string
	Param  types.Type
	Result types.Type
}

func (f *Func) IsGetter() bool {
	return f.Param == nil && f.Result != nil
}

func (f *Func) IsSetter() bool {
	if !strings.HasPrefix(f.Name, Set) {
		return false
	}
	return f.Param != nil && f.Result == nil
}
