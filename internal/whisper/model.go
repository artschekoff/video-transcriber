package whisper

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

type Segment struct {
	Start time.Duration
	End   time.Duration
	Text  string
}

// readPCMFloat32 reads a raw float32 LE PCM file into []float32.
func readPCMFloat32(path string) ([]float32, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	count := info.Size() / 4
	samples := make([]float32, count)
	if err := binary.Read(f, binary.LittleEndian, &samples); err != nil && err != io.EOF {
		return nil, err
	}
	return samples, nil
}

// Transcribe transcribes all PCM chunks and returns merged segments with
// correct time offsets. onProgress(chunkIdx, total) is called before each chunk.
func Transcribe(
	pcmPaths []string,
	modelPath string,
	language string,
	chunkDuration int,
	onProgress func(int, int),
) ([]Segment, error) {
	if err := ValidateLanguage(language); err != nil {
		return nil, err
	}

	model, err := whisper.New(modelPath)
	if err != nil {
		return nil, fmt.Errorf("load model: %w", err)
	}
	defer model.Close()

	var all []Segment
	total := len(pcmPaths)

	for idx, pcmPath := range pcmPaths {
		if onProgress != nil {
			onProgress(idx, total)
		}

		samples, err := readPCMFloat32(pcmPath)
		if err != nil {
			return nil, fmt.Errorf("read chunk %d: %w", idx, err)
		}

		ctx, err := model.NewContext()
		if err != nil {
			return nil, fmt.Errorf("new context: %w", err)
		}

		if language != "auto" && language != "" {
			if err := ctx.SetLanguage(language); err != nil {
				return nil, fmt.Errorf("set language: %w", err)
			}
		}

		if err := ctx.Process(samples, nil, nil, nil); err != nil {
			return nil, fmt.Errorf("process chunk %d: %w", idx, err)
		}

		offset := time.Duration(idx*chunkDuration) * time.Second
		for {
			seg, err := ctx.NextSegment()
			if err != nil {
				break // io.EOF signals no more segments
			}
			all = append(all, Segment{
				Start: seg.Start + offset,
				End:   seg.End + offset,
				Text:  seg.Text,
			})
		}
	}

	if onProgress != nil {
		onProgress(total, total)
	}
	return all, nil
}
