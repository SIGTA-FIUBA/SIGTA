// Package acceptance runs the Gherkin scenarios in features/ with go test. It
// uses the real services over the in-memory and development adapters, so it
// needs no external service.
package acceptance

import (
	"testing"

	"github.com/cucumber/godog"
)

// TestScenarios fails on a scenario that fails and, because of Strict, on a
// step that has no implementation: a scenario cannot pass without running.
func TestScenarios(t *testing.T) {
	suite := godog.TestSuite{
		// godog calls this once per scenario, so each one starts from scratch.
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			registerProcedureTypeSteps(sc)
			registerAuthenticationSteps(sc)
			registerCatalogSteps(sc)
		},
		Options: &godog.Options{
			Format:   "progress",
			Paths:    []string{"../../../../features"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("the scenarios failed")
	}
}
