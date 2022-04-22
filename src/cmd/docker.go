package cmd

import (
	"encoding/json"
	"os/exec"
	"strings"
)

type Context struct {
	Current            bool   `json:"Current"`
	Description        string `json:"Description"`
	DockerEndpoint     string `json:"DockerEndpoint"`
	KubernetesEndpoint string `json:"KubernetesEndpoint"`
	ContextType        string `json:"ContextType"`
	Name               string `json:"Name"`
	StackOrchestrator  string `json:"StackOrchestrator"`
}

type Service struct {
	ID       string
	Name     string
	Mode     string
	Replicas string
	Image    string
	Ports    string
}

type DockerClient struct {
	Context  string
	Contexts []Context
	Services []Service
}

type DockerSecret struct {
	Name  string
	Owner string
}

func (d *DockerClient) GetSecrets() []DockerSecret {
	var dockerSecrets []DockerSecret

	return dockerSecrets
}

func (d *DockerClient) GetContexts() error {
	result, error := RunCmd([]string{"context", "list", "--format", "json"})
	if error == nil {
		json.Unmarshal([]byte(result), &d.Contexts)
		for _, context := range d.Contexts {
			if context.Current {
				d.Context = context.Name
			}
		}
	}
	return error
}

func (d *DockerClient) SwitchContext(context string) {
	RunCmd([]string{"context", "use", context})
}

func (d *DockerClient) GetServices() {
	output, err := d.RunCmdForCurrentContext([]string{"service", "list", "--format", `"{{.ID}}|~|{{.Name}}|~|{{.Mode}}|~|{{.Replicas}}|~|{{.Image}}|~|{{.Ports}}"`})
	d.Services = []Service{}
	if err == nil {
		for _, line := range strings.Split(string(output), "\n") {
			mep := strings.Split(line, "|~|")
			if len(mep) >= 6 {
				d.Services = append(d.Services, Service{
					ID:       mep[0],
					Name:     mep[1],
					Mode:     mep[2],
					Replicas: mep[3],
					Image:    mep[4],
					Ports:    mep[5],
				})
			}
		}
	}
}

var (
	RunCmd func(commandArray []string) ([]byte, error)
)

func (d *DockerClient) RunCmdForCurrentContext(commandArray []string) ([]byte, error) {
	var l DockerClient
	var stdout []byte
	var err error
	l.GetContexts()
	restoreContext := ""
	if l.Context != d.Context {
		restoreContext = d.Context
		d.SwitchContext(d.Context)
		stdout, err = RunCmd(commandArray)
		d.SwitchContext(restoreContext)
	} else {
		stdout, err = RunCmd(commandArray)

	}
	return stdout, err
}

func init() {
	RunCmd = func(commandArray []string) ([]byte, error) {
		cmd := exec.Command("docker", commandArray...)
		stdout, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		return stdout, err
	}
}
