package audio

import (
	"os"
	"testing"
)

func TestCheckFFmpeg_NotOnPath(t *testing.T) {
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", origPath)

	err := CheckFFmpeg()
	if err == nil {
		t.Fatal("expected error when PATH is empty, got nil")
	}
}

func TestCheckFFmpeg_Present(t *testing.T) {
	// This test only runs if ffmpeg is actually installed.
	if _, err := findExecutable("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed, skipping")
	}
	if err := CheckFFmpeg(); err != nil {
		t.Fatalf("CheckFFmpeg: %v", err)
	}
}

func TestCleanup_DeletesFiles(t *testing.T) {
	f, err := os.CreateTemp("", "test-*.tmp")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	path := f.Name()

	Cleanup([]string{path})

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("file still exists after Cleanup: %s", path)
	}
}

func TestCleanup_ToleratesNonexistent(t *testing.T) {
	Cleanup([]string{"/tmp/nonexistent-abc123.tmp"}) // must not panic
}
