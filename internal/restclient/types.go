package restclient

import (
	"github.com/loopopen/shoot/internal/shoot"
)

type TmplData struct {
	*shoot.TmplDataBase
	MethodList      []string
	SigMap          map[string]string
	HTTPMethodMap   map[string]string
	PathMap         map[string]string
	AliasMap        map[string]map[string]string
	PathParamsMap   map[string][]string
	QueryParamsMap  map[string][]string
	IsParamPtrMap   map[string]map[string]bool
	ReturnResultMap map[string]struct {
		Type  string
		IsPtr bool
	} // method may return one result or not
	ErrReturnMap    map[string]string
	BodyParamMap    map[string]string
	QueryDictMap    map[string]string
	DefaultHeaders  map[string]map[string]string
	CtxParamMap     map[string]string
	BodyHTTPMethods []string
}

func NewTmplData(cmdline, version string) *TmplData {
	return &TmplData{
		TmplDataBase: shoot.NewTmplDataBase(cmdline, version),
	}
}

// type Flags struct{}
