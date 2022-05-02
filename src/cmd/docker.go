package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
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

type Volume struct {
	Name       string
	Driver     string
	Scope      string
	Mountpoint string
	Labels     string
}

type Secret struct {
	ID        string
	Name      string
	CreatedAt string
	UpdatedAt string
	Labels    string
}

type DockerClient struct {
	Context  string
	Contexts []Context
	Services []Service
	Volumes  []Volume
	Secrets  []Secret
	Prefixes []string
}

type DockerSecret struct {
	Name  string
	Owner string
}

func (d *DockerClient) GetContexts() error {
	result, error := RunCmd(d.Context, []string{"context", "list", "--format", "json"})
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

func (d *DockerClient) GetVersion() (string, error) {
	output, err := d.RunCmdForCurrentContext([]string{"version", "--format", "{{.Server.Version}}"})
	return string(output), err
}

func (d *DockerClient) GetPrefixes() ([]string, error) {
	output, err := d.RunCmdForCurrentContext([]string{"service", "logs", "bob"})
	output_string := string(output)
	if output_string[0:26] == "Error response from daemon" {
		if output_string[0:61] != "Error response from daemon: This node is not a swarm manager." {
			index := strings.Index(output_string, "(")
			if index != -1 {
				d.Prefixes = []string{}
				translate := "[" + strings.Replace(output_string[index+1:len(output_string)-2], `'`, `"`, -1) + "]"
				err := json.Unmarshal([]byte(translate), &d.Prefixes)
				fmt.Printf("[%v]|[%v]\n", translate, d.Prefixes)
				return d.Prefixes, err
			}
		}
	}
	return []string{}, err
}

func (d *DockerClient) GetSecrets(stillActive func(string, string) bool, command string, me string) {
	output, err := d.RunCmdForCurrentContext([]string{"secret", "ls", "--format", `"{{.ID}}|~|{{.Name}}|~|{{.CreatedAt}}|~|{{.UpdatedAt}}|~|{{.Labels}}"`})
	if stillActive(command, me) {
		d.Secrets = []Secret{}
		if err == nil {
			for _, line := range strings.Split(string(output), "\n") {
				mep := strings.Split(line, "|~|")
				if len(mep) >= 5 && d.matchPrefixes(mep[1]) {
					d.Secrets = append(d.Secrets, Secret{
						ID:        mep[0],
						Name:      mep[1],
						CreatedAt: mep[2],
						UpdatedAt: mep[3],
						Labels:    mep[4],
					})
				}
			}
		}
	}
}

func (d *DockerClient) InspectSecret(secret string) (string, error) {
	output, err := d.RunCmdForCurrentContext([]string{"secret", "inspect", secret})
	return string(output), err
}

func (d *DockerClient) GetServices(stillActive func(string, string) bool, command string, me string) {
	output, err := d.RunCmdForCurrentContext([]string{"service", "list", "--format", `"{{.ID}}|~|{{.Name}}|~|{{.Mode}}|~|{{.Replicas}}|~|{{.Image}}|~|{{.Ports}}"`})
	if stillActive(command, me) {
		d.Services = []Service{}
		if err == nil {
			for _, line := range strings.Split(string(output), "\n") {
				mep := strings.Split(line, "|~|")
				if len(mep) >= 6 && d.matchPrefixes(mep[1]) {
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
}

func (d *DockerClient) matchPrefixes(lookAt string) bool {
	for _, x := range d.Prefixes {
		if len(lookAt) > len(x) && x == lookAt[0:len(x)] {
			return true
		}
	}
	return false
}

func (d *DockerClient) GetVolumes(stillActive func(string, string) bool, command string, me string) {
	output, err := d.RunCmdForCurrentContext([]string{"volume", "list", "--format", `{{.Name}}|~|{{.Driver}}|~|{{.Scope}}|~|{{.Mountpoint}}|~|{{.Labels}}`})
	if stillActive(command, me) {
		d.Volumes = []Volume{}
		if err == nil {
			for _, line := range strings.Split(string(output), "\n") {
				mep := strings.Split(line, "|~|")
				if len(mep) >= 5 && d.matchPrefixes(mep[0]) {
					d.Volumes = append(d.Volumes, Volume{
						Name:       mep[0],
						Driver:     mep[1],
						Scope:      mep[2],
						Mountpoint: mep[3],
						Labels:     mep[4],
					})
				}
			}
		}
	}
}

func (d *DockerClient) InspectVolume(volume string) (string, error) {
	output, err := d.RunCmdForCurrentContext([]string{"volume", "inspect", volume})
	return string(output), err
}

func (d *DockerClient) GetPS(service string) (string, error) {
	output, err := d.RunCmdForCurrentContext([]string{"service", "ps", service, "--no-trunc"})
	return string(output), err

}

var StopStream bool

// This should be altered to just return content
// The function calling this should have the window.
func (d *DockerClient) MakeWindowFollowCommand(a fyne.App, title string, command []string) {
	// Build the window
	w := a.NewWindow(title)
	w.Resize(fyne.NewSize(800, 600))
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
	cmd := RunCmdStream(d.Context, command)
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
}

// Create a new window for the logs to go in and get said logs
func (d *DockerClient) FollowLogs(a fyne.App, service string) {
	d.MakeWindowFollowCommand(
		a,
		fmt.Sprintf("Logs for %s", service),
		[]string{"service", "logs", "--no-trunc", "--follow", "--no-task-ids", service},
	)
}

/**************/
var (
	RunCmd       func(context string, commandArray []string) ([]byte, error)
	RunCmdStream func(context string, commandArray []string) *cmd.Cmd
)

func (d *DockerClient) RunCmdForCurrentContext(commandArray []string) ([]byte, error) {
	var stdout []byte
	var err error
	stdout, err = RunCmd(d.Context, commandArray)
	return stdout, err
}

func init() {
	StopStream = false
	RunCmd = func(context string, commandArray []string) ([]byte, error) {
		cmdPath := "docker"
		if exists("/usr/local/bin/docker") {
			cmdPath = "/usr/local/bin/docker"
		}
		cmd := exec.Command(cmdPath, append([]string{"--context", context}, commandArray...)...)
		stdout, err := cmd.CombinedOutput()
		return stdout, err
	}
	RunCmdStream = func(context string, commandArray []string) *cmd.Cmd {
		cmdPath := "docker"
		if exists("/usr/local/bin/docker") {
			cmdPath = "/usr/local/bin/docker"
		}
		cmdOptions := cmd.Options{
			Buffered:  false,
			Streaming: true,
		}
		cmd := cmd.NewCmdOptions(cmdOptions, cmdPath, append([]string{"--context", context}, commandArray...)...)
		cmd.Start()
		return cmd
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
