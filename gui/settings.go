package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/artschekoff/video-transcriber/internal/config"
)

func ShowSettings(parent fyne.Window, cfg config.AppConfig, onSave func(config.AppConfig)) {
	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetText(cfg.OpenAIAPIKey)

	chunkEntry := widget.NewEntry()
	chunkEntry.SetText(fmt.Sprintf("%d", cfg.ChunkDuration))

	form := dialog.NewForm("Settings", "Save", "Cancel",
		[]*widget.FormItem{
			{Text: "OpenAI API Key", Widget: apiKeyEntry},
			{Text: "Chunk duration (seconds)", Widget: chunkEntry},
		},
		func(ok bool) {
			if !ok {
				return
			}
			chunkSecs := cfg.ChunkDuration
			fmt.Sscanf(chunkEntry.Text, "%d", &chunkSecs)
			updated := config.AppConfig{
				OpenAIAPIKey:  apiKeyEntry.Text,
				ChunkDuration: chunkSecs,
			}
			config.Save(updated)
			onSave(updated)
		},
		parent,
	)
	form.Show()
}
