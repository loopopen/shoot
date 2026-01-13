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

func (f *Func) Equals(fn *Func) bool {
	if f == nil || fn == nil {
		return f == fn
	}
	if f.Name != fn.Name {
		return false
	}
	if f.IsGetter() == fn.IsGetter() {
		return TypeEquals(f.Result, fn.Result)
	}
	if f.IsSetter() == fn.IsSetter() {
		return TypeEquals(f.Param, fn.Param)
	}
	return false
}
