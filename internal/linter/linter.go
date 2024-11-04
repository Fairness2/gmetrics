package linter

import (
	"github.com/kisielk/errcheck/errcheck"
	"github.com/tdakkota/asciicheck"
	"gmetrics/internal/linter/osexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

var (
	styleChecks = map[string]bool{
		"ST1017": true, // Don’t use Yoda conditions
		"ST1013": true, // Should use constants for HTTP error codes, not magic numbers
	}
	quickFixes = map[string]bool{
		"QF1003": true, // Convert if/else-if chain to tagged switch
		"QF1012": true, // Use fmt.Fprintf(x, ...) instead of x.Write(fmt.Sprintf(...))
		"QF1009": true, // Use time.Time.Equal instead of == operator
		"QF1011": true, // Omit redundant type from variable declaration
	}
)

var Analyzers = []*analysis.Analyzer{
	asmdecl.Analyzer,
	assign.Analyzer,
	atomic.Analyzer,
	atomicalign.Analyzer,
	bools.Analyzer,
	buildssa.Analyzer,
	buildtag.Analyzer,
	cgocall.Analyzer,
	composite.Analyzer,
	copylock.Analyzer,
	ctrlflow.Analyzer,
	deepequalerrors.Analyzer,
	directive.Analyzer,
	errorsas.Analyzer,
	//fieldalignment.Analyzer, // Проверяет структуры на выравнивание памяти
	findcall.Analyzer,
	framepointer.Analyzer,
	httpresponse.Analyzer,
	ifaceassert.Analyzer,
	inspect.Analyzer,
	loopclosure.Analyzer,
	lostcancel.Analyzer,
	nilfunc.Analyzer,
	nilness.Analyzer,
	pkgfact.Analyzer,
	printf.Analyzer,
	reflectvaluecompare.Analyzer,
	shadow.Analyzer,
	shift.Analyzer,
	sigchanyzer.Analyzer,
	sortslice.Analyzer,
	stdmethods.Analyzer,
	stringintconv.Analyzer,
	structtag.Analyzer,
	testinggoroutine.Analyzer,
	tests.Analyzer,
	timeformat.Analyzer,
	unmarshal.Analyzer,
	unreachable.Analyzer,
	unsafeptr.Analyzer,
	unusedresult.Analyzer,
	unusedwrite.Analyzer,
	usesgenerics.Analyzer,
	errcheck.Analyzer,
	asciicheck.NewAnalyzer(),
	osexit.Analyzer,
}

func Check() {
	mychecks := Analyzers

	for _, v := range staticcheck.Analyzers {
		// Analyzers включает в себя только SA анализиторы
		mychecks = append(mychecks, v.Analyzer)
	}
	for _, st := range stylecheck.Analyzers {
		// Analyzers включает в себя только ST анализиторы
		if styleChecks[st.Analyzer.Name] {
			mychecks = append(mychecks, st.Analyzer)
		}
	}
	for _, qf := range quickfix.Analyzers {
		// Analyzers включает в себя только QF анализиторы
		if quickFixes[qf.Analyzer.Name] {
			mychecks = append(mychecks, qf.Analyzer)
		}
	}

	multichecker.Main(
		mychecks...,
	)
}
