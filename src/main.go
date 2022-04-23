package main

import (
	"docker-swarm-visualiser/cmd"
	"docker-swarm-visualiser/utils/mocks"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var Docker cmd.DockerClient

func init() {
	Docker = cmd.DockerClient{}
}

func main() {
	setupApp()
	MainWindow.ShowAndRun()
}

var statusBar = [][]string{{"Context", "Busy", "Active"}}
var statusBarTable *widget.Table
var MainApp fyne.App
var MainWindow fyne.Window

func setupApp() {
	// TESTING
	mocks.PatchDockerForTesting(&Docker)
	mocks.AddCommandLines([]mocks.CommandStruct{
		// List context
		{Out: []byte(`[{"Current":true,"Description":"Current DOCKER_HOST based configuration","DockerEndpoint":"npipe:////./pipe/docker_engine","KubernetesEndpoint":"","ContextType":"moby","Name":"default","StackOrchestrator":"swarm"},{"Current":false,"Description":"","DockerEndpoint":"npipe:////./pipe/dockerDesktopLinuxEngine","KubernetesEndpoint":"","ContextType":"moby","Name":"desktop-linux","StackOrchestrator":""}]`), Err: nil},
		// Set context
		{Out: []byte(`[{"Current":true,"Description":"Current DOCKER_HOST based configuration","DockerEndpoint":"npipe:////./pipe/docker_engine","KubernetesEndpoint":"","ContextType":"moby","Name":"default","StackOrchestrator":"swarm"},{"Current":false,"Description":"","DockerEndpoint":"npipe:////./pipe/dockerDesktopLinuxEngine","KubernetesEndpoint":"","ContextType":"moby","Name":"desktop-linux","StackOrchestrator":""}]`), Err: nil},
		// List services
		{Out: []byte("36xvvwwauej0|~|frontend|~|replicated|~|5/5|~|nginx:alpine|~|80\n74nzcxxjv6fq|~|backend|~|replicated|~|3/3|~|redis:3.0.6|~|443\n"), Err: nil},
		// Reset context
		{Out: []byte(""), Err: nil},
	})
	// END TESTING
	MainApp = app.New()
	MainWindow = MainApp.NewWindow("Hello World")
	MainWindow.Resize(fyne.NewSize(400, 400))
	mainStatusbar()
	var originalContent fyne.CanvasObject

	err := Docker.GetContexts()
	if err == nil {
		Docker.GetServices()
		x := ""
		for _, context := range Docker.Contexts {
			x = x + context.Name + "\n"
		}
		originalContent = serviceToVBox() // serviceToList()
	} else {
		originalContent = widget.NewLabel("Can't access docker")
	}

	statusBar[0][0] = Docker.Context
	content := container.NewBorder(
		mainToolbar(),
		statusBarTable,
		nil,
		nil,
		originalContent,
	)
	MainWindow.SetContent(content)
}

func mainToolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.HomeIcon(), func() {
			log.Printf("Switch Context")
		}),
	)
}

func mainStatusbar() {
	statusBarTable = widget.NewTable(
		func() (int, int) {
			return 1, 3
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("wide content")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(statusBar[i.Row][i.Col])
		})
}

func serviceToVBox() *fyne.Container {
	me := container.New(layout.NewVBoxLayout())
	for _, item := range Docker.Services {
		service := item.Name
		me.Add(container.New(
			layout.NewHBoxLayout(),
			widget.NewLabel(service),
			layout.NewSpacer(),
			widget.NewButton("PS", func() { log.Printf("PS: %s\n", service) }),
			widget.NewButton("Logs", func() {
				go func() {
					Docker.FollowLogs(MainApp, service)
				}()
				log.Printf("Logs: %s\n", service)
			}),
		))
	}
	return me
}
