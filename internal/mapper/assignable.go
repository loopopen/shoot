package mapper

import (
	"go/types"
)

func assignableToIface(ptrTyp, ifaceTyp types.Type) (types.Type, bool) {
	// ptrTyp: *entity.Base[uint]
	// ifacTyp: entity.BaseGetter[T ID]

	ptr, ok := ptrTyp.(*types.Pointer)
	if !ok {
		return nil, false
	}

	named, ok := ptr.Elem().(*types.Named)
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
			return ifaceTyp, types.AssignableTo(ptrTyp, ifaceTyp)
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

	return instantiated, types.AssignableTo(ptrTyp, iface)
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
