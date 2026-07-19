package linter

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer проверяет, что анализатор выдаёт нужные диагностики в тестовых пакетах.
func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "exitpkg", "fatalpkg", "panicpkg", "mainok")
}
