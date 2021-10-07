package gomodbom_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitGoModBOM(t *testing.T) {
	suite := spec.New("go-mod-bom", spec.Report(report.Terminal{}))
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite("ModuleBOM", testModuleBOM)
	suite.Run(t)
}
