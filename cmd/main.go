package main

import (
	"flag"

	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/pecolynx/pkgforbid"
)

func main() {
	pkgforbid.ConfigFile = flag.String("config", "pkgforbid.yaml", "config")
	flag.Parse()
	singlechecker.Main(pkgforbid.Analyzer)
}
