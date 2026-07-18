package exporter

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/artschekoff/video-transcriber/internal/whisper"
)

func formatTS(d time.Duration) string {
	ms := d.Milliseconds()
	h := ms / 3_600_000
	ms -= h * 3_600_000
	m := ms / 60_000
	ms -= m * 60_000
	s := ms / 1_000
	ms -= s * 1_000
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}

func ToSRT(segments []whisper.Segment) string {
	var sb strings.Builder
	for i, seg := range segments {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		fmt.Fprintf(&sb, "%d\n%s --> %s\n%s",
			i+1,
			formatTS(seg.Start),
			formatTS(seg.End),
			strings.TrimSpace(seg.Text),
		)
	}
	sb.WriteString("\n")
	return sb.String()
}

func ToTXT(segments []whisper.Segment) string {
	lines := make([]string, len(segments))
	for i, seg := range segments {
		lines[i] = strings.TrimSpace(seg.Text)
	}
	return strings.Join(lines, "\n")
}

func Save(content, path string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
