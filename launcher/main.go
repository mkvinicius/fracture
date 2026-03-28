package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var fractureProcess *exec.Cmd

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("FRACTURE")
	w.Resize(fyne.NewSize(400, 280))
	w.SetFixedSize(true)

	// título
	title := canvas.NewText("FRACTURE", theme.ForegroundColor())
	title.TextSize = 22
	title.TextStyle = fyne.TextStyle{Bold: true}
	subtitle := canvas.NewText("Market Disruption Simulation Engine", theme.DisabledColor())
	subtitle.TextSize = 12

	// status
	statusDot := canvas.NewCircle(theme.ErrorColor())
	statusDot.SetMinSize(fyne.NewSize(12, 12))
	statusLabel := widget.NewLabel("Parado")

	statusRow := container.NewHBox(statusDot, statusLabel)

	// botões
	var btnStart, btnStop, btnOpen *widget.Button

	btnOpen = widget.NewButton("Abrir Dashboard", func() {
		openBrowser("http://localhost:3000")
	})
	btnOpen.Disable()

	btnStop = widget.NewButton("Parar", func() {
		if fractureProcess != nil && fractureProcess.Process != nil {
			fractureProcess.Process.Kill()
			fractureProcess = nil
		}
		statusDot.FillColor = theme.ErrorColor()
		statusDot.Refresh()
		statusLabel.SetText("Parado")
		btnStart.Enable()
		btnStop.Disable()
		btnOpen.Disable()
	})
	btnStop.Disable()

	btnStart = widget.NewButton("Iniciar FRACTURE", func() {
		btnStart.Disable()
		statusLabel.SetText("Iniciando...")
		statusDot.FillColor = theme.WarningColor()
		statusDot.Refresh()

		go func() {
			// encontra o binário na mesma pasta do launcher
			exe, _ := os.Executable()
			dir := filepath.Dir(exe)
			binary := filepath.Join(dir, binaryName())

			fractureProcess = exec.Command(binary)
			fractureProcess.Start()

			// aguarda o servidor subir
			time.Sleep(2 * time.Second)

			statusDot.FillColor = theme.SuccessColor()
			statusDot.Refresh()
			statusLabel.SetText("Rodando em localhost:3000")
			btnStop.Enable()
			btnOpen.Enable()
		}()
	})

	// layout
	content := container.NewVBox(
		container.NewCenter(title),
		container.NewCenter(subtitle),
		widget.NewSeparator(),
		container.NewCenter(statusRow),
		widget.NewSeparator(),
		btnStart,
		btnOpen,
		btnStop,
	)

	w.SetContent(container.NewPadded(content))
	w.ShowAndRun()
}

func binaryName() string {
	if runtime.GOOS == "windows" {
		return "fracture.exe"
	}
	return "fracture"
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
		exec.Command("xdg-open", url).Start()
	}
}
