package shoot

// import (
// 	"bytes"
// 	"fmt"
// 	"go/types"
// 	"strings"

// 	"github.com/lopolopen/shoot/internal/tools/logx"
// )

// type Iface struct {
// 	Name string
// 	// TypeParams
// 	Funcs map[string]*Func
// }

// func NewIface(name string) *Iface {
// 	return &Iface{
// 		Name: name,
// 	}
// }

// func (i *Iface) String() string {
// 	var buf bytes.Buffer
// 	buf.WriteString(fmt.Sprintf("type %s interface {\n", i.Name))
// 	qf := func(*types.Package) string { return "" }
// 	for _, f := range i.Funcs {
// 		if f.IsGetter() {
// 			buf.WriteString(fmt.Sprintf("\t%s() %s\n", f.Name, types.TypeString(f.Result, qf)))
// 		} else if f.IsSetter() {
// 			buf.WriteString(fmt.Sprintf("\t%s(%s)\n", f.Name, types.TypeString(f.Param, qf)))
// 		}
// 	}
// 	buf.WriteString("}\n")
// 	return buf.String()
// }

// func (i *Iface) Exists(fn *Func) bool {
// 	if i.Funcs == nil {
// 		return false
// 	}
// 	fn0, ok := i.Funcs[fn.Name]
// 	if !ok {
// 		return false
// 	}
// 	return fn0.Equals(fn)
// }

// func (i *Iface) Add(fn *Func) {
// 	if fn == nil {
// 		return
// 	}
// 	if i.Funcs == nil {
// 		i.Funcs = make(map[string]*Func)
// 	}
// 	//same name, but not identical?
// 	i.Funcs[fn.Name] = fn
// }

// type IfaceRegiser struct {
// 	ifaceMap map[string]*Iface
// }

// func NewIfaceRegister() *IfaceRegiser {
// 	return &IfaceRegiser{
// 		ifaceMap: make(map[string]*Iface),
// 	}
// }

// func (ir *IfaceRegiser) AddGetter(name string, fn *Func) {
// 	if !strings.HasSuffix(name, Getter) {
// 		logx.Fatalf("invalid getter name: %s", name)
// 	}
// 	if fn == nil {
// 		return
// 	}
// 	if !fn.IsGetter() {
// 		logx.Fatalf("add a non-getter func to %s: %s", name, fn.Name)
// 	}
// 	iface, ok := ir.ifaceMap[name]
// 	if !ok {
// 		iface = NewIface(name)
// 		ir.ifaceMap[name] = iface
// 	}
// 	iface.Add(fn)
// }

// func (ir *IfaceRegiser) AddSetter(name string, fn *Func) {
// 	if !strings.HasSuffix(name, Setter) {
// 		logx.Fatalf("invalid setter name: %s", name)
// 	}
// 	if fn == nil {
// 		return
// 	}
// 	if !fn.IsSetter() {
// 		logx.Fatalf("add a non-setter func to %s: %s", name, fn.Name)
// 	}
// 	iface, ok := ir.ifaceMap[name]
// 	if !ok {
// 		iface = NewIface(name)
// 		ir.ifaceMap[name] = iface
// 	}
// 	iface.Add(fn)
// }

// func (ir *IfaceRegiser) Get(name string) (*Iface, bool) {
// 	i, ok := ir.ifaceMap[name]
// 	return i, ok
// }

// func (ir *IfaceRegiser) IsAssignableTo(name1, name2 string) bool {
// 	iface1 := ir.ifaceMap[name1]
// 	iface2 := ir.ifaceMap[name2]
// 	if iface1 == nil || iface2 == nil {
// 		return false
// 	}
// 	for _, f := range iface2.Funcs {
// 		if !iface1.Exists(f) {
// 			return false
// 		}
// 	}
// 	return true
// }

// func (ir *IfaceRegiser) Exists(name string) bool {
// 	_, ok := ir.ifaceMap[name]
// 	return ok
// }

// func (ir *IfaceRegiser) Embeds(derivedName string, baseNmae string) {
// 	base, ok := ir.Get(baseNmae)
// 	if !ok {
// 		return
// 	}

// 	for _, f := range base.Funcs {
// 		if f.IsGetter() {
// 			ir.AddGetter(derivedName, f)
// 		} else if f.IsSetter() {
// 			ir.AddSetter(derivedName, f)
// 		}
// 	}
// }
