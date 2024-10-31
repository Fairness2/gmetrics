// Статический анализатор, который проверяет код, содержит следующие анализаторы
//
// Состоит из
// - стандартных статических анализаторов пакета golang.org/x/tools/go/analysis/passes
// - всех анализаторов класса SA пакета staticcheck.io
// - ST1017 Don’t use Yoda conditions
// - ST1013 Should use constants for HTTP error codes, not magic numbers
// - QF1003 Convert if/else-if chain to tagged switch
// - QF1012 Use fmt.Fprintf(x, ...) instead of x.Write(fmt.Sprintf(...))
// - QF1009 Use time.Time.Equal instead of == operator
// - QF1011 Omit redundant type from variable declaration
// - github.com/kisielk/errcheck/errcheck check for unchecked errors
// - github.com/tdakkota/asciicheck checks that all code identifiers does not have non-ASCII symbols in the name
// - noosexit Check os.Exit() is used in main function of package
//
// Для сборки нужно запустить go build -o staticlint
// Для проверки кода достаточно запустить staticlint ./...
package main

import "gmetrics/internal/linter"

func main() {
	linter.Check()
}
