//go:build go1.27

package mapper

import (
	"go/types"

	"github.com/loopopen/shoot/internal/shoot"
)

func mapperFuncMatches(fn shoot.Func, param, result types.Type) bool {
	if shoot.TypeEquals(fn.Param, param) && shoot.TypeEquals(fn.Result, result) {
		return true
	}

	sig := fn.Signature
	if sig == nil || sig.TypeParams() == nil || sig.TypeParams().Len() == 0 {
		return false
	}
	if sig.Params().Len() != 1 || sig.Results().Len() != 1 {
		return false
	}

	tparams := sig.TypeParams()
	allowed := make(map[*types.TypeParam]bool, tparams.Len())
	bindings := make(map[*types.TypeParam]types.Type, tparams.Len())
	for i := 0; i < tparams.Len(); i++ {
		allowed[tparams.At(i)] = true
	}
	if !inferMapperTypeArgs(sig.Params().At(0).Type(), param, allowed, bindings) {
		return false
	}

	targs := make([]types.Type, tparams.Len())
	for i := 0; i < tparams.Len(); i++ {
		arg, ok := bindings[tparams.At(i)]
		if !ok {
			return false
		}
		targs[i] = arg
	}

	instantiated, err := types.Instantiate(nil, sig, targs, true)
	if err != nil {
		return false
	}
	instantiatedSig, ok := instantiated.(*types.Signature)
	if !ok {
		return false
	}
	return shoot.TypeEquals(instantiatedSig.Params().At(0).Type(), param) &&
		shoot.TypeEquals(instantiatedSig.Results().At(0).Type(), result)
}

func inferMapperTypeArgs(pattern, actual types.Type, allowed map[*types.TypeParam]bool, bindings map[*types.TypeParam]types.Type) bool {
	if tp, ok := pattern.(*types.TypeParam); ok && allowed[tp] {
		if bound, exists := bindings[tp]; exists {
			return shoot.TypeEquals(bound, actual)
		}
		bindings[tp] = actual
		return true
	}

	switch p := pattern.(type) {
	case *types.Pointer:
		a, ok := actual.(*types.Pointer)
		return ok && inferMapperTypeArgs(p.Elem(), a.Elem(), allowed, bindings)
	case *types.Slice:
		a, ok := actual.(*types.Slice)
		return ok && inferMapperTypeArgs(p.Elem(), a.Elem(), allowed, bindings)
	case *types.Array:
		a, ok := actual.(*types.Array)
		return ok && p.Len() == a.Len() && inferMapperTypeArgs(p.Elem(), a.Elem(), allowed, bindings)
	case *types.Map:
		a, ok := actual.(*types.Map)
		return ok &&
			inferMapperTypeArgs(p.Key(), a.Key(), allowed, bindings) &&
			inferMapperTypeArgs(p.Elem(), a.Elem(), allowed, bindings)
	case *types.Chan:
		a, ok := actual.(*types.Chan)
		return ok && p.Dir() == a.Dir() && inferMapperTypeArgs(p.Elem(), a.Elem(), allowed, bindings)
	case *types.Named:
		a, ok := actual.(*types.Named)
		if !ok || !sameNamedOrigin(p, a) || p.TypeArgs().Len() != a.TypeArgs().Len() {
			return false
		}
		for i := 0; i < p.TypeArgs().Len(); i++ {
			if !inferMapperTypeArgs(p.TypeArgs().At(i), a.TypeArgs().At(i), allowed, bindings) {
				return false
			}
		}
		return true
	default:
		return shoot.TypeEquals(pattern, actual)
	}
}

func sameNamedOrigin(a, b *types.Named) bool {
	aobj := a.Origin().Obj()
	bobj := b.Origin().Obj()
	if aobj.Name() != bobj.Name() {
		return false
	}
	if aobj.Pkg() == nil || bobj.Pkg() == nil {
		return aobj.Pkg() == bobj.Pkg()
	}
	return aobj.Pkg().Path() == bobj.Pkg().Path()
}
