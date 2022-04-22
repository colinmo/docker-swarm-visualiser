package mocks

type MockDocker struct {
}

var (
	RunCmdForCurrentContext func(commandArray []string) ([]byte, error)
)
