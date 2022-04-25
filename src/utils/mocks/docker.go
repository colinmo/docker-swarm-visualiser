package mocks

import (
	"docker-swarm-visualiser/cmd"
	"log"

	cmd2 "github.com/go-cmd/cmd"
)

type CommandStruct struct {
	Out []byte
	Err error
}

var CommandLines []CommandStruct

func AddCommandLines(lines []CommandStruct) {
	CommandLines = append(CommandLines, lines...)
}

func PatchDockerForTesting(d *cmd.DockerClient) {
	cmd.RunCmd = func(context string, commandArray []string) ([]byte, error) {
		if len(CommandLines) > 0 {
			toReturn := CommandLines[0]
			CommandLines = CommandLines[1:]
			return toReturn.Out, toReturn.Err
		} else {
			log.Print("No commands to return")
		}
		return nil, nil
	}

	cmd.RunCmdStream = func(contest string, commandArray []string) *cmd2.Cmd {
		cmdOptions := cmd2.Options{
			Buffered:  false,
			Streaming: true,
		}
		cmd := cmd2.NewCmdOptions(cmdOptions, "php", "-r", `echo "hi\n";sleep(10);echo "dude\n";sleep(10);`)
		cmd.Start()
		return cmd
	}
}

func TestMode(docker *cmd.DockerClient) {
	PatchDockerForTesting(docker)
	AddCommandLines([]CommandStruct{
		// List context
		{Out: []byte(`[{"Current":true,"Description":"Current DOCKER_HOST based configuration","DockerEndpoint":"npipe:////./pipe/docker_engine","KubernetesEndpoint":"","ContextType":"moby","Name":"default","StackOrchestrator":"swarm"},{"Current":false,"Description":"","DockerEndpoint":"npipe:////./pipe/dockerDesktopLinuxEngine","KubernetesEndpoint":"","ContextType":"moby","Name":"desktop-linux","StackOrchestrator":""}]`), Err: nil},
		// Set context
		{Out: []byte(`[{"Current":true,"Description":"Current DOCKER_HOST based configuration","DockerEndpoint":"npipe:////./pipe/docker_engine","KubernetesEndpoint":"","ContextType":"moby","Name":"default","StackOrchestrator":"swarm"},{"Current":false,"Description":"","DockerEndpoint":"npipe:////./pipe/dockerDesktopLinuxEngine","KubernetesEndpoint":"","ContextType":"moby","Name":"desktop-linux","StackOrchestrator":""}]`), Err: nil},
		// List services
		{Out: []byte("36xvvwwauej0|~|frontend|~|replicated|~|5/5|~|nginx:alpine|~|80\n74nzcxxjv6fq|~|backend|~|replicated|~|3/3|~|redis:3.0.6|~|443\n"), Err: nil},
		// Reset context
		{Out: []byte(""), Err: nil},
	})
}
