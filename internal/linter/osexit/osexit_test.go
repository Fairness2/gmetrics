package osexit

import (
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

func TestRun(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "./...")
}
