package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lopolopen/shoot/internal/constructor"
	"github.com/lopolopen/shoot/internal/enumer"
	"github.com/lopolopen/shoot/internal/mapper"
	"github.com/lopolopen/shoot/internal/restclient"
	"github.com/lopolopen/shoot/internal/shoot"
	"github.com/lopolopen/shoot/internal/tools/logx"
)

var subCmdMap = map[string]string{
	constructor.SubCmd: "[-opt] [-getset] [-json] [-exp] [-tagcase=<case>] [-type=<Type> | -file=<GoFile>] [dir] [-s] [-v]",
	enumer.SubCmd:      "[-json] [-text] -[bit] [-json] [-type=<Type> | -file=<GoFile>] [dir] [-s] [-v]",
	restclient.SubCmd:  "[-type=<Type> | -file=<GoFile>] [dir] [-s] [-v]",
	mapper.SubCmd:      "[-path=<path>] [-alias=<alias>] [-to=<DestType>] [-type=<SrcType> | -file=<GoFile>] [dir] [-s] [-v]",
}

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		log.Println(`Usage: shoot <subcommand> [options]`)
		log.Println()
		log.Println(`These are all the sub commands supported as of now:`)
		log.Println()

		for _, sc := range []string{
			constructor.SubCmd,
			enumer.SubCmd,
			restclient.SubCmd,
			mapper.SubCmd,
		} {
			log.Printf("%s %s\n", sc, subCmdMap[sc])
			log.Println()
		}

		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	subCmd := flag.Args()[0]

	var g shoot.Generator
	switch subCmd {
	case "version":
		fmt.Println(shoot.Version)
		os.Exit(0)
	case constructor.SubCmd:
		g = constructor.New()
	case enumer.SubCmd:
		g = enumer.New()
	case restclient.SubCmd:
		g = restclient.New()
	case mapper.SubCmd:
		g = mapper.New()
	default:
		flag.Usage()
		os.Exit(2)
	}

	g.ParseFlags()
	g.LoadPackage()
	srcMap := g.Generate(g)
	var fileNames []string
	for fname, src := range srcMap {
		notedownSrc(fname, src)
		fileNames = append(fileNames, fname)
	}

	if len(srcMap) == 0 {
		logx.Warnf("nothing generated: [%s]", strings.Join(flag.Args(), " "))
		return
	}

	log.Printf("🎉 go generate successfully: [%s]\n", strings.Join(flag.Args(), " "))
	for _, fn := range fileNames {
		log.Printf("\t%s\n", fn)
	}
}

func notedownSrc(fileName string, src []byte) {
	// write to tmpfile first
	tmpFile, err := os.CreateTemp(".", fmt.Sprintf(".%s_", fileName))
	defer func() {
		if tmpFile != nil {
			_ = tmpFile.Close()
		}
	}()
	if err != nil {
		logx.Fatalf("creating temporary file for output: %s", err)
	}
	_, err = tmpFile.Write(src)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		logx.Fatalf("writing output: %s", err)
	}
	tmpFile.Close()

	// rename tmpfile to output file
	err = os.Rename(tmpFile.Name(), fileName)
	if err != nil {
		logx.Fatalf("moving tempfile to output file: %s", err)
	}
}
