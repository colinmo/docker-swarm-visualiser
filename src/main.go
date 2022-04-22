package main

import (
	"docker-swarm-visualiser/cmd"
	"docker-swarm-visualiser/utils/mocks"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var Docker cmd.DockerClient

func init() {
	Docker = cmd.DockerClient{}
}

func main() {
	_, w := setupApp()
	w.ShowAndRun()
}

var statusBar = [][]string{{"Context", "Busy", "Active"}}
var statusBarTable *widget.Table

func setupApp() (fyne.App, fyne.Window) {
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
	a := app.New()
	w := a.NewWindow("Hello World")
	w.Resize(fyne.NewSize(400, 400))
	mainStatusbar()
	var originalContent fyne.CanvasObject

	err := Docker.GetContexts()
	if err == nil {
		Docker.GetServices()
		x := ""
		for _, context := range Docker.Contexts {
			x = x + context.Name + "\n"
		}
		originalContent = serviceToList()
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
	w.SetContent(content)
	return a, w
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

func serviceToList() *widget.List {
	data := binding.BindStringList(
		&[]string{},
	)
	for _, item := range Docker.Services {
		data.Append(item.Name)
	}
	return widget.NewListWithData(
		data,
		func() fyne.CanvasObject {
			return widget.NewButton("template", func() {})
		},
		func(item binding.DataItem, co fyne.CanvasObject) {
			x, _ := item.(binding.String).Get()
			co.(*widget.Button).SetText(x)
			co.(*widget.Button).OnTapped = func() {
				log.Printf("Touched %s\n", x)
			}
		},
	)
}
