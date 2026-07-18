package main

import (
	_ "embed"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/artschekoff/video-transcriber/internal/audio"
	guipkg "github.com/artschekoff/video-transcriber/gui"
)

//go:embed models/ggml-base.bin
var modelBytes []byte

// ensurePATH prepends common CLI install dirs. Apps launched from Finder inherit
// launchd's minimal PATH (no Homebrew), so ffmpeg installed via brew isn't found.
// ponytail: static list covers Homebrew on Apple Silicon + Intel; extend if needed.
func ensurePATH() {
	path := os.Getenv("PATH")
	for _, dir := range []string{"/opt/homebrew/bin", "/usr/local/bin"} {
		if !strings.Contains(path, dir) {
			path = dir + string(os.PathListSeparator) + path
		}
	}
	os.Setenv("PATH", path)
}

// fatalWindow shows a visible error window and blocks until the user quits.
// GUI apps launched from Finder have no visible stderr, so log.Fatal would look
// like a silent crash — surface startup failures on screen instead.
func fatalWindow(a fyne.App, title, message string) {
	w := a.NewWindow("Video Transcriber")
	w.Resize(fyne.NewSize(480, 200))
	heading := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	body := widget.NewLabel(message)
	body.Wrapping = fyne.TextWrapWord
	quit := widget.NewButton("Quit", func() { a.Quit() })
	w.SetContent(container.NewBorder(heading, quit, nil, nil, container.NewVScroll(body)))
	w.ShowAndRun()
}

func main() {
	ensurePATH()

	a := app.NewWithID("net.artschekoff.videotranscriber")

	if err := audio.CheckFFmpeg(); err != nil {
		fatalWindow(a, "ffmpeg not found",
			err.Error()+"\n\nInstall it with:\n    brew install ffmpeg\n\nThen reopen Video Transcriber.")
		return
	}

	// Extract bundled model to a temp file; whisper.cpp needs a path.
	tmp, err := os.CreateTemp("", "whisper-model-*.bin")
	if err != nil {
		fatalWindow(a, "Startup failed", "Could not create temp model file:\n"+err.Error())
		return
	}
	modelPath := tmp.Name()
	defer os.Remove(modelPath)

	if _, err := tmp.Write(modelBytes); err != nil {
		fatalWindow(a, "Startup failed", "Could not write model file:\n"+err.Error())
		return
	}
	tmp.Close()

	win := guipkg.NewApp(modelPath)
	win.ShowAndRun()
}
