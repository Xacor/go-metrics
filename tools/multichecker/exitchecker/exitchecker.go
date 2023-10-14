package exitchecker

import "golang.org/x/tools/go/analysis"

var ExitChecker = &analysis.Analyzer{
	Name: "exitchecker",
	Doc:  "check for unchecked errors",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// реализация будет ниже
	return nil, nil
}
