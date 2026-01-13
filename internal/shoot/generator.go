package shoot

import "golang.org/x/tools/go/packages"

type Generator interface {
	DataMaker
	TypeLister

	ParseFlags()

	LoadPackage(patterns ...string) map[string]*packages.Package

	Generate(gen interface {
		TypeLister
		DataMaker
	}) map[string][]byte

	Clean() error
}

type DataMaker interface {
	MakeData(typName string) (any, bool)
}

type TypeLister interface {
	ListTypes() []string
}
