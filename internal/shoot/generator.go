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
}

type DataMaker interface {
	MakeData(typName string) any
}

type TypeLister interface {
	ListTypes() []string
}
