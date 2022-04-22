package mocks

import (
	"docker-swarm-visualiser/cmd"
	"log"
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
	cmd.RunCmd = func(commandArray []string) ([]byte, error) {
		if len(CommandLines) > 0 {
			toReturn := CommandLines[0]
			CommandLines = CommandLines[1:]
			return toReturn.Out, toReturn.Err
		} else {
			log.Print("No commands to return")
		}
		return nil, nil
	}
}
