package shoot

import (
	"go/types"

	"golang.org/x/tools/go/packages"
)

func AssignableToIface(theTyp, ifaceTyp types.Type) (types.Type, bool) {
	stTyp := theTyp
	ptrTyp, ok := theTyp.(*types.Pointer)
	if ok {
		stTyp = ptrTyp.Elem()
	} else {
		ptrTyp = types.NewPointer(theTyp)
	}

	named, ok := stTyp.(*types.Named)
	if !ok {
		return nil, false
	}

	typeArgs := extractTypeArgs(named)

	ifaceNamed, ok := ifaceTyp.(*types.Named)
	if !ok {
		return nil, false
	}

	ifaceTPs := ifaceNamed.TypeParams()
	if ifaceTPs == nil || ifaceTPs.Len() == 0 {
		if len(typeArgs) == 0 {
			return ifaceTyp, types.AssignableTo(ptrTyp, ifaceTyp) || types.AssignableTo(stTyp, ifaceTyp)
		}
	}

	if ifaceTPs.Len() != len(typeArgs) {
		return nil, false
	}

	instantiated, err := types.Instantiate(nil, ifaceNamed, typeArgs, true)
	if err != nil {
		return nil, false
	}

	iface, ok := instantiated.Underlying().(*types.Interface)
	if !ok {
		return nil, false
	}

	return instantiated, types.AssignableTo(ptrTyp, iface) || types.AssignableTo(stTyp, iface)
}

func extractTypeArgs(named *types.Named) []types.Type {
	if named.TypeArgs() == nil {
		return nil
	}
	var args []types.Type
	for i := 0; i < named.TypeArgs().Len(); i++ {
		args = append(args, named.TypeArgs().At(i))
	}
	return args
}

func FindGetterSetterIfac(pkg *packages.Package, name string) (types.Type, types.Type) {
	getterName := name + Getter
	setterName := name + Setter
	var get, set types.Type

	scope := pkg.Types.Scope()

	if obj := scope.Lookup(getterName); obj != nil {
		if typeName, ok := obj.(*types.TypeName); ok {
			get = typeName.Type()
		}
	}

	if obj := scope.Lookup(setterName); obj != nil {
		if typeName, ok := obj.(*types.TypeName); ok {
			set = typeName.Type()
		}
	}
	return get, set
}

func ParseGetSetIface(pkg *packages.Package, theType types.Type, typeName string, funcs *[]Func) {
	get, set := FindGetterSetterIfac(pkg, typeName)
	if get == nil && set == nil {
		return
	}

	if named, ok := AssignableToIface(theType, get); ok {
		//todo: refac: REPEAT01
		iface, ok := named.Underlying().(*types.Interface)
		if ok {
			for i := 0; i < iface.NumMethods(); i++ {
				fn := iface.Method(i)
				sig := fn.Type().(*types.Signature)
				params := sig.Params()
				results := sig.Results()
				if params != nil && params.Len() > 0 {
					continue
				}
				if results == nil || results.Len() == 0 || results.Len() > 1 {
					continue
				}
				*funcs = append(*funcs, Func{
					Name:   fn.Name(),
					Result: results.At(0).Type(),
					Path:   fn.Name(),
				})
			}
		}
	}
	if named, ok := AssignableToIface(theType, set); ok {
		iface, ok := named.Underlying().(*types.Interface)
		if ok {
			for i := 0; i < iface.NumMethods(); i++ {
				fn := iface.Method(i)
				sig := fn.Type().(*types.Signature)
				params := sig.Params()
				results := sig.Results()
				if results != nil && results.Len() > 0 {
					continue
				}
				if params == nil || params.Len() == 0 || params.Len() > 1 {
					continue
				}
				*funcs = append(*funcs, Func{
					Name:  fn.Name(),
					Param: params.At(0).Type(),
					Path:  fn.Name(),
				})
			}
		}
	}
}
