package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/go-cmd/cmd"
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

var StopStream bool

func (d *DockerClient) MakeWindowFollowCommand(a fyne.App, title string, command []string) {
	// Build the window
	w := a.NewWindow(title)
	data := binding.BindStringList(
		&[]string{},
	)
	w.SetContent(widget.NewListWithData(
		data,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(item binding.DataItem, co fyne.CanvasObject) {
			x, _ := item.(binding.String).Get()
			co.(*widget.Label).SetText(x)
		},
	))
	w.Show()
	// Run the command
	cmd := RunCmdStream(command)
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		// Done when both channels have been closed
		// https://dave.cheney.net/2013/04/30/curious-channels
		for cmd.Stdout != nil || cmd.Stderr != nil {
			select {
			case line, open := <-cmd.Stdout:
				if !open {
					cmd.Stdout = nil
					continue
				}
				data.Append(line)
			case line, open := <-cmd.Stderr:
				if !open {
					cmd.Stderr = nil
					continue
				}
				data.Append(line)
			}
		}
	}()

	// Run and wait for Cmd to return, discard Status
	<-cmd.Start()
	w.SetOnClosed(func() {
		cmd.Stop()
	})

	// Wait for goroutine to print everything
	<-doneChan
	log.Printf("We're done here %s", title)

}

// Create a new window for the logs to go in and get said logs
func (d *DockerClient) FollowLogs(a fyne.App, service string) {
	d.MakeWindowFollowCommand(
		a,
		fmt.Sprintf("Logs for %s", service),
		[]string{"service", "logs", "--no-trunc", "--follow", service},
	)
}

/**************/
var (
	RunCmd       func(commandArray []string) ([]byte, error)
	RunCmdStream func(commandArray []string) *cmd.Cmd
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
	StopStream = false
	RunCmd = func(commandArray []string) ([]byte, error) {
		cmd := exec.Command("docker", commandArray...)
		stdout, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		return stdout, err
	}
	RunCmdStream = func(commandArray []string) *cmd.Cmd {
		cmdOptions := cmd.Options{
			Buffered:  false,
			Streaming: true,
		}
		cmd := cmd.NewCmdOptions(cmdOptions, "docker", commandArray...)
		cmd.Start()
		return cmd
	}
}
