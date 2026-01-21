package constructor

import (
	"go/types"

	"github.com/lopolopen/shoot/internal/shoot"
)

func (g *Generator) makeGetSet() {
	var getList []string
	var setList []string
	var getIfaces []string
	var setIfaces []string
	onceSet := shoot.MakeSet[string]()
	for _, f := range g.fields {
		if onceSet.Has(f.name) {
			continue
		}
		onceSet[f.name] = struct{}{}

		if f.isEmbeded {
			var get, set types.Type
			if g.getter || g.setter {
				get, set = shoot.FindGetterSetterIfac(g.Pkg(), f.name)
			}
			if g.getter && get != nil {
				named, ok := shoot.AssignableToIface(f.typ, get)
				if ok {
					getIfaces = append(getIfaces, types.TypeString(named, g.qualifier))

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
							g.getsetMethods = append(g.getsetMethods, shoot.Func{
								Name:   fn.Name(),
								Result: results.At(0).Type(),
								Path:   fn.Name(),
							})
						}
					}
				}
			}
			if g.setter && set != nil {
				named, ok := shoot.AssignableToIface(f.typ, set)
				if ok {
					setIfaces = append(setIfaces, types.TypeString(named, g.qualifier))
				}

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
						g.getsetMethods = append(g.getsetMethods, shoot.Func{
							Name:  fn.Name(),
							Param: params.At(0).Type(),
							Path:  fn.Name(),
						})
					}
				}
			}
			continue
		}

		if f.isGet && g.getter {
			getList = append(getList, f.name)
		}
		if f.isSet && g.setter {
			setList = append(setList, f.name)
		}
	}

	if g.getter {
		g.data.GetterIfaces = getIfaces
		g.data.GetterList = getList
	}
	if g.setter {
		g.data.SetterIfaces = setIfaces
		g.data.SetterList = setList
	}
	g.data.GetSet = g.flags.getset
}
