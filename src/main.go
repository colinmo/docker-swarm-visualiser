package main

import (
	"docker-swarm-visualiser/cmd"
	"docker-swarm-visualiser/utils/mocks"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var Docker cmd.DockerClient
var statusBarTable *fyne.Container
var MainApp fyne.App
var MainWindow fyne.Window
var ActiveWindows map[string]fyne.Window
var originalContent fyne.CanvasObject

func init() {
	Docker = cmd.DockerClient{}
}

func main() {
	setupApp()
	MainWindow.ShowAndRun()
}

func setupApp() {
	mocks.TestMode(&Docker)
	MainApp = app.New()
	MainWindow = MainApp.NewWindow("Griffith Docker GUI")
	MainWindow.Resize(fyne.NewSize(400, 400))
	mainStatusbar()

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

	statusBarTable.Objects[0] = widget.NewButton(Docker.Context, func() { selectContextPopup() })
	content := container.NewBorder(
		nil, //mainToolbar(),
		statusBarTable,
		nil,
		nil,
		originalContent,
	)
	MainWindow.SetContent(content)
	ActiveWindows = make(map[string]fyne.Window)
}

func mainStatusbar() {
	statusBarTable = container.New(
		layout.NewHBoxLayout(),
		widget.NewButton("Service", func() { log.Print("Service") }),
		layout.NewSpacer(),
		widget.NewLabel("Busy"),
		layout.NewSpacer(),
		widget.NewLabel("Active"),
	)
}

func selectContextPopup() {
	var maxWidth int
	maxWidth = 0.0
	data := binding.BindStringList(&[]string{})
	for _, d := range Docker.Contexts {
		data.Append(d.Name)
		if len(d.Name) > maxWidth {
			maxWidth = len(d.Name)
		}
	}
	list := widget.NewListWithData(
		data,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		},
	)
	modal := widget.NewModalPopUp(
		list,
		MainWindow.Canvas(),
	)
	list.OnSelected = func(id int) {
		x, _ := data.GetItem(id)
		y, _ := x.(binding.String).Get()
		statusBarTable.Objects[0] = widget.NewButton(y, func() { selectContextPopup() })
		// @todo Refresh the pages
		modal.Hide()
	}
	log.Printf("Max width is %v\n", maxWidth)
	modal.Resize(fyne.Size{Width: float32(maxWidth * 12), Height: 300})
	modal.Show()
}

func serviceToVBox() *container.AppTabs {
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
	aboutBit := widget.NewLabel("Hi\nThis app's purpose is to provide a GUI over Docker Swarm specifically for Griffith University's use. This is because it ties into the 'vlad' access control system.\n\nFor comments, questions, or gifting me large sacks of unmarked bills, contact relapse@gmail.com.")
	aboutBit.Wrapping = fyne.TextWrapWord
	return container.NewAppTabs(
		container.NewTabItem("Services", me),
		container.NewTabItem("Storage", container.New(layout.NewVBoxLayout())),
		container.NewTabItem("Secret", container.New(layout.NewVBoxLayout())),
		container.NewTabItem("About", aboutBit),
	)
}

/*
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
*/
