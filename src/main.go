package main

import (
	"docker-swarm-visualiser/cmd"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var Docker cmd.DockerClient
var statusBar = [][]string{{"Context", "Busy", "Active"}}
var statusBarTable *widget.Table
var MainApp fyne.App
var MainWindow fyne.Window
var ActiveWindows map[string]fyne.Window

func init() {
	Docker = cmd.DockerClient{}
}

func main() {
	setupApp()
	MainWindow.ShowAndRun()
}

func setupApp() {
	// mocks.TestMode(&Docker)
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
	ActiveWindows = make(map[string]fyne.Window)
}

func mainToolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.HomeIcon(), func() {
			makeNewWindow("Home?", `meh`)
		}),
		widget.NewToolbarAction(theme.StorageIcon(), func() {
			makeNewWindow("Storage", `Have some storage`)
		}),
		widget.NewToolbarAction(theme.VisibilityOffIcon(), func() {
			makeNewWindow("Secrets", `Have some secrets`)
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.InfoIcon(), func() {
			dialog.ShowInformation(
				"About",
				"Hi\nThis app's purpose is to provide a GUI over\nDocker Swarm specifically for Griffith University's\nuse. This is because it ties into the\n'vlad' access control system.\n\nFor comments, questions, or\ngifting me large sacks\nof unmarked bills, contact\nrelapse@gmail.com.",
				MainWindow,
			)
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

func makeNewWindow(title string, content string) {
	if ActiveWindows[title] == nil {
		ActiveWindows[title] = MainApp.NewWindow(title)
		ActiveWindows[title].Resize(fyne.NewSize(400, 400))
		ActiveWindows[title].SetContent(widget.NewLabel(content))
		ActiveWindows[title].SetOnClosed(func() {
			ActiveWindows[title] = nil
		})
		ActiveWindows[title].Show()
	} else {
		ActiveWindows[title].RequestFocus()
	}
}
