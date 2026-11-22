// Package acceptance runs the Gherkin scenarios in features/ with go test. It
// uses the real services over the in-memory adapters and a test identity
// provider, so it needs no external service.
package acceptance

import (
	"testing"

	"github.com/cucumber/godog"

	authdomain "backend/src/auth/domain"
)

// scenarioAdministrator creates the areas, permissions and memberships the
// scenarios need.
const scenarioAdministrator = "admin@fi.uba.ar"

// signedIn is someone who entered SIGTA once, as a membership requires.
func signedIn(id string) authdomain.User {
	return authdomain.User{ID: id, Email: id, Name: id}
}

// TestScenarios fails on a scenario that fails and, because of Strict, on a
// step that has no implementation: a scenario cannot pass without running.
func TestScenarios(t *testing.T) {
	suite := godog.TestSuite{
		// godog calls this once per scenario, so each one starts from scratch.
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			registerProcedureTypeSteps(sc)
			registerAuthenticationSteps(sc)
			registerCatalogSteps(sc)
			authorization := registerAuthorizationSteps(sc)
			registerStaffAreaSteps(sc, authorization)
			registerStaffAdministrationSteps(sc, authorization)
			confirmations := registerConfirmationSteps(sc, authorization)
			registerHistorySteps(sc, confirmations)
			registerCorrectionRequestSteps(sc, confirmations)
			myDetail := registerMyProcedureDetailSteps(sc, confirmations)
			registerMyProceduresSteps(sc, myDetail)
			registerCorrectionAnswerSteps(sc, myDetail)
			registerDetailSteps(sc, registerDocumentSteps(sc))
			registerSubmissionSteps(sc)
			registerInboxSteps(sc)
			registerProcedureTypeDraftSteps(sc)
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
