package main

import (
	"docker-swarm-visualiser/cmd"
	"fmt"
	"log"
	"strings"
	"time"

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
var activeBackgroundTasks map[string]string

var stillActive func(command string, me string) bool

func init() {
	Docker = cmd.DockerClient{}
	activeBackgroundTasks = make(map[string]string)
	stillActive = func(command string, me string) bool {
		fmt.Printf("Still active %v|%s", activeBackgroundTasks, activeBackgroundTasks[command])
		return activeBackgroundTasks[command] == me
	}
}

func main() {
	setupApp()
	MainWindow.ShowAndRun()
}

func setupApp() {
	MainApp = app.New()
	MainWindow = MainApp.NewWindow("Griffith Docker GUI")
	MainWindow.Resize(fyne.NewSize(480, 640))
	mainStatusbar()
	originalContent = widget.NewLabel("Starting up, looking for contexts")
	content := container.NewBorder(
		nil,
		statusBarTable,
		nil,
		nil,
		originalContent,
	)
	MainWindow.SetContent(content)
	MainWindow.Resize(fyne.NewSize(480, 640))
	ActiveWindows = make(map[string]fyne.Window)

	go func() {
		me := NewBackgroundProcess("Getting contexts")
		err := Docker.GetContexts()
		if err == nil {
			statusBarTable.Objects[0] = widget.NewButton(Docker.Context, func() { selectContextPopup() })
			version, _ := Docker.GetVersion()
			auths, _ := Docker.GetPrefixes()
			if activeBackgroundTasks["Getting contexts"] == me {
				statusBarTable.Objects[2] = widget.NewLabel("Docker v." + strings.TrimSpace(version) + "|" + strings.Join(auths, ","))
				updateContentDisplay()
				populateServices()
			}
		} else {
			originalContent = widget.NewLabel("Can't access docker")
		}
		updateContentDisplay()
		EndBackgroundProcess("Getting contexts")
	}()
}

func populateServices() {
	originalContent = serviceToVBox()
}

func EndBackgroundProcess(name string) {
	delete(activeBackgroundTasks, name)
	updateBackgroundProcesses()
}
func NewBackgroundProcess(name string) string {
	forWho := fmt.Sprintf("%d", time.Now().UTC())
	activeBackgroundTasks[name] = forWho
	updateBackgroundProcesses()
	return forWho
}

func updateBackgroundProcesses() {
	if len(activeBackgroundTasks) == 0 {
		statusBarTable.Objects[4] = widget.NewLabel("Idle")
	} else {
		statusBarTable.Objects[4] = widget.NewLabel(fmt.Sprintf("Active %d", len(activeBackgroundTasks)))
	}
}

func updateContentDisplay() {
	content := container.NewBorder(
		nil, //mainToolbar(),
		statusBarTable,
		nil,
		nil,
		originalContent,
	)
	MainWindow.SetContent(content)
}

func mainStatusbar() {
	statusBarTable = container.New(
		layout.NewHBoxLayout(),
		widget.NewButton("Contexts", func() { log.Print("Service") }),
		layout.NewSpacer(),
		widget.NewLabel("."),
		layout.NewSpacer(),
		widget.NewLabel("."),
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
		Docker.Context = y
		version, _ := Docker.GetVersion()
		auths, _ := Docker.GetPrefixes()
		statusBarTable.Objects[2] = widget.NewLabel("Docker v." + strings.TrimSpace(version) + "|" + strings.Join(auths, ","))
		populateServices()
		updateContentDisplay()
		modal.Hide()
	}
	modal.Resize(fyne.Size{Width: float32(maxWidth * 12), Height: 300})
	modal.Show()
}

func serviceToVBox() *container.AppTabs {
	// Services Tab
	services := container.New(layout.NewVBoxLayout())
	go func() {
		me := NewBackgroundProcess("Service Tab")
		Docker.GetServices(stillActive, "Service Tab", me)
		if stillActive("Service Tab", me) {
			for _, item := range Docker.Services {
				service := item.Name
				services.Add(container.New(
					layout.NewHBoxLayout(),
					widget.NewLabel(service),
					layout.NewSpacer(),
					widget.NewButton("PS", func() {
						ps, error := Docker.GetPS(service)
						if error == nil {
							makeNewWindow(
								fmt.Sprintf("[%s] PS for "+service, Docker.Context, service),
								ps,
							)
						}
					}),
					widget.NewButton("Logs", func() {
						go func() {
							processName := fmt.Sprintf("[%s] Logs %s", Docker.Context, service)
							NewBackgroundProcess(processName)
							Docker.FollowLogs(MainApp, service)
							EndBackgroundProcess(processName)
						}()
					}),
				))
			}
			EndBackgroundProcess("Service Tab")
		}
	}()
	volumes := container.New(layout.NewVBoxLayout())
	// Storage Tab
	go func() {
		me := NewBackgroundProcess("Storage Tab")
		Docker.GetVolumes(stillActive, "Storage Tab", me)
		if stillActive("Storage Tab", me) {
			for _, item := range Docker.Volumes {
				volume := item.Name
				volumes.Add(container.New(
					layout.NewHBoxLayout(),
					widget.NewLabel(volume),
					layout.NewSpacer(),
					widget.NewButton("Inspect", func() {
						output, error := Docker.InspectVolume(volume)
						if error == nil {
							makeNewWindow(
								fmt.Sprintf("[%s] Volume %s", Docker.Context, volume),
								output,
							)
						} else {
							log.Printf("Failed volume %s\n", volume)
						}
					}),
				))
			}
			EndBackgroundProcess("Storage Tab")
		}
	}()
	// Secrets Tab
	secrets := container.New(layout.NewVBoxLayout())
	go func() {
		me := NewBackgroundProcess("Secrets Tab")
		Docker.GetSecrets(stillActive, "Secrets Tab", me)
		if stillActive("Secrets Tab", me) {
			for _, item := range Docker.Secrets {
				secret := item.Name
				secrets.Add(container.New(
					layout.NewHBoxLayout(),
					widget.NewLabel(secret),
					layout.NewSpacer(),
					widget.NewButton("Inspect", func() {
						output, error := Docker.InspectSecret(secret)
						if error == nil {
							makeNewWindow(
								fmt.Sprintf("[%s] Secret %s", Docker.Context, secret),
								output,
							)
						} else {
							log.Printf("Failed secret %s\n", secret)
						}
					}),
				))
			}
			EndBackgroundProcess("Secrets Tab")
		}
	}()
	// About Tab
	aboutBit := widget.NewLabel("Hi\nThis app's purpose is to provide a GUI over Docker Swarm specifically for Griffith University's use. This is because it ties into the 'vlad' access control system.\n\nFor comments, questions, or gifting me large sacks of unmarked bills, contact relapse@gmail.com.")
	aboutBit.Wrapping = fyne.TextWrapWord
	// Return Tab Interface
	return container.NewAppTabs(
		container.NewTabItem("Services", container.NewVScroll(services)),
		container.NewTabItem("Storage", container.NewVScroll(volumes)),
		container.NewTabItem("Secret", container.NewVScroll(secrets)),
		container.NewTabItem("About", aboutBit),
	)
}

func makeNewWindow(title string, content string) {
	if ActiveWindows[title] == nil {
		contentLabel := widget.NewLabel(content)
		contentLabel.TextStyle.Monospace = true
		ActiveWindows[title] = MainApp.NewWindow(title)
		ActiveWindows[title].Resize(fyne.NewSize(400, 400))
		ActiveWindows[title].SetContent(contentLabel)
		ActiveWindows[title].SetOnClosed(func() {
			ActiveWindows[title] = nil
		})
		ActiveWindows[title].Show()
	} else {
		ActiveWindows[title].RequestFocus()
	}
}
