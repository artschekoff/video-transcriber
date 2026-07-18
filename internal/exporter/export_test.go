package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/artschekoff/video-transcriber/internal/whisper"
)

var testSegs = []whisper.Segment{
	{Start: 0, End: 2500 * time.Millisecond, Text: " Hello world"},
	{Start: 2500 * time.Millisecond, End: 5 * time.Second, Text: " Goodbye"},
}

func TestToSRT_BlockCount(t *testing.T) {
	result := ToSRT(testSegs)
	blocks := splitBlocks(result)
	if len(blocks) != 2 {
		t.Errorf("want 2 blocks, got %d\n---\n%s", len(blocks), result)
	}
}

func TestToSRT_FirstBlockFormat(t *testing.T) {
	result := ToSRT(testSegs)
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if lines[0] != "1" {
		t.Errorf("first line: want '1', got %q", lines[0])
	}
	if lines[1] != "00:00:00,000 --> 00:00:02,500" {
		t.Errorf("timestamp: want '00:00:00,000 --> 00:00:02,500', got %q", lines[1])
	}
	if lines[2] != "Hello world" {
		t.Errorf("text: want 'Hello world', got %q", lines[2])
	}
}

func TestToSRT_SecondBlockIndex(t *testing.T) {
	result := ToSRT(testSegs)
	blocks := splitBlocks(result)
	if !strings.HasPrefix(blocks[1], "2\n") {
		t.Errorf("second block should start with '2\\n', got %q", blocks[1])
	}
}

func TestToTXT_JoinsText(t *testing.T) {
	result := ToTXT(testSegs)
	if result != "Hello world\nGoodbye" {
		t.Errorf("want 'Hello world\\nGoodbye', got %q", result)
	}
}

func TestSave_WritesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")
	if err := Save("hello", path); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "hello" {
		t.Errorf("want 'hello', got %q", string(data))
	}
}

func splitBlocks(s string) []string {
	var out []string
	for _, b := range strings.Split(strings.TrimSpace(s), "\n\n") {
		if strings.TrimSpace(b) != "" {
			out = append(out, b)
		}
	}
	return out
}
