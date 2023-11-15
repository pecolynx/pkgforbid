package pkgforbid

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer is a analyzerTest for Analyzer.
func TestAnalyzer(t *testing.T) {
	Dependencies = map[string]map[string]bool{
		"a,b": {
			"net/http": true,
		},
	}
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "a")
}
