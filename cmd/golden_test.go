package main

import (
	"encoding/json"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/lopolopen/shoot/cmd/test/ctor"
	"github.com/lopolopen/shoot/internal/constructor"
	"github.com/lopolopen/shoot/internal/enumer"
	"github.com/lopolopen/shoot/internal/mapper"
	"github.com/lopolopen/shoot/internal/restclient"
	"github.com/lopolopen/shoot/internal/shoot"
	"github.com/sebdah/goldie/v2"
)

type Golden struct {
	cmd   string
	names []string
}

var goldens_new = []Golden{
	// {
	// 	cmd: "shoot new -getset -type=User ./notexists",
	// 	names: []string{
	// 		"new_getset.go.user.go",
	// 	},
	// },
	{
		cmd: "shoot new -getset -type=User",
		names: []string{
			"new_getset.shootnew.user.go",
		},
	},
	{
		cmd: "shoot new -opt -type=Conf",
		names: []string{
			"new_opt.shootnew.conf.go",
		},
	},
	{
		cmd: "shoot new -file=new_file.go",
		names: []string{
			"new_file.shootnew.go",
		},
	},
	{
		cmd: "shoot new -getset -file=new_getset2.go",
		names: []string{
			"new_getset2.shootnew.go",
		},
	},
}

var goldens_enum = []Golden{
	{
		cmd:   "shoot enum -bit -file=nothing.go",
		names: []string{},
	},
	{
		cmd: "shoot enum -bit -type=FormatStyle",
		names: []string{
			"enum_bit.shootenum.formatstyle.go",
		},
	},
	{
		cmd: "shoot enum -json -type=Color",
		names: []string{
			"enum_json.shootenum.color.go",
		},
	},
}

var goldens_rest = []Golden{
	{
		cmd: "shoot rest -type=Client",
		names: []string{
			"rest.shootrest.client.go",
		},
	},
}

var goldens_map = []Golden{
	{
		cmd: "shoot map -path=../dest -type=Order",
		names: []string{
			"map_src.shootmap.order.go",
		},
	},
	{
		cmd: "shoot map -path=../dest -type=Order2",
		names: []string{
			"map_src.shootmap.order2.go",
		},
	},
	{
		cmd: "shoot map -path=../dest -alias=target -to=Dest -type=Src",
		names: []string{
			"map_src.shootmap.src.go",
		},
	},
}

const datadir = "./testdata"

func TestShootNew_Golden(t *testing.T) {
	const codedir = "./test/ctor"

	g := goldie.New(t,
		goldie.WithFixtureDir(datadir),
	)

	for _, test := range goldens_new {
		srcMap := generate(test, constructor.New(), codedir)

		for _, name := range test.names {
			got, ok := srcMap[name]
			if !ok {
				var keys []string
				for key := range srcMap {
					keys = append(keys, key)
				}
				t.Errorf("expected file: %s, got: %v", name, keys)
			}
			g.Assert(t, name, got)
		}
	}
}

func TestShootEnum_Golden(t *testing.T) {
	const codedir = "./test/enumer"
	g := goldie.New(t,
		goldie.WithFixtureDir(datadir),
	)

	for _, test := range goldens_enum {
		srcMap := generate(test, enumer.New(), codedir)

		if len(srcMap) != len(test.names) {
			t.Errorf("expected count: %d, got: %d", len(test.names), len(srcMap))
		}

		for _, name := range test.names {
			got, ok := srcMap[name]
			if !ok {
				var keys []string
				for key := range srcMap {
					keys = append(keys, key)
				}
				t.Errorf("expected file: %s, got: %v", name, keys)
			}
			g.Assert(t, name, got)
		}
	}
}

func TestShootRest_Golden(t *testing.T) {
	const codedir = "./test/restclient"
	g := goldie.New(t,
		goldie.WithFixtureDir(datadir),
	)

	for _, test := range goldens_rest {
		srcMap := generate(test, restclient.New(), codedir)

		if len(srcMap) != len(test.names) {
			t.Errorf("expected count: %d, got: %d", len(test.names), len(srcMap))
		}

		for _, name := range test.names {
			got, ok := srcMap[name]
			if !ok {
				var keys []string
				for key := range srcMap {
					keys = append(keys, key)
				}
				t.Errorf("expected file: %s, got: %v", name, keys)
			}
			g.Assert(t, name, got)
		}
	}
}

func TestShootMap_Golden(t *testing.T) {
	const codedir = "./test/mapper/src"
	g := goldie.New(t,
		goldie.WithFixtureDir(datadir),
	)

	for _, test := range goldens_map {
		srcMap := generate(test, mapper.New(), codedir)

		if len(srcMap) != len(test.names) {
			t.Errorf("expected count: %d, got: %d", len(test.names), len(srcMap))
		}

		for _, name := range test.names {
			got, ok := srcMap[name]
			if !ok {
				var keys []string
				for key := range srcMap {
					keys = append(keys, key)
				}
				t.Errorf("expected file: %s, got: %v", name, keys)
			}
			g.Assert(t, name, got)
		}
	}
}

func TestShootNew_JSON_Golden(t *testing.T) {
	const name = "new_json.json"
	g := goldie.New(t,
		goldie.WithFixtureDir(datadir),
	)
	obj := ctor.NewSonJSON("Z", 1, "A", 2, 3, ctor.Other{
		Name: "cz",
		Age:  18,
	})
	j, _ := json.MarshalIndent(obj, "", "  ")
	g.Assert(t, name, j)
}

func generate(test Golden, g shoot.Generator, dir string) map[string][]byte {
	os.Args = strings.Fields(test.cmd)
	os.Args = append(os.Args, "-version=test", dir)
	flag.Parse()

	g.ParseFlags()
	g.LoadPackage()
	srcMap := g.Generate(g)
	return srcMap
}
