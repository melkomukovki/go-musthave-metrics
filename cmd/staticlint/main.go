package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/melkomukovki/go-musthave-metrics/internal/customanalyzers"
)

func main() {
	analyzers := make([]*analysis.Analyzer, 10)

	analyzers = append(analyzers, shadow.Analyzer, bools.Analyzer, defers.Analyzer, nilness.Analyzer)

	for _, a := range staticcheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	for _, a := range stylecheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	for _, a := range simple.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	for _, a := range quickfix.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	analyzers = append(analyzers, customanalyzers.OsExitCheckAnalyzer)

	multichecker.Main(analyzers...)
}
