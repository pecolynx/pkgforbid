package main

import (
	"flag"
	"fmt"

	"golang.org/x/tools/go/analysis/singlechecker"
	"golang.org/x/tools/go/packages"

	"github.com/pecolynx/pkgforbid"
)

func makeDependencies(srcPkgPath string, dependencyList map[string]map[string]bool, dependencyMap map[string]map[string]bool) error {
	depList1, ok := dependencyMap[srcPkgPath]
	if !ok {
		return nil
	}

	if _, ok := dependencyList[srcPkgPath]; !ok {
		dependencyList[srcPkgPath] = make(map[string]bool)
	} else {
		return nil
	}

	for dep1 := range depList1 {
		if _, ok := dependencyList[dep1]; !ok {
			if err := makeDependencies(dep1, dependencyList, dependencyMap); err != nil {
				return err
			}
		}

		if depList2, ok := dependencyList[dep1]; ok {
			for dep2 := range depList2 {
				dependencyList[srcPkgPath][dep2] = true
			}
		}

		dependencyList[srcPkgPath][dep1] = true
	}

	return nil
}

func main() {
	conf := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(conf, "./...")
	if err != nil {
		panic(err)
	}

	dependencyMap := make(map[string]map[string]bool)
	for _, pkg := range pkgs {
		dependencyMap[pkg.PkgPath] = make(map[string]bool)
		for _, i := range pkg.Imports {
			dependencyMap[pkg.PkgPath][i.PkgPath] = true
		}
		if pkg.Name == "main" {
			dependencyMap[pkg.Name] = make(map[string]bool)
			for _, i := range pkg.Imports {
				dependencyMap[pkg.Name][i.PkgPath] = true
			}
		}
	}

	dependencyList := make(map[string]map[string]bool)
	for _, pkg := range pkgs {
		if err := makeDependencies(pkg.PkgPath, dependencyList, dependencyMap); err != nil {
			panic(err)
		}
	}
	// for src, v := range dependencyList {
	// 	for dst := range v {
	// 		fmt.Printf("%s -> %s\n", src, dst)
	// 	}
	// }

	dependencies := make(map[string]map[string]bool)
	for _, pkg := range pkgs {
		for _, i := range pkg.Imports {
			path := fmt.Sprintf("%s,%s", pkg.PkgPath, i.PkgPath)
			dependencies[path] = make(map[string]bool)
			for dep := range dependencyList[i.PkgPath] {
				dependencies[path][dep] = true
			}
		}
		if pkg.Name == "main" {
			for _, i := range pkg.Imports {
				path := fmt.Sprintf("%s,%s", pkg.Name, i.PkgPath)
				dependencies[path] = make(map[string]bool)
				for dep := range dependencyList[i.PkgPath] {
					dependencies[path][dep] = true
				}
			}
		}
	}

	// for src, v := range dependencies {
	// 	for dst := range v {
	// 		fmt.Printf("%s: %s\n", src, dst)
	// 	}
	// }

	pkgforbid.Dependencies = dependencies
	pkgforbid.ConfigFile = flag.String("config", "pkgforbid.yaml", "config")
	flag.Parse()
	singlechecker.Main(pkgforbid.Analyzer)
}
