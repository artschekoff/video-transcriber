package main

import (
	_ "embed"
	"log"
	"os"
	"strings"

	"fyne.io/fyne/v2/app"

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

func main() {
	ensurePATH()
	if err := audio.CheckFFmpeg(); err != nil {
		log.Fatalf("startup check failed: %v", err)
	}

	// Extract bundled model to a temp file; whisper.cpp needs a path.
	tmp, err := os.CreateTemp("", "whisper-model-*.bin")
	if err != nil {
		log.Fatalf("create temp model: %v", err)
	}
	modelPath := tmp.Name()
	defer os.Remove(modelPath)

	if _, err := tmp.Write(modelBytes); err != nil {
		log.Fatalf("write model: %v", err)
	}
	tmp.Close()

	app.New()
	win := guipkg.NewApp(modelPath)
	win.ShowAndRun()
}
