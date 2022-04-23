package main

import (
	"docker-swarm-visualiser/utils/mocks"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"fyne.io/fyne/v2/app"
	"github.com/cucumber/godog"
)

func TestMainStatusBar(t *testing.T) {
	MainApp = app.New()
	MainWindow = MainApp.NewWindow("Griffith Docker GUI")
	initial := statusBarTable
	mainStatusbar()
	if statusBarTable == initial {
		t.Fatalf("Status bar didn't update")
	}
}

func TestSetupApp(t *testing.T) {
	mocks.TestMode(&Docker)
	setupApp()
	tO := fmt.Sprintf("%v", reflect.TypeOf(statusBarTable.Objects[0]))
	if tO != "*widget.Button" {
		t.Fatalf("Failed to set the status bar initial context %v\n", tO)
	}
}

func TestBadDocker(t *testing.T) {
	mocks.PatchDockerForTesting(&Docker)
	mocks.AddCommandLines([]mocks.CommandStruct{
		// List context
		{Out: []byte(``), Err: errors.New("mep")},
	})
	setupApp()
	tO := fmt.Sprintf("%v", reflect.TypeOf(statusBarTable.Objects[0]))
	if tO == "*widget.Button" {
		t.Fatalf("Still set the status to a button even with docker unavailable %v\n", tO)
	}
}

// GO TESTS

func iLoadTheSecretsPage() error {
	return godog.ErrPending
}

func iShouldBeNotOwnedByMe(arg1 int) error {
	return godog.ErrPending
}

func iShouldSeeOwnedByMe(arg1 int) error {
	return godog.ErrPending
}

func iShouldSeeSecrets(arg1 int) error {
	return godog.ErrPending
}

func secretsNotAttachedToMyUser(arg1 int) error {
	return godog.ErrPending
}

func thereAreSecretsAttachedToMyUser(arg1 int) error {
	return godog.ErrPending
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^I load the secrets page$`, iLoadTheSecretsPage)
	ctx.Step(`^I should be (\d+) not owned by me$`, iShouldBeNotOwnedByMe)
	ctx.Step(`^I should see (\d+) owned by me$`, iShouldSeeOwnedByMe)
	ctx.Step(`^I should see (\d+) secrets$`, iShouldSeeSecrets)
	ctx.Step(`^(\d+) secrets not attached to my user$`, secretsNotAttachedToMyUser)
	ctx.Step(`^there are (\d+) secrets attached to my user$`, thereAreSecretsAttachedToMyUser)
}
