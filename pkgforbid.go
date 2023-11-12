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
	debug          bool `yaml:"debug"`
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if appConfig.debug {
		fmt.Println("---")
	}

	if packageConfig, ok := appConfig.packageConfigs[pass.Pkg.Path()]; ok {
		for _, f := range pass.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				if x, ok := n.(*ast.ImportSpec); ok {
					importedPackage := trimQuotes(x.Path.Value)

					if _, ok := packageConfig.forbiddenPackages[importedPackage]; ok {
						pass.Reportf(x.Pos(), "imported forbidden package: %s", importedPackage)
					} else if appConfig.debug {
						fmt.Printf("%s is not forbidden in %s\n", importedPackage, pass.Pkg.Path())
					}
				}
				return true
			})
		}
	} else if appConfig.debug {
		fmt.Printf("config for %s is not found\n", pass.Pkg.Path())
	}

	return nil, nil
}
