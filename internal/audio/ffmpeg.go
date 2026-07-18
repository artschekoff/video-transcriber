package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func findExecutable(name string) (string, error) {
	return exec.LookPath(name)
}

func CheckFFmpeg() error {
	if _, err := findExecutable("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found on PATH: install ffmpeg to use this app")
	}
	if _, err := findExecutable("ffprobe"); err != nil {
		return fmt.Errorf("ffprobe not found on PATH: install ffmpeg (includes ffprobe)")
	}
	return nil
}

func Cleanup(paths []string) {
	for _, p := range paths {
		os.Remove(p)
	}
}

func getDuration(wavPath string) (float64, error) {
	out, err := exec.Command(
		"ffprobe", "-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		wavPath,
	).Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe: %w", err)
	}
	return strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
}

// SplitAudio extracts audio from inputPath and splits into chunkDuration-second
// raw PCM files (float32 LE, 16kHz, mono). Returns paths to temp files.
func SplitAudio(inputPath string, chunkDuration int) ([]string, error) {
	tmpDir, err := os.MkdirTemp("", "vtranscriber-*")
	if err != nil {
		return nil, err
	}

	// Step 1: extract full audio as WAV for probing duration
	fullWAV := filepath.Join(tmpDir, "full.wav")
	if err := exec.Command(
		"ffmpeg", "-y", "-i", inputPath,
		"-vn", "-acodec", "pcm_s16le", "-ar", "16000", "-ac", "1",
		fullWAV,
	).Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg extract audio: %w", err)
	}

	duration, err := getDuration(fullWAV)
	if err != nil {
		return nil, err
	}

	var paths []string
	for i, start := 0, 0.0; start < duration; i, start = i+1, start+float64(chunkDuration) {
		chunkPath := filepath.Join(tmpDir, fmt.Sprintf("chunk_%04d.pcm", i))
		err := exec.Command(
			"ffmpeg", "-y",
			"-ss", strconv.FormatFloat(start, 'f', 3, 64),
			"-t", strconv.Itoa(chunkDuration),
			"-i", fullWAV,
			"-f", "f32le", "-ac", "1", "-ar", "16000",
			chunkPath,
		).Run()
		if err != nil {
			Cleanup(paths)
			return nil, fmt.Errorf("ffmpeg chunk %d: %w", i, err)
		}
		paths = append(paths, chunkPath)
	}

	// ponytail: fullWAV is only needed for splitting; clean it up now
	os.Remove(fullWAV)
	os.Remove(tmpDir) // only removes if empty after fullWAV removal; chunk files keep tmpDir alive
	return paths, nil
}
