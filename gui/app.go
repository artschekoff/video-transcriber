package gui

import (
	"fmt"
	"path/filepath"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/artschekoff/video-transcriber/internal/audio"
	"github.com/artschekoff/video-transcriber/internal/config"
	"github.com/artschekoff/video-transcriber/internal/exporter"
	"github.com/artschekoff/video-transcriber/internal/translator"
	"github.com/artschekoff/video-transcriber/internal/whisper"
)

type appState struct {
	win       fyne.Window
	modelPath string
	cfg       config.AppConfig
	segments  []whisper.Segment
	filePath  string
}

func NewApp(modelPath string) fyne.Window {
	cfg, _ := config.Load()
	state := &appState{modelPath: modelPath, cfg: cfg}

	a := fyne.CurrentApp()
	state.win = a.NewWindow("Video Transcriber")
	state.win.SetMaster()
	state.win.Resize(fyne.NewSize(580, 420))

	state.win.SetContent(state.buildUI())
	return state.win
}

func (s *appState) buildUI() fyne.CanvasObject {
	// File picker
	fileLabel := widget.NewLabel("No file selected")
	chooseBtn := widget.NewButton("Choose File", func() {
		fd := dialog.NewFileOpen(func(f fyne.URIReadCloser, err error) {
			if err != nil || f == nil {
				return
			}
			f.Close()
			s.filePath = f.URI().Path()
			fileLabel.SetText(filepath.Base(s.filePath))
		}, s.win)
		fd.SetFilter(newMediaFilter())
		fd.Show()
	})

	// Language options
	langNames, langCodes := buildLanguageList()
	langSelect := widget.NewSelect(langNames, nil)
	langSelect.SetSelected("Auto-detect")

	targetNames, _ := buildLanguageList()
	targetNames = append([]string{"None"}, targetNames...)
	targetSelect := widget.NewSelect(targetNames, nil)
	targetSelect.SetSelected("None")

	formatSelect := widget.NewSelect([]string{"SRT", "TXT"}, nil)
	formatSelect.SetSelected("SRT")

	// Progress
	progress := widget.NewProgressBar()
	progressLabel := widget.NewLabel("")

	// Save button (disabled until done)
	saveBtn := widget.NewButton("Save Result", nil)
	saveBtn.Disable()

	// Start button
	startBtn := widget.NewButton("▶  Start Transcription", nil)
	startBtn.Importance = widget.HighImportance

	startBtn.OnTapped = func() {
		if s.filePath == "" {
			dialog.ShowError(fmt.Errorf("please select a file first"), s.win)
			return
		}
		startBtn.Disable()
		saveBtn.Disable()
		s.segments = nil
		progress.SetValue(0)
		progressLabel.SetText("Extracting audio...")

		langCode := "auto"
		if idx := findIndex(langNames, langSelect.Selected); idx >= 0 {
			langCode = langCodes[idx]
		}

		go func() {
			pcmPaths, err := audio.SplitAudio(s.filePath, s.cfg.ChunkDuration)
			if err != nil {
				dialog.ShowError(err, s.win)
				startBtn.Enable()
				return
			}
			defer audio.Cleanup(pcmPaths)

			segs, err := whisper.Transcribe(
				pcmPaths, s.modelPath, langCode, s.cfg.ChunkDuration,
				func(cur, total int) {
					if total > 0 {
						progress.SetValue(float64(cur) / float64(total))
						progressLabel.SetText(fmt.Sprintf("Chunk %d/%d", cur, total))
					}
				},
			)
			if err != nil {
				startBtn.Enable()
				dialog.ShowError(err, s.win)
				return
			}

			if targetSelect.Selected != "None" {
				if s.cfg.OpenAIAPIKey == "" {
					dialog.ShowError(fmt.Errorf("set OpenAI API key in Settings to translate"), s.win)
					startBtn.Enable()
					return
				}
				segs, err = translator.Translate(segs, targetSelect.Selected, s.cfg.OpenAIAPIKey)
				if err != nil {
					startBtn.Enable()
					dialog.ShowError(err, s.win)
					return
				}
			}

			s.segments = segs
			progress.SetValue(1)
			progressLabel.SetText(fmt.Sprintf("Done — %d segments", len(segs)))
			startBtn.Enable()
			saveBtn.Enable()
		}()
	}

	saveBtn.OnTapped = func() {
		if len(s.segments) == 0 {
			return
		}
		ext := "." + formatSelect.Selected
		fd := dialog.NewFileSave(func(f fyne.URIWriteCloser, err error) {
			if err != nil || f == nil {
				return
			}
			f.Close()
			var content string
			if formatSelect.Selected == "SRT" {
				content = exporter.ToSRT(s.segments)
			} else {
				content = exporter.ToTXT(s.segments)
			}
			if err := exporter.Save(content, f.URI().Path()); err != nil {
				dialog.ShowError(err, s.win)
				return
			}
			dialog.ShowInformation("Saved", fmt.Sprintf("Saved to %s", f.URI().Path()), s.win)
		}, s.win)
		fd.SetFileName("output" + ext)
		fd.Show()
	}

	settingsBtn := widget.NewButton("⚙ Settings", func() {
		ShowSettings(s.win, s.cfg, func(updated config.AppConfig) {
			s.cfg = updated
		})
	})

	titleRow := container.NewBorder(nil, nil, nil, settingsBtn,
		widget.NewLabelWithStyle("Video Transcriber", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	options := container.NewGridWithColumns(3,
		container.NewVBox(widget.NewLabel("Input language"), langSelect),
		container.NewVBox(widget.NewLabel("Translate to"), targetSelect),
		container.NewVBox(widget.NewLabel("Save as"), formatSelect),
	)

	fileRow := container.NewBorder(nil, nil, chooseBtn, nil, fileLabel)

	return container.NewVBox(
		titleRow,
		widget.NewSeparator(),
		fileRow,
		options,
		startBtn,
		progressLabel,
		progress,
		saveBtn,
	)
}

func buildLanguageList() (names []string, codes []string) {
	type pair struct{ name, code string }
	var pairs []pair
	for code, name := range whisper.WHISPER_LANGUAGES {
		pairs = append(pairs, pair{name, code})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].name < pairs[j].name })

	names = []string{"Auto-detect"}
	codes = []string{"auto"}
	for _, p := range pairs {
		names = append(names, p.name)
		codes = append(codes, p.code)
	}
	return
}

func findIndex(slice []string, val string) int {
	for i, v := range slice {
		if v == val {
			return i
		}
	}
	return -1
}

type mediaFilter struct{}

func (f *mediaFilter) Matches(uri fyne.URI) bool {
	ext := filepath.Ext(uri.Name())
	for _, e := range []string{".mp4", ".mkv", ".avi", ".mov", ".mp3", ".wav", ".m4a", ".flac", ".ogg", ".webm"} {
		if ext == e {
			return true
		}
	}
	return false
}
func (f *mediaFilter) MimeTypes() []string { return []string{"video/*", "audio/*"} }

func newMediaFilter() storage.FileFilter {
	return &mediaFilter{}
}
