package main

import (
	"flag"
	"os"
	"strings"
	"testing"

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
		cmd: "shoot map -path=./dest -type=Order",
		names: []string{
			"map_src.shootmap.order.go",
		},
	},
	{
		cmd: "shoot map -path=./dest -alias=target -to=Dest -type=Src",
		names: []string{
			"map_src.shootmap.src.go",
		},
	},
}

func TestShootNew_Golden(t *testing.T) {
	const dir = "./testdata/constructor"
	g := goldie.New(t,
		goldie.WithFixtureDir(dir),
	)

	for _, test := range goldens_new {
		srcMap := generate(test, constructor.New(), dir)

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
	const dir = "./testdata/enumer"
	g := goldie.New(t,
		goldie.WithFixtureDir(dir),
	)

	for _, test := range goldens_enum {
		srcMap := generate(test, enumer.New(), dir)

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
	const dir = "./testdata/restclient"
	g := goldie.New(t,
		goldie.WithFixtureDir(dir),
	)

	for _, test := range goldens_rest {
		srcMap := generate(test, restclient.New(), dir)

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
	const dir = "./testdata/mapper"
	g := goldie.New(t,
		goldie.WithFixtureDir(dir),
	)

	for _, test := range goldens_map {
		srcMap := generate(test, mapper.New(), dir)

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

func generate(test Golden, g shoot.Generator, dir string) map[string][]byte {
	os.Args = strings.Fields(test.cmd)
	os.Args = append(os.Args, "-version=test", dir)
	flag.Parse()

	g.ParseFlags()
	g.LoadPackage()
	srcMap := g.Generate(g)
	return srcMap
}
