package pkgforbid

import (
	"fmt"
	"go/ast"
	"io"
	"os"

	"github.com/go-playground/validator/v10"
	"golang.org/x/tools/go/analysis"
	"gopkg.in/yaml.v3"
)

var (
	appValidator = validator.New()

	ConfigFile *string

	Dependencies map[string]map[string]bool

	Analyzer = &analysis.Analyzer{
		Name: "pkgforbid",
		Doc:  "pkgforbid",
		Run:  run,
	}
)

func trimQuotes(s string) string {
	if len(s) >= 2 {
		switch {
		case s[0] == '"' && s[len(s)-1] == '"':
			return s[1 : len(s)-1]
		case s[0] == '\'' && s[len(s)-1] == '\'':
			return s[1 : len(s)-1]
		}
	}
	return s
}

type filePackageConfig struct {
	ForbiddenPackages []string `yaml:"forbiddenPackages" validate:"required"`
}

type fileConfig struct {
	PackageConfigs map[string]filePackageConfig `yaml:"packageConfigs" validate:"required"`
	Debug          bool                         `yaml:"debug"`
}

type appPackageConfig struct {
	forbiddenPackages map[string]bool
}

type appConfig struct {
	packageConfigs map[string]appPackageConfig
	debug          bool
}

func loadConfig() (*appConfig, error) {
	configFilePath := "pkgforbid.yaml"
	if ConfigFile != nil {
		configFilePath = *ConfigFile
	}

	f, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	confContent, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	fileConfig := fileConfig{}
	if err := yaml.Unmarshal(confContent, &fileConfig); err != nil {
		return nil, err
	}

	if err := appValidator.Struct(&fileConfig); err != nil {
		return nil, err
	}

	packageConfigs := map[string]appPackageConfig{}
	for importer, filePackageConfig := range fileConfig.PackageConfigs {
		appPackageConfig := appPackageConfig{
			forbiddenPackages: make(map[string]bool),
		}
		for _, importee := range filePackageConfig.ForbiddenPackages {
			appPackageConfig.forbiddenPackages[importee] = true
		}
		packageConfigs[importer] = appPackageConfig
	}

	appConfig := appConfig{
		packageConfigs: packageConfigs,
		debug:          fileConfig.Debug,
	}

	return &appConfig, nil
}

func run(pass *analysis.Pass) (interface{}, error) {
	appConfig, err := loadConfig()
	if err != nil {
		return nil, err
	}

	if appConfig.debug {
		fmt.Println("---")
	}

	_run := func(packageConfig appPackageConfig, importerPackage string) {
		for _, f := range pass.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				if x, ok := n.(*ast.ImportSpec); ok {
					importedPackage := trimQuotes(x.Path.Value)

					if _, ok := packageConfig.forbiddenPackages[importedPackage]; ok {
						pass.Reportf(x.Pos(), "imported forbidden package: %s", importedPackage)
					} else if appConfig.debug {
						fmt.Printf("%s is not forbidden in %s\n", importedPackage, importerPackage)
					}

					path := fmt.Sprintf("%s,%s", importerPackage, importedPackage)
					for dep := range Dependencies[path] {
						if _, ok := packageConfig.forbiddenPackages[dep]; ok {
							pass.Reportf(x.Pos(), "imported forbidden package: %s", dep)
						} else if appConfig.debug {
							fmt.Printf("%s is not forbidden in %s\n", dep, importerPackage)
						}
					}
				}
				return true
			})
		}
	}

	found := false
	srcs := make([]string, 0)
	srcs = append(srcs, pass.Pkg.Path())
	if pass.Pkg.Name() == "main" {
		srcs = append(srcs, pass.Pkg.Name())
	}

	for _, src := range srcs {
		if packageConfig, ok := appConfig.packageConfigs[src]; ok {
			found = true
			_run(packageConfig, src)
		}
	}

	if !found && appConfig.debug {
		fmt.Printf("config for %s is not found\n", pass.Pkg.Path())
	}

	return nil, nil
}
