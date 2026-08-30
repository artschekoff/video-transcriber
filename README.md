# Video Transcriber

**On-device video & audio subtitling for macOS — Whisper running locally, optional OpenAI translation.**

Desktop app (Fyne GUI) that extracts subtitles from video/audio using Whisper
(whisper.cpp, running locally), with optional OpenAI translation of the result.

Pick a file → transcribe on-device → save as SRT or TXT. Optionally translate
each subtitle to another language via OpenAI (timings are preserved).

## Requirements

- Go 1.23+
- **ffmpeg** on PATH — `brew install ffmpeg` (checked at startup)
- **cmake** — `brew install cmake` (builds the whisper.cpp static library)
- Xcode command line tools — `xcode-select --install`
- For releases: `gh` CLI authenticated

> Native build is macOS/Metal only. Linux needs different link flags (drop the
> Metal frameworks in the Makefile's `EXT_LDFLAGS`).

## Build

```bash
make build            # builds whisper.cpp lib, downloads ggml-base.bin (~142MB), builds the app
./dist/VideoTranscriber
```

`make help` lists all targets.

## Install to /Applications

```bash
make install          # builds, packages VideoTranscriber.app, copies it to /Applications
```

Then launch **Video Transcriber** from Finder/Spotlight. `make bundle` alone just
produces `dist/VideoTranscriber.app` without installing.

ffmpeg **must** be installed (`brew install ffmpeg`) — the app checks for it at
startup and exits if missing. Apps launched from Finder don't inherit your shell
PATH, so the app also probes the standard Homebrew locations automatically.

## Test

```bash
make test
```

## Translation (optional)

Open **⚙ Settings** in the app and paste an OpenAI API key. Then pick a target
language in "Translate to". Without a key, transcription still works; only
translation is unavailable. The key is stored in the OS config dir, not in git.

## Release

```bash
make release          # guards clean tree + pushed commits, prompts for semver bump, tags + publishes via gh
```

## Bundled model

The `ggml-base.bin` Whisper model is embedded in the binary at build time
(`//go:embed`). Downloaded by `make download-model` from HuggingFace
(ggerganov/whisper.cpp). Not committed to git (see `.gitignore`).
